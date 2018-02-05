package zpm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path/filepath"
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

func (c *Cache) GetConfig(uri string) string {
	return filepath.Join(c.path, fmt.Sprint(c.getId(uri), ".config.json"))
}

func (c *Cache) GetPackages(osarch string, uri string) string {
	return filepath.Join(c.path, fmt.Sprint(c.getId(uri), "-", osarch, ".packages.json"))
}

func (c *Cache) GetFile(name string) string {
	return filepath.Join(c.path, name)
}

func (c *Cache) Clean() error {
	pkgs, _ := filepath.Glob(filepath.Join(c.path, "*.zpkg"))

	for _, f := range pkgs {
		os.Remove(f)
	}

	return nil
}

func (c *Cache) Clear() error {
	err := c.Clean()
	if err != nil {
		return err
	}

	metafiles, _ := filepath.Glob(filepath.Join(c.path, "*.json"))

	for _, f := range metafiles {
		os.Remove(f)
	}

	return nil
}

func (c *Cache) getId(id string) string {
	c.hasher.Reset()
	c.hasher.Write([]byte(id))
	return hex.EncodeToString(c.hasher.Sum(nil))
}
