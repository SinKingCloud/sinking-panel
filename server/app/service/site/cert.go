package site

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"server/app/enum/cert_type"
	"server/app/enum/secret_provider"
	"server/app/model"
	repositoryCert "server/app/repository/cert"
	serviceSecret "server/app/service/secret"
	"server/app/util/page"
	webServer "server/app/util/server"
	"server/app/util/str"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
)

// CreateCert 导入并保存一组证书和私钥，允许导入已过期的证书。
func (s *service) CreateCert(data *model.Cert) error {
	if data == nil {
		return errors.New("证书数据不能为空")
	}
	candidate := *data
	candidate.Type = cert_type.Manual
	candidate.Challenge = ""
	candidate.SecretId = 0
	candidate.AutoRenew = 0
	if err := s.prepareCertificate(&candidate, true); err != nil {
		return err
	}
	candidate.Id = str.GetSnowWorkIns().GetId()

	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.database.Transaction(func(tx *gorm.DB) error {
		return s.repositoryCert.Create(&candidate, tx)
	}); err != nil {
		return fmt.Errorf("创建证书失败: %w", err)
	}
	*data = candidate
	return nil
}

// UpdateCert 更新证书。域名和有效期始终从实际证书内容重新提取。
func (s *service) UpdateCert(id int64, data *repositoryCert.UpdateCert) error {
	if id <= 0 {
		return errors.New("证书 ID 不合法")
	}
	if data == nil {
		return errors.New("证书更新数据不能为空")
	}
	if data.Type != nil || data.Domains != nil || data.StartTime != nil || data.ExpireTime != nil {
		return errors.New("证书类型、域名和有效期不能直接修改")
	}
	settingsChanged := data.Challenge != nil || data.SecretId != nil || data.AutoRenew != nil
	if data.Name == nil && data.Certificate == nil && data.PrivateKey == nil && !settingsChanged {
		return nil
	}

	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	var previous, candidate model.Cert
	contentChanged := data.Certificate != nil || data.PrivateKey != nil
	err := s.database.Transaction(func(tx *gorm.DB) error {
		current, err := s.repositoryCert.FindById(id, tx)
		if err != nil {
			return err
		}
		previous = *current
		candidate = *current
		if data.Name != nil {
			candidate.Name = *data.Name
		}
		if data.Certificate != nil {
			candidate.Certificate = *data.Certificate
		}
		if data.PrivateKey != nil {
			candidate.PrivateKey = *data.PrivateKey
		}
		if data.AutoRenew != nil {
			candidate.AutoRenew = *data.AutoRenew
		}
		// 单独关闭自动续签不依赖原密钥仍然存在或有效。
		if data.Challenge != nil || data.SecretId != nil || (data.AutoRenew != nil && candidate.AutoRenew != 0) {
			request := CertificateRequest{SecretId: data.SecretId}
			if data.Challenge != nil {
				request.Challenge = webServer.CertificateChallenge(*data.Challenge)
			}
			if _, err = s.prepareCertificateRequest(&candidate, &request, tx); err != nil {
				return err
			}
		}
		if contentChanged {
			if err = s.prepareCertificate(&candidate, false); err != nil {
				return err
			}
		} else {
			candidate.Name, err = s.validateCertificateMetadata(candidate.Name, candidate.Type)
			if err != nil {
				return err
			}
		}
		return s.repositoryCert.UpdateById(id, &repositoryCert.UpdateCert{
			Name:        &candidate.Name,
			Type:        &candidate.Type,
			Challenge:   &candidate.Challenge,
			SecretId:    &candidate.SecretId,
			AutoRenew:   &candidate.AutoRenew,
			Domains:     &candidate.Domains,
			Certificate: &candidate.Certificate,
			PrivateKey:  &candidate.PrivateKey,
			StartTime:   &candidate.StartTime,
			ExpireTime:  &candidate.ExpireTime,
		}, tx)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("证书不存在")
		}
		return fmt.Errorf("更新证书失败: %w", err)
	}
	if !contentChanged {
		return nil
	}
	return s.syncCertificateLocked(id, "更新证书", func() error {
		return s.restoreCertificate(&previous, &candidate)
	})
}

// DeleteCert 删除未被网站域名使用的证书。
func (s *service) DeleteCert(id int64) error {
	if id <= 0 {
		return errors.New("证书 ID 不合法")
	}

	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	err := s.database.Transaction(func(tx *gorm.DB) error {
		if _, err := s.repositoryCert.FindById(id, tx); err != nil {
			return err
		}
		count, err := s.repositorySiteDomain.CountByCertId(id, tx)
		if err != nil {
			return fmt.Errorf("查询证书引用失败: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("证书正在被 %d 个网站域名使用，不能删除", count)
		}
		return s.repositoryCert.DeleteById(id, tx)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("证书不存在")
		}
		return fmt.Errorf("删除证书失败: %w", err)
	}
	return nil
}

// FindCert 查询证书详情。
func (s *service) FindCert(id int64) (*model.Cert, error) {
	if id <= 0 {
		return nil, errors.New("证书 ID 不合法")
	}
	result, err := s.repositoryCert.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("证书不存在")
		}
		return nil, fmt.Errorf("查询证书失败: %w", err)
	}
	return result, nil
}

// SelectCert 分页查询证书。
func (s *service) SelectCert(where *repositoryCert.SelectCert, queryPage *page.Query) (*page.Result[*repositoryCert.Cert], error) {
	var filter *repositoryCert.SelectCert
	if where != nil {
		value := *where
		if value.Keyword != nil {
			keyword := strings.TrimSpace(*value.Keyword)
			value.Keyword = nil
			if keyword != "" {
				value.Keyword = &keyword
			}
		}
		if value.Name != nil {
			name := strings.TrimSpace(*value.Name)
			value.Name = nil
			if name != "" {
				value.Name = &name
			}
		}
		if value.Type != nil {
			if _, exists := cert_type.Map()[*value.Type]; !exists {
				return nil, errors.New("证书类型不合法")
			}
		}
		filter = &value
	}
	result, err := s.repositoryCert.Select(filter, queryPage)
	if err != nil {
		return nil, fmt.Errorf("查询证书列表失败: %w", err)
	}
	return result, nil
}

// ObtainCert 通过 ACME 申请证书并保存。
func (s *service) ObtainCert(ctx context.Context, name string, request CertificateRequest) (*model.Cert, error) {
	if ctx == nil {
		return nil, errors.New("证书申请上下文不能为空")
	}
	name, err := s.validateCertificateMetadata(name, cert_type.ACME)
	if err != nil {
		return nil, err
	}
	request.Domain, err = s.normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("证书域名无效: %w", err)
	}

	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	candidate := &model.Cert{
		Id:   str.GetSnowWorkIns().GetId(),
		Name: name,
		Type: cert_type.ACME,
	}
	acme, err := s.prepareCertificateRequest(candidate, &request)
	if err != nil {
		return nil, err
	}
	issued, err := s.http.ObtainCertificate(ctx, acme)
	if err != nil {
		return nil, err
	}
	candidate.Certificate = string(issued.CertificatePEM)
	candidate.PrivateKey = string(issued.PrivateKeyPEM)
	if err = s.prepareCertificate(candidate, false); err != nil {
		return nil, fmt.Errorf("申请得到的证书无效: %w", err)
	}
	if err = s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.checkCertificateSecret(candidate, acme, tx); err != nil {
			return err
		}
		return s.repositoryCert.Create(candidate, tx)
	}); err != nil {
		return nil, fmt.Errorf("保存申请的证书失败: %w", err)
	}
	if err = s.syncCertificateLocked(candidate.Id, "部署申请的证书", func() error {
		return s.database.Transaction(func(tx *gorm.DB) error {
			return s.repositoryCert.DeleteById(candidate.Id, tx)
		})
	}); err != nil {
		return nil, err
	}
	return candidate, nil
}

// RenewCert 续签 ACME 证书并更新现有记录。
func (s *service) RenewCert(ctx context.Context, id int64, request CertificateRequest) (*model.Cert, error) {
	if ctx == nil {
		return nil, errors.New("证书续签上下文不能为空")
	}
	if id <= 0 {
		return nil, errors.New("证书 ID 不合法")
	}
	if strings.TrimSpace(request.Domain) == "" {
		return nil, errors.New("续签证书域名不能为空")
	}
	requestedDomain, err := s.normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("续签证书域名无效: %w", err)
	}
	request.Domain = requestedDomain

	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	current, err := s.repositoryCert.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("证书不存在")
		}
		return nil, fmt.Errorf("查询待续签证书失败: %w", err)
	}
	if current.Type != cert_type.ACME {
		return nil, errors.New("只有 ACME 证书可以续签")
	}
	var domains []string
	if err = json.Unmarshal([]byte(current.Domains), &domains); err != nil {
		return nil, fmt.Errorf("解析证书域名失败: %w", err)
	}
	covered := false
	for _, domain := range domains {
		if strings.EqualFold(domain, requestedDomain) {
			covered = true
			break
		}
	}
	if !covered {
		return nil, errors.New("续签域名不属于当前证书")
	}

	var previous model.Cert
	candidate := *current
	acme, err := s.prepareCertificateRequest(&candidate, &request)
	if err != nil {
		return nil, err
	}
	issued, err := s.http.RenewCertificate(ctx, acme)
	if err != nil {
		return nil, err
	}
	candidate.Certificate = string(issued.CertificatePEM)
	candidate.PrivateKey = string(issued.PrivateKeyPEM)
	if err = s.prepareCertificate(&candidate, false); err != nil {
		return nil, fmt.Errorf("续签得到的证书无效: %w", err)
	}
	if err = s.database.Transaction(func(tx *gorm.DB) error {
		latest, findErr := s.repositoryCert.FindById(id, tx)
		if findErr != nil {
			return findErr
		}
		if latest.Type != cert_type.ACME {
			return errors.New("证书类型已发生变化，无法续签")
		}
		previous = *latest
		if latest.SecretId != current.SecretId || latest.AutoRenew != current.AutoRenew {
			candidate.SecretId = latest.SecretId
			candidate.AutoRenew = latest.AutoRenew
		}
		if err := s.checkCertificateSecret(&candidate, acme, tx); err != nil {
			return err
		}
		return s.repositoryCert.UpdateById(id, &repositoryCert.UpdateCert{
			Name:        &candidate.Name,
			Type:        &candidate.Type,
			Challenge:   &candidate.Challenge,
			SecretId:    &candidate.SecretId,
			AutoRenew:   &candidate.AutoRenew,
			Domains:     &candidate.Domains,
			Certificate: &candidate.Certificate,
			PrivateKey:  &candidate.PrivateKey,
			StartTime:   &candidate.StartTime,
			ExpireTime:  &candidate.ExpireTime,
		}, tx)
	}); err != nil {
		return nil, fmt.Errorf("保存续签证书失败: %w", err)
	}
	if err = s.syncCertificateLocked(id, "部署续签证书", func() error {
		return s.restoreCertificate(&previous, &candidate)
	}); err != nil {
		return nil, err
	}
	result, err := s.repositoryCert.FindById(id)
	if err != nil {
		return nil, fmt.Errorf("读取续签证书失败: %w", err)
	}
	return result, nil
}

// prepareCertificateRequest 合并证书申请设置，并从关联密钥读取 DNS 凭据。
func (s *service) prepareCertificateRequest(data *model.Cert, request *CertificateRequest, tx ...*gorm.DB) (webServer.CertificateRequest, error) {
	acme := webServer.CertificateRequest{
		Domain: request.Domain,
		Email:  request.Email,
		CA:     request.CA,
	}
	if challenge := strings.ToLower(strings.TrimSpace(string(request.Challenge))); challenge != "" {
		data.Challenge = challenge
	}
	if request.SecretId != nil {
		data.SecretId = *request.SecretId
	}
	if request.AutoRenew != nil {
		data.AutoRenew = *request.AutoRenew
	}
	if data.AutoRenew != 0 && data.AutoRenew != 1 {
		return acme, errors.New("自动续签只能设置为 0 或 1")
	}
	if data.SecretId < 0 {
		return acme, errors.New("密钥 ID 不合法")
	}
	if data.Type == cert_type.Manual {
		if data.Challenge != "" || data.SecretId != 0 || data.AutoRenew != 0 {
			return acme, errors.New("导入的证书不能设置 ACME 申请方式、密钥或自动续签")
		}
		return acme, nil
	}
	if data.Challenge == "" {
		data.Challenge = string(webServer.CertificateChallengeHTTP)
	}
	acme.Challenge = webServer.CertificateChallenge(data.Challenge)
	domains := []string{request.Domain}
	if request.Domain == "" && data.Domains != "" {
		if err := json.Unmarshal([]byte(data.Domains), &domains); err != nil {
			return acme, errors.New("证书域名数据不合法")
		}
	}
	for _, domain := range domains {
		if net.ParseIP(domain) != nil && acme.Challenge != webServer.CertificateChallengeHTTP {
			return acme, errors.New("IP 地址证书必须使用 HTTP 验证")
		}
		if strings.HasPrefix(domain, "*.") && acme.Challenge != webServer.CertificateChallengeDNS {
			return acme, errors.New("通配符证书必须使用 DNS 验证")
		}
	}
	switch acme.Challenge {
	case webServer.CertificateChallengeHTTP:
		if request.SecretId != nil && *request.SecretId != 0 {
			return acme, errors.New("HTTP 验证不需要关联 DNS 密钥")
		}
		data.SecretId = 0
	case webServer.CertificateChallengeDNS:
		if data.SecretId == 0 {
			// 申请和续签必须关联密钥，编辑证书设置时仍可解除关联。
			if request.Domain != "" {
				return acme, errors.New("请先添加并选择 DNS 验证密钥")
			}
			if data.AutoRenew != 0 {
				return acme, errors.New("DNS 自动续签需要关联已保存的密钥")
			}
			return acme, nil
		}
		secret, err := s.repositorySecret.FindById(data.SecretId, tx...)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return acme, errors.New("关联的密钥不存在")
			}
			return acme, fmt.Errorf("读取 DNS 密钥失败: %w", err)
		}
		var credentials webServer.DNSCredentials
		var provider webServer.DNSProvider
		switch secret.Provider {
		case secret_provider.TencentCloud:
			provider = webServer.DNSProviderTencentCloud
			var values serviceSecret.TencentCloud
			err = json.Unmarshal([]byte(secret.Data), &values)
			credentials.TencentSecretID = values.SecretId
			credentials.TencentSecretKey = values.SecretKey
		case secret_provider.Aliyun:
			provider = webServer.DNSProviderAliDNS
			var values serviceSecret.Aliyun
			err = json.Unmarshal([]byte(secret.Data), &values)
			credentials.AliyunAccessKeyID = values.AccessKeyId
			credentials.AliyunAccessKeySecret = values.AccessKeySecret
		case secret_provider.HuaweiCloud:
			provider = webServer.DNSProviderHuaweiCloud
			var values serviceSecret.HuaweiCloud
			err = json.Unmarshal([]byte(secret.Data), &values)
			credentials.HuaweiAccessKeyID = values.AccessKeyId
			credentials.HuaweiSecretAccessKey = values.SecretAccessKey
		case secret_provider.Volcengine:
			provider = webServer.DNSProviderVolcengine
			var values serviceSecret.Volcengine
			err = json.Unmarshal([]byte(secret.Data), &values)
			credentials.VolcengineAccessKeyID = values.AccessKeyId
			credentials.VolcengineAccessKeySecret = values.AccessKeySecret
		case secret_provider.BaiduCloud:
			provider = webServer.DNSProviderBaiduCloud
			var values serviceSecret.BaiduCloud
			err = json.Unmarshal([]byte(secret.Data), &values)
			credentials.BaiduAccessKeyID = values.AccessKeyId
			credentials.BaiduSecretAccessKey = values.SecretAccessKey
		case secret_provider.DNSPod:
			provider = webServer.DNSProviderDNSPod
			var values serviceSecret.DNSPod
			err = json.Unmarshal([]byte(secret.Data), &values)
			credentials.DNSPodAPIToken = values.APIToken
		default:
			return acme, errors.New("关联密钥的服务商不合法")
		}
		if err != nil {
			return acme, errors.New("关联密钥的 DNS 凭据格式不正确")
		}
		if err = credentials.Validate(provider); err != nil {
			return acme, fmt.Errorf("关联密钥不可用: %w", err)
		}
		// 使用已保存的密钥时，服务商和凭据必须同时来自该记录。
		acme.DNSProvider = provider
		acme.DNSCredentials = credentials
	default:
		return acme, errors.New("证书申请方式只支持 http 或 dns")
	}
	return acme, nil
}

// checkCertificateSecret 在保存签发结果的事务内重查关联，避免签发期间密钥被删除或换厂商。
func (s *service) checkCertificateSecret(data *model.Cert, request webServer.CertificateRequest, tx *gorm.DB) error {
	if data.SecretId == 0 {
		return nil
	}
	acme, err := s.prepareCertificateRequest(data, &CertificateRequest{Domain: request.Domain}, tx)
	if err != nil {
		return err
	}
	if acme.DNSProvider != request.DNSProvider {
		return errors.New("关联密钥的服务商已改变，请重新申请或续签")
	}
	return nil
}

func (s *service) prepareCertificate(data *model.Cert, allowExpired bool) error {
	if data == nil {
		return errors.New("证书数据不能为空")
	}
	name, err := s.validateCertificateMetadata(data.Name, data.Type)
	if err != nil {
		return err
	}
	data.Name = name
	certificate := strings.TrimSpace(data.Certificate)
	privateKey := strings.TrimSpace(data.PrivateKey)
	if certificate == "" || privateKey == "" {
		return errors.New("证书和私钥不能为空")
	}
	leaf, domains, err := s.inspectCertificate(certificate, privateKey)
	if err != nil {
		return err
	}
	now := time.Now()
	if now.Before(leaf.NotBefore) {
		return errors.New("证书尚未生效")
	}
	if !allowExpired && !now.Before(leaf.NotAfter) {
		return errors.New("证书已经过期")
	}
	serverAuth := len(leaf.ExtKeyUsage) == 0
	for _, usage := range leaf.ExtKeyUsage {
		if usage == x509.ExtKeyUsageAny || usage == x509.ExtKeyUsageServerAuth {
			serverAuth = true
			break
		}
	}
	if !serverAuth {
		return errors.New("证书不允许用于 TLS 服务器认证")
	}
	encodedDomains, err := json.Marshal(domains)
	if err != nil {
		return fmt.Errorf("格式化证书域名失败: %w", err)
	}
	data.Certificate = certificate + "\n"
	data.PrivateKey = privateKey + "\n"
	data.Domains = string(encodedDomains)
	data.StartTime = str.DateTime(leaf.NotBefore)
	data.ExpireTime = str.DateTime(leaf.NotAfter)
	return nil
}

func (s *service) validateCertificateMetadata(name string, certificateType int) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("证书名称不能为空")
	}
	if utf8.RuneCountInString(name) > 100 {
		return "", errors.New("证书名称不能超过 100 个字符")
	}
	if _, exists := cert_type.Map()[certificateType]; !exists {
		return "", errors.New("证书类型不合法")
	}
	return name, nil
}

func (s *service) inspectCertificate(certificatePEM, privateKeyPEM string) (*x509.Certificate, []string, error) {
	certificateBytes := bytes.TrimSpace([]byte(certificatePEM))
	privateKeyBytes := bytes.TrimSpace([]byte(privateKeyPEM))
	if len(certificateBytes) == 0 || len(privateKeyBytes) == 0 {
		return nil, nil, errors.New("证书和私钥不能为空")
	}

	remaining := certificateBytes
	parsedCertificates := make([]*x509.Certificate, 0, 2)
	for len(remaining) > 0 {
		remaining = bytes.TrimSpace(remaining)
		if len(remaining) == 0 {
			break
		}
		if !bytes.HasPrefix(remaining, []byte("-----BEGIN CERTIFICATE-----")) {
			return nil, nil, errors.New("证书 PEM 包含非证书内容")
		}
		block, rest := pem.Decode(remaining)
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, nil, errors.New("证书 PEM 格式错误")
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("解析证书失败: %w", err)
		}
		parsedCertificates = append(parsedCertificates, certificate)
		remaining = rest
	}
	if len(parsedCertificates) == 0 {
		return nil, nil, errors.New("证书内容为空")
	}

	if !bytes.HasPrefix(privateKeyBytes, []byte("-----BEGIN ")) {
		return nil, nil, errors.New("私钥 PEM 格式错误")
	}
	keyBlock, rest := pem.Decode(privateKeyBytes)
	if keyBlock == nil || len(bytes.TrimSpace(rest)) != 0 {
		return nil, nil, errors.New("私钥 PEM 必须且只能包含一个私钥")
	}
	var keyErr error
	switch keyBlock.Type {
	case "PRIVATE KEY":
		_, keyErr = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	case "RSA PRIVATE KEY":
		_, keyErr = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		_, keyErr = x509.ParseECPrivateKey(keyBlock.Bytes)
	default:
		return nil, nil, fmt.Errorf("不支持的私钥 PEM 类型: %s", keyBlock.Type)
	}
	if keyErr != nil {
		return nil, nil, fmt.Errorf("解析私钥失败: %w", keyErr)
	}
	if _, err := tls.X509KeyPair(certificateBytes, privateKeyBytes); err != nil {
		return nil, nil, fmt.Errorf("证书和私钥不匹配: %w", err)
	}

	leaf := parsedCertificates[0]
	domainSet := make(map[string]struct{}, len(leaf.DNSNames)+len(leaf.IPAddresses))
	domains := make([]string, 0, len(leaf.DNSNames)+len(leaf.IPAddresses))
	for _, value := range leaf.DNSNames {
		domain, err := s.normalizeDomain(value)
		if err != nil {
			return nil, nil, fmt.Errorf("证书域名 %q 无效: %w", value, err)
		}
		if _, exists := domainSet[domain]; !exists {
			domainSet[domain] = struct{}{}
			domains = append(domains, domain)
		}
	}
	for _, value := range leaf.IPAddresses {
		ip := net.IP(value).String()
		if ip == "<nil>" {
			return nil, nil, errors.New("证书包含无效 IP 地址")
		}
		if _, exists := domainSet[ip]; !exists {
			domainSet[ip] = struct{}{}
			domains = append(domains, ip)
		}
	}
	if len(domains) == 0 {
		return nil, nil, errors.New("证书未包含 DNS 或 IP SAN")
	}
	sort.Strings(domains)
	return leaf, domains, nil
}

func (s *service) restoreCertificate(data, expected *model.Cert) error {
	if data == nil || expected == nil {
		return errors.New("缺少待恢复的证书数据")
	}
	return s.database.Transaction(func(tx *gorm.DB) error {
		current, err := s.repositoryCert.FindById(data.Id, tx)
		if err != nil {
			return err
		}
		restored := *data
		if current.SecretId != expected.SecretId || current.AutoRenew != expected.AutoRenew {
			restored.SecretId = current.SecretId
			restored.AutoRenew = current.AutoRenew
			restored.UpdateTime = current.UpdateTime
		} else if restored.SecretId != 0 && restored.SecretId != expected.SecretId {
			if _, err = s.repositorySecret.FindById(restored.SecretId, tx); err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				restored.SecretId = 0
				restored.AutoRenew = 0
			}
		}
		return s.repositoryCert.Restore(&restored, tx)
	})
}

func (s *service) syncCertificateLocked(id int64, action string, rollback func() error) error {
	if count, err := s.repositorySiteDomain.CountByCertId(id); err == nil && count == 0 {
		return nil
	}
	if err := s.syncLocked(); err != nil {
		syncErr := fmt.Errorf("%s后同步网站运行时失败: %w", action, err)
		rollbackErr := rollback()
		if rollbackErr != nil {
			rollbackErr = fmt.Errorf("补偿恢复证书数据失败: %w", rollbackErr)
		}
		restoreErr := s.syncLocked()
		if restoreErr != nil {
			restoreErr = fmt.Errorf("补偿恢复网站运行时失败: %w", restoreErr)
		}
		if rollbackErr == nil && restoreErr == nil {
			return fmt.Errorf("%w，证书数据和网站运行时已恢复", syncErr)
		}
		return errors.Join(syncErr, rollbackErr, restoreErr)
	}
	return nil
}
