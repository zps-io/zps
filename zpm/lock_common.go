package zpm

import (
	"net/url"
)

type Locker interface {
	LockWithEtag() ([16]byte, error)
	UnlockWithEtag([16]byte) error
	Lock() error
	Unlock() error
}

func NewLocker(uri *url.URL) Locker {
	switch uri.Scheme {
	case "file":
		return NewFileLocker(uri)
	case "dynamo":
		return NewDynamoLocker(uri)
	case "none":
		return NewNoneLocker(uri)
	default:
		return nil
	}
}
