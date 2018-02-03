package zpm

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"path/filepath"
	"os"
	"fmt"
)

type Cache struct {
	path   string
	hasher hash.Hash
}

func NewCache(path string) *Cache {
	return &Cache{path, sha256.New()}
}

func (c *Cache) Exists(name string) bool {
	if _, err := os.Stat(c.GetFile(name)); os.IsNotExist(err) {
		return false
	}

	return true
}

func (c *Cache) GetConfig(osarch string, uri string) string {
	return filepath.Join(c.path, fmt.Sprint(c.getId(uri), ".config.json"))
}

func (c *Cache) GetPackages(osarch string, uri string) string {
	return filepath.Join(c.path, fmt.Sprint(c.getId(uri), ".packages.json"))
}

func (c *Cache) GetFile(name string) string {
	return filepath.Join(c.path, name)
}

func (c *Cache) getId(id string) string {
	c.hasher.Reset()
	c.hasher.Write([]byte(id))
	return hex.EncodeToString(c.hasher.Sum(nil))
}
