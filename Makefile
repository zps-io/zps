ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CONFIG='mode = "ancillary"'

define REPO_A
priority = 10
enabled = true

fetch {
	uri = "file://$(ROOT_DIR)/test/testrepo"
}

publish {
	uri = "file://$(ROOT_DIR)/test/testrepo"
	name = "Test Repo"
	prune = 3
}
endef

define REPO_B
priority = 10
enabled = true

fetch {
	uri = "file://$(ROOT_DIR)/test/anotherrepo"
}

publish {
	uri = "file://$(ROOT_DIR)/test/anotherrepo"
	name = "Another Repo"
	prune = 3
}
endef

define REPO_F
priority = 10
enabled = true

channels = [
	"spoon"
]

fetch {
	uri = "file://$(ROOT_DIR)/test/filterrepo"
}

publish {
	uri = "file://$(ROOT_DIR)/test/filterrepo"
	name = "Filtered Repo"
	prune = 3
}
endef

export REPO_A
export REPO_B
export REPO_F


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
	go get github.com/coreos/bbolt/...
	go get github.com/chuckpreslar/emission
	go get github.com/spf13/cobra/cobra
	go get github.com/mitchellh/colorstring
	go get github.com/hashicorp/hcl
	go get github.com/hashicorp/hil
	go get github.com/hashicorp/hcl2/...
	go get github.com/hashicorp/go-multierror
	go get github.com/ryanuber/columnize
	go get github.com/blang/semver
	go get github.com/solvent-io/sat
	go get gonum.org/v1/gonum/graph
	go get github.com/asdine/storm
	go get github.com/segmentio/ksuid
	go get github.com/davecgh/go-spew/spew
	go get github.com/nightlyone/lockfile

zps: clean deps
	mkdir -p dist/etc/zps/image.d
	mkdir -p dist/etc/zps/policy.d
	mkdir -p dist/etc/zps/repo.d
	mkdir -p dist/var/lib/zps
	mkdir -p dist/var/cache/zps
	echo $(CONFIG) > dist/etc/zps/main.conf
	echo "$$REPO_A" > dist/etc/zps/repo.d/testrepo.conf
	echo "$$REPO_B" > dist/etc/zps/repo.d/anotherrepo.conf
	echo "$$REPO_F" > dist/etc/zps/repo.d/filteredrepo.conf
	go build -o dist/usr/bin/zps github.com/solvent-io/zps/cli/zps
	ln dist/usr/bin/zps dist/usr/bin/zpkg
#	ln dist/usr/bin/zps dist/usr/bin/zpm


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
