package mongo

// entrypoint 先准备数据目录并降权，再交给官方入口，避免尝试 chown 宿主机的日志文件描述符。
// 旧 mongo 使用脚本文件初始化，避免交互模式改变密码中的转义字符。
// 初始实例仍开启认证，利用 MongoDB 的 localhost exception 创建首个用户。
const entrypoint = `#!/bin/sh
set -eu
entrypoint=$(command -v "$1")
shift
if [ "$(id -u)" = 0 ]; then
    find /data/db /data/configdb ! -user mongodb -exec chown -h mongodb '{}' +
    exec gosu mongodb sh "$0" "$entrypoint" "$@"
fi
if ! command -v mongo >/dev/null 2>&1 && grep -q MONGO_INITDB_ROOT_USERNAME "$entrypoint"; then
    if ! grep -q 'file_env.*MONGO_INITDB_ROOT_PASSWORD' "$entrypoint"; then
        MONGO_INITDB_ROOT_PASSWORD=$(cat /run/sinking-database/password)
        export MONGO_INITDB_ROOT_PASSWORD
        unset MONGO_INITDB_ROOT_PASSWORD_FILE
    fi
    exec "$entrypoint" "$@"
fi
if [ ! -f /data/db/.sinking-initialized ]; then
    mongod --config /etc/sinking-database/init.conf --fork --logpath /tmp/sinking-mongo-init.log --pidfilepath /tmp/sinking-mongo-init.pid
    trap 'if [ -s /tmp/sinking-mongo-init.pid ]; then kill "$(cat /tmp/sinking-mongo-init.pid)" 2>/dev/null || :; fi' EXIT HUP INT TERM
    attempts=0
    until mongo --norc --nodb --quiet /etc/sinking-database/init.js >/dev/null 2>&1; do
        attempts=$((attempts + 1))
        if [ "$attempts" -ge 30 ] || ! kill -0 "$(cat /tmp/sinking-mongo-init.pid)" 2>/dev/null; then
            cat /tmp/sinking-mongo-init.log >&2
            echo 'MongoDB 管理员初始化失败' >&2
            exit 1
        fi
        sleep 1
    done
    mongod --config /etc/sinking-database/init.conf --shutdown
    trap - EXIT HUP INT TERM
    touch /data/db/.sinking-initialized
    rm -f /tmp/sinking-mongo-init.log /tmp/sinking-mongo-init.pid
fi
unset MONGO_INITDB_ROOT_USERNAME MONGO_INITDB_ROOT_PASSWORD_FILE
exec "$entrypoint" "$@"
`
