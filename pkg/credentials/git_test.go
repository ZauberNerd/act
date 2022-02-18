package credentials

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCredentialsGet(t *testing.T) {
	credentials := &Credentials{
		hostname: "github.com",
		username: "username",
		password: "password",
	}

	username, err := credentials.Get("github.com", "user")
	if assert.NoError(t, err) {
		assert.Equal(t, "username", username)
	}

	password, err := credentials.Get("github.com", "oauth_token")
	if assert.NoError(t, err) {
		assert.Equal(t, "password", password)
	}
}

func TestCredentialsSet(t *testing.T) {
	credentials := &Credentials{}

	err := credentials.Set("github.com", "user", "username")
	if assert.NoError(t, err) {
		assert.Equal(t, "username", credentials.username)
	}

	err = credentials.Set("github.com", "oauth_token", "password")
	if assert.NoError(t, err) {
		assert.Equal(t, "password", credentials.password)
	}
}

type mockCommand struct {
	mock.Mock
	commandImpl
}

func (mc *mockCommand) Output() ([]byte, error) {
	args := mc.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func TestCredentialsWrite(t *testing.T) {
	mc := &mockCommand{}

	credentials := &Credentials{
		helper:   "libsecret",
		hostname: "github.com",
		username: "username",
		password: "password",
		newCommand: func(name string, args ...string) command {
			mc.cmd = *exec.Command(name, args...)
			return mc
		},
	}

	mc.On("Output").Return([]byte(""), nil)

	err := credentials.Write()
	if assert.NoError(t, err) {
		assert.Equal(t, "/bin/sh", mc.cmd.Path)
		assert.Equal(t, []string{"/bin/sh", "-c", "git credential-libsecret store"}, mc.cmd.Args)
		assert.Contains(t, mc.cmd.Env, "GCM_INTERACTIVE=0")

		stdin := new(strings.Builder)
		_, err = io.Copy(stdin, mc.cmd.Stdin)
		if assert.NoError(t, err) {
			parts := strings.Split(stdin.String(), "\n")
			assert.ElementsMatch(t, []string{"protocol=https", "host=github.com", "username=username", "password=password", "path=nektos/act#github-token", "", ""}, parts)
		}
	}

	mc.AssertExpectations(t)
}

func TestLoadCredentials(t *testing.T) {
	mc := &mockCommand{}

	credentials := &Credentials{
		helper:   "libsecret",
		hostname: "github.com",
		newCommand: func(name string, args ...string) command {
			mc.cmd = *exec.Command(name, args...)
			return mc
		},
	}

	mc.On("Output").Return([]byte("protocol=https\nhost=github.com\nusername=username\npassword=password\npath=nektos/act#github-token\n\n"), nil)

	err := credentials.loadCredentials()
	if assert.NoError(t, err) {
		assert.Equal(t, "username", credentials.username)
		assert.Equal(t, "password", credentials.password)
		assert.Contains(t, mc.cmd.Env, "GCM_INTERACTIVE=0")

		stdin := new(strings.Builder)
		_, err = io.Copy(stdin, mc.cmd.Stdin)
		if assert.NoError(t, err) {
			parts := strings.Split(stdin.String(), "\n")
			assert.ElementsMatch(t, []string{"protocol=https", "host=github.com", "path=nektos/act#github-token", "", ""}, parts)
		}
	}
}

func TestNewCredentialsStore(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), ".git-credentials")
	if assert.NoError(t, err) {
		_, err = tmpFile.Write([]byte("https://user:pass@github.com/nektos/act#github-token\n"))
		if assert.NoError(t, err) {
			credentials, err := NewCredentials("github.com", "store --file "+tmpFile.Name())
			if assert.NoError(t, err) {
				assert.Equal(t, "user", credentials.username)
				assert.Equal(t, "pass", credentials.password)
			}
		}
	}
}

func TestNewCredentialsShell(t *testing.T) {
	t.Setenv("GIT_USERNAME", "user")
	t.Setenv("GIT_PASSWORD", "pass")

	credentials, err := NewCredentials("github.com", `!f() { echo "username=${GIT_USERNAME}"; echo "password=${GIT_PASSWORD}"; }; f`)
	if assert.NoError(t, err) {
		assert.Equal(t, "user", credentials.username)
		assert.Equal(t, "pass", credentials.password)
	}
}
