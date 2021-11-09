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
	_, err := n.LockWithEtag()
	return err
}

func (n *NoneLocker) Unlock() error {
	return n.UnlockWithEtag([16]byte{})
}

func (n *NoneLocker) LockWithEtag() ([16]byte, error) {
	return [16]byte{}, nil
}

func (n *NoneLocker) UnlockWithEtag(etag [16]byte) error {
	return nil
}
