ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CONFIG='mode = "ancillary"'

define REPO
name = "Test Repo"
priority = 0
enabled = true

fetch {
	uri = "file://$(ROOT_DIR)/test/testrepo"
}

publish {
	uri = "file://$(ROOT_DIR)/test/testrepo"
	prune = 3
}
endef

export REPO


all: zps

clean:
	rm -rf dist

deps:
	go get golang.org/x/tools/cmd/goimports
	go get golang.org/x/sys/unix
	go get golang.org/x/net/context
	go get github.com/kardianos/osext
	go get github.com/naegelejd/go-acl/os/group
	go get github.com/lunixbochs/struc
	go get github.com/dsnet/compress
	go get github.com/boltdb/bolt/...
	go get github.com/chuckpreslar/emission
	go get github.com/spf13/cobra/cobra
	go get github.com/mitchellh/colorstring
	go get github.com/hashicorp/hcl
	go get github.com/hashicorp/hil
	go get github.com/hashicorp/go-multierror
	go get github.com/ryanuber/columnize
	go get github.com/blang/semver
	go get github.com/solvent-io/sat
	go get github.com/davecgh/go-spew/spew

zps: clean deps
	mkdir -p dist/etc/zps/image.d
	mkdir -p dist/etc/zps/policy.d
	mkdir -p dist/etc/zps/repo.d
	mkdir -p dist/var/lib/zps
	mkdir -p dist/var/cache/zps
	echo $(CONFIG) > dist/etc/zps/main.conf
	echo "$$REPO" > dist/etc/zps/repo.d/testrepo.conf
	go build -o dist/usr/bin/zps github.com/solvent-io/zps/cli/zps
	ln dist/usr/bin/zps dist/usr/bin/zpkg
	ln dist/usr/bin/zps dist/usr/bin/zpm


illumos: clean deps
	mkdir -p dist/illumos/etc/zps/image.d
	mkdir -p dist/illumos/etc/zps/policy.d
	mkdir -p dist/illumos/etc/zps/repo.d
	mkdir -p dist/var/cache/zps/repo
	echo $(CONFIG) >> dist/illumos/etc/zps/config.conf
	GOOS=solaris go build -o dist/illumos/usr/bin/zps github.com/solvent-io/zps/cmd/zps
	ln dist/illumos/usr/bin/zps dist/illumos/usr/bin/zpkg
	ln dist/illumos/usr/bin/zps dist/illumos/usr/bin/zpm

fmt:
	goimports -w .

.PHONY: deps fmt
