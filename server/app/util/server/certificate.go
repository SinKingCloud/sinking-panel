package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"
)

func (m *Manager) obtainCertificate(ctx context.Context, request CertificateRequest, renew bool) (*Certificate, error) {
	if ctx == nil {
		return nil, errors.New("证书申请上下文不能为空")
	}
	domain, err := m.normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("证书域名无效: %w", err)
	}
	if strings.HasPrefix(domain, "*.") {
		return nil, errors.New("HTTP-01 验证不支持通配符证书")
	}
	if net.ParseIP(domain) != nil || !certmagic.SubjectQualifiesForPublicCert(domain) {
		return nil, errors.New("域名不能申请公开证书")
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
	if m.storage == nil {
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
	logger := caddy.Log().Named("certificate")
	var magic *certmagic.Config
	cache := certmagic.NewCache(certmagic.CacheOptions{
		Logger: logger,
		GetConfigForCert: func(certmagic.Certificate) (*certmagic.Config, error) {
			return magic, nil
		},
	})
	defer cache.Stop()

	magic = certmagic.New(cache, certmagic.Config{
		Storage: m.storage,
		Logger:  logger,
	})
	issuer := certmagic.NewACMEIssuer(magic, certmagic.ACMEIssuer{
		CA:                      caURL,
		Email:                   email,
		Agreed:                  true,
		DisableTLSALPNChallenge: true,
		ListenHost:              strings.TrimSpace(m.options.HTTPChallengeHost),
		AltHTTPPort:             m.options.HTTPChallengePort,
	})
	magic.Issuers = []certmagic.Issuer{issuer}

	issuerKey := issuer.IssuerKey()
	certificateKey := certmagic.StorageKeys.SiteCert(issuerKey, domain)
	privateKey := certmagic.StorageKeys.SitePrivateKey(issuerKey, domain)
	metadataKey := certmagic.StorageKeys.SiteMeta(issuerKey, domain)
	if renew && m.storage.Exists(ctx, certificateKey) && m.storage.Exists(ctx, privateKey) && m.storage.Exists(ctx, metadataKey) {
		err = magic.RenewCertSync(ctx, domain, true)
	} else {
		err = magic.ObtainCertSync(ctx, domain)
	}
	if err != nil {
		if renew {
			return nil, fmt.Errorf("续签证书失败: %w", err)
		}
		return nil, fmt.Errorf("申请证书失败: %w", err)
	}

	certificatePEM, err := m.storage.Load(ctx, certificateKey)
	if err != nil {
		return nil, fmt.Errorf("读取证书失败: %w", err)
	}
	privateKeyPEM, err := m.storage.Load(ctx, privateKey)
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

	return &Certificate{
		Domain:          domain,
		Issuer:          string(caName),
		DNSNames:        append([]string(nil), leaf.DNSNames...),
		SerialNumber:    leaf.SerialNumber.String(),
		NotBefore:       leaf.NotBefore,
		NotAfter:        leaf.NotAfter,
		CertificateFile: m.storage.Filename(certificateKey),
		KeyFile:         m.storage.Filename(privateKey),
		CertificatePEM:  append([]byte(nil), certificatePEM...),
		PrivateKeyPEM:   append([]byte(nil), privateKeyPEM...),
	}, nil
}
