module server

go 1.27.0

require (
	github.com/DeRuina/timberjack v1.4.8
	github.com/SinKingCloud/sinking-go/sinking-web v0.0.0-20260927045123-4f43fcb3834b
	github.com/afocus/captcha v0.0.0-20191010092841-4bd1f21c8868
	github.com/axgle/mahonia v0.0.0-20180208002826-3358181d7394
	github.com/baidubce/bce-sdk-go v0.9.276
	github.com/caddy-dns/alidns v1.0.29
	github.com/caddy-dns/huaweicloud v1.0.1
	github.com/caddy-dns/tencentcloud v0.4.3
	github.com/caddy-dns/volcengine v0.0.1
	github.com/caddyserver/cache-handler v0.17.0
	github.com/caddyserver/caddy/v2 v2.11.4
	github.com/caddyserver/certmagic v0.25.4
	github.com/containerd/console v1.0.5
	github.com/corazawaf/coraza-caddy/v2 v2.6.1
	github.com/corazawaf/coraza/v3 v3.7.0
	github.com/cyphar/filepath-securejoin v0.7.0
	github.com/darkweak/storages/core v0.0.20
	github.com/dustin/go-humanize v1.1.0
	github.com/glebarez/sqlite v1.11.0
	github.com/go-playground/locales v0.14.2
	github.com/go-playground/universal-translator v0.18.2
	github.com/go-playground/validator/v10 v10.30.5
	github.com/godbus/dbus/v5 v5.2.2
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/gorilla/websocket v1.5.3
	github.com/libdns/alidns v1.0.7
	github.com/libdns/huaweicloud v1.0.1
	github.com/libdns/libdns v1.1.1
	github.com/libdns/tencentcloud v1.4.3
	github.com/libdns/volcengine v0.0.1
	github.com/mholt/acmez/v3 v3.1.7
	github.com/mholt/caddy-ratelimit v0.1.1-0.20260612195517-5625512f24f6
	github.com/moby/profiles/seccomp v0.2.3
	github.com/moby/sys/mountinfo v0.7.2
	github.com/moby/sys/user v0.4.1
	github.com/opencontainers/cgroups v0.1.0
	github.com/opencontainers/runc v1.5.2
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pierrec/lz4/v4 v4.1.31
	github.com/robfig/cron/v3 v3.0.1
	github.com/sevlyar/go-daemon v0.1.7
	github.com/shirou/gopsutil/v4 v4.26.8
	github.com/spf13/viper v1.21.0
	github.com/vmihailenco/msgpack/v5 v5.4.1
	github.com/wenlng/go-captcha-assets v1.0.7
	github.com/wenlng/go-captcha/v2 v2.0.5
	go.etcd.io/bbolt v1.5.0
	go.uber.org/zap v1.28.0
	golang.org/x/crypto v0.57.0
	golang.org/x/net v0.59.0
	golang.org/x/sys v0.48.0
	golang.org/x/term v0.46.0
	golang.org/x/text v0.42.0
	golang.org/x/time v0.16.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/gorm v1.31.2
)

require (
	cel.dev/expr v0.25.3 // indirect
	cloud.google.com/go/auth v0.24.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.3.0 // indirect
	cloud.google.com/go/compute/metadata v0.10.0 // indirect
	cyphar.com/go-pathrs v0.2.6 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/bigmod v0.1.0 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/AndreasBriese/bbloom v0.0.0-20190825152654-46b345b51c96 // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	// Caddy v2.11.4 仍使用 v1.0.0 已移除的内存限制 API。
	github.com/KimMachineGun/automemlimit v0.7.5 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/alecthomas/chroma/v2 v2.27.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/aperturerobotics/protobuf-go-lite v0.19.0 // indirect
	github.com/aryann/difflib v0.0.0-20210328193216-ff5ff6dc229b // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/caddyserver/zerossl v0.1.5 // indirect
	github.com/ccoveille/go-safecast/v2 v2.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/checkpoint-restore/go-criu/v8 v8.4.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cilium/ebpf v0.22.0 // indirect
	github.com/cloudflare/circl v1.6.5 // indirect
	github.com/corazawaf/coraza-coreruleset/v4 v4.25.0 // indirect
	github.com/corazawaf/libinjection-go v0.3.3 // indirect
	github.com/coreos/go-oidc/v3 v3.21.0 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/darkweak/souin v1.7.9 // indirect
	github.com/dgraph-io/badger v1.6.2 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.2.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da // indirect
	github.com/dlclark/regexp2/v2 v2.8.0 // indirect
	github.com/ebitengine/purego v0.11.1 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/glebarez/go-sqlite v1.23.0 // indirect
	github.com/go-chi/chi/v5 v5.3.2 // indirect
	github.com/go-jose/go-jose/v3 v3.0.5 // indirect
	github.com/go-jose/go-jose/v4 v4.1.5 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-sql-driver/mysql v1.10.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	// Caddy v2.11.4 尚未兼容 v0.29.0 起的 NewCall 参数变更。
	github.com/google/cel-go v0.28.1 // indirect
	github.com/google/certificate-transparency-go v1.3.3 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/go-tspi v0.3.0 // indirect
	github.com/google/s2a-go v0.1.10 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.22 // indirect
	github.com/googleapis/gax-go/v2 v2.26.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.31.0 // indirect
	github.com/huandu/xstrings v1.6.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jcchavezs/mergefs v0.1.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kaptinlin/jsonpointer v0.4.28 // indirect
	github.com/kaptinlin/jsonschema v0.9.10 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20260802145828-341c2f0c90b5 // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/miekg/dns v1.1.73 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/sys/capability v0.4.0 // indirect
	github.com/moby/sys/userns v0.2.1 // indirect
	github.com/mrunalp/fileutils v0.5.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/opencontainers/selinux v1.15.1 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/petar-dambovaliev/aho-corasick v0.0.0-20250424160509-463d218d4745 // indirect
	github.com/pires/go-proxyproto v0.15.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20260916203055-22a1a467d9f0 // indirect
	github.com/pquerna/cachecontrol v0.2.0 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.71.0 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.63.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/seccomp/libseccomp-golang v0.12.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.10.2 // indirect
	github.com/slackhq/nebula v1.11.2 // indirect
	github.com/smallstep/certificates v0.30.2 // indirect
	github.com/smallstep/cli-utils v0.12.2 // indirect
	github.com/smallstep/go-attestation v0.4.9 // indirect
	github.com/smallstep/linkedca v0.26.0 // indirect
	github.com/smallstep/nosql v0.8.0 // indirect
	github.com/smallstep/pkcs7 v0.2.3 // indirect
	github.com/smallstep/scep v0.0.0-20260331191114-261f960a40d1 // indirect
	github.com/smallstep/truststore v0.13.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tailscale/go-winio v0.0.0-20231025203758-c4f33415bf55 // indirect
	github.com/tailscale/tscert v0.0.0-20251216020129-aea342f6d747 // indirect
	github.com/tidwall/gjson v1.19.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.2 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/urfave/cli v1.22.17 // indirect
	github.com/valllabh/ocsf-schema-golang v1.0.3 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	github.com/yuin/goldmark v1.8.6 // indirect
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/bridges/prometheus v0.71.0 // indirect
	go.opentelemetry.io/contrib/exporters/autoexport v0.71.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.71.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.71.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.46.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.46.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.46.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.46.0 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.22.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.22.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.68.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.22.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.46.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.46.0 // indirect
	go.opentelemetry.io/otel/log v0.22.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.22.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.step.sm/crypto v0.91.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto/x509roots/fallback v0.0.0-20260921070245-7a4a4d6beae2 // indirect
	golang.org/x/exp v0.0.0-20260908205506-85c1c2202aba // indirect
	golang.org/x/image v0.46.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	google.golang.org/api v0.299.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260921155816-b14227669459 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260921155816-b14227669459 // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.6.2 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	howett.net/plist v1.0.1 // indirect
	modernc.org/libc v1.77.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.59.0 // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
)
