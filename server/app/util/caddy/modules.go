package caddy

import (
	_ "server/app/util/caddy/cache"
	_ "server/app/util/caddy/dns/baidu"
	_ "server/app/util/caddy/traffic"
	_ "server/app/util/caddy/waf"

	_ "github.com/caddy-dns/alidns"
	_ "github.com/caddy-dns/huaweicloud"
	_ "github.com/caddy-dns/tencentcloud"
	_ "github.com/caddy-dns/volcengine"
	_ "github.com/caddyserver/cache-handler"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/corazawaf/coraza-caddy/v2"
	_ "github.com/mholt/caddy-ratelimit"
)
