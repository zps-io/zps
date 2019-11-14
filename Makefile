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

zps: clean
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
