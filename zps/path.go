package zps

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
)

func ConfigPath() (string, error) {
	prefix, err := InstallPrefix()
	if err != nil {
		return "", err
	}

	return filepath.Join(prefix, "etc", "zps"), nil
}

func DbPath() (string, error) {
	prefix, err := InstallPrefix()
	if err != nil {
		return "", err
	}

	return filepath.Join(prefix, "var", "lib", "zps"), nil
}

func CachePath() (string, error) {
	prefix, err := InstallPrefix()
	if err != nil {
		return "", err
	}

	return filepath.Join(prefix, "var", "cache", "zps"), nil
}

func InstallPrefix() (string, error) {
	binPath, err := osext.ExecutableFolder()
	if err != nil {
		return "", err
	}

	return strings.Replace(binPath, "usr"+string(os.PathSeparator)+"bin", "", 1), nil
}
