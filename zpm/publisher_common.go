package zpm

import (
	"net/url"
)

type Publisher interface {
	Init() error
	Update() error
	Publish(...string) error
}

func NewPublisher(uri *url.URL, name string, prune int) Publisher {
	switch uri.Scheme {
	case "file":
		return NewFilePublisher(uri, name, prune)
	default:
		return nil
	}
}
