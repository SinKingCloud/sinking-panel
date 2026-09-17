package baidu

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/dns"
	"github.com/caddyserver/caddy/v2"
	"github.com/libdns/libdns"
)

// 百度 SDK 每次请求都会修改共享 HTTP 客户端的超时配置。
var requestMu sync.Mutex

// Provider 使用独立凭据管理百度云 DNS 验证记录。
type Provider struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`

	clientOnce sync.Once
	client     *dns.Client
	clientErr  error
}

func init() {
	caddy.RegisterModule(&Provider{})
}

func (*Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.baiducloud",
		New: func() caddy.Module { return &Provider{} },
	}
}

func (p *Provider) getClient() (*dns.Client, error) {
	p.clientOnce.Do(func() {
		if p.AccessKeyID == "" || p.SecretAccessKey == "" {
			p.clientErr = fmt.Errorf("baiducloud: access_key_id and secret_access_key are required")
			return
		}
		p.client, p.clientErr = dns.NewClient(p.AccessKeyID, p.SecretAccessKey, "https://dns.baidubce.com")
		if p.clientErr != nil {
			return
		}
		// SDK 不支持请求级 context；在调用前检查取消，并限制单次请求耗时。
		p.client.Config.ConnectionTimeoutInMillis = 30 * 1000
		p.client.Config.Retry = bce.NewNoRetryPolicy()
	})
	return p.client, p.clientErr
}

// AppendRecords 添加记录，返回服务商实际使用的 TTL。
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	requestMu.Lock()
	defer requestMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	zone = strings.TrimSuffix(zone, ".")
	var created []libdns.Record
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return created, err
		}
		rr := record.RR()
		seconds := max(rr.TTL/time.Second, 300)
		if seconds > math.MaxInt32 {
			return created, fmt.Errorf("baiducloud: record TTL exceeds the supported range")
		}
		ttl := int32(seconds)
		rr.TTL = time.Duration(ttl) * time.Second
		data, err := rr.Parse()
		if err != nil {
			return created, err
		}
		err = client.CreateRecord(zone, &dns.CreateRecordRequest{
			Rr:    rr.Name,
			Type:  rr.Type,
			Value: rr.Data,
			Ttl:   &ttl,
		}, "")
		if err != nil {
			return created, fmt.Errorf("baiducloud: create record %q in %q: %w", rr.Name, zone, err)
		}
		created = append(created, data)
	}
	return created, nil
}

// DeleteRecords 仅删除名称、类型和值全部匹配的记录。
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	requestMu.Lock()
	defer requestMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}
	zone = strings.TrimSuffix(zone, ".")
	var deleted []libdns.Record
	for _, record := range records {
		rr := record.RR()
		var matches []dns.Record
		marker := ""
		for {
			if err := ctx.Err(); err != nil {
				return deleted, err
			}
			page, err := client.ListRecord(zone, &dns.ListRecordRequest{Rr: rr.Name, Marker: marker})
			if err != nil {
				return deleted, fmt.Errorf("baiducloud: list records in %q: %w", zone, err)
			}
			for _, item := range page.Records {
				if item.Rr == rr.Name && item.Type == rr.Type && item.Value == rr.Data {
					matches = append(matches, item)
				}
			}
			if !page.IsTruncated {
				break
			}
			if page.NextMarker == "" || page.NextMarker == marker {
				return deleted, fmt.Errorf("baiducloud: invalid record pagination marker")
			}
			marker = page.NextMarker
		}
		// 查完分页再删除，避免当前页变化导致后续记录被跳过。
		for _, item := range matches {
			if err := ctx.Err(); err != nil {
				return deleted, err
			}
			data, err := (libdns.RR{
				Name: item.Rr,
				Type: item.Type,
				Data: item.Value,
				TTL:  time.Duration(item.Ttl) * time.Second,
			}).Parse()
			if err != nil {
				return deleted, err
			}
			err = client.DeleteRecord(zone, item.Id, "")
			if err != nil {
				return deleted, fmt.Errorf("baiducloud: delete record %q in %q: %w", item.Id, zone, err)
			}
			deleted = append(deleted, data)
		}
	}
	return deleted, nil
}
