/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package config

type ImageConfig struct {
	Name string `hcl:"name"`
	Path string `hcl:"path"`
	Os   string `hcl:"os"`
	Arch string `hcl:"arch"`
}
