package site

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"server/app/constant"
	"server/app/enum/cert_auto_renew"
	"server/app/enum/cert_type"
	"server/app/enum/task_status"
	"server/app/model"
	repositoryTask "server/app/repository/task"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// initCertTask 在业务启动时按保存的设置注册内部续签任务。
func (s *service) initCertTask() error {
	if s.taskService == nil {
		return errors.New("任务服务未初始化")
	}
	settings, err := s.GetCertSettings()
	if err != nil {
		return err
	}
	status := task_status.Stop
	if settings[constant.CertAutoRenewEnabled] == "true" {
		status = task_status.Running
	}
	content, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return s.taskService.Add(&model.Task{
		Id:     constant.TaskCertAutoRenew,
		Name:   "证书自动续签",
		Spec:   settings[constant.CertAutoRenewSpec],
		Script: string(content),
		Status: status,
	}, s.autoRenewCerts)
}

// GetCertSettings 读取续签设置，未配置时每天凌晨三点执行且不重试。
func (s *service) GetCertSettings() (map[string]string, error) {
	if s.configService == nil {
		return nil, errors.New("配置服务未初始化")
	}
	settings := map[string]string{
		constant.CertAutoRenewEnabled: "true",
		constant.CertAutoRenewSpec:    "0 0 3 * * *",
		constant.CertAutoRenewRetries: "0",
	}
	configs := s.configService.Group(constant.CertGroup)
	for key := range settings {
		if value, exists := configs[key]; exists {
			settings[key] = value
		}
	}
	if err := s.validateCertSettings(settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// validateCertSettings 校验并规范化独立配置项，供保存和任务快照复用。
func (s *service) validateCertSettings(settings map[string]string) error {
	enabled, err := strconv.ParseBool(settings[constant.CertAutoRenewEnabled])
	if err != nil {
		return errors.New("自动续签开关配置不合法")
	}
	retries, err := strconv.Atoi(settings[constant.CertAutoRenewRetries])
	if err != nil || retries < 0 || retries > 10 {
		return errors.New("自动续签重试次数必须为 0–10")
	}
	spec := strings.TrimSpace(settings[constant.CertAutoRenewSpec])
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(spec); err != nil {
		return errors.New("自动续签执行周期不合法")
	}
	settings[constant.CertAutoRenewEnabled] = strconv.FormatBool(enabled)
	settings[constant.CertAutoRenewRetries] = strconv.Itoa(retries)
	settings[constant.CertAutoRenewSpec] = spec
	return nil
}

// SetCertSettings 更新内部任务及配置，配置写入失败时恢复原调度。
func (s *service) SetCertSettings(configs map[string]string) error {
	if s.configService == nil || s.taskService == nil {
		return errors.New("配置或任务服务未初始化")
	}
	settings := map[string]string{
		constant.CertAutoRenewEnabled: configs[constant.CertAutoRenewEnabled],
		constant.CertAutoRenewSpec:    configs[constant.CertAutoRenewSpec],
		constant.CertAutoRenewRetries: configs[constant.CertAutoRenewRetries],
	}
	if err := s.validateCertSettings(settings); err != nil {
		return err
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.closed {
		return errors.New("网站服务已关闭")
	}
	previous, err := s.GetCertSettings()
	if err != nil {
		return err
	}
	status := task_status.Stop
	if settings[constant.CertAutoRenewEnabled] == "true" {
		status = task_status.Running
	}
	content, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	script := string(content)
	spec := settings[constant.CertAutoRenewSpec]
	// 周期和执行参数一起更新，已经开始的任务继续使用本次配置快照。
	if s.certTask {
		if err = s.taskService.UpdateByIds([]int64{constant.TaskCertAutoRenew}, &repositoryTask.UpdateTask{
			Spec:   &spec,
			Script: &script,
			Status: &status,
		}); err != nil {
			return err
		}
	}
	if err = s.configService.Sets(settings); err != nil {
		if s.certTask {
			previousContent, _ := json.Marshal(previous)
			previousScript := string(previousContent)
			previousSpec := previous[constant.CertAutoRenewSpec]
			status = task_status.Stop
			if previous[constant.CertAutoRenewEnabled] == "true" {
				status = task_status.Running
			}
			err = errors.Join(err, s.taskService.UpdateByIds([]int64{constant.TaskCertAutoRenew}, &repositoryTask.UpdateTask{
				Spec:   &previousSpec,
				Script: &previousScript,
				Status: &status,
			}))
		}
		return err
	}
	return nil
}

// autoRenewCerts 使用本次任务配置续签证书，执行途中仍检查开关与证书状态。
func (s *service) autoRenewCerts(ctx context.Context, _ int64, content string, writeLog func(string)) error {
	if ctx == nil {
		return nil
	}
	ctx, release, err := s.operationContext(ctx)
	if err != nil {
		return err
	}
	defer release()
	if writeLog == nil {
		writeLog = func(string) {}
	}
	log := func(format string, args ...interface{}) { writeLog(fmt.Sprintf(format, args...)) }
	var settings map[string]string
	if err = json.Unmarshal([]byte(content), &settings); err != nil || settings == nil {
		return errors.New("自动续签任务配置不合法")
	}
	if err := s.validateCertSettings(settings); err != nil {
		return err
	}
	if settings[constant.CertAutoRenewEnabled] != "true" {
		return nil
	}
	retries, _ := strconv.Atoi(settings[constant.CertAutoRenewRetries])
	before := time.Now().Add(72 * time.Hour)
	log("开始检查自动续签证书：剩余有效期不超过 3 天，失败最多重试 %d 次，单次申请最多 30 秒", retries)
	certificates, err := s.repositoryCert.SelectAutoRenewTask(before)
	if err != nil {
		return fmt.Errorf("查询待续签证书失败：%w", err)
	}
	success, failed := 0, 0
	defer func() {
		log("本次检查结束：待续签 %d 张，成功 %d 张，失败 %d 张，未执行 %d 张", len(certificates), success, failed, len(certificates)-success-failed)
	}()
certificatesLoop:
	for _, item := range certificates {
		previousCertificate := ""
		for attempt := 0; attempt <= retries; attempt++ {
			if ctx.Err() != nil {
				log("任务已停止，取消剩余续签")
				return nil
			}
			current, err := s.GetCertSettings()
			if err != nil {
				return fmt.Errorf("读取自动续签设置失败：%w", err)
			}
			if current[constant.CertAutoRenewEnabled] != "true" {
				log("自动续签已关闭，取消剩余续签")
				return nil
			}
			certificate, err := s.repositoryCert.FindById(item.Id)
			if err != nil {
				failed++
				log("读取证书 #%d 失败：%v", item.Id, err)
				continue certificatesLoop
			}
			if certificate == nil {
				failed++
				log("读取证书 #%d 失败：证书不存在", item.Id)
				continue certificatesLoop
			}
			if attempt > 0 && certificate.Certificate != previousCertificate {
				log("证书 #%d 已更新，取消剩余重试", certificate.Id)
				continue certificatesLoop
			}
			previousCertificate = certificate.Certificate
			// 自动任务独立检查条件，不限制用户主动续签未到期的证书。
			if certificate.Type != cert_type.Auto || certificate.AutoRenew != cert_auto_renew.Enabled || time.Time(certificate.ExpireTime).After(before) {
				log("证书 #%d 已不满足自动续签条件，本次跳过", certificate.Id)
				continue certificatesLoop
			}
			var domains []string
			if err = json.Unmarshal([]byte(certificate.Domains), &domains); err != nil || len(domains) == 0 {
				failed++
				log("证书 #%d 续签失败：证书域名数据不合法", certificate.Id)
				continue certificatesLoop
			}
			if len(domains) != 1 {
				// 一次仅申请一个域名，不能在自动续签时丢失原有的其他 SAN。
				log("证书 #%d 为多域名证书，需要手动更新，本次跳过", certificate.Id)
				continue certificatesLoop
			}
			log("开始续签证书 #%d [%s]，第 %d/%d 次，原到期时间：%s", certificate.Id, strings.Join(domains, ", "), attempt+1, retries+1, time.Time(certificate.ExpireTime).Format("2006-01-02 15:04:05"))
			// 复用续签业务，继承验证方式、关联密钥和原 CA，保存后更新已部署的网站。
			renewContext, cancel := context.WithTimeout(ctx, CertificateOperationTimeout)
			result, renewErr := s.RenewCert(renewContext, certificate.Id, CertificateRequest{Domain: domains[0]})
			cancel()
			if renewErr != nil {
				log("证书 #%d [%s] 续签失败：%v", certificate.Id, strings.Join(domains, ", "), renewErr)
				if attempt == retries {
					failed++
				}
				continue
			}
			success++
			log("证书 #%d [%s] 续签成功，到期时间：%s", certificate.Id, strings.Join(domains, ", "), result.ExpireTime.Format("2006-01-02 15:04:05"))
			break
		}
	}
	return nil
}

// ReadCertLog 读取内部证书续签任务日志。
func (s *service) ReadCertLog(after int64, before int64, pageSize int) (map[string]interface{}, error) {
	return s.taskService.ReadLog(constant.TaskCertAutoRenew, after, before, pageSize)
}

// ClearCertLog 清空内部证书续签任务日志。
func (s *service) ClearCertLog() error {
	return s.taskService.ClearLog(constant.TaskCertAutoRenew)
}
