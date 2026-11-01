#!/bin/sh
set -eu

lesson=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
root=$(dirname "$lesson")
work="$root/tmp/tutorial"

if [ -e "$work" ]; then
    echo "The tutorial workspace already exists: $work" >&2
    echo "Move it aside before starting a new walkthrough." >&2
    exit 1
fi

GOWORK=off go run "$lesson/render.go" -dir "$lesson"
mkdir -p "$root/tmp"
mkdir "$work"
cd "$work"
export GOWORK=off

go mod init example.com/blog
go mod edit -go=1.27.3
go get -tool \
    github.com/go-kanna/orm/cmd/kanna-orm@v0.0.9 \
    github.com/go-kanna/validate/cmd/kanna-validate@v0.0.15 \
    github.com/go-kanna/mapper/cmd/kanna-mapper@v0.0.14 \
    github.com/go-kanna/di/cmd/kanna-di@v0.0.7 \
    github.com/go-kanna/fixture/cmd/kanna-fixture@v0.0.11
go get github.com/jackc/pgx/v5/pgxpool@v5.11.0
cp "$lesson/files/http.go.txt" http.go
cp "$lesson/files/api_test.go.txt" api_test.go

docker compose -p kanna-tutorial -f "$lesson/compose.yaml" up -d --wait
docker compose -p kanna-tutorial -f "$lesson/compose.yaml" exec -T db \
    psql -U tutorial -d tutorial -v ON_ERROR_STOP=1 < "$lesson/posts.sql"
address=$(docker compose -p kanna-tutorial -f "$lesson/compose.yaml" port db 5432)
port=${address##*:}

cat > env.sh <<EOF
export GOWORK=off
export DATABASE_URL='postgres://tutorial:tutorial@127.0.0.1:$port/tutorial?sslmode=disable'
export LESSON='$lesson'
EOF

echo "Ready: $work"
echo 'Run: cd tmp/tutorial && . ./env.sh'
