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
	return nil
}

func (n *NoneLocker) LockWithEtag() (string, error) {
	return "", nil
}

func (n *NoneLocker) UnlockWithEtag(etag *string) error {
	return nil
}
