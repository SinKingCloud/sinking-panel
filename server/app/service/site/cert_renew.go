package site

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server/app/constant"
	"server/app/enum/cert_auto_renew"
	"server/app/enum/cert_type"
	webServer "server/app/util/server"
	"strings"
	"time"
)

// autoRenewCerts 每日检查临近到期的证书，每张仅续签一次，失败留待次日检查。
func (s *service) autoRenewCerts(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	writeLog := func(format string, args ...interface{}) {
		if err := s.taskService.WriteLog(constant.TaskCertAutoRenew, fmt.Sprintf(format, args...)); err != nil {
			log.Printf("写入证书续签日志失败: %v", err)
		}
	}

	before := time.Now().Add(72 * time.Hour)
	writeLog("开始检查自动续签证书：剩余有效期不超过 3 天，每张仅尝试一次，单次申请最多 30 秒")
	certificates, err := s.repositoryCert.SelectAutoRenewTask(before)
	if err != nil {
		writeLog("查询待续签证书失败：%v", err)
		return
	}
	success, failed := 0, 0
	defer func() {
		writeLog("本次检查结束：待续签 %d 张，成功 %d 张，失败 %d 张，未执行 %d 张", len(certificates), success, failed, len(certificates)-success-failed)
	}()
	for _, item := range certificates {
		if ctx.Err() != nil {
			writeLog("任务已停止，取消剩余续签")
			return
		}
		certificate, err := s.repositoryCert.FindById(item.Id)
		if err != nil {
			failed++
			writeLog("读取证书 #%d 失败：%v", item.Id, err)
			continue
		}
		// 自动任务执行前重新检查，避免覆盖期间已续签或已关闭自动续签的证书。
		if certificate.Type != cert_type.Auto || certificate.AutoRenew != cert_auto_renew.Enabled || time.Time(certificate.ExpireTime).After(before) {
			writeLog("证书 #%d 已不满足自动续签条件，本次跳过", certificate.Id)
			continue
		}
		var domains []string
		if err = json.Unmarshal([]byte(certificate.Domains), &domains); err != nil || len(domains) == 0 {
			failed++
			writeLog("证书 #%d 续签失败：证书域名数据不合法", certificate.Id)
			continue
		}
		if len(domains) != 1 {
			// 当前签发接口一次只申请一个域名，自动任务不能丢失原有的其他 SAN。
			writeLog("证书 #%d 为多域名证书，需要手动更新，本次跳过", certificate.Id)
			continue
		}
		writeLog("开始续签证书 #%d [%s]，原到期时间：%s", certificate.Id, strings.Join(domains, ", "), time.Time(certificate.ExpireTime).Format("2006-01-02 15:04:05"))
		// 复用续签业务，继承验证方式、关联密钥和原 CA，保存后更新已部署的网站。
		renewContext, cancel := context.WithTimeout(ctx, webServer.CertificateRequestTimeout)
		result, renewErr := s.RenewCert(renewContext, certificate.Id, CertificateRequest{Domain: domains[0]})
		cancel()
		if renewErr != nil {
			failed++
			writeLog("证书 #%d [%s] 续签失败：%v", certificate.Id, strings.Join(domains, ", "), renewErr)
			continue
		}
		success++
		writeLog("证书 #%d [%s] 续签成功，到期时间：%s", result.Id, strings.Join(domains, ", "), time.Time(result.ExpireTime).Format("2006-01-02 15:04:05"))
	}
}

// ReadCertLog 读取内部证书续签任务日志。
func (s *service) ReadCertLog(after int64, before int64, pageSize int) (map[string]interface{}, error) {
	return s.taskService.ReadLog(constant.TaskCertAutoRenew, after, before, pageSize)
}

// ClearCertLog 清空内部证书续签任务日志。
func (s *service) ClearCertLog() error {
	return s.taskService.ClearLog(constant.TaskCertAutoRenew)
}
