/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zpm

import (
	"net/url"

	"github.com/chuckpreslar/emission"
)

type Publisher interface {
	Init() error
	Update() error
	Channel(pkg string, channel string) error
	Publish(...string) error
}

func NewPublisher(emitter *emission.Emitter, uri *url.URL, name string, prune int) Publisher {
	switch uri.Scheme {
	case "file":
		return NewFilePublisher(emitter, uri, name, prune)
	default:
		return nil
	}
}
