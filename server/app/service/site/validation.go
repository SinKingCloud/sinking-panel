package site

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"server/app/enum/site_status"
	"server/app/enum/site_type"
	"server/app/enum/type_module"
	"server/app/model"
	"server/app/util/str"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/idna"
	"gorm.io/gorm"
)

// prepareSite 校验网站、规范配置，并生成可写入数据库的域名列表。
func (s *service) prepareSite(data *model.Site, input []Domain, previous []*model.Domain, validateExistingCertificates bool, tx ...*gorm.DB) (interface{}, []*model.Domain, error) {
	if data == nil {
		return nil, nil, errors.New("网站数据不能为空")
	}
	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" {
		return nil, nil, errors.New("网站名称不能为空")
	}
	if utf8.RuneCountInString(data.Name) > 100 {
		return nil, nil, errors.New("网站名称不能超过 100 个字符")
	}
	config, value, err := s.checkConfig(data.Config, data.Type)
	if err != nil {
		return nil, nil, err
	}
	data.Config = config
	data.Root = strings.TrimSpace(data.Root)
	data.RunPath = strings.TrimSpace(data.RunPath)
	if data.Type == site_type.Proxy {
		if data.Root != "" || data.RunPath != "" {
			return nil, nil, errors.New("反向代理网站不能配置网站目录")
		}
	} else {
		if data.Root == "" {
			return nil, nil, errors.New("网站根目录不能为空")
		}
		root, err := filepath.Abs(data.Root)
		if err != nil {
			return nil, nil, fmt.Errorf("解析网站根目录失败: %w", err)
		}
		data.Root = filepath.Clean(root)
		runPath := strings.TrimLeft(data.RunPath, "/\\")
		if filepath.VolumeName(runPath) != "" {
			return nil, nil, errors.New("网站运行目录必须位于网站根目录内")
		}
		if runPath == "" || runPath == "." {
			runPath = ""
		} else {
			runPath = filepath.Clean(runPath)
			if runPath == ".." || strings.HasPrefix(runPath, ".."+string(filepath.Separator)) {
				return nil, nil, errors.New("网站运行目录不能超出网站根目录")
			}
		}
		relative, err := filepath.Rel(data.Root, filepath.Join(data.Root, runPath))
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return nil, nil, errors.New("网站运行目录不能超出网站根目录")
		}
		if relative == "." {
			relative = ""
		}
		data.RunPath = filepath.ToSlash(relative)
	}
	if len(input) == 0 {
		return nil, nil, errors.New("网站至少需要绑定一个域名")
	}
	previousByName := make(map[string]*model.Domain, len(previous))
	for _, domain := range previous {
		previousByName[strings.ToLower(domain.Domain)] = domain
	}
	result := make([]*model.Domain, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	certIds := make(map[int64]struct{})
	for _, item := range input {
		if item.CertId < 0 {
			return nil, nil, errors.New("证书 ID 不能小于 0")
		}
		domain, err := s.normalizeDomain(item.Domain)
		if err != nil {
			return nil, nil, err
		}
		if _, exists := seen[domain]; exists {
			return nil, nil, fmt.Errorf("网站域名重复: %s", domain)
		}
		seen[domain] = struct{}{}
		exists, err := s.repositoryDomain.Exists(domain, data.Id, tx...)
		if err != nil {
			return nil, nil, fmt.Errorf("检查网站域名失败: %w", err)
		}
		if exists {
			return nil, nil, fmt.Errorf("域名已被其他网站使用: %s", domain)
		}
		id := str.GetSnowWorkIns().GetId()
		if old := previousByName[domain]; old != nil {
			id = old.Id
		}
		result = append(result, &model.Domain{Id: id, SiteId: data.Id, CertId: item.CertId, Domain: domain})
		if item.CertId > 0 {
			certIds[item.CertId] = struct{}{}
		}
	}
	ids := make([]int64, 0, len(certIds))
	for id := range certIds {
		ids = append(ids, id)
	}
	certificates, err := s.repositoryCert.SelectByIds(ids, tx...)
	if err != nil {
		return nil, nil, fmt.Errorf("查询网站证书失败: %w", err)
	}
	certMap := make(map[int64]*model.Cert, len(certificates))
	for _, certificate := range certificates {
		certMap[certificate.Id] = certificate
	}
	if len(certMap) != len(ids) {
		return nil, nil, errors.New("网站绑定的证书不存在")
	}
	for _, domain := range result {
		if domain.CertId == 0 {
			continue
		}
		if old := previousByName[strings.ToLower(domain.Domain)]; !validateExistingCertificates && old != nil && old.CertId == domain.CertId {
			continue
		}
		if err = s.validateCertificate(certMap[domain.CertId], domain.Domain); err != nil {
			return nil, nil, fmt.Errorf("域名 %s 的证书无效: %w", domain.Domain, err)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Domain < result[j].Domain })
	return value, result, nil
}

func (s *service) validateSiteEnums(siteType, status int) error {
	if _, ok := site_type.Map()[siteType]; !ok {
		return errors.New("网站类型不合法")
	}
	if _, ok := site_status.Map()[status]; !ok {
		return errors.New("网站状态不合法")
	}
	return nil
}

// validateTypeId 校验网站分类。
func (s *service) validateTypeId(typeId int64) error {
	if typeId < 0 {
		return errors.New("网站分类不合法")
	}
	if typeId == 0 {
		return nil
	}
	data, err := s.typeService.FindById(typeId)
	if err != nil || data == nil || data.Module != type_module.Site {
		return errors.New("网站分类不合法")
	}
	return nil
}

func (s *service) normalizeDomain(value string) (string, error) {
	domain := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(value, ".")))
	if strings.HasPrefix(domain, "[") && strings.HasSuffix(domain, "]") {
		domain = strings.TrimSuffix(strings.TrimPrefix(domain, "["), "]")
	}
	if domain == "" || strings.ContainsAny(domain, "/@?#[]\r\n") {
		return "", errors.New("域名格式错误")
	}
	wildcard := strings.HasPrefix(domain, "*.")
	if wildcard {
		domain = strings.TrimPrefix(domain, "*.")
	}
	if address := net.ParseIP(domain); address != nil {
		if wildcard {
			return "", errors.New("IP 地址不支持通配符")
		}
		return address.String(), nil
	}
	if strings.Contains(domain, ":") {
		return "", errors.New("域名格式错误")
	}
	ascii, err := idna.Lookup.ToASCII(domain)
	if err != nil {
		return "", fmt.Errorf("域名格式错误: %w", err)
	}
	if len(ascii) > 253 {
		return "", errors.New("域名过长")
	}
	for _, label := range strings.Split(ascii, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", errors.New("域名标签格式错误")
		}
		for _, character := range label {
			if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '-' {
				return "", errors.New("域名包含无效字符")
			}
		}
	}
	if wildcard {
		return "*." + ascii, nil
	}
	return ascii, nil
}

func (s *service) validateCertificate(certificate *model.Cert, domain string) error {
	if certificate == nil || certificate.Id <= 0 {
		return errors.New("证书不存在")
	}
	pair, err := tls.X509KeyPair([]byte(certificate.Certificate), []byte(certificate.PrivateKey))
	if err != nil {
		return fmt.Errorf("证书和私钥不匹配: %w", err)
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}
	if time.Now().Before(leaf.NotBefore) || time.Now().After(leaf.NotAfter) {
		return errors.New("证书不在有效期内")
	}
	if strings.HasPrefix(domain, "*.") {
		for _, name := range leaf.DNSNames {
			if strings.EqualFold(name, domain) {
				return nil
			}
		}
		return errors.New("证书不覆盖通配符域名")
	}
	if net.ParseIP(domain) != nil {
		return errors.New("当前证书管理不支持为 IP 地址启用 SSL")
	}
	if err = leaf.VerifyHostname(domain); err != nil {
		return errors.New("证书不覆盖当前域名")
	}
	return nil
}
