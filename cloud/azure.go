package cloud

import (
	"net/url"
	"strings"
)

type AzureMeta struct {
	Compute *AzureMetaCompute `json:"compute"`
}

type AzureMetaCompute struct {
	Tags string `json:"tags"`
}

func AzureStorageAccountFromURL(url *url.URL) string {
	return strings.ReplaceAll(url.Hostname(), ".", "")
}

// Azure does not allow "." in container names
func AzureBlobContainerFromURL(url *url.URL) string {
	return strings.ReplaceAll(strings.Split(url.Path, "/")[1], ".", "-")
}

func AzureBlobObjectPrefixFromURL(url *url.URL) string {
	fragments := strings.Split(url.Path, "/")
	if len(fragments) <= 2 {
		return ""
	}

	return strings.TrimPrefix(url.Path, "/"+strings.Split(url.Path, "/")[1]+"/")
}