VERSION = 0.1.0

define REPO
priority = 10
enabled = true

fetch {
	uri = "https://packages.zps.io/zps.io/zps"
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
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/image.d
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/policy.d
	mkdir -p -m 0750 dist/${@}-x86_64/etc/zps/repo.d
	mkdir -p -m 0750 dist/${@}-x86_64/var/lib/zps
	mkdir -p -m 0750 dist/${@}-x86_64/var/cache/zps
	mkdir -p -m 0750 dist/${@}-x86_64/var/tmp/zps
	mkdir -p -m 0755 dist/${@}-x86_64/usr/share/zps/certs/zps.io
	cp ../zps.io/ca.pem ../zps.io/intermediate.pem ../zps.io/zps.pem dist/${@}-x86_64/usr/share/zps/certs/zps.io
	echo "$$REPO" > dist/${@}-x86_64/etc/zps/repo.d/zps.conf
	chmod 640 dist/${@}-x86_64/etc/zps/repo.d/*
	GOOS=${@} go build -ldflags "-X github.com/fezz-io/zps/cli/zps/commands.Version=${VERSION}" -o dist/${@}-x86_64/usr/bin/zps \
	github.com/fezz-io/zps/cli/zps
	OS=${@} Version=${VERSION} zps zpkg build --target-path dist/${@}-x86_64

fmt:
	goimports -w .

.PHONY: fmt
