package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/config"
)

func authenticate(githubInstance string) error {
	fmt.Println(githubInstance)

	config, err := config.LoadConfig(config.GlobalScope)
	if err != nil {
		return err
	}

	credential := config.Raw.Section("credential")
	helper := credential.Options.Get("helper")

	var cmd exec.Cmd
	switch {
	case strings.HasPrefix(helper, "!"):
		cmd = *exec.Command("/bin/sh", "-c", strings.TrimLeft(helper, "!"), "get")
	case filepath.IsAbs(helper):
		if strings.Contains(helper, " ") {
			fields := strings.Fields(helper)
			fields = append(fields, "get")
			for i, field := range fields {
				if strings.HasPrefix(field, "~") {
					fields[i] = strings.Replace(field, "~", os.Getenv("HOME"), 1)
				}
			}
			cmd = *exec.Command(fields[0], fields[1:]...)
		} else {
			cmd = *exec.Command(helper, "get")
		}
	default:
		cmd = *exec.Command("git", fmt.Sprintf("credential-%s", helper), "get")
	}

	cmd.Stdin = strings.NewReader(fmt.Sprintf(
		"protocol=%s\nhost=%s\n\n",
		"https",
		"github.com",
	))

	cmd.Env = append(os.Environ(), "GCM_INTERACTIVE=0")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
	// return shared.Login(&shared.LoginOptions{
	// 	IO:          iostreams.System(),
	// 	Config:      &GhConfig{},
	// 	HTTPClient:  &http.Client{Transport: http.DefaultTransport},
	// 	Hostname:    githubInstance,
	// 	Interactive: false,
	// 	Web:         true,
	// 	Scopes:      nil,
	// 	Executable:  "act",
	// })
}
