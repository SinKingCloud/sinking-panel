package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/mholt/acmez/v3"
	"github.com/mholt/acmez/v3/acme"
)

// Acme 分步申请证书：Obtain 或 Renew 返回验证信息，配置好 HTTP 文件或 DNS 记录后调用 Submit。
// 可序列化到缓存并恢复；JSON 包含账户密钥和证书私钥，仅供服务端保存。
// 一个实例处理一个订单，context 只控制当前调用，不保存在订单中。
type Acme struct {
	AccountURL           string                    `json:"account_url"`
	AccountPrivateKeyPEM []byte                    `json:"account_private_key_pem"`
	Order                acme.Order                `json:"order"`
	OrderURL             string                    `json:"order_url"`  // Order.Location 不参与 JSON 序列化。
	Challenges           map[string]acme.Challenge `json:"challenges"` // 授权 URL 对应所选验证方式。
	Domain               string                    `json:"domain"`
	CA                   CertificateCA             `json:"ca"`
	Email                string                    `json:"email"`
	CSR                  []byte                    `json:"csr"`
	CertificatePEM       []byte                    `json:"certificate_pem"`
	PrivateKeyPEM        []byte                    `json:"private_key_pem"`

	mu     sync.Mutex
	client *acme.Client
}

// NewAcme 创建证书申请器；缓存的 JSON 可直接反序列化到此实例，再调用 Submit。
func NewAcme() *Acme {
	return &Acme{}
}

// Obtain 创建新订单并返回待配置的验证信息；授权已有效时返回空列表，仍调用 Submit 获取证书。
func (m *Acme) Obtain(ctx context.Context, request CertificateRequest) ([]acme.Challenge, error) {
	return m.start(ctx, request, false)
}

// Renew 创建续签订单；同一实例内续签同一证书时会附带上次签发的替换信息。
func (m *Acme) Renew(ctx context.Context, request CertificateRequest) ([]acme.Challenge, error) {
	return m.start(ctx, request, true)
}

func (m *Acme) start(ctx context.Context, request CertificateRequest, renew bool) ([]acme.Challenge, error) {
	if ctx == nil {
		return nil, errors.New("证书申请上下文不能为空")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("Acme 尚未初始化")
	}
	if !m.mu.TryLock() {
		return nil, errors.New("当前证书订单正在处理")
	}
	defer m.mu.Unlock()
	if m.OrderURL != "" && m.Order.Status != acme.StatusInvalid && len(m.CertificatePEM) == 0 &&
		(m.Order.Expires.IsZero() || time.Now().Before(m.Order.Expires)) {
		return nil, errors.New("当前订单尚未完成，请先提交验证或取消原订单")
	}
	domain, err := normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("证书域名无效: %w", err)
	}
	if !certmagic.SubjectQualifiesForPublicCert(domain) {
		return nil, errors.New("域名或 IP 地址不能申请公开证书")
	}
	challenge := CertificateChallenge(strings.ToLower(strings.TrimSpace(string(request.Challenge))))
	if challenge == "" {
		challenge = CertificateChallengeHTTP
	}
	if challenge != CertificateChallengeHTTP && challenge != CertificateChallengeDNS {
		return nil, errors.New("证书验证类型只支持 HTTP 或 DNS")
	}
	if net.ParseIP(domain) != nil && challenge != CertificateChallengeHTTP {
		return nil, errors.New("IP 地址证书必须使用 HTTP 验证")
	}
	if strings.HasPrefix(domain, "*.") && challenge != CertificateChallengeDNS {
		return nil, errors.New("通配符证书必须使用 DNS 验证")
	}
	caName := CertificateCA(strings.ToLower(strings.TrimSpace(string(request.CA))))
	caURL := certmagic.LetsEncryptProductionCA
	switch caName {
	case "", CertificateCAProd:
		caName = CertificateCAProd
	case CertificateCAStaging:
		caURL = certmagic.LetsEncryptStagingCA
	default:
		return nil, errors.New("证书 CA 只支持 production 或 staging")
	}
	email := strings.TrimSpace(request.Email)
	if email != "" {
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email {
			return nil, errors.New("ACME 邮箱地址无效")
		}
	}
	sameAccount := m.AccountURL != "" && len(m.AccountPrivateKeyPEM) > 0 && m.CA == caName && strings.EqualFold(m.Email, email)
	client := m.client
	if client == nil || client.Directory != caURL {
		client = &acme.Client{Directory: caURL, HTTPClient: &http.Client{Timeout: 30 * time.Second}, PollTimeout: 5 * time.Minute, UserAgent: "sinking-panel"}
	}
	account := acme.Account{TermsOfServiceAgreed: true}
	if email != "" {
		account.Contact = []string{"mailto:" + email}
	}
	if sameAccount {
		account.Location = m.AccountURL
		account.PrivateKey, err = certmagic.PEMDecodePrivateKey(m.AccountPrivateKeyPEM)
	} else {
		account.PrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}
	if err != nil {
		return nil, fmt.Errorf("读取 ACME 账户密钥失败: %w", err)
	}
	accountPrivateKeyPEM, err := certmagic.PEMEncodePrivateKey(account.PrivateKey)
	if err != nil {
		return nil, err
	}
	if account.Location == "" {
		account, err = client.NewAccount(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("注册 ACME 账户失败: %w", err)
		}
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	csr, err := acmez.NewCSR(key, []string{domain})
	if err != nil {
		return nil, err
	}
	privateKeyPEM, err := certmagic.PEMEncodePrivateKey(key)
	if err != nil {
		return nil, err
	}
	params, err := acmez.OrderParametersFromCSR(account, csr)
	if err != nil {
		return nil, err
	}
	order := acme.Order{Identifiers: params.Identifiers}
	if net.ParseIP(domain) != nil {
		order.Profile = "shortlived"
	}
	if renew && sameAccount && m.Domain == domain && len(m.CertificatePEM) > 0 {
		pair, err := tls.X509KeyPair(m.CertificatePEM, m.PrivateKeyPEM)
		if err != nil {
			return nil, err
		}
		previous, err := x509.ParseCertificate(pair.Certificate[0])
		if err != nil {
			return nil, err
		}
		order.Replaces, err = acme.ARIUniqueIdentifier(previous)
		if err != nil {
			return nil, err
		}
	}
	created, err := client.NewOrder(ctx, account, order)
	var problem acme.Problem
	if err != nil && order.Replaces != "" && errors.As(err, &problem) && problem.Type == acme.ProblemTypeAlreadyReplaced {
		order.Replaces = ""
		created, err = client.NewOrder(ctx, account, order)
	}
	if err != nil {
		return nil, fmt.Errorf("创建证书订单失败: %w", err)
	}
	selected := make(map[string]acme.Challenge)
	var challenges []acme.Challenge
	for _, address := range created.Authorizations {
		authorization, err := client.GetAuthorization(ctx, account, address)
		if err != nil {
			return nil, err
		}
		if authorization.Status == acme.StatusValid {
			continue
		}
		if authorization.Status != acme.StatusPending {
			return nil, fmt.Errorf("证书授权状态不可用: %s", authorization.Status)
		}
		found := false
		for _, value := range authorization.Challenges {
			if value.Type == string(challenge)+"-01" {
				selected[address] = value
				challenges = append(challenges, value)
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("CA 未提供所选的验证方式")
		}
	}
	m.Domain, m.CA, m.Email = domain, caName, email
	m.AccountURL, m.AccountPrivateKeyPEM = account.Location, accountPrivateKeyPEM
	m.client, m.Order, m.OrderURL, m.Challenges = client, created, created.Location, selected
	m.CSR, m.PrivateKeyPEM, m.CertificatePEM = csr.Raw, privateKeyPEM, nil
	return challenges, nil
}

// Submit 提交验证并返回证书 PEM、私钥和有效期等信息；下载失败后可再次提交同一订单。
func (m *Acme) Submit(ctx context.Context) (*Certificate, error) {
	if ctx == nil {
		return nil, errors.New("证书验证上下文不能为空")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("Acme 尚未初始化")
	}
	if !m.mu.TryLock() {
		return nil, errors.New("当前证书订单正在处理")
	}
	defer m.mu.Unlock()
	if m.OrderURL == "" || m.AccountURL == "" || len(m.AccountPrivateKeyPEM) == 0 || m.Domain == "" || len(m.CSR) == 0 || len(m.PrivateKeyPEM) == 0 {
		return nil, errors.New("请先申请或续签证书")
	}
	caURL := certmagic.LetsEncryptProductionCA
	switch m.CA {
	case CertificateCAProd:
	case CertificateCAStaging:
		caURL = certmagic.LetsEncryptStagingCA
	default:
		return nil, errors.New("证书 CA 只支持 production 或 staging")
	}
	if m.client == nil || m.client.Directory != caURL {
		m.client = &acme.Client{Directory: caURL, HTTPClient: &http.Client{Timeout: 30 * time.Second}, PollTimeout: 5 * time.Minute, UserAgent: "sinking-panel"}
	}
	account := acme.Account{Location: m.AccountURL}
	var err error
	account.PrivateKey, err = certmagic.PEMDecodePrivateKey(m.AccountPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("读取 ACME 账户密钥失败: %w", err)
	}
	m.Order.Location = m.OrderURL
	certificatePEM := m.CertificatePEM
	if len(certificatePEM) == 0 {
		var err error
		m.Order, err = m.client.GetOrder(ctx, account, m.Order)
		if err != nil {
			return nil, err
		}
		// 上次提交可能已被 CA 接收；处理中的订单只查询结果，不重复提交 CSR。
		if m.Order.Status == acme.StatusProcessing {
			pollCtx, stopPolling := context.WithTimeout(ctx, 5*time.Minute)
			defer stopPolling()
			for m.Order.Status == acme.StatusProcessing {
				select {
				case <-pollCtx.Done():
					return nil, pollCtx.Err()
				case <-time.After(time.Second):
				}
				m.Order, err = m.client.GetOrder(pollCtx, account, m.Order)
				if err != nil {
					return nil, err
				}
			}
		}
		if m.Order.Status == acme.StatusInvalid {
			return nil, fmt.Errorf("证书订单已失效: %v", m.Order.Error)
		}
		if m.Order.Status != acme.StatusValid && !m.Order.Expires.IsZero() && time.Now().After(m.Order.Expires) {
			return nil, errors.New("证书订单已过期，请重新申请")
		}
		if m.Order.Status != acme.StatusValid {
			for address, selected := range m.Challenges {
				authorization, err := m.client.GetAuthorization(ctx, account, address)
				if err != nil {
					return nil, err
				}
				if authorization.Status == acme.StatusValid {
					continue
				}
				if authorization.Status != acme.StatusPending {
					m.Order.Status = acme.StatusInvalid
					return nil, fmt.Errorf("证书授权状态不可用: %s", authorization.Status)
				}
				found := false
				for _, value := range authorization.Challenges {
					if value.URL == selected.URL && value.Token == selected.Token {
						selected = value
						found = true
						break
					}
				}
				if !found || selected.Status == acme.StatusInvalid {
					m.Order.Status = acme.StatusInvalid
					return nil, fmt.Errorf("验证信息已失效，请重新申请: %v", selected.Error)
				}
				if selected.Status == acme.StatusPending || selected.Status == "" {
					if _, err = m.client.InitiateChallenge(ctx, account, selected); err != nil {
						return nil, fmt.Errorf("提交验证失败: %w", err)
					}
				}
				authorization, err = m.client.PollAuthorization(ctx, account, authorization)
				if err != nil {
					if authorization.Status != acme.StatusPending && authorization.Status != acme.StatusValid {
						m.Order.Status = acme.StatusInvalid
					}
					return nil, fmt.Errorf("证书验证失败: %w", err)
				}
			}
			m.Order, err = m.client.FinalizeOrder(ctx, account, m.Order, m.CSR)
			if err != nil {
				return nil, fmt.Errorf("签发证书失败: %w", err)
			}
		}
		chains, err := m.client.GetCertificateChain(ctx, account, m.Order.Certificate)
		if err != nil {
			return nil, fmt.Errorf("下载证书失败: %w", err)
		}
		if len(chains) == 0 {
			return nil, errors.New("ACME 未返回证书")
		}
		certificatePEM = chains[0].ChainPEM
	}
	pair, err := tls.X509KeyPair(certificatePEM, m.PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return nil, err
	}
	if err = leaf.VerifyHostname(m.Domain); err != nil {
		return nil, fmt.Errorf("签发证书的域名不匹配: %w", err)
	}
	m.CertificatePEM = certificatePEM
	ipAddresses := make([]string, 0, len(leaf.IPAddresses))
	for _, address := range leaf.IPAddresses {
		ipAddresses = append(ipAddresses, address.String())
	}
	return &Certificate{
		Domain: m.Domain, Issuer: string(m.CA), DNSNames: append([]string(nil), leaf.DNSNames...), IPAddresses: ipAddresses,
		SerialNumber: leaf.SerialNumber.String(), NotBefore: leaf.NotBefore, NotAfter: leaf.NotAfter,
		CertificatePEM: append([]byte(nil), certificatePEM...), PrivateKeyPEM: append([]byte(nil), m.PrivateKeyPEM...),
	}, nil
}
