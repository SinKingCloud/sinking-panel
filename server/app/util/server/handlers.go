package server

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func (m *Manager) buildSiteRoute(site *Site, excludedDomains []string, scope string) (map[string]interface{}, error) {
	handlers := make([]interface{}, 0, 9+len(site.Handlers))
	if handler := m.buildRateLimitHandler(site, scope); handler != nil {
		handlers = append(handlers, handler)
	}
	if handler := m.buildTrafficLimitHandler(site); handler != nil {
		handlers = append(handlers, handler)
	}
	if handler := m.buildWAFHandler(site.WAF); handler != nil {
		handlers = append(handlers, handler)
	}
	if handler := m.buildHeaderOperations(site.Headers); len(handler) > 0 {
		handler["handler"] = "headers"
		handlers = append(handlers, handler)
	}
	if handler := m.buildCompressionHandler(site.Compression); handler != nil {
		handlers = append(handlers, handler)
	}
	for _, handler := range site.Handlers {
		handlers = append(handlers, handler)
	}
	if handler := m.buildCacheHandler(site, scope); handler != nil {
		handlers = append(handlers, handler)
	}

	routes := make([]interface{}, 0, len(site.Routes)+1)
	for index := range site.Routes {
		route, err := m.buildRoute(&site.Routes[index])
		if err != nil {
			return nil, fmt.Errorf("生成站点 %s 路由 %d 失败: %w", site.ID, index+1, err)
		}
		routes = append(routes, route)
	}
	if site.Root != "" {
		routes = append(routes, map[string]interface{}{
			"handle":   []interface{}{m.buildStaticHandler(site.Root, site.Index, site.Browse, site.TryFiles, site.Hide, site.Precompressed)},
			"terminal": true,
		})
	} else if site.Proxy != nil {
		handler, err := m.buildProxyHandler(site.Proxy)
		if err != nil {
			return nil, fmt.Errorf("生成站点 %s 反向代理失败: %w", site.ID, err)
		}
		routes = append(routes, map[string]interface{}{
			"handle":   []interface{}{handler},
			"terminal": true,
		})
	} else if len(routes) > 0 {
		routes = append(routes, map[string]interface{}{
			"handle":   []interface{}{map[string]interface{}{"handler": "static_response", "status_code": 404, "body": "Not Found"}},
			"terminal": true,
		})
	}
	if len(routes) > 0 {
		handlers = append(handlers, map[string]interface{}{"handler": "subroute", "routes": routes})
	}
	if len(handlers) == 0 {
		return nil, errors.New("站点没有可执行的处理器")
	}
	matcher := map[string]interface{}{"host": append([]string(nil), site.Domains...)}
	if len(excludedDomains) > 0 {
		matcher["not"] = []interface{}{map[string]interface{}{"host": append([]string(nil), excludedDomains...)}}
	}
	return map[string]interface{}{
		"match":    []interface{}{matcher},
		"handle":   handlers,
		"terminal": true,
	}, nil
}

func (m *Manager) buildRoute(route *Route) (map[string]interface{}, error) {
	handlers := make([]interface{}, 0, 3+len(route.Handlers))
	if route.StripPrefix != "" {
		handlers = append(handlers, map[string]interface{}{
			"handler":           "rewrite",
			"strip_path_prefix": route.StripPrefix,
		})
	}
	if route.Rewrite != "" {
		handlers = append(handlers, map[string]interface{}{"handler": "rewrite", "uri": route.Rewrite})
	}
	for _, handler := range route.Handlers {
		handlers = append(handlers, handler)
	}
	if route.Root != "" {
		handlers = append(handlers, m.buildStaticHandler(route.Root, route.Index, route.Browse, route.TryFiles, route.Hide, route.Precompressed))
	} else if route.Proxy != nil {
		handler, err := m.buildProxyHandler(route.Proxy)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	} else if route.Response != nil {
		handler := map[string]interface{}{
			"handler":     "static_response",
			"status_code": route.Response.Status,
		}
		if route.Response.Body != "" {
			handler["body"] = route.Response.Body
		}
		if len(route.Response.Headers) > 0 {
			handler["headers"] = route.Response.Headers
		}
		handlers = append(handlers, handler)
	}
	if len(handlers) == 0 {
		return nil, errors.New("路由没有可执行的处理器")
	}
	result := map[string]interface{}{
		"handle":   handlers,
		"terminal": true,
	}
	matcher := map[string]interface{}{}
	if len(route.Paths) > 0 {
		matcher["path"] = append([]string(nil), route.Paths...)
	}
	if len(route.Methods) > 0 {
		matcher["method"] = append([]string(nil), route.Methods...)
	}
	if len(matcher) > 0 {
		result["match"] = []interface{}{matcher}
	}
	return result, nil
}

func (m *Manager) buildRateLimitHandler(site *Site, scope string) map[string]interface{} {
	limit := site.RateLimit
	if !limit.Enabled {
		return nil
	}
	zone := map[string]interface{}{
		"key":        limit.Key,
		"window":     limit.Window.String(),
		"max_events": limit.MaxEvents,
	}
	matcher := map[string]interface{}{}
	if len(limit.Paths) > 0 {
		matcher["path"] = append([]string(nil), limit.Paths...)
	}
	if len(limit.Methods) > 0 {
		matcher["method"] = append([]string(nil), limit.Methods...)
	}
	if len(matcher) > 0 {
		zone["match"] = []interface{}{matcher}
	}
	if limit.IPv4Prefix > 0 {
		zone["ipv4_prefix"] = limit.IPv4Prefix
	}
	if limit.IPv6Prefix > 0 {
		zone["ipv6_prefix"] = limit.IPv6Prefix
	}
	policy, _ := json.Marshal(limit)
	fingerprint := sha256.Sum256(policy)
	return map[string]interface{}{
		"handler":         "rate_limit",
		"disable_metrics": true,
		"rate_limits": map[string]interface{}{
			fmt.Sprintf("sinking_cloud_site_%s_%s_%x", site.ID, scope, fingerprint[:8]): zone,
		},
	}
}

func (m *Manager) buildTrafficLimitHandler(site *Site) map[string]interface{} {
	limit := site.TrafficLimit
	if !limit.Enabled {
		return nil
	}
	handler := map[string]interface{}{
		"handler": "traffic_limit",
		"scope":   site.ID,
	}
	if limit.MaxConnections > 0 {
		handler["max_connections"] = limit.MaxConnections
	}
	if limit.MaxConnectionsPerIP > 0 {
		handler["max_connections_per_ip"] = limit.MaxConnectionsPerIP
	}
	if limit.RatePerRequest > 0 {
		handler["rate_per_request"] = limit.RatePerRequest
	}
	return handler
}

func (m *Manager) buildWAFHandler(options WAFOptions) map[string]interface{} {
	if !options.Enabled {
		return nil
	}
	directives := make([]string, 0, 12+len(options.Directives)+len(options.Rules))
	if options.OWASP {
		directives = append(directives,
			"Include @coraza.conf-recommended",
			"Include @crs-setup.conf.example",
			"Include @owasp_crs/*.conf",
		)
	}
	for _, directive := range options.Directives {
		if directive != "" {
			directives = append(directives, directive)
		}
	}
	for _, rule := range options.Rules {
		if rule.Enabled {
			directives = append(directives, rule.Directive)
		}
	}
	if options.RequestBodyLimit > 0 {
		directives = append(directives, fmt.Sprintf("SecRequestBodyLimit %d", options.RequestBodyLimit))
		if options.Mode == WAFModeBlock {
			directives = append(directives, "SecRequestBodyLimitAction Reject")
		} else {
			directives = append(directives, "SecRequestBodyLimitAction ProcessPartial")
		}
	}
	if options.AuditLog {
		directives = append(directives,
			"SecAuditEngine RelevantOnly",
			`SecAuditLogRelevantStatus "^(?:4|5)"`,
			"SecAuditLogParts ABCHIJKZ",
			"SecAuditLogType Serial",
			"SecAuditLogFormat JSON",
			"SecAuditLog "+strconv.Quote(m.wafLogPath),
		)
	} else {
		directives = append(directives, "SecAuditEngine Off")
	}
	if options.Mode == WAFModeBlock {
		directives = append(directives, "SecRuleEngine On")
	} else {
		directives = append(directives, "SecRuleEngine DetectionOnly")
	}
	return map[string]interface{}{
		"handler":        "waf",
		"directives":     strings.Join(directives, "\n"),
		"load_owasp_crs": options.OWASP,
	}
}

func (m *Manager) buildHeaderOperations(options HeaderOptions) map[string]interface{} {
	result := map[string]interface{}{}
	request := map[string]interface{}{}
	if len(options.RequestAdd) > 0 {
		request["add"] = options.RequestAdd
	}
	if len(options.RequestSet) > 0 {
		request["set"] = options.RequestSet
	}
	if len(options.RequestDelete) > 0 {
		request["delete"] = append([]string(nil), options.RequestDelete...)
	}
	if len(request) > 0 {
		result["request"] = request
	}
	response := map[string]interface{}{}
	if len(options.ResponseAdd) > 0 {
		response["add"] = options.ResponseAdd
	}
	if len(options.ResponseSet) > 0 {
		response["set"] = options.ResponseSet
	}
	if len(options.ResponseDelete) > 0 {
		response["delete"] = append([]string(nil), options.ResponseDelete...)
	}
	if len(response) > 0 {
		response["deferred"] = true
		result["response"] = response
	}
	return result
}

func (m *Manager) buildCompressionHandler(options CompressionOptions) map[string]interface{} {
	if !options.Enabled {
		return nil
	}
	encodings := make(map[string]interface{}, len(options.Algorithms))
	for _, algorithm := range options.Algorithms {
		encodings[algorithm] = map[string]interface{}{}
	}
	handler := map[string]interface{}{
		"handler":   "encode",
		"encodings": encodings,
		"prefer":    append([]string(nil), options.Algorithms...),
	}
	if options.MinLength > 0 {
		handler["minimum_length"] = options.MinLength
	}
	return handler
}

func (m *Manager) buildCacheHandler(site *Site, scope string) map[string]interface{} {
	options := site.Cache
	if !options.Enabled {
		return nil
	}
	cacheDirectory := filepath.Join(m.cachePath, site.ID, scope)
	configuration := map[string]interface{}{
		"allowed_http_verbs":       append([]string(nil), options.Methods...),
		"ttl":                      options.TTL.String(),
		"storers":                  []string{"simplefs"},
		"default_cache_control":    options.DefaultCacheControl,
		"cache_name":               options.Name,
		"max_cacheable_body_bytes": options.MaxSize,
		"simplefs": map[string]interface{}{
			"found": true,
			"path":  cacheDirectory,
			"configuration": map[string]interface{}{
				"directory_size": options.MaxSize,
				"allowed_root":   m.cachePath,
			},
		},
	}
	if options.Stale > 0 {
		configuration["stale"] = options.Stale.String()
	}
	if len(options.KeyHeaders) > 0 {
		configuration["headers"] = append([]string(nil), options.KeyHeaders...)
	}
	return map[string]interface{}{
		"handler": "cache",
		"configuration": map[string]interface{}{
			"DefaultCache": configuration,
		},
	}
}

func (m *Manager) buildStaticHandler(root string, index []string, browse bool, tryFiles, hide, precompressed []string) map[string]interface{} {
	fileServer := map[string]interface{}{
		"handler":     "file_server",
		"root":        root,
		"index_names": append([]string(nil), index...),
	}
	if browse {
		fileServer["browse"] = map[string]interface{}{}
	}
	if len(hide) > 0 {
		fileServer["hide"] = append([]string(nil), hide...)
	}
	if len(precompressed) > 0 {
		encodings := make(map[string]interface{}, len(precompressed))
		for _, algorithm := range precompressed {
			encodings[algorithm] = map[string]interface{}{}
		}
		fileServer["precompressed"] = encodings
		fileServer["precompressed_order"] = append([]string(nil), precompressed...)
	}
	if len(tryFiles) == 0 {
		return fileServer
	}
	return map[string]interface{}{
		"handler": "subroute",
		"routes": []interface{}{
			map[string]interface{}{
				"match": []interface{}{map[string]interface{}{
					"file": map[string]interface{}{
						"root":      root,
						"try_files": append([]string(nil), tryFiles...),
					},
				}},
				"handle": []interface{}{map[string]interface{}{
					"handler": "rewrite",
					"uri":     "{http.matchers.file.relative}",
				}},
			},
			map[string]interface{}{
				"handle":   []interface{}{fileServer},
				"terminal": true,
			},
		},
	}
}

func (m *Manager) buildProxyHandler(options *ProxyOptions) (map[string]interface{}, error) {
	if options == nil || len(options.Upstreams) == 0 {
		return nil, errors.New("反向代理没有上游")
	}
	upstreams := make([]interface{}, 0, len(options.Upstreams))
	for _, upstream := range options.Upstreams {
		item := map[string]interface{}{"dial": upstream.Dial}
		if upstream.MaxRequests > 0 {
			item["max_requests"] = upstream.MaxRequests
		}
		upstreams = append(upstreams, item)
	}
	handler := map[string]interface{}{
		"handler":   "reverse_proxy",
		"upstreams": upstreams,
		"load_balancing": map[string]interface{}{
			"selection_policy": map[string]interface{}{"policy": options.Policy},
		},
	}
	loadBalancing := handler["load_balancing"].(map[string]interface{})
	if options.Retries > 0 {
		loadBalancing["retries"] = options.Retries
	}
	if options.TryDuration > 0 {
		loadBalancing["try_duration"] = options.TryDuration.String()
	}
	if options.TryInterval > 0 {
		loadBalancing["try_interval"] = options.TryInterval.String()
	}
	if options.FlushInterval != 0 {
		handler["flush_interval"] = options.FlushInterval.String()
	}
	if options.StreamTimeout > 0 {
		handler["stream_timeout"] = options.StreamTimeout.String()
	}
	if options.StreamCloseDelay > 0 {
		handler["stream_close_delay"] = options.StreamCloseDelay.String()
	}
	if headers := m.buildHeaderOperations(options.Headers); len(headers) > 0 {
		handler["headers"] = headers
	}
	if options.HealthURI != "" {
		active := map[string]interface{}{"uri": options.HealthURI}
		if options.HealthInterval > 0 {
			active["interval"] = options.HealthInterval.String()
		}
		if options.HealthTimeout > 0 {
			active["timeout"] = options.HealthTimeout.String()
		}
		if options.HealthStatus > 0 {
			active["expect_status"] = options.HealthStatus
		}
		handler["health_checks"] = map[string]interface{}{"active": active}
	}
	if options.Transport == ProxyTransportFastCGI {
		if options.Root == "" {
			return nil, errors.New("FastCGI 缺少站点根目录")
		}
		paths := make([]string, 0, len(options.SplitPath)*2)
		for _, extension := range options.SplitPath {
			paths = append(paths, "*"+extension, "*"+extension+"/*")
		}
		if len(paths) == 0 {
			return nil, errors.New("FastCGI 缺少脚本扩展名")
		}
		handler["transport"] = m.buildFastCGITransport(options)
		directoryIndex := "{http.request.uri.path}/" + options.Index
		return map[string]interface{}{
			"handler": "subroute",
			"routes": []interface{}{
				map[string]interface{}{
					"match": []interface{}{map[string]interface{}{
						"file": map[string]interface{}{
							"root":      options.Root,
							"try_files": []string{directoryIndex},
						},
						"not": []interface{}{map[string]interface{}{"path": []string{"*/"}}},
					}},
					"handle": []interface{}{map[string]interface{}{
						"handler":     "static_response",
						"status_code": 308,
						"headers": map[string][]string{
							"Location": {"{http.request.orig_uri.path}/{http.request.orig_uri.prefixed_query}"},
						},
					}},
					"terminal": true,
				},
				map[string]interface{}{
					"match": []interface{}{map[string]interface{}{
						"file": map[string]interface{}{
							"root":       options.Root,
							"try_files":  append([]string(nil), options.TryFiles...),
							"try_policy": options.TryPolicy,
							"split_path": append([]string(nil), options.SplitPath...),
						},
					}},
					"handle": []interface{}{map[string]interface{}{"handler": "rewrite", "uri": "{http.matchers.file.relative}"}},
				},
				map[string]interface{}{
					"match": []interface{}{map[string]interface{}{
						"path": paths,
					}},
					"handle":   []interface{}{handler},
					"terminal": true,
				},
				map[string]interface{}{
					"handle": []interface{}{m.buildStaticHandler(
						options.Root,
						[]string{"index.html", "index.htm"},
						false,
						nil,
						options.Hide,
						nil,
					)},
					"terminal": true,
				},
			},
		}, nil
	} else {
		handler["transport"] = m.buildHTTPTransport(options)
	}
	return handler, nil
}

func (m *Manager) buildHTTPTransport(options *ProxyOptions) map[string]interface{} {
	transport := map[string]interface{}{"protocol": "http"}
	if options.DialTimeout > 0 {
		transport["dial_timeout"] = options.DialTimeout.String()
	}
	if options.ReadTimeout > 0 {
		transport["read_timeout"] = options.ReadTimeout.String()
	}
	if options.WriteTimeout > 0 {
		transport["write_timeout"] = options.WriteTimeout.String()
	}
	if options.ResponseHeaderTimeout > 0 {
		transport["response_header_timeout"] = options.ResponseHeaderTimeout.String()
	}
	if len(options.Versions) > 0 {
		versions := make([]string, 0, len(options.Versions))
		for _, version := range options.Versions {
			version = strings.ToLower(strings.TrimSpace(version))
			switch version {
			case "h1", "http/1.1":
				versions = append(versions, "1.1")
			case "h2", "http/2":
				versions = append(versions, "2")
			case "h3", "http/3":
				versions = append(versions, "3")
			default:
				versions = append(versions, version)
			}
		}
		transport["versions"] = versions
	}
	if options.Scheme == ProxySchemeHTTPS {
		tls := map[string]interface{}{}
		if options.TLSServerName != "" {
			tls["server_name"] = options.TLSServerName
		}
		if options.TLSInsecureSkipVerify == nil || *options.TLSInsecureSkipVerify {
			tls["insecure_skip_verify"] = true
		}
		transport["tls"] = tls
	}
	return transport
}

func (m *Manager) buildFastCGITransport(options *ProxyOptions) map[string]interface{} {
	transport := map[string]interface{}{
		"protocol":   "fastcgi",
		"root":       options.Root,
		"split_path": append([]string(nil), options.SplitPath...),
	}
	if options.ResolveRootSymlink {
		transport["resolve_root_symlink"] = true
	}
	if options.DialTimeout > 0 {
		transport["dial_timeout"] = options.DialTimeout.String()
	}
	if options.ReadTimeout > 0 {
		transport["read_timeout"] = options.ReadTimeout.String()
	}
	if options.WriteTimeout > 0 {
		transport["write_timeout"] = options.WriteTimeout.String()
	}
	if len(options.Env) > 0 {
		transport["env"] = options.Env
	}
	if options.CaptureStderr {
		transport["capture_stderr"] = true
	}
	return transport
}
