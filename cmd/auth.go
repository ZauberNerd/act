package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cli/cli/v2/pkg/cmd/auth/shared"
	"github.com/cli/cli/v2/pkg/iostreams"
	"gopkg.in/yaml.v3"
)

func authenticate(githubInstance string) error {
	fmt.Println(githubInstance)
	return shared.Login(&shared.LoginOptions{
		IO:          iostreams.System(),
		Config:      &GhConfig{},
		HTTPClient:  &http.Client{Transport: http.DefaultTransport},
		Hostname:    githubInstance,
		Interactive: false,
		Web:         true,
		Scopes:      nil,
		Executable:  "act",
	})
}

type tokenBackendType string

const (
	tokenBackendTypeNone          tokenBackendType = "none"
	tokenBackendTypeKeychain      tokenBackendType = "keychain"
	tokenBackendTypeWincred       tokenBackendType = "wincred"
	tokenBackendTypeSecretService tokenBackendType = "secret-service"
	tokenBackendTypePass          tokenBackendType = "pass"
	tokenBackendTypeGhHosts       tokenBackendType = "hosts"
)

func (backend tokenBackendType) String() string {
	switch backend {
	case tokenBackendTypeNone:
		return "none"
	case tokenBackendTypeKeychain:
		return "MacOS Keychain"
	case tokenBackendTypeWincred:
		return "Windows Credential Manager"
	case tokenBackendTypeSecretService:
		return "DBus Secret Service"
	case tokenBackendTypePass:
		return "pass"
	case tokenBackendTypeGhHosts:
		return "gh CLI hosts file"
	}
	return "unknown"
}

type ghHostsEntry struct {
	User  string `yaml:"user"`
	Token string `yaml:"oauth_token"`
}

type GhConfig struct {
	hosts     map[string]*ghHostsEntry
	hostsFile string
	backend   tokenBackendType
}

type filesystem interface {
	fs.FS
	fs.StatFS
}

type configParser interface {
	hostsFilename() string
	parseHostsFile(io.Reader) error
}

func NewGHConfig(fs filesystem, parser configParser) *GhConfig {
	config := &GhConfig{}
	config.hostsFile = parser.hostsFilename()

	if _, err := fs.Stat(config.hostsFile); err == nil {
		f, err := fs.Open(config.hostsFile)
		if err != nil {
			return config
		}

		defer f.Close()

		err = parser.parseHostsFile(f)
		if err == nil {
			config.backend = tokenBackendTypeGhHosts
		}
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

	return "", fmt.Errorf("host %s not found in %s", hostname, config.backend)
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
	// TODO: github token implementation
	// write the token to the OS keyring
	// I would like to use:
	// windows: wincred
	// mac: keychain
	// linux: secret-service or pass
	// OS agnostic keyring implementations:
	// - https://github.com/zalando/go-keyring (does not support pass)
	// - https://github.com/keybase/go-keychain (API is awkward and only supports mac and secret-service)
	// - https://github.com/99designs/keyring (seems like the most likely fit)
	// or maybe implement something simiar to the git/docker credential helper
	// https://github.com/docker/docker-credential-helpers (might be worth considering also - but would be more work on the user's side)
	// in any case - we should go mod vendor these dependencies to keep the code on our side
	return nil
}

func (config *GhConfig) hostsFilename() string {
	if a := os.Getenv("GH_CONFIG_DIR"); a != "" {
		return filepath.Join(a, "hosts.yml")
	} else if b := os.Getenv("XDG_CONFIG_HOME"); b != "" {
		return filepath.Join(b, "gh", "hosts.yml")
	} else if c := os.Getenv("AppData"); runtime.GOOS == "windows" && c != "" {
		return filepath.Join(c, "GitHub CLI", "hosts.yml")
	} else {
		d, _ := os.UserHomeDir()
		return filepath.Join(d, ".config", "gh", "hosts.yml")
	}
}

func (config *GhConfig) parseHostsFile(reader io.Reader) error {
	return yaml.NewDecoder(reader).Decode(&config.hosts)
}
