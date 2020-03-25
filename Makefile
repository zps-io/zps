VERSION = 0.1.1

define REPO
priority = 10
enabled = true

fetch {
	uri = "https://zps.io/packages/zps.io/core"
}
endef

export REPO
export VERSION

os = darwin linux

all: clean $(os)

clean:
	rm -rf dist
	rm -rf *.zpkg

$(os):
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/config.d
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/image.d
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/policy.d
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/tpl.d
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/repo.d
	mkdir -p -m 0750 dist/${@}-x86_64/var/lib/zps
	mkdir -p -m 0750 dist/${@}-x86_64/var/cache/zps
	mkdir -p -m 0750 dist/${@}-x86_64/var/tmp/zps
	mkdir -p -m 0755 dist/${@}-x86_64/usr/share/zps/certs/zps.io
	cp ../zps.io/ca.pem ../zps.io/intermediate.pem ../zps.io/zps.pem dist/${@}-x86_64/usr/share/zps/certs/zps.io
	echo "$$REPO" > dist/${@}-x86_64/etc/zps/repo.d/zps.conf
	chmod 640 dist/${@}-x86_64/etc/zps/repo.d/*
	GOOS=${@} go build -ldflags "-s -w -X github.com/fezz-io/zps/cli/zps/commands.Version=${VERSION}" -o dist/${@}-x86_64/usr/bin/zps \
	github.com/fezz-io/zps/cli/zps
	OS=${@} VERSION=${VERSION} zps zpkg build --secure --target-path dist/${@}-x86_64
	tar -zcf zps-${@}-x86_64.tar.gz -C dist/${@}-x86_64 .

release: clean $(os)
	mkdir -p dist/release/downloads
	cp -Rp site/* dist/release/
	cp *.tar.gz dist/release/downloads
	zps publish "ZPS Core" *.zpkg
	aws s3 sync dist/release/ s3://zps.io/

fmt:
	goimports -w .

.PHONY: fmt
