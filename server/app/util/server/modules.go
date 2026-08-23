package server

import (
	_ "server/app/util/server/cache"
	_ "server/app/util/server/traffic"
	_ "server/app/util/server/waf"

	_ "github.com/caddyserver/cache-handler"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/corazawaf/coraza-caddy/v2"
	_ "github.com/mholt/caddy-ratelimit"
)
