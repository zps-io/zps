package zpm

import (
	"net/url"
	"strings"

	"github.com/fezz-io/zps/zps"
)

func FilterPackagesByArch(osarch *zps.OsArch, zpkgs map[string]*zps.Pkg) ([]string, []*zps.Pkg) {
	var files []string
	var pkgs []*zps.Pkg

	for k, v := range zpkgs {
		if v.Os() == osarch.Os && v.Arch() == osarch.Arch {
			files = append(files, k)
			pkgs = append(pkgs, zpkgs[k])
		}
	}

	return files, pkgs
}

func PublisherFromUri(uri *url.URL) string {
	parts := strings.Split(uri.Path, "/")

	if len(parts) < 2 {
		return ""
	} else {
		return parts[len(parts)-2]
	}
}
