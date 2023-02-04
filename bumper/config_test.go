package bumper

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
)

func TestReadConfig_PathOk(t *testing.T) {
	bumperConfigDirPath := filepath.Join(t.TempDir(), "some/non-standard/dir")
	os.MkdirAll(bumperConfigDirPath, 0o755)
	configPath := filepath.Join(bumperConfigDirPath, "config.yaml")
	err := ioutil.WriteFile(configPath, []byte("providers: {test_key: test_value}"), 0o644)

	actualConfig, err := ReadConfig(configPath)

	assert.Nil(t, err)
	assert.NotNil(t, actualConfig)
	assert.Equal(t, "test_value", actualConfig.Get("providers.test_key").String())
}

func TestReadConfig_PathNoConfig(t *testing.T) {
	_, err := ReadConfig("/some/non-existing/path/config.yaml")

	assert.ErrorIs(t, err, ErrInvalidConfigPath)
}

func TestReadConfig_DefaultOk(t *testing.T) {
	configDirPath := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", configDirPath)
	bumperConfigDirPath := filepath.Join(configDirPath, "bumper")
	os.MkdirAll(bumperConfigDirPath, 0o755)
	configPath := filepath.Join(bumperConfigDirPath, "config.yaml")
	err := ioutil.WriteFile(configPath, []byte("providers: {test_key: test_value}"), 0o644)

	actualConfig, err := ReadConfig("")

	assert.Nil(t, err)
	assert.NotNil(t, actualConfig)
	assert.Equal(t, "test_value", actualConfig.Get("providers.test_key").String())
}

func TestReadConfig_DefaultNoConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	actualConfig, err := ReadConfig("")

	assert.Equal(t, actualConfig, config.NopProvider{})
	assert.Nil(t, err)
}

func TestReadConfig_DefaultNoPath(t *testing.T) {
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("HOME")

	_, err := ReadConfig("")

	assert.ErrorIs(t, err, ErrUnknownConfigPath)
}
