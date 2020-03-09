package zpm

import (
	"github.com/fezz-io/zps/action"
)

const (
	SecurityModeNone = "none"
)

type SecurityNone struct{}

func (s *SecurityNone) Mode() string {
	return SecurityModeNone
}

func (s *SecurityNone) KeyPair(publisher string) (*KeyPairEntry, error) {
	return nil, nil
}

// TODO warn on the presence of invalid signatures
func (s *SecurityNone) Verify(content *[]byte, signatures []*action.Signature) (*action.Signature, error) {
	return nil, nil
}
