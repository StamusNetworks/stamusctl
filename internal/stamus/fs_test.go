package stamus

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/go-playground/assert/v2"
)

// mockFileOpener implements FileOpener for testing.
type mockFileOpener struct {
	mkdirAllFn func(string, os.FileMode) error
	openFileFn func(string, int, os.FileMode) (*os.File, error)
	readAllFn  func(io.Reader) ([]byte, error)
}

func (m *mockFileOpener) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirAllFn != nil {
		return m.mkdirAllFn(path, perm)
	}
	return nil
}

func (m *mockFileOpener) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if m.openFileFn != nil {
		return m.openFileFn(name, flag, perm)
	}
	return nil, nil
}

func (m *mockFileOpener) ReadAll(r io.Reader) ([]byte, error) {
	if m.readAllFn != nil {
		return m.readAllFn(r)
	}
	return nil, nil
}

func TestGetOrCreateStamusConfigFile(t *testing.T) {
	app.ConfigFolder = "~"

	testPath := ""
	var testPerm os.FileMode

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		mkdirAllFn: func(path string, perm os.FileMode) error {
			testPath = path
			testPerm = perm
			return nil
		},
		openFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			testName = name
			testFlag = flag
			testOsOpenFilePerm = perm
			return nil, nil
		},
	}, nil)

	cm.getOrCreateConfigFile()

	assert.Equal(t, testPath, "~")
	assert.Equal(t, testPerm.String(), "-rwxr-xr-x")

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDWR|os.O_CREATE)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")
}

func TestGetOrCreateStamusConfigFileErrorMkdir(t *testing.T) {
	app.ConfigFolder = "~"

	testPath := ""
	var testPerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		mkdirAllFn: func(path string, perm os.FileMode) error {
			testPath = path
			testPerm = perm
			return errors.New("mock error")
		},
	}, nil)

	_, err := cm.getOrCreateConfigFile()

	assert.Equal(t, testPath, "~")
	assert.Equal(t, testPerm.String(), "-rwxr-xr-x")

	assert.Equal(t, err.Error(), "mock error")
}

func TestGetOrCreateStamusConfigFileErrorOpen(t *testing.T) {
	app.ConfigFolder = "~"

	testPath := ""
	var testPerm os.FileMode

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		mkdirAllFn: func(path string, perm os.FileMode) error {
			testPath = path
			testPerm = perm
			return nil
		},
		openFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			testName = name
			testFlag = flag
			testOsOpenFilePerm = perm
			return nil, errors.New("mock error")
		},
	}, nil)

	_, err := cm.getOrCreateConfigFile()

	assert.Equal(t, testPath, "~")
	assert.Equal(t, testPerm.String(), "-rwxr-xr-x")

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDWR|os.O_CREATE)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, err.Error(), "mock error")
}

func TestGetStamusConfig(t *testing.T) {
	app.ConfigFolder = "~"
	testConfig := &Config{
		Registries: Registries{
			"test": Logins{
				"foo": "bar",
			},
		},
	}

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		openFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			testName = name
			testFlag = flag
			testOsOpenFilePerm = perm
			return nil, nil
		},
		readAllFn: func(_ io.Reader) ([]byte, error) {
			return json.Marshal(testConfig)
		},
	}, nil)

	config, err := cm.GetConfig()

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDONLY)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, config, testConfig)
	assert.Equal(t, err, nil)
}

func TestGetStamusConfigErrorOpenFile(t *testing.T) {
	app.ConfigFolder = "~"

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		openFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			testName = name
			testFlag = flag
			testOsOpenFilePerm = perm
			return nil, errors.New("mock error")
		},
	}, nil)

	config, err := cm.GetConfig()

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDONLY)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, config, &Config{})
	assert.Equal(t, err, nil)
}

func TestGetStamusConfigErrorReadAll(t *testing.T) {
	app.ConfigFolder = "~"

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		openFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			testName = name
			testFlag = flag
			testOsOpenFilePerm = perm
			return nil, nil
		},
		readAllFn: func(_ io.Reader) ([]byte, error) {
			return []byte(""), errors.New("mock error")
		},
	}, nil)

	config, err := cm.GetConfig()

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDONLY)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, config, &Config{})
	assert.Equal(t, err, nil)
}

func TestGetStamusConfigErrorUnvalidConfig(t *testing.T) {
	app.ConfigFolder = "~"

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		openFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			testName = name
			testFlag = flag
			testOsOpenFilePerm = perm
			return nil, nil
		},
		readAllFn: func(_ io.Reader) ([]byte, error) {
			return []byte("foobar"), nil
		},
	}, nil)

	config, err := cm.GetConfig()

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDONLY)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, config, &Config{})
	assert.Equal(t, err, nil)
}

func TestGetStamusConfigErrorEmptyConfig(t *testing.T) {
	app.ConfigFolder = "~"

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	cm := NewConfigManager(&mockFileOpener{
		openFileFn: func(name string, flag int, perm os.FileMode) (*os.File, error) {
			testName = name
			testFlag = flag
			testOsOpenFilePerm = perm
			return nil, nil
		},
		readAllFn: func(_ io.Reader) ([]byte, error) {
			return []byte(""), nil
		},
	}, nil)

	config, err := cm.GetConfig()

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDONLY)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, config, &Config{})
	assert.Equal(t, err, nil)
}
