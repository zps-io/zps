package config

import (
	"os"
	"strings"

	"github.com/kardianos/osext"
)

func InstallPrefix() (string, error) {
	binPath, err := osext.ExecutableFolder()
	if err != nil {
		return "", err
	}

	return strings.Replace(binPath, "usr"+string(os.PathSeparator)+"bin", "", 1), nil
}
