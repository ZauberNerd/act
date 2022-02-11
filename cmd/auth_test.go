package cmd

import (
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ghCom = `
github.com:
  user: user
  oauth_token: ghcomtoken
`
	gheCom = `
ghe.com:
  user: user
  oauth_token: ghetoken
`
)

func TestHostsFilename(t *testing.T) {
	table := []struct {
		name     string
		env      []string
		expected string
	}{
		{
			name:     "home",
			env:      []string{"HOME", "/home/user"},
			expected: "/home/user/.config/gh/hosts.yml",
		},
		{
			name:     "xdg",
			env:      []string{"XDG_CONFIG_HOME", "/home/user/.config"},
			expected: "/home/user/.config/gh/hosts.yml",
		},
		{
			name:     "custom",
			env:      []string{"GH_CONFIG_DIR", "/home/user"},
			expected: "/home/user/hosts.yml",
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GH_CONFIG_DIR", "")
			t.Setenv("XDG_CONFIG_HOME", "")
			t.Setenv("HOME", "")

			t.Setenv(tt.env[0], tt.env[1])

			config := &GhConfig{}
			filename := config.hostsFilename()
			assert.Equal(t, tt.expected, filename)
		})
	}
}

func TestParseHostsFile(t *testing.T) {
	table := []struct {
		name     string
		reader   io.Reader
		host     string
		expected string
	}{
		{
			name:     "singleEntry",
			reader:   strings.NewReader(ghCom),
			host:     "github.com",
			expected: "ghcomtoken",
		},
		{
			name:     "multipleEntries",
			reader:   strings.NewReader(ghCom + gheCom),
			host:     "ghe.com",
			expected: "ghetoken",
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			config := &GhConfig{}

			err := config.parseHostsFile(tt.reader)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, config.hosts[tt.host].Token)
		})
	}
}

type configParserMock struct {
	mock.Mock
}

func (m *configParserMock) hostsFilename() string {
	args := m.Called()
	return args.String(0)
}
func (m *configParserMock) parseHostsFile(reader io.Reader) error {
	args := m.Called(reader)
	return args.Error(0)
}

type fsMock struct {
	mock.Mock
}

func (fsm *fsMock) Open(name string) (fs.File, error) {
	args := fsm.Called(name)
	return args.Get(0).(fs.File), args.Error(1)
}

func (fsm *fsMock) Stat(name string) (fs.FileInfo, error) {
	args := fsm.Called(name)
	return args.Get(0).(fs.FileInfo), args.Error(1)
}

type fileMock struct {
	mock.Mock
}

func (fm *fileMock) Stat() (fs.FileInfo, error) {
	args := fm.Called()
	return args.Get(0).(fs.FileInfo), args.Error(1)
}
func (fm *fileMock) Read(bytes []byte) (int, error) {
	args := fm.Called(bytes)
	return args.Int(0), args.Error(1)
}
func (fm *fileMock) Close() error {
	args := fm.Called()
	return args.Error(0)
}

type fileInfoMock struct {
	mock.Mock
}

func (fim *fileInfoMock) Name() string {
	args := fim.Called()
	return args.String(0)
}
func (fim *fileInfoMock) Size() int64 {
	args := fim.Called()
	return int64(args.Int(0))
}
func (fim *fileInfoMock) Mode() fs.FileMode {
	args := fim.Called()
	return args.Get(0).(fs.FileMode)
}
func (fim *fileInfoMock) ModTime() time.Time {
	args := fim.Called()
	return args.Get(0).(time.Time)
}
func (fim *fileInfoMock) IsDir() bool {
	args := fim.Called()
	return args.Bool(0)
}
func (fim *fileInfoMock) Sys() interface{} {
	args := fim.Called()
	return args.Get(0)
}

func TestNewGHConfig(t *testing.T) {
	fsm := &fsMock{}
	cpm := &configParserMock{}
	fm := &fileMock{}
	fim := &fileInfoMock{}

	fm.On("Close").Return(nil)

	cpm.On("hostsFilename").Return("/home/user/.config/gh/hosts.yml")
	cpm.On("parseHostsFile", fm).Return(nil)

	fsm.On("Stat", "/home/user/.config/gh/hosts.yml").Return(fim, nil)
	fsm.On("Open", "/home/user/.config/gh/hosts.yml").Return(fm, nil)

	config := NewGHConfig(fsm, cpm)
	assert.Equal(t, "/home/user/.config/gh/hosts.yml", config.hostsFile)
	assert.Equal(t, tokenBackendTypeGhHosts, config.backend)

	cpm.AssertExpectations(t)
}
