ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CONFIG='mode = "ancillary"'

define REPO_A
priority = 10
enabled = true

fetch {
	uri = "file://$(ROOT_DIR)/test/fezz.io/testrepo"
}

publish {
	uri = "file://$(ROOT_DIR)/test/fezz.io/testrepo"
	name = "Test Repo"
	prune = 3
}
endef

define REPO_B
priority = 10
enabled = true

fetch {
	uri = "file://$(ROOT_DIR)/test/fezz.io/anotherrepo"
}

publish {
	uri = "file://$(ROOT_DIR)/test/fezz.io/anotherrepo"
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
	uri = "file://$(ROOT_DIR)/test/fezz.io/filterrepo"
}

publish {
	uri = "file://$(ROOT_DIR)/test/fezz.io/filterrepo"
	name = "Filtered Repo"
	prune = 3
}
endef

define REPO_S
priority = 10
enabled = true

fetch {
	uri = "s3://packages.fezz.io/fezz.io/s3repo"
}

publish {
	uri = "s3://packages.fezz.io/fezz.io/s3repo"
	name = "S3 Repo"
	prune = 3
}
endef

export REPO_A
export REPO_B
export REPO_F
export REPO_S


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
	go get github.com/hashicorp/hcl2/...
	go get github.com/hashicorp/go-multierror
	go get github.com/ryanuber/columnize
	go get github.com/blang/semver
	go get github.com/fezz-io/sat
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
	mkdir -p dist/var/tmp/zps
	echo $(CONFIG) > dist/etc/zps/main.conf
	echo "$$REPO_A" > dist/etc/zps/repo.d/testrepo.conf
	echo "$$REPO_B" > dist/etc/zps/repo.d/anotherrepo.conf
	echo "$$REPO_F" > dist/etc/zps/repo.d/filteredrepo.conf
	echo "$$REPO_S" > dist/etc/zps/repo.d/s3repo.conf
	go build -o dist/usr/bin/zps github.com/fezz-io/zps/cli/zps
	ln dist/usr/bin/zps dist/usr/bin/zpkg
	ln dist/usr/bin/zps dist/usr/bin/zpm

fmt:
	goimports -w .

.PHONY: deps fmt
