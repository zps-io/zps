package zpm

import "net/url"

type NoneLocker struct {
	uri *url.URL
}

func NewNoneLocker(uri *url.URL) *NoneLocker {
	return &NoneLocker{
		uri: uri,
	}
}

func (n *NoneLocker) Lock() error {
	return nil
}

func (n *NoneLocker) Unlock() error {
	return nil
}
