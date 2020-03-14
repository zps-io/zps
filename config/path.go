/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

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

	return strings.Replace(binPath, string(os.PathSeparator)+"usr"+string(os.PathSeparator)+"bin", "", 1), nil
}
