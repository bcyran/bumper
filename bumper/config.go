package bumper

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/config"
)

const (
	defaultConfigDir = ".config"
	relConfigPath    = "bumper/config.yaml"
)

var (
	UnknownConfigPath = errors.New("could not determine config file path")
)

func ReadConfig() (config.Provider, error) {
	configPath, err := getConfigPath()
	fmt.Printf("config path: %s\n", configPath)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); err != nil {
		return nil, nil
	}

	return config.NewYAML(config.File(configPath))
}

func getConfigPath() (string, error) {
	configHome, configHomeSet := os.LookupEnv("XDG_CONFIG_HOME")
	if configHomeSet {
		return filepath.Join(configHome, relConfigPath), nil
	}
	userHome, userHomeSet := os.LookupEnv("HOME")
	if userHomeSet {
		return filepath.Join(userHome, defaultConfigDir, relConfigPath), nil
	}
	return "", UnknownConfigPath
}
