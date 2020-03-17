VERSION = "0.1.0"

define REPO
priority = 10
enabled = true

fetch {
	uri = "https://packages.zps.io/zps.io/zps"
}
endef

export REPO
export VERSION

all: zps

clean:
	rm -rf dist

zps: clean
	mkdir -p dist/etc/zps/image.d
	mkdir -p dist/etc/zps/policy.d
	mkdir -p dist/etc/zps/repo.d
	mkdir -p dist/var/lib/zps
	mkdir -p dist/var/cache/zps
	mkdir -p dist/var/tmp/zps
	echo "$$REPO" > dist/etc/zps/repo.d/zps.conf
	go build -ldflags "-X github.com/fezz-io/zps/cli/zps/commands.Version=${VERSION}" -o dist/usr/bin/zps github.com/fezz-io/zps/cli/zps

fmt:
	goimports -w .

.PHONY: fmt
