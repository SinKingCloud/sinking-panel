# waf：Coraza 审计日志轮转适配

导入路径：`server/app/util/caddy/waf`。此包向 Coraza 注册名为 `rotating` 的审计日志写入器，负责每个站点 WAF 审计日志的格式化、并发写入、权限和轮转。它不定义业务 WAF 规则，也不直接执行请求拦截。

使用 [caddy Manager](../README.md) 时，根包 [modules.go](../modules.go) 会自动导入本包。写入器类型未导出，没有供业务直接调用的 NewWriter；通过 `Site.WAF` 配置接入即可。

## 示例：开启检测模式和审计日志

`manager` 为应用持有的 Manager，`siteID` 是已有站点。此完整函数启用内嵌 OWASP CRS，并返回站点审计日志路径。

```go
package example

import panelcaddy "server/app/util/caddy"

func EnableWAFAudit(manager *panelcaddy.Manager, siteID string) (string, error) {
    site, err := manager.Site(siteID)
    if err != nil {
        return "", err
    }
    site.WAF = panelcaddy.WAFOptions{
        Enabled: true,
        Mode: panelcaddy.WAFModeDetection,
        AuditLog: true,
        OWASP: panelcaddy.OWASPOptions{Enabled: true},
    }
    if err = manager.UpdateSite(*site); err != nil {
        return "", err
    }
    return manager.LogPath(siteID, panelcaddy.LogWAF)
}
```

检测模式记录规则命中但不以拦截模式阻断请求。需要阻断时，由业务明确改为 `WAFModeBlock`，并按网站情况配置 OWASP 参数和自定义规则。上面的赋值会替换已有 WAF 配置；如需保留规则，应只修改原 site.WAF 的相关字段。

## 注册链和实际配置

注册链为：根包导入 waf → init 注册 Coraza AuditLogWriter 工厂 → Manager 生成 Coraza 指令 → Coraza 初始化并调用写入器。

当 `Site.WAF.Enabled` 与 `AuditLog` 均开启时，Manager 生成以下指令。路径只是示意，真实路径通过 `LogPath(siteID, LogWAF)` 获取：

```text
SecAuditEngine RelevantOnly
SecAuditLogParts ABCHIJKZ
SecAuditLogType rotating
SecAuditLogFormat JSON
SecAuditLogFileMode 0600
SecAuditLog "/path/to/root/site-id/waf.log"
```

审计采用 RelevantOnly，并非每个普通请求都记录；一般访问记录查看站点 access 日志。关闭 AuditLog 时生成 SecAuditEngine Off。全局 Options.WAFLogPath 是 WAF 模块日志路径，不等同于这里的单站点审计日志。

## 写入器生命周期

| 内部方法 | 行为 |
| --- | --- |
| `Init(AuditLogConfig) error` | 要求非空 Target 和 Formatter；创建日志父目录并检查文件可写，设置 `0600` 权限。 |
| `Write(AuditLog) error` | 在互斥锁中格式化并追加一条记录及换行；格式化或文件写入错误向上传递，短写返回 io.ErrShortWrite。 |
| `Close() error` | 关闭日志资源并清空状态，可重复调用；后续 Write 会返回未初始化错误。 |

上述方法由 Coraza 生命周期调用，不应绕过 Caddy 热加载与关闭流程手动管理。应用退出仍由 Manager.Stop 完成整体资源释放。

## 轮转、权限与排查

轮转参数固定为单文件 50 MB、最多 10 份备份、保留 30 天、gzip 压缩，文件权限 `0600`，新建目录权限 `0700`。这些参数目前没有独立公开配置字段；不要在 README 中的 Site.WAF 示例添加不存在的日志大小参数。

日志可能包含请求头、参数和请求体等审计信息，业务下载日志接口应沿用系统权限校验。目录不可写、磁盘不足或格式化失败会返回具体错误；审计文件不存在时先确认站点已启用 WAF/AuditLog、配置已成功应用，再检查目录权限和模块日志。

相关工具：[log 通用业务日志](../../log/README.md)、[file 文件操作](../../file/README.md)。
