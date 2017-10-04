package zps

import (
	"errors"
	"strings"
)

const (
	Proto string = "zpkg://"
)

// zpkg://solvent/tests/testpkg@0.0.1:7:20151106T023134Z

type ZpkgUri struct {
	Publisher string
	Category  string
	Name      string
	Version   *Version
}

func NewZpkgUri() *ZpkgUri {
	zu := &ZpkgUri{}
	return zu
}

func (z *ZpkgUri) Parse(uri string) error {
	// TODO change this to regex
	// match on all uri components
	// exclude garbage characters
	if !(strings.Contains(uri, Proto) && strings.Contains(uri, "@")) {
		return errors.New("meta uri does not look like a zpkg uri")
	}

	suffix := strings.Replace(uri, Proto, "", 1)

	split := strings.Split(suffix, "@")
	if len(split) != 2 {
		return errors.New("meta uri does not look like a zpkg uri")
	}

	identifier := split[0]
	version := split[1]

	z.Version = &Version{}
	err := z.Version.Parse(version)
	if err != nil {
		return err
	}

	rest := strings.Split(identifier, "/")

	z.Publisher, rest = rest[0], rest[1:]
	z.Name, rest = rest[len(rest)-1], rest[:len(rest)-1]
	z.Category = strings.Join(rest, "/")

	return err
}

func (z *ZpkgUri) String() string {
	identifier := strings.Join([]string{z.Publisher, z.Category, z.Name}, "/")
	return strings.Join([]string{Proto, identifier, "@", z.Version.String()}, "")
}
