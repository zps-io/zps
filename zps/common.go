package zps

import (
	"fmt"
)

func ZpkgFileName(name string, version string, os string, arch string) string {
	return fmt.Sprintf("%s@%s-%s-%s.zpkg", name, version, os, arch)
}
