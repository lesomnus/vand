package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/lesomnus/z"
)

var DefaultConfigPaths = []string{
	"vand.yaml",
	"vand.yml",
}

type Config struct {
	path string

	Server   ServerConfig
	Packages PackageConfig

	Otel OtelConfig
}

func ReadFromFile(p string) (*Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, z.Err(err, "open")
	}

	var c Config
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		return nil, z.Err(err, "decode")
	}

	c.path = p
	return &c, nil
}

func (c *Config) Path() string {
	return c.path
}

func (c *Config) Evaluate() error {
	z.FallbackP(&c.Server.Addr, ":8080")
	if c.Packages == nil {
		c.Packages = make(PackageConfig)
	}
	for name, pkg := range c.Packages {
		if pkg.Root == nil {
			continue
		}
		root := *pkg.Root
		if root != "" && root != name && !strings.HasPrefix(name, root+"/") {
			return fmt.Errorf("package %q: root %q must be a path prefix of the package path", name, root)
		}
	}
	return nil
}
