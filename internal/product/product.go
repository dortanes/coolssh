// Package product describes the app itself: its name, version, copyright, license and source.
package product

import (
	"errors"
	"fmt"

	"go.yaml.in/yaml/v3"
)

const (
	license    = "MIT"
	repository = "https://github.com/dortanes/coolssh"
	donate     = "https://anthy.net/donate"
)

// Product is what the app says about itself.
type Product struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Copyright  string `json:"copyright"`
	License    string `json:"license"`
	Repository string `json:"repository"`
	// Donate is the page listing every way to support the app.
	Donate string `json:"donate"`
	// Icon is a PNG of the app's icon.
	Icon []byte `json:"icon"`
}

// Read takes the name, version and copyright from the `info` block of build/config.yml, the file
// the build assets and their version are rendered from.
func Read(config []byte) (Product, error) {
	var file struct {
		Info struct {
			ProductName string `yaml:"productName"`
			Version     string `yaml:"version"`
			Copyright   string `yaml:"copyright"`
		} `yaml:"info"`
	}
	if err := yaml.Unmarshal(config, &file); err != nil {
		return Product{}, fmt.Errorf("build config: %w", err)
	}
	info := file.Info
	if info.ProductName == "" || info.Version == "" {
		return Product{}, errors.New("build config: info names no product or version")
	}
	return Product{
		Name:       info.ProductName,
		Version:    info.Version,
		Copyright:  info.Copyright,
		License:    license,
		Repository: repository,
		Donate:     donate,
	}, nil
}
