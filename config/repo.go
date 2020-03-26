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
	"net/url"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type RepoConfig struct {
	// Used only for Imagefile
	Name string `hcl:"name,label"`

	Priority int  `hcl:"priority"`
	Enabled  bool `hcl:"enabled"`

	Channels []string `hcl:"channels,optional"`

	Fetch   *FetchConfig   `hcl:"fetch,block"`
	Publish *PublishConfig `hcl:"publish,block"`
}

type FetchConfig struct {
	Uri       *url.URL
	UriString string `hcl:"uri"`
}

type PublishConfig struct {
	Uri       *url.URL
	UriString string `hcl:"uri"`
	Name      string `hcl:"name"`
	Prune     int    `hcl:"prune"`
}

// Sadly there is no way yet to dump a struct to HCL
// so when the struct changes we have to update the HCL encoding
func (r *RepoConfig) ToHclFile() *hclwrite.File {
	file := hclwrite.NewEmptyFile()

	file.Body().SetAttributeValue("priority", cty.NumberIntVal(int64(r.Priority)))
	file.Body().SetAttributeValue("enabled", cty.BoolVal(r.Enabled))

	if len(r.Channels) > 0 {
		var channels []cty.Value

		for _, ch := range r.Channels {
			channels = append(channels, cty.StringVal(ch))
		}

		file.Body().AppendNewline()
		file.Body().SetAttributeValue("channels", cty.ListVal(channels))
	}

	if r.Fetch != nil {
		file.Body().AppendNewline()

		fetch := file.Body().AppendNewBlock("fetch", nil)
		fetch.Body().SetAttributeValue("uri", cty.StringVal(r.Fetch.UriString))
	}

	if r.Publish != nil {
		file.Body().AppendNewline()

		publish := file.Body().AppendNewBlock("fetch", nil)
		publish.Body().SetAttributeValue("uri", cty.StringVal(r.Publish.UriString))
		publish.Body().SetAttributeValue("name", cty.StringVal(r.Publish.Name))
		publish.Body().SetAttributeValue("prune", cty.NumberIntVal(int64(r.Publish.Prune)))
	}

	return file
}
