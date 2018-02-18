package zpm

import (
	"net/url"

	"github.com/chuckpreslar/emission"
)

type Publisher interface {
	Init() error
	Update() error
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
