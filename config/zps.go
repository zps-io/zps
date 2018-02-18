package config

import (
	"io/ioutil"
	"path"
	"path/filepath"

	"errors"
	"fmt"

	"net/url"

	"runtime"

	"github.com/hashicorp/hcl"
)

type ZpsConfig struct {
	Mode string

	Root         string
	CurrentImage *ImageConfig

	Images []*ImageConfig
	Repos  []*RepoConfig
}

func LoadConfig(image string) (*ZpsConfig, error) {
	var err error

	config := &ZpsConfig{}

	err = config.SetRoot()
	if err != nil {
		return nil, err
	}

	// Load configured image list based on current root
	err = config.LoadImages()
	if err != nil {
		return nil, err
	}

	// If image name, path, or hash id provided override image prefix
	err = config.SelectImage(image)
	if err != nil {
		return nil, err
	}

	// Load image zps config
	err = config.LoadMain()
	if err != nil {
		return nil, err
	}

	// Load repository configs
	err = config.LoadRepos()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (z *ZpsConfig) SetRoot() error {
	var err error

	z.Root, err = InstallPrefix()
	if err != nil {
		return err
	}

	return nil
}

func (z *ZpsConfig) ConfigPath() string {
	if z.CurrentImage == nil {
		return filepath.Join(z.Root, "etc", "zps")
	}

	return filepath.Join(z.CurrentImage.Path, "etc", "zps")
}

func (z *ZpsConfig) StatePath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "lib", "zps")
}

func (z *ZpsConfig) CachePath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "cache", "zps")
}

func (z *ZpsConfig) LockPath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "lib", "zps")
}

func (z *ZpsConfig) LoadImages() error {
	defaultOs := runtime.GOOS
	var defaultArch string
	switch runtime.GOARCH {
	case "amd64":
		defaultArch = "x86_64"
	default:
		defaultArch = runtime.GOARCH
	}

	defaultImage := &ImageConfig{"default", z.Root, defaultOs, defaultArch}

	z.Images = append(z.Images, defaultImage)

	// Load defined images
	imageConfigs, err := filepath.Glob(path.Join(z.ConfigPath(), "image.d", "*.conf"))
	if err != nil {
		return nil
	}

	for _, iconfig := range imageConfigs {
		image := &ImageConfig{}

		bytes, err := ioutil.ReadFile(iconfig)
		if err != nil {
			return nil
		}

		imageHcl, err := hcl.Parse(string(bytes))
		if err != nil {
			return nil
		}

		err = hcl.DecodeObject(&image, imageHcl)
		if err != nil {
			return nil
		}

		z.Images = append(z.Images, image)
	}

	return nil
}

func (z *ZpsConfig) SelectImage(image string) error {
	if image == "" {
		z.CurrentImage = z.Images[0]
		return nil
	}

	// Select image

	for index, img := range z.Images {
		if img.Path == image {
			z.CurrentImage = z.Images[index]
			break
		}

		if img.Name == image {
			z.CurrentImage = z.Images[index]
			break
		}
	}

	return nil
}

func (z *ZpsConfig) LoadMain() error {
	var config map[string]interface{}

	bytes, err := ioutil.ReadFile(path.Join(z.ConfigPath(), "main.conf"))
	if err != nil {
		return nil
	}

	mainHcl, err := hcl.Parse(string(bytes))
	if err != nil {
		return nil
	}

	err = hcl.DecodeObject(&config, mainHcl)
	if err != nil {
		return nil
	}

	if val, ok := config["mode"]; ok {
		z.Mode = val.(string)
	} else {
		z.Mode = "ancillary"
	}

	return nil
}

func (z *ZpsConfig) LoadRepos() error {
	// Load defined repos
	repoConfigs, err := filepath.Glob(path.Join(z.ConfigPath(), "repo.d", "*.conf"))
	if err != nil {
		return nil
	}

	for _, rconfig := range repoConfigs {
		var repoMap map[string]interface{}
		repo := &RepoConfig{}
		repo.Fetch = &FetchConfig{}
		repo.Publish = &PublishConfig{}

		bytes, err := ioutil.ReadFile(rconfig)
		if err != nil {
			return nil
		}

		repoHcl, err := hcl.Parse(string(bytes))
		if err != nil {
			return nil
		}

		err = hcl.DecodeObject(&repoMap, repoHcl)
		if err != nil {
			return nil
		}

		if val, ok := repoMap["enabled"]; ok {
			repo.Enabled = val.(bool)
		} else {
			repo.Enabled = true
		}

		if val, ok := repoMap["enabled"]; ok {
			repo.Enabled = val.(bool)
		} else {
			repo.Enabled = true
		}

		if val, ok := repoMap["priority"]; ok {
			repo.Priority = val.(int)
		} else {
			repo.Priority = 10
		}

		if val, ok := repoMap["fetch"]; ok {
			if uri, ok := val.([]map[string]interface{})[0]["uri"]; ok {
				repo.Fetch.Uri, err = url.Parse(uri.(string))
				if err != nil {
					return err
				}
			} else {
				return errors.New(fmt.Sprint("config: repo fetch.uri required in ", rconfig))
			}
		} else {
			return errors.New(fmt.Sprint("config: repo fetch section required in ", rconfig))
		}

		if val, ok := repoMap["publish"]; ok {
			if uri, ok := val.([]map[string]interface{})[0]["uri"]; ok {
				repo.Publish.Uri, err = url.Parse(uri.(string))
				if err != nil {
					return err
				}
			} else {
				return errors.New(fmt.Sprint("config: repo publish.uri required in ", rconfig))
			}

			if val, ok := val.([]map[string]interface{})[0]["name"]; ok {
				repo.Publish.Name = val.(string)
			} else {
				return errors.New(fmt.Sprint("config: repo publish.name required in ", rconfig))
			}

			if prune, ok := val.([]map[string]interface{})[0]["prune"]; ok {
				repo.Publish.Prune = prune.(int)
			} else {
				repo.Publish.Prune = 10
			}
		}

		z.Repos = append(z.Repos, repo)
	}

	return nil
}
