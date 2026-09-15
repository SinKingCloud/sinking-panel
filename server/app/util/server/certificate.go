package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/mail"
	baiduDNS "server/app/util/server/dns/baidu"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	libdnsAliDNS "github.com/libdns/alidns"
	libdnsHuaweiCloud "github.com/libdns/huaweicloud"
	libdnsTencentCloud "github.com/libdns/tencentcloud"
	libdnsVolcengine "github.com/libdns/volcengine"
	"go.uber.org/zap"
)

const (
	// CertificateRequestTimeout 限制单次自动申请或续签的总等待时间。
	CertificateRequestTimeout    = 30 * time.Second
	certificateValidationTimeout = 3 * time.Second
)

func (m *Manager) obtainCertificate(ctx context.Context, request CertificateRequest, renew bool) (certificate *Certificate, err error) {
	operation := "申请"
	if renew {
		operation = "续签"
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			certificate = nil
			err = fmt.Errorf("%s证书时发生异常，请检查验证参数后重试", operation)
		}
	}()

	challenge := CertificateChallenge(strings.ToLower(strings.TrimSpace(string(request.Challenge))))
	action := CertificateAction(strings.ToLower(strings.TrimSpace(string(request.Action))))
	if action == "" {
		action = CertificateActionAuto
	}
	if action != CertificateActionAuto && action != CertificateActionManual {
		return nil, errors.New("证书验证模式只能为自动或手动")
	}
	if action == CertificateActionManual {
		return nil, errors.New("手动验证请使用 Acme 申请并提交验证")
	}
	// 签发仅使用配置快照，等待 CA 或 DNS 时不占用网站运行时的操作锁。
	m.mu.RLock()
	options, storage := m.options, m.acmeStorage
	m.mu.RUnlock()

	if ctx == nil {
		return nil, errors.New("证书申请上下文不能为空")
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, CertificateRequestTimeout)
	defer cancel()
	domain, err := normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("证书域名无效: %w", err)
	}
	isIP := net.ParseIP(domain) != nil
	if !certmagic.SubjectQualifiesForPublicCert(domain) {
		return nil, errors.New("域名或 IP 地址不能申请公开证书")
	}
	if challenge == "" {
		challenge = CertificateChallengeHTTP
	}
	if challenge != CertificateChallengeHTTP && challenge != CertificateChallengeDNS {
		return nil, errors.New("证书验证类型只支持 HTTP 或 DNS")
	}
	if isIP && challenge != CertificateChallengeHTTP {
		return nil, errors.New("IP 地址证书必须使用 HTTP 验证")
	}
	if strings.HasPrefix(domain, "*.") && challenge == CertificateChallengeHTTP {
		return nil, errors.New("通配符证书必须使用 DNS 验证")
	}

	caName := CertificateCA(strings.ToLower(strings.TrimSpace(string(request.CA))))
	var caURL string
	switch caName {
	case "", CertificateCAProd:
		caName = CertificateCAProd
		caURL = certmagic.LetsEncryptProductionCA
	case CertificateCAStaging:
		caURL = certmagic.LetsEncryptStagingCA
	default:
		return nil, errors.New("证书 CA 只支持 production 或 staging")
	}
	if storage == nil {
		return nil, errors.New("证书存储尚未初始化")
	}

	email := strings.TrimSpace(request.Email)
	if email == "" {
		email = strings.TrimSpace(options.ACMEEmail)
	}
	if email != "" {
		address, parseErr := mail.ParseAddress(email)
		if parseErr != nil || address.Address != email {
			return nil, errors.New("ACME 邮箱地址无效")
		}
	}
	logger := zap.NewNop()
	var magic *certmagic.Config
	cache := certmagic.NewCache(certmagic.CacheOptions{
		Logger: logger,
		GetConfigForCert: func(certmagic.Certificate) (*certmagic.Config, error) {
			return magic, nil
		},
	})
	defer cache.Stop()

	magic = certmagic.New(cache, certmagic.Config{
		Storage: storage,
		Logger:  logger,
	})
	issuerOptions := certmagic.ACMEIssuer{
		CA:                      caURL,
		Email:                   email,
		AccountKeyPEM:           request.AccountKeyPEM,
		Agreed:                  true,
		Logger:                  logger,
		CertObtainTimeout:       CertificateRequestTimeout,
		DisableTLSALPNChallenge: true,
		ListenHost:              strings.TrimSpace(options.HTTPChallengeHost),
		AltHTTPPort:             options.HTTPChallengePort,
	}
	if isIP {
		// Let's Encrypt 的 IP 证书必须使用短期配置，有效期为 160 小时。
		issuerOptions.Profile = "shortlived"
	}
	if challenge == CertificateChallengeDNS && action == CertificateActionAuto {
		credentials := request.DNSCredentials
		credentials.AliyunAccessKeyID = strings.TrimSpace(credentials.AliyunAccessKeyID)
		credentials.AliyunAccessKeySecret = strings.TrimSpace(credentials.AliyunAccessKeySecret)
		credentials.TencentSecretID = strings.TrimSpace(credentials.TencentSecretID)
		credentials.TencentSecretKey = strings.TrimSpace(credentials.TencentSecretKey)
		credentials.HuaweiAccessKeyID = strings.TrimSpace(credentials.HuaweiAccessKeyID)
		credentials.HuaweiSecretAccessKey = strings.TrimSpace(credentials.HuaweiSecretAccessKey)
		credentials.VolcengineAccessKeyID = strings.TrimSpace(credentials.VolcengineAccessKeyID)
		credentials.VolcengineAccessKeySecret = strings.TrimSpace(credentials.VolcengineAccessKeySecret)
		credentials.BaiduAccessKeyID = strings.TrimSpace(credentials.BaiduAccessKeyID)
		credentials.BaiduSecretAccessKey = strings.TrimSpace(credentials.BaiduSecretAccessKey)

		providerName := DNSProvider(strings.ToLower(strings.TrimSpace(string(request.DNSProvider))))
		if err := credentials.Validate(providerName); err != nil {
			return nil, err
		}
		var provider certmagic.DNSProvider
		manager := certmagic.DNSManager{
			Logger:             logger,
			PropagationTimeout: certificateValidationTimeout,
		}
		switch providerName {
		case DNSProviderAliDNS:
			provider = &libdnsAliDNS.Provider{CredentialInfo: libdnsAliDNS.CredentialInfo{
				AccessKeyID:     credentials.AliyunAccessKeyID,
				AccessKeySecret: credentials.AliyunAccessKeySecret,
			}}
		case DNSProviderTencentCloud:
			provider = &libdnsTencentCloud.Provider{SecretId: credentials.TencentSecretID, SecretKey: credentials.TencentSecretKey}
		case DNSProviderHuaweiCloud:
			provider = &libdnsHuaweiCloud.Provider{
				AccessKeyId:     credentials.HuaweiAccessKeyID,
				SecretAccessKey: credentials.HuaweiSecretAccessKey,
			}
		case DNSProviderVolcengine:
			provider = &libdnsVolcengine.Provider{CredentialInfo: libdnsVolcengine.CredentialInfo{
				AccessKeyID:     credentials.VolcengineAccessKeyID,
				AccessKeySecret: credentials.VolcengineAccessKeySecret,
			}}
			// 保留原有记录 TTL，传播等待与其他服务商保持一致。
			manager.TTL = 600 * time.Second
		case DNSProviderBaiduCloud:
			provider = &baiduDNS.Provider{
				AccessKeyID:     credentials.BaiduAccessKeyID,
				SecretAccessKey: credentials.BaiduSecretAccessKey,
			}
		}
		manager.DNSProvider = provider
		issuerOptions.DisableHTTPChallenge = true
		issuerOptions.DNS01Solver = &certmagic.DNS01Solver{DNSManager: manager}
	}
	issuer := certmagic.NewACMEIssuer(magic, issuerOptions)
	magic.Issuers = []certmagic.Issuer{issuer}

	issuerKey := issuer.IssuerKey()
	// 同域名的签发和结果读取串行，避免不同 CA 的 HTTP 挑战覆盖同一个内存令牌。
	lockKey := "certificate_" + domain
	if err = storage.Lock(ctx, lockKey); err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, fmt.Errorf("锁定证书失败: %w", err)
	}
	defer storage.Unlock(context.WithoutCancel(ctx), lockKey)
	certificateKey := certmagic.StorageKeys.SiteCert(issuerKey, domain)
	privateKey := certmagic.StorageKeys.SitePrivateKey(issuerKey, domain)
	metadataKey := certmagic.StorageKeys.SiteMeta(issuerKey, domain)
	if storage.Exists(ctx, certificateKey) && storage.Exists(ctx, privateKey) && storage.Exists(ctx, metadataKey) {
		// 已有证书由 CertMagic 根据 ARI 和有效期判断续签，显式续签则强制执行。
		err = magic.RenewCertSync(ctx, domain, renew)
	} else {
		err = magic.ObtainCertSync(ctx, domain)
	}
	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, fmt.Errorf("%s证书失败: %w", operation, err)
	}

	certificatePEM, err := storage.Load(ctx, certificateKey)
	if err != nil {
		return nil, fmt.Errorf("读取证书失败: %w", err)
	}
	privateKeyPEM, err := storage.Load(ctx, privateKey)
	if err != nil {
		return nil, fmt.Errorf("读取证书私钥失败: %w", err)
	}
	pair, err := tls.X509KeyPair(certificatePEM, privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}
	if len(pair.Certificate) == 0 {
		return nil, errors.New("证书内容为空")
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("解析证书信息失败: %w", err)
	}
	ipAddresses := make([]string, 0, len(leaf.IPAddresses))
	for _, address := range leaf.IPAddresses {
		ipAddresses = append(ipAddresses, address.String())
	}

	return &Certificate{
		Domain:          domain,
		Issuer:          string(caName),
		DNSNames:        append([]string(nil), leaf.DNSNames...),
		IPAddresses:     ipAddresses,
		SerialNumber:    leaf.SerialNumber.String(),
		NotBefore:       leaf.NotBefore,
		NotAfter:        leaf.NotAfter,
		CertificateFile: storage.Filename(certificateKey),
		KeyFile:         storage.Filename(privateKey),
		CertificatePEM:  append([]byte(nil), certificatePEM...),
		PrivateKeyPEM:   append([]byte(nil), privateKeyPEM...),
	}, nil
}

// Validate 校验 DNS 服务商所需的凭据，供证书申请和密钥绑定共用。
func (credentials DNSCredentials) Validate(provider DNSProvider) error {
	switch provider {
	case DNSProviderAliDNS:
		if strings.TrimSpace(credentials.AliyunAccessKeyID) == "" || strings.TrimSpace(credentials.AliyunAccessKeySecret) == "" {
			return errors.New("阿里云 DNS 验证需要 AccessKey ID 和 AccessKey Secret")
		}
	case DNSProviderTencentCloud:
		if strings.TrimSpace(credentials.TencentSecretID) == "" || strings.TrimSpace(credentials.TencentSecretKey) == "" {
			return errors.New("腾讯云 DNS 验证需要 Secret ID 和 Secret Key")
		}
	case DNSProviderHuaweiCloud:
		if strings.TrimSpace(credentials.HuaweiAccessKeyID) == "" || strings.TrimSpace(credentials.HuaweiSecretAccessKey) == "" {
			return errors.New("华为云 DNS 验证需要 AccessKey ID 和 Secret Access Key")
		}
	case DNSProviderVolcengine:
		if strings.TrimSpace(credentials.VolcengineAccessKeyID) == "" || strings.TrimSpace(credentials.VolcengineAccessKeySecret) == "" {
			return errors.New("火山引擎 DNS 验证需要 AccessKey ID 和 AccessKey Secret")
		}
	case DNSProviderBaiduCloud:
		if strings.TrimSpace(credentials.BaiduAccessKeyID) == "" || strings.TrimSpace(credentials.BaiduSecretAccessKey) == "" {
			return errors.New("百度云 DNS 验证需要 AccessKey ID 和 Secret Access Key")
		}
	default:
		return errors.New("不支持的 DNS 服务商")
	}
	return nil
}
