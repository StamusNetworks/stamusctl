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

func TestGetOrCreateStamusConfigFile(t *testing.T) {
	app.ConfigFolder = "~"

	testPath := ""
	var testPerm os.FileMode

	osMkdirAll = func(path string, perm os.FileMode) error {
		testPath = path
		testPerm = perm

		return nil
	}

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		testName = name
		testFlag = flag
		testOsOpenFilePerm = perm

		return nil, nil
	}

	getOrCreateStamusConfigFile()

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

	osMkdirAll = func(path string, perm os.FileMode) error {
		testPath = path
		testPerm = perm

		return errors.New("mock error")
	}

	_, err := getOrCreateStamusConfigFile()

	assert.Equal(t, testPath, "~")
	assert.Equal(t, testPerm.String(), "-rwxr-xr-x")

	assert.Equal(t, err.Error(), "mock error")
}

func TestGetOrCreateStamusConfigFileErrorOpen(t *testing.T) {
	app.ConfigFolder = "~"

	testPath := ""
	var testPerm os.FileMode

	osMkdirAll = func(path string, perm os.FileMode) error {
		testPath = path
		testPerm = perm

		return nil
	}

	testName := ""
	testFlag := 0
	var testOsOpenFilePerm os.FileMode

	osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		testName = name
		testFlag = flag
		testOsOpenFilePerm = perm

		return nil, errors.New("mock error")
	}

	_, err := getOrCreateStamusConfigFile()

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

	osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		testName = name
		testFlag = flag
		testOsOpenFilePerm = perm

		return nil, nil
	}

	ioReadAll = func(_ io.Reader) ([]byte, error) {
		return json.Marshal(testConfig)
	}

	config, err := GetStamusConfig()

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDONLY)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, config, testConfig)
	assert.Equal(t, err, nil)
}

func TestGetStamusConfigErrorOpenFile(t *testing.T) {
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

	osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		testName = name
		testFlag = flag
		testOsOpenFilePerm = perm

		return nil, errors.New("mock error")
	}

	ioReadAll = func(_ io.Reader) ([]byte, error) {
		return json.Marshal(testConfig)
	}

	config, err := GetStamusConfig()

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

	osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		testName = name
		testFlag = flag
		testOsOpenFilePerm = perm

		return nil, nil
	}

	ioReadAll = func(_ io.Reader) ([]byte, error) {
		return []byte(""), errors.New("mock error")
	}

	config, err := GetStamusConfig()

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

	osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		testName = name
		testFlag = flag
		testOsOpenFilePerm = perm

		return nil, nil
	}

	ioReadAll = func(_ io.Reader) ([]byte, error) {
		return []byte("foobar"), nil
	}

	config, err := GetStamusConfig()

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

	osOpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		testName = name
		testFlag = flag
		testOsOpenFilePerm = perm

		return nil, nil
	}

	ioReadAll = func(_ io.Reader) ([]byte, error) {
		return []byte(""), nil
	}

	config, err := GetStamusConfig()

	assert.Equal(t, testName, "~/config.json")
	assert.Equal(t, testFlag, os.O_RDONLY)
	assert.Equal(t, testOsOpenFilePerm.String(), "-rwxr-xr-x")

	assert.Equal(t, config, &Config{})
	assert.Equal(t, err, nil)
}
