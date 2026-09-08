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
	"server/app/model"
	certRepository "server/app/repository/cert"
	"server/app/util/page"
	webServer "server/app/util/server"
	"server/app/util/str"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
)

// CreateCert 导入并保存一组证书和私钥。
func (s *service) CreateCert(data *model.Cert) error {
	if data == nil {
		return errors.New("证书数据不能为空")
	}
	candidate := s.cloneCertificate(data)
	candidate.Type = cert_type.Manual
	if err := s.prepareCertificate(candidate); err != nil {
		return err
	}
	candidate.Id = str.GetSnowWorkIns().GetId()

	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.database.Transaction(func(tx *gorm.DB) error {
		return s.repositoryCert.Create(candidate, tx)
	}); err != nil {
		return fmt.Errorf("创建证书失败: %w", err)
	}
	*data = *s.cloneCertificate(candidate)
	return nil
}

// UpdateCert 更新证书。域名和有效期始终从实际证书内容重新提取。
func (s *service) UpdateCert(id int64, data *certRepository.UpdateCert) error {
	if id <= 0 {
		return errors.New("证书 ID 不合法")
	}
	if data == nil {
		return errors.New("证书更新数据不能为空")
	}
	if data.Type != nil || data.Domains != nil || data.StartTime != nil || data.ExpireTime != nil {
		return errors.New("证书类型、域名和有效期不能直接修改")
	}
	if data.Name == nil && data.Certificate == nil && data.PrivateKey == nil {
		return nil
	}

	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	var previous *model.Cert
	contentChanged := data.Certificate != nil || data.PrivateKey != nil
	err := s.database.Transaction(func(tx *gorm.DB) error {
		current, err := s.repositoryCert.FindById(id, tx)
		if err != nil {
			return err
		}
		previous = s.cloneCertificate(current)
		candidate := s.cloneCertificate(current)
		if data.Name != nil {
			candidate.Name = *data.Name
		}
		if data.Certificate != nil {
			candidate.Certificate = *data.Certificate
		}
		if data.PrivateKey != nil {
			candidate.PrivateKey = *data.PrivateKey
		}
		if contentChanged {
			if err = s.prepareCertificate(candidate); err != nil {
				return err
			}
		} else {
			candidate.Name, err = s.validateCertificateMetadata(candidate.Name, candidate.Type)
			if err != nil {
				return err
			}
		}
		return s.repositoryCert.UpdateById(id, s.certificateUpdate(candidate), tx)
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
		return s.restoreCertificate(previous)
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
		count, err := s.repositoryDomain.CountByCertId(id, tx)
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
func (s *service) SelectCert(where *certRepository.SelectCert, queryPage *page.Query) (*page.Result[*certRepository.Cert], error) {
	var filter *certRepository.SelectCert
	if where != nil {
		value := *where
		value.Keyword = strings.TrimSpace(value.Keyword)
		value.Name = strings.TrimSpace(value.Name)
		value.Type = strings.TrimSpace(value.Type)
		if value.Type != "" {
			certificateType, err := strconv.Atoi(value.Type)
			if err != nil {
				return nil, errors.New("证书类型参数错误")
			}
			if _, exists := cert_type.Map()[certificateType]; !exists {
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
func (s *service) ObtainCert(ctx context.Context, name string, request webServer.CertificateRequest) (*model.Cert, error) {
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
	issued, err := s.http.ObtainCertificate(ctx, request)
	if err != nil {
		return nil, err
	}
	candidate := &model.Cert{
		Id:          str.GetSnowWorkIns().GetId(),
		Name:        name,
		Type:        cert_type.ACME,
		Certificate: string(issued.CertificatePEM),
		PrivateKey:  string(issued.PrivateKeyPEM),
	}
	if err = s.prepareCertificate(candidate); err != nil {
		return nil, fmt.Errorf("申请得到的证书无效: %w", err)
	}
	if err = s.database.Transaction(func(tx *gorm.DB) error {
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
	return s.cloneCertificate(candidate), nil
}

// RenewCert 续签 ACME 证书并更新现有记录。
func (s *service) RenewCert(ctx context.Context, id int64, request webServer.CertificateRequest) (*model.Cert, error) {
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

	issued, err := s.http.RenewCertificate(ctx, request)
	if err != nil {
		return nil, err
	}
	previous := s.cloneCertificate(current)
	candidate := s.cloneCertificate(current)
	candidate.Certificate = string(issued.CertificatePEM)
	candidate.PrivateKey = string(issued.PrivateKeyPEM)
	if err = s.prepareCertificate(candidate); err != nil {
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
		return s.repositoryCert.UpdateById(id, s.certificateUpdate(candidate), tx)
	}); err != nil {
		return nil, fmt.Errorf("保存续签证书失败: %w", err)
	}
	if err = s.syncCertificateLocked(id, "部署续签证书", func() error {
		return s.restoreCertificate(previous)
	}); err != nil {
		return nil, err
	}
	result, err := s.repositoryCert.FindById(id)
	if err != nil {
		return nil, fmt.Errorf("读取续签证书失败: %w", err)
	}
	return result, nil
}

func (s *service) prepareCertificate(data *model.Cert) error {
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
	if !now.Before(leaf.NotAfter) {
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

func (s *service) cloneCertificate(data *model.Cert) *model.Cert {
	if data == nil {
		return nil
	}
	result := *data
	return &result
}

func (s *service) certificateUpdate(data *model.Cert) *certRepository.UpdateCert {
	name := data.Name
	certificateType := data.Type
	domains := data.Domains
	certificate := data.Certificate
	privateKey := data.PrivateKey
	startTime := data.StartTime
	expireTime := data.ExpireTime
	return &certRepository.UpdateCert{
		Name:        &name,
		Type:        &certificateType,
		Domains:     &domains,
		Certificate: &certificate,
		PrivateKey:  &privateKey,
		StartTime:   &startTime,
		ExpireTime:  &expireTime,
	}
}

func (s *service) restoreCertificate(data *model.Cert) error {
	if data == nil {
		return errors.New("缺少待恢复的证书数据")
	}
	return s.database.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&model.Cert{}).Where("id = ?", data.Id).UpdateColumns(map[string]interface{}{
			"name":        data.Name,
			"type":        data.Type,
			"domains":     data.Domains,
			"certificate": data.Certificate,
			"private_key": data.PrivateKey,
			"start_time":  data.StartTime,
			"expire_time": data.ExpireTime,
			"create_time": data.CreateTime,
			"update_time": data.UpdateTime,
		}).Error
	})
}

func (s *service) syncCertificateLocked(id int64, action string, rollback func() error) error {
	if count, err := s.repositoryDomain.CountByCertId(id); err == nil && count == 0 {
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
