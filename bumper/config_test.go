package bumper

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig_Ok(t *testing.T) {
	configDirPath := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", configDirPath)
	bumperConfigDirPath := filepath.Join(configDirPath, "bumper")
	os.MkdirAll(bumperConfigDirPath, 0755)
	configPath := filepath.Join(bumperConfigDirPath, "config.yaml")
	err := ioutil.WriteFile(configPath, []byte("providers: {test_key: test_value}"), 0644)

	config, err := ReadConfig()

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test_value", config.Get("providers.test_key").String())
}

func TestReadConfig_NoConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	config, err := ReadConfig()

	assert.Nil(t, config)
	assert.Nil(t, err)
}
func TestReadConfig_UnknownPath(t *testing.T) {
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("HOME")

	_, err := ReadConfig()

	assert.ErrorIs(t, err, UnknownConfigPath)
}
