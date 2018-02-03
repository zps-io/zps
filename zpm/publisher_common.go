package zpm

import (
	"net/url"
)

type Publisher interface {
	Init() error
	Publish(...string) error
}

func NewPublisher(uri *url.URL, prune int) Publisher {
	switch uri.Scheme {
	case "file":
		return NewFilePublisher(uri, prune)
	default:
		return nil
	}
}
