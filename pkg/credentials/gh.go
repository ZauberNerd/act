package credentials

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

type ghHostsEntry struct {
	User  string `yaml:"user"`
	Token string `yaml:"oauth_token"`
}

type GhConfig struct {
	hosts     map[string]*ghHostsEntry
	hostsFile string
}

type filesystem interface {
	fs.FS
	fs.StatFS
}

func NewGHConfig(fs filesystem) *GhConfig {
	d, _ := os.UserHomeDir()
	hostsFile := filepath.Join(d, ".config", "gh", "hosts.yml")

	if a := os.Getenv("GH_CONFIG_DIR"); a != "" {
		hostsFile = filepath.Join(a, "hosts.yml")
	} else if b := os.Getenv("XDG_CONFIG_HOME"); b != "" {
		hostsFile = filepath.Join(b, "gh", "hosts.yml")
	} else if c := os.Getenv("AppData"); runtime.GOOS == "windows" && c != "" {
		hostsFile = filepath.Join(c, "GitHub CLI", "hosts.yml")
	}

	config := &GhConfig{
		hostsFile: hostsFile,
	}

	if _, err := fs.Stat(config.hostsFile); err == nil {
		f, err := fs.Open(config.hostsFile)
		if err != nil {
			return config
		}

		defer f.Close()

		_ = yaml.NewDecoder(f).Decode(&config.hosts)
	}

	return config
}

func (config *GhConfig) Get(hostname string, key string) (string, error) {
	if host, ok := config.hosts[hostname]; ok {
		switch key {
		case "user":
			return host.User, nil
		case "oauth_token":
			return host.Token, nil
		}

		return "", fmt.Errorf("unknown key %s", key)
	}

	return "", fmt.Errorf("host %s not found", hostname)
}

func (config *GhConfig) Set(hostname string, key string, value string) error {
	host, ok := config.hosts[hostname]
	if !ok {
		host = &ghHostsEntry{}
		config.hosts[hostname] = host
	}

	switch key {
	case "user":
		host.User = value
	case "oauth_token":
		host.Token = value
	}

	return nil
}

func (config *GhConfig) Write() error {
	return nil
}
