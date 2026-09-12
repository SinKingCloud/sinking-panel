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

	caddyDNSPod "github.com/caddy-dns/dnspod"
	"github.com/caddyserver/certmagic"
	libdnsAliDNS "github.com/libdns/alidns"
	libdnsHuaweiCloud "github.com/libdns/huaweicloud"
	libdnsTencentCloud "github.com/libdns/tencentcloud"
	libdnsVolcengine "github.com/libdns/volcengine"
	"go.uber.org/zap"
)

func (m *Manager) obtainCertificate(ctx context.Context, request CertificateRequest, renew bool) (certificate *Certificate, err error) {
	action := "申请"
	if renew {
		action = "续签"
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			certificate = nil
			err = fmt.Errorf("%s证书时发生异常，请检查验证参数后重试", action)
		}
	}()

	if ctx == nil {
		return nil, errors.New("证书申请上下文不能为空")
	}
	domain, err := m.normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("证书域名无效: %w", err)
	}
	isIP := net.ParseIP(domain) != nil
	if !certmagic.SubjectQualifiesForPublicCert(domain) {
		return nil, errors.New("域名或 IP 地址不能申请公开证书")
	}
	challenge := CertificateChallenge(strings.ToLower(strings.TrimSpace(string(request.Challenge))))
	if challenge == "" {
		challenge = CertificateChallengeHTTP
	}
	if challenge != CertificateChallengeHTTP && challenge != CertificateChallengeDNS {
		return nil, errors.New("证书验证方式只支持 http 或 dns")
	}
	if isIP && challenge != CertificateChallengeHTTP {
		return nil, errors.New("IP 地址证书必须使用 HTTP 验证")
	}
	if strings.HasPrefix(domain, "*.") && challenge != CertificateChallengeDNS {
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
	if m.acmeStorage == nil {
		return nil, errors.New("证书存储尚未初始化")
	}

	email := strings.TrimSpace(request.Email)
	if email == "" {
		email = strings.TrimSpace(m.options.ACMEEmail)
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
		Storage: m.acmeStorage,
		Logger:  logger,
	})
	issuerOptions := certmagic.ACMEIssuer{
		CA:                      caURL,
		Email:                   email,
		Agreed:                  true,
		Logger:                  logger,
		DisableTLSALPNChallenge: true,
		ListenHost:              strings.TrimSpace(m.options.HTTPChallengeHost),
		AltHTTPPort:             m.options.HTTPChallengePort,
	}
	if isIP {
		// Let's Encrypt 的 IP 证书必须使用短期配置，有效期为 160 小时。
		issuerOptions.Profile = "shortlived"
	}
	if challenge == CertificateChallengeDNS {
		credentials := request.DNSCredentials
		credentials.AliyunAccessKeyID = strings.TrimSpace(credentials.AliyunAccessKeyID)
		credentials.AliyunAccessKeySecret = strings.TrimSpace(credentials.AliyunAccessKeySecret)
		credentials.DNSPodAPIToken = strings.TrimSpace(credentials.DNSPodAPIToken)
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
		manager := certmagic.DNSManager{Logger: logger}
		switch providerName {
		case DNSProviderAliDNS:
			provider = &libdnsAliDNS.Provider{CredentialInfo: libdnsAliDNS.CredentialInfo{
				AccessKeyID:     credentials.AliyunAccessKeyID,
				AccessKeySecret: credentials.AliyunAccessKeySecret,
			}}
		case DNSProviderDNSPod:
			parts := strings.Split(credentials.DNSPodAPIToken, ",")
			credentials.DNSPodAPIToken = strings.TrimSpace(parts[0]) + "," + strings.TrimSpace(parts[1])
			provider = &caddyDNSPod.Provider{APIToken: credentials.DNSPodAPIToken}
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
			// 火山引擎记录生效较慢，延长传播检测时间。
			manager.TTL = 600 * time.Second
			manager.PropagationTimeout = 15 * time.Minute
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
	certificateKey := certmagic.StorageKeys.SiteCert(issuerKey, domain)
	privateKey := certmagic.StorageKeys.SitePrivateKey(issuerKey, domain)
	metadataKey := certmagic.StorageKeys.SiteMeta(issuerKey, domain)
	if m.acmeStorage.Exists(ctx, certificateKey) && m.acmeStorage.Exists(ctx, privateKey) && m.acmeStorage.Exists(ctx, metadataKey) {
		// 已有证书由 CertMagic 根据 ARI 和有效期判断续签，显式续签则强制执行。
		err = magic.RenewCertSync(ctx, domain, renew)
	} else {
		err = magic.ObtainCertSync(ctx, domain)
	}
	if err != nil {
		return nil, fmt.Errorf("%s证书失败: %w", action, err)
	}

	certificatePEM, err := m.acmeStorage.Load(ctx, certificateKey)
	if err != nil {
		return nil, fmt.Errorf("读取证书失败: %w", err)
	}
	privateKeyPEM, err := m.acmeStorage.Load(ctx, privateKey)
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
		CertificateFile: m.acmeStorage.Filename(certificateKey),
		KeyFile:         m.acmeStorage.Filename(privateKey),
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
	case DNSProviderDNSPod:
		parts := strings.Split(credentials.DNSPodAPIToken, ",")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return errors.New("DNSPod API Token 格式必须为 ID,TOKEN")
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
