package zps

import (
	"sort"
	"strings"
)

type OsArch struct {
	Os   string
	Arch string
}

type OsArches []*OsArch

func (slice OsArches) Len() int {
	return len(slice)
}

func (slice OsArches) Less(i, j int) bool {
	return slice[i].String() < slice[j].String()
}

func (slice OsArches) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func supportedPlatforms() map[string][]string {
	return map[string][]string{
		"darwin":  {"x86_64"},
		"freebsd": {"arm64", "x86_64"},
		"linux":   {"arm64", "x86_64"},
	}
}

func Platforms() OsArches {
	var platforms OsArches
	for os, arches := range supportedPlatforms() {
		for _, arch := range arches {
			platforms = append(platforms, &OsArch{os, arch})
		}
	}

	sort.Sort(platforms)

	return platforms
}

func (oa *OsArch) String() string {
	return strings.Join([]string{oa.Os, oa.Arch}, "-")
}
