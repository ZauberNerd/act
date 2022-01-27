package cmd

import (
	"fmt"
	"net/http"

	"github.com/cli/cli/v2/pkg/cmd/auth/shared"
	"github.com/cli/cli/v2/pkg/iostreams"
)

type Config struct {
	user        string
	oauth_token string // nolint:revive
}

func (config *Config) Get(hostname string, key string) (string, error) {
	switch key {
	case "user":
		return config.user, nil
	case "oauth_token":
		return config.oauth_token, nil
	default:
		return "", nil
	}
}
func (config *Config) Set(hostname string, key string, value string) error {
	switch key {
	case "user":
		config.user = value
	case "oauth_token":
		config.oauth_token = value
	default:
		return nil
	}
	return nil
}
func (config *Config) Write() error {
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

func authenticate(githubInstance string) error {
	fmt.Println(githubInstance)
	return shared.Login(&shared.LoginOptions{
		IO:          iostreams.System(),
		Config:      &Config{},
		HTTPClient:  &http.Client{Transport: http.DefaultTransport},
		Hostname:    githubInstance,
		Interactive: false,
		Web:         true,
		Scopes:      nil,
		Executable:  "act",
	})
}
