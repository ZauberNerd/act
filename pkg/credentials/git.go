package credentials

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type command interface {
	Output() ([]byte, error)
	SetEnv(env []string)
	SetStdin(in io.Reader)
}

type commandImpl struct {
	cmd exec.Cmd
}

func (c *commandImpl) Output() ([]byte, error) {
	return c.cmd.Output()
}

func (c *commandImpl) SetEnv(env []string) {
	c.cmd.Env = env
}

func (c *commandImpl) SetStdin(in io.Reader) {
	c.cmd.Stdin = in
}

type Credentials struct {
	helper     string
	hostname   string
	username   string
	password   string
	newCommand func(string, ...string) command
}

func NewCredentials(hostname string, helper string) (*Credentials, error) {
	credentials := &Credentials{
		helper:   helper,
		hostname: hostname,
		newCommand: func(name string, args ...string) command {
			return &commandImpl{*exec.Command(name, args...)}
		},
	}

	err := credentials.loadCredentials()
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

func (credentials *Credentials) Get(hostname string, key string) (string, error) {
	if hostname == credentials.hostname {
		switch key {
		case "user":
			return credentials.username, nil
		case "oauth_token":
			return credentials.password, nil
		}
	}
	return "", fmt.Errorf("No credentials for host \"%s\"", hostname)
}

func (credentials *Credentials) Set(hostname string, key string, value string) error {
	credentials.hostname = hostname
	switch key {
	case "user":
		credentials.username = value
	case "oauth_token":
		credentials.password = value
	}

	return nil
}

func (credentials *Credentials) Write() error {
	_, err := credentials.execHelper("store", map[string]string{
		"protocol": "https",
		"host":     credentials.hostname,
		"path":     "nektos/act#github-token",
		"username": credentials.username,
		"password": credentials.password,
	})

	return err
}

func (credentials *Credentials) loadCredentials() error {
	out, err := credentials.execHelper("get", map[string]string{
		"protocol": "https",
		"host":     credentials.hostname,
		"path":     "nektos/act#github-token",
	})
	if err != nil {
		return err
	}

	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		kv := strings.Split(line, "=")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "username":
			credentials.username = value
		case "password":
			credentials.password = value
		}
	}

	return nil
}

func (credentials *Credentials) execHelper(operation string, input map[string]string) (string, error) {
	helper := credentials.helper
	if strings.HasPrefix(helper, "!") {
		helper = strings.TrimLeft(helper, "!")
	} else if !filepath.IsAbs(helper) {
		helper = fmt.Sprintf("git credential-%s", helper)
	}

	cmd := credentials.newCommand("/bin/sh", "-c", helper+" "+operation)
	cmd.SetEnv(append(os.Environ(), "GCM_INTERACTIVE=0"))
	cmd.SetStdin(strings.NewReader(buildQuery(input)))

	fmt.Println(cmd)

	out, err := cmd.Output()

	return string(out), err
}

func buildQuery(input map[string]string) string {
	query := ""
	for key, value := range input {
		query += fmt.Sprintf("%s=%s\n", key, value)
	}
	query += "\n"

	return query
}
