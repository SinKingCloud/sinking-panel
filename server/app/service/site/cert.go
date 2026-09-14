package site

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"path"
	"server/app/constant"
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

	"github.com/caddyserver/certmagic"
	acmeTypes "github.com/mholt/acmez/v3/acme"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CertificateRequestTimeout 为 DNS 传播和用户手动验证预留时间。
const CertificateRequestTimeout = constant.CacheTimeWithCertOrder

// CreateCert 导入并保存一组证书和私钥，允许导入已过期的证书。
func (s *service) CreateCert(data *model.Cert) error {
	if data == nil {
		return errors.New("证书数据不能为空")
	}
	candidate := *data
	candidate.Type = cert_type.Import
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
	if data.Domains != nil || data.StartTime != nil || data.ExpireTime != nil {
		return errors.New("证书域名和有效期不能直接修改")
	}
	settingsChanged := data.Type != nil || data.Challenge != nil || data.SecretId != nil || data.AutoRenew != nil
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
		if data.Type != nil || data.Challenge != nil || data.SecretId != nil || (data.AutoRenew != nil && candidate.AutoRenew != 0) {
			request := CertificateRequest{Type: data.Type, SecretId: data.SecretId, AutoRenew: data.AutoRenew}
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
	name, err := s.validateCertificateMetadata(name, cert_type.Auto)
	if err != nil {
		return nil, err
	}
	request.Domain, err = s.normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("证书域名无效: %w", err)
	}

	// 外部验证不持有网站操作锁，签发完成后再保存和同步。
	manual := request.Type != nil && *request.Type == cert_type.Manual
	candidate := &model.Cert{
		Id:   str.GetSnowWorkIns().GetId(),
		Name: name,
		Type: cert_type.Auto,
	}
	acme, err := s.prepareCertificateRequest(candidate, &request)
	if err != nil {
		return nil, err
	}
	var issued *webServer.Certificate
	if manual {
		if request.order == nil {
			return nil, errors.New("请先获取手动验证信息，再提交验证")
		}
		candidate.Id = request.order.CertificateId
		issued, err = request.order.Acme.Submit(ctx)
	} else {
		issued, err = s.http.ObtainCertificate(ctx, acme)
	}
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	candidate.Certificate = string(issued.CertificatePEM)
	candidate.PrivateKey = string(issued.PrivateKeyPEM)
	if err = s.prepareCertificate(candidate, false); err != nil {
		return nil, fmt.Errorf("申请得到的证书无效: %w", err)
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	created := false
	if err = s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.checkCertificateSave(candidate, acme, tx, request.order); err != nil {
			return err
		}
		if manual {
			current, err := s.repositoryCert.FindById(candidate.Id, tx)
			if err == nil {
				if current.Certificate != candidate.Certificate || current.PrivateKey != candidate.PrivateKey {
					return errors.New("证书记录已改变，不能覆盖")
				}
				*candidate = *current
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if err := s.repositoryCert.Create(candidate, tx); err != nil {
			return err
		}
		created = true
		return nil
	}); err != nil {
		return nil, fmt.Errorf("保存申请的证书失败: %w", err)
	}
	if err = s.syncCertificateLocked(candidate.Id, "部署申请的证书", func() error {
		if !created {
			return nil
		}
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
	current, err := s.repositoryCert.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("证书不存在")
		}
		return nil, fmt.Errorf("查询待续签证书失败: %w", err)
	}
	if current.Type != cert_type.Manual && current.Type != cert_type.Auto {
		return nil, errors.New("只有手动申请或自动申请的证书可以续签")
	}
	if request.order != nil && len(request.order.Acme.CertificatePEM) > 0 &&
		strings.TrimSpace(current.Certificate) == strings.TrimSpace(string(request.order.Acme.CertificatePEM)) &&
		strings.TrimSpace(current.PrivateKey) == strings.TrimSpace(string(request.order.Acme.PrivateKeyPEM)) {
		// 上次已保存但响应失败，复用同一订单的结果。
		s.operationMu.Lock()
		defer s.operationMu.Unlock()
		current, err = s.repositoryCert.FindById(id)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(current.Certificate) != strings.TrimSpace(string(request.order.Acme.CertificatePEM)) ||
			strings.TrimSpace(current.PrivateKey) != strings.TrimSpace(string(request.order.Acme.PrivateKeyPEM)) {
			return nil, errors.New("证书已修改，请重新续签")
		}
		if err = s.syncCertificateLocked(id, "部署续签证书", func() error { return nil }); err != nil {
			return nil, err
		}
		return current, nil
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
	manual := candidate.Type == cert_type.Manual
	if manual {
		if request.order == nil || request.order.Previous == nil {
			return nil, errors.New("请先获取手动续签验证信息，再提交验证")
		}
		started := request.order.Previous
		if started.Id != id || current.Certificate != started.Certificate || current.PrivateKey != started.PrivateKey || current.Challenge != started.Challenge || current.Type != started.Type || current.SecretId != started.SecretId || current.AutoRenew != started.AutoRenew {
			return nil, errors.New("证书在验证期间已修改，请重新续签")
		}
	}
	var issued *webServer.Certificate
	if manual {
		issued, err = request.order.Acme.Submit(ctx)
	} else {
		issued, err = s.http.RenewCertificate(ctx, acme)
	}
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	candidate.Certificate = string(issued.CertificatePEM)
	candidate.PrivateKey = string(issued.PrivateKeyPEM)
	if err = s.prepareCertificate(&candidate, false); err != nil {
		return nil, fmt.Errorf("续签得到的证书无效: %w", err)
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = s.database.Transaction(func(tx *gorm.DB) error {
		latest, findErr := s.repositoryCert.FindById(id, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		if findErr != nil {
			return findErr
		}
		if latest.Type != cert_type.Manual && latest.Type != cert_type.Auto {
			return errors.New("证书类型已发生变化，无法续签")
		}
		if latest.Certificate != current.Certificate || latest.PrivateKey != current.PrivateKey || latest.Challenge != current.Challenge || latest.Type != current.Type ||
			(manual && (latest.SecretId != current.SecretId || latest.AutoRenew != current.AutoRenew)) {
			return errors.New("证书在验证期间已修改，请重新续签")
		}
		candidate.Name = latest.Name
		previous = *latest
		if latest.SecretId != current.SecretId || latest.AutoRenew != current.AutoRenew {
			candidate.SecretId = latest.SecretId
			candidate.AutoRenew = latest.AutoRenew
		}
		if err := s.checkCertificateSave(&candidate, acme, tx, request.order); err != nil {
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

// ManualCert 通过普通 HTTP 分步申请证书，使用缓存锁保护共享的 JSON 订单。
func (s *service) ManualCert(ctx context.Context, id int64, name string, request CertificateRequest) (interface{}, error) {
	if ctx == nil {
		return nil, errors.New("手动验证上下文不能为空")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if id < 0 || request.Type == nil || *request.Type != cert_type.Manual {
		return nil, errors.New("手动验证请求不合法")
	}
	if request.Action == "" {
		request.Action = "start"
	}
	if request.Action != "start" {
		if request.Action != "submit" && request.Action != "cancel" {
			return nil, errors.New("手动验证步骤只支持 start、submit 或 cancel")
		}
		if request.SessionId == "" {
			return nil, errors.New("手动验证订单不能为空")
		}
		key := constant.CacheNameWithCertOrder + request.SessionId
		cancelKey := key + ":cancel"
		lockKey := key + ":submit"
		if request.Action == "submit" {
			if s.cache.Get(cancelKey) != nil {
				return nil, errors.New("手动验证订单已取消，请重新申请")
			}
			// 锁的有效期覆盖整个订单，异常退出后也不会让其他节点提前接管。
			if !s.cache.Lock(lockKey, constant.CacheTimeWithCertOrder+time.Minute) {
				return nil, errors.New("正在提交验证，请勿重复操作")
			}
			defer s.cache.UnLock(lockKey)
		}
		// 提交取得锁后再读取，避免使用其他节点更新之前的旧结果。
		raw, ok := s.cache.Get(key).(string)
		if !ok || raw == "" {
			if request.Action == "cancel" {
				return map[string]bool{"cancelled": true}, nil
			}
			return nil, errors.New("手动验证订单已过期或取消，请重新申请")
		}
		var order certificateOrder
		if err := json.Unmarshal([]byte(raw), &order); err != nil || order.Acme == nil || order.SessionId != request.SessionId {
			return nil, errors.New("缓存的验证订单无效，请重新申请")
		}
		if order.Id != id {
			return nil, errors.New("手动验证订单与当前证书不匹配")
		}
		if request.Action == "cancel" {
			// 取消标记不等待提交锁，且保留到订单和提交锁过期，防止并发写回恢复已取消订单。
			s.cache.SetWithExpire(cancelKey, request.SessionId, constant.CacheTimeWithCertOrder+time.Minute)
			if value, _ := s.cache.Get(cancelKey).(string); value != request.SessionId {
				return nil, errors.New("取消验证订单失败，请重试")
			}
			s.cache.Delete(key)
			return map[string]bool{"cancelled": true}, nil
		}
		if s.cache.Get(cancelKey) != nil || !order.ExpiresAt.After(time.Now()) {
			return nil, errors.New("手动验证订单已过期或取消，请重新申请")
		}
		if order.Result != nil {
			return order.Result, nil
		}
		ctx, cancel := context.WithDeadline(ctx, order.ExpiresAt)
		defer cancel()
		// 仅提交期间检查共享缓存，其他节点取消订单后也能中断当前请求。
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if s.cache.Get(cancelKey) != nil || s.cache.Get(key) == nil || !s.cache.IsLock(lockKey) {
						cancel()
						return
					}
				}
			}
		}()
		request = order.Request
		request.order = &order
		var data *model.Cert
		var issueErr error
		if id > 0 {
			data, issueErr = s.RenewCert(ctx, id, request)
		} else {
			data, issueErr = s.ObtainCert(ctx, order.Name, request)
		}
		if s.cache.Get(cancelKey) != nil || s.cache.Get(key) == nil || !order.ExpiresAt.After(time.Now()) || !s.cache.IsLock(lockKey) {
			s.cache.Delete(key)
			return nil, errors.Join(issueErr, errors.New("手动验证订单已过期或取消，请重新申请"))
		}
		order.Result = data
		encoded, err := json.Marshal(&order)
		if err != nil {
			return nil, errors.Join(issueErr, err)
		}
		remaining := time.Until(order.ExpiresAt)
		if remaining <= 0 {
			return nil, errors.Join(issueErr, errors.New("手动验证订单已过期，请重新申请"))
		}
		s.cache.SetWithExpire(key, string(encoded), remaining)
		if s.cache.Get(cancelKey) != nil || !order.ExpiresAt.After(time.Now()) {
			s.cache.Delete(key)
			return nil, errors.Join(issueErr, errors.New("手动验证订单已过期或取消，请重新申请"))
		}
		if saved, _ := s.cache.Get(key).(string); saved != string(encoded) {
			return nil, errors.Join(issueErr, errors.New("保存验证进度失败，请重试"))
		}
		if issueErr != nil {
			return nil, issueErr
		}
		return data, nil
	}
	if request.SessionId != "" {
		return nil, errors.New("重新申请时不能携带旧的验证订单")
	}
	domain, err := s.normalizeDomain(request.Domain)
	if err != nil {
		return nil, fmt.Errorf("证书域名无效: %w", err)
	}
	request.Domain = domain
	candidate := &model.Cert{Name: name, Type: cert_type.Auto}
	var previous *model.Cert
	if id > 0 {
		current, err := s.repositoryCert.FindById(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("证书不存在")
			}
			return nil, err
		}
		if current.Type != cert_type.Manual && current.Type != cert_type.Auto {
			return nil, errors.New("只有手动申请或自动申请的证书可以续签")
		}
		var domains []string
		if err = json.Unmarshal([]byte(current.Domains), &domains); err != nil {
			return nil, errors.New("证书域名数据不合法")
		}
		covered := false
		for _, value := range domains {
			if strings.EqualFold(value, domain) {
				covered = true
				break
			}
		}
		if !covered {
			return nil, errors.New("续签域名不属于当前证书")
		}
		previous = current
		copy := *current
		candidate = &copy
	}
	candidate.Name, err = s.validateCertificateMetadata(candidate.Name, candidate.Type)
	if err != nil {
		return nil, err
	}
	acmeRequest, err := s.prepareCertificateRequest(candidate, &request)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(constant.CacheTimeWithCertOrder)
	ctx, cancel := context.WithDeadline(ctx, expiresAt)
	defer cancel()
	client := request.acme
	if client == nil {
		client = webServer.NewAcme()
	}
	var challenges []acmeTypes.Challenge
	if id > 0 {
		challenges, err = client.Renew(ctx, acmeRequest)
	} else {
		challenges, err = client.Obtain(ctx, acmeRequest)
	}
	if err != nil {
		return nil, err
	}
	values := make([]map[string]string, 0, len(challenges))
	for _, challenge := range challenges {
		value := map[string]string{"id": challenge.Token}
		switch challenge.Type {
		case acmeTypes.ChallengeTypeHTTP01:
			host := challenge.Identifier.Value
			if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
				host = "[" + host + "]"
			}
			address := url.URL{Scheme: "http", Host: host, Path: challenge.HTTP01ResourcePath()}
			value["name"], value["type"], value["value"] = challenge.Token, "HTTP", challenge.KeyAuthorization
			value["path"], value["url"] = address.Path, address.String()
		case acmeTypes.ChallengeTypeDNS01:
			value["name"], value["type"], value["value"] = challenge.DNS01TXTRecordName(), "TXT", challenge.DNS01KeyAuthorization()
		default:
			cancel()
			return nil, errors.New("不支持的手动验证方式")
		}
		values = append(values, value)
	}
	certificateType, secretId, autoRenew := cert_type.Manual, int64(0), 0
	sessionId := rand.Text()
	certificateId := id
	if certificateId == 0 {
		certificateId = str.GetSnowWorkIns().GetId()
	}
	order := &certificateOrder{
		Acme: client, Id: id, CertificateId: certificateId, SessionId: sessionId, ExpiresAt: expiresAt,
		Name: candidate.Name, Previous: previous,
		Request: CertificateRequest{
			Domain: domain, Email: client.Email, CA: client.CA, Type: &certificateType,
			Challenge: acmeRequest.Challenge, SecretId: &secretId, AutoRenew: &autoRenew,
		},
	}
	encoded, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	remaining := time.Until(expiresAt)
	if remaining <= 0 {
		return nil, errors.New("手动验证订单已过期，请重新申请")
	}
	key := constant.CacheNameWithCertOrder + sessionId
	s.cache.SetWithExpire(key, string(encoded), remaining)
	if saved, _ := s.cache.Get(key).(string); saved != string(encoded) {
		return nil, errors.New("保存验证订单失败，请重试")
	}
	return map[string]interface{}{"session_id": sessionId, "challenges": values, "expires_at": expiresAt.Unix()}, nil
}

// prepareCertificateRequest 合并证书申请设置，并从关联密钥读取 DNS 凭据。
func (s *service) prepareCertificateRequest(data *model.Cert, request *CertificateRequest, tx ...*gorm.DB) (webServer.CertificateRequest, error) {
	acme := webServer.CertificateRequest{
		Domain: request.Domain,
		Email:  request.Email,
		CA:     request.CA,
	}
	options := s.http.Options()
	if data.Id > 0 && data.Certificate != "" && request.Domain != "" && request.order == nil {
		storage := &certmagic.FileStorage{Path: options.ACMEPath}
		key := fmt.Sprintf("renewals/%d/%x.json", data.Id, sha256.Sum256([]byte(strings.TrimSpace(data.Certificate))))
		encoded, err := storage.Load(context.Background(), key)
		if err == nil {
			request.acme = webServer.NewAcme()
			if err = json.Unmarshal(encoded, request.acme); err != nil {
				return acme, fmt.Errorf("读取证书续签信息失败: %w", err)
			}
		} else if errors.Is(err, fs.ErrNotExist) {
			// 原签发文件已有账户与 CA 信息，按证书和账户 URL 精确恢复。
			var names []string
			if err := json.Unmarshal([]byte(data.Domains), &names); err != nil {
				return acme, fmt.Errorf("读取证书域名失败: %w", err)
			}
			resource := certmagic.CertificateResource{SANs: append([]string(nil), names...)}
			names = append(names, request.Domain, resource.NamesKey())
			fallbackCtx := context.Background()
			for _, caName := range []webServer.CertificateCA{webServer.CertificateCAProd, webServer.CertificateCAStaging} {
				caURL := certmagic.LetsEncryptProductionCA
				if caName == webServer.CertificateCAStaging {
					caURL = certmagic.LetsEncryptStagingCA
				}
				issuerKey := (&certmagic.ACMEIssuer{CA: caURL}).IssuerKey()
				for _, name := range names {
					if name == "" {
						continue
					}
					certificatePEM, err := storage.Load(fallbackCtx, certmagic.StorageKeys.SiteCert(issuerKey, name))
					if errors.Is(err, fs.ErrNotExist) {
						continue
					}
					if err != nil {
						return acme, fmt.Errorf("读取原 ACME 证书失败: %w", err)
					}
					if !bytes.Equal(bytes.TrimSpace(certificatePEM), bytes.TrimSpace([]byte(data.Certificate))) {
						continue
					}
					metadata, err := storage.Load(fallbackCtx, certmagic.StorageKeys.SiteMeta(issuerKey, name))
					if err != nil {
						return acme, fmt.Errorf("读取原 ACME 证书信息失败: %w", err)
					}
					var saved certmagic.CertificateResource
					if err = json.Unmarshal(metadata, &saved); err != nil {
						return acme, fmt.Errorf("解析原 ACME 证书信息失败: %w", err)
					}
					var issued acmeTypes.Certificate
					if err = json.Unmarshal(saved.IssuerData, &issued); err != nil {
						return acme, fmt.Errorf("解析原 ACME 签发信息失败: %w", err)
					}
					if issued.CA != caURL || issued.Account == "" {
						return acme, errors.New("原 ACME 签发环境或账户信息不完整")
					}
					accounts, err := storage.List(fallbackCtx, path.Join("acme", certmagic.StorageKeys.Safe(issuerKey), "users"), true)
					if err != nil {
						return acme, fmt.Errorf("读取原 ACME 账户目录失败: %w", err)
					}
					for _, accountKey := range accounts {
						if !strings.HasSuffix(accountKey, ".json") {
							continue
						}
						encoded, err := storage.Load(fallbackCtx, accountKey)
						if err != nil {
							return acme, fmt.Errorf("读取 ACME 账户信息失败: %w", err)
						}
						var account acmeTypes.Account
						if err = json.Unmarshal(encoded, &account); err != nil {
							return acme, fmt.Errorf("解析 ACME 账户信息失败: %w", err)
						}
						if account.Location != issued.Account {
							continue
						}
						accountPrivateKeyPEM, err := storage.Load(fallbackCtx, strings.TrimSuffix(accountKey, ".json")+".key")
						if err != nil {
							return acme, fmt.Errorf("读取原 ACME 账户密钥失败: %w", err)
						}
						if _, err = certmagic.PEMDecodePrivateKey(accountPrivateKeyPEM); err != nil {
							return acme, fmt.Errorf("解析原 ACME 账户密钥失败: %w", err)
						}
						email := ""
						if len(account.Contact) > 0 {
							email = strings.TrimPrefix(account.Contact[0], "mailto:")
						}
						request.acme = &webServer.Acme{
							Domain: request.Domain, CA: caName, Email: email,
							AccountURL: account.Location, AccountPrivateKeyPEM: accountPrivateKeyPEM,
							CertificatePEM: []byte(data.Certificate), PrivateKeyPEM: []byte(data.PrivateKey),
						}
						break
					}
					if request.acme == nil {
						return acme, errors.New("找不到原 ACME 账户，请检查 ACME 数据目录")
					}
					break
				}
				if request.acme != nil {
					break
				}
			}
		} else {
			return acme, fmt.Errorf("读取证书续签信息失败: %w", err)
		}
		if request.acme != nil {
			if acme.CA == "" {
				acme.CA = request.acme.CA
			}
			if acme.Email == "" {
				acme.Email = request.acme.Email
			}
		}
	}
	if acme.CA == "" {
		acme.CA = webServer.CertificateCAProd
	}
	if acme.Email == "" && request.acme == nil {
		acme.Email = strings.TrimSpace(options.ACMEEmail)
	}
	if request.acme != nil && acme.CA == request.acme.CA && strings.EqualFold(acme.Email, request.acme.Email) {
		acme.AccountKeyPEM = string(request.acme.AccountPrivateKeyPEM)
	}
	if request.Type != nil {
		if data.Type == cert_type.Import {
			return acme, errors.New("导入的证书不能修改为申请证书")
		}
		if *request.Type != cert_type.Manual && *request.Type != cert_type.Auto {
			return acme, errors.New("证书申请类型只支持手动申请或自动申请")
		}
		data.Type = *request.Type
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
	if data.Type == cert_type.Import {
		if data.Challenge != "" || data.SecretId != 0 || data.AutoRenew != 0 {
			return acme, errors.New("导入的证书不能设置 ACME 申请方式、密钥或自动续签")
		}
		return acme, nil
	}
	if data.Challenge == "" {
		data.Challenge = string(webServer.CertificateChallengeHTTP)
	}
	switch data.Type {
	case cert_type.Manual:
		acme.Action = webServer.CertificateActionManual
	case cert_type.Auto:
		acme.Action = webServer.CertificateActionAuto
	default:
		return acme, errors.New("证书申请类型只支持手动申请或自动申请")
	}
	acme.Challenge = webServer.CertificateChallenge(data.Challenge)
	if acme.Challenge != webServer.CertificateChallengeHTTP && acme.Challenge != webServer.CertificateChallengeDNS {
		return acme, errors.New("证书验证方式只支持 http 或 dns")
	}
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
	if data.Type == cert_type.Manual {
		if request.SecretId != nil && *request.SecretId != 0 {
			return acme, errors.New("手动验证不需要关联密钥")
		}
		if request.AutoRenew != nil && *request.AutoRenew != 0 {
			return acme, errors.New("手动验证不支持自动续签")
		}
		data.SecretId = 0
		data.AutoRenew = 0
		return acme, nil
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
		return acme, errors.New("证书验证方式不合法")
	}
	return acme, nil
}

// checkCertificateSave 确认订单归属和关联密钥，保存签发进度及供下次续签使用的信息。
func (s *service) checkCertificateSave(data *model.Cert, request webServer.CertificateRequest, tx *gorm.DB, order *certificateOrder) error {
	if order != nil {
		key := constant.CacheNameWithCertOrder + order.SessionId
		if s.cache.Get(key+":cancel") != nil || s.cache.Get(key) == nil || !s.cache.IsLock(key+":submit") || !order.ExpiresAt.After(time.Now()) {
			return errors.New("验证订单已过期或取消，不能保存证书")
		}
		encoded, err := json.Marshal(order)
		if err != nil {
			return err
		}
		remaining := time.Until(order.ExpiresAt)
		if remaining <= 0 {
			return errors.New("验证订单已过期，不能保存证书")
		}
		s.cache.SetWithExpire(key, string(encoded), remaining)
		if s.cache.Get(key+":cancel") != nil || !order.ExpiresAt.After(time.Now()) {
			s.cache.Delete(key)
			return errors.New("验证订单已过期或取消，不能保存证书")
		}
		if saved, _ := s.cache.Get(key).(string); saved != string(encoded) {
			return errors.New("保存签发进度失败，请重试")
		}
	}
	settings := &CertificateRequest{Domain: request.Domain, CA: request.CA, Email: request.Email, order: order}
	if data.Type == cert_type.Auto && data.Challenge == string(webServer.CertificateChallengeDNS) && data.SecretId == 0 {
		// 密钥在续签期间被删除时，保留解除关联的结果，仍可保存已签发的证书。
		settings.Domain = ""
	}
	acme, err := s.prepareCertificateRequest(data, settings, tx)
	if err != nil {
		return err
	}
	if data.SecretId != 0 && acme.DNSProvider != request.DNSProvider {
		return errors.New("关联密钥的服务商已改变，请重新申请或续签")
	}
	renewal := settings.acme
	if order != nil {
		renewal = order.Acme
	}
	if renewal == nil {
		renewal = &webServer.Acme{Domain: request.Domain, CA: request.CA, Email: request.Email}
	}
	encoded, err := json.Marshal(renewal)
	if err != nil {
		return fmt.Errorf("序列化证书续签信息失败: %w", err)
	}
	// 按证书内容保存，失败的续签或数据库回滚不会覆盖上一张证书的账户信息。
	storage := &certmagic.FileStorage{Path: s.http.Options().ACMEPath}
	key := fmt.Sprintf("renewals/%d/%x.json", data.Id, sha256.Sum256([]byte(strings.TrimSpace(data.Certificate))))
	if err = storage.Store(context.Background(), key, encoded); err != nil {
		return fmt.Errorf("保存证书续签信息失败: %w", err)
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
		if current.Type != expected.Type || current.Challenge != expected.Challenge {
			restored.Type = current.Type
			restored.Challenge = current.Challenge
			restored.SecretId = current.SecretId
			restored.AutoRenew = current.AutoRenew
			restored.UpdateTime = current.UpdateTime
		} else if current.SecretId != expected.SecretId || current.AutoRenew != expected.AutoRenew {
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
