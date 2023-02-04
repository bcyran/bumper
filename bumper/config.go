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
	ErrInvalidConfigPath = errors.New("invalid configuration path")
	ErrUnknownConfigPath = errors.New("could not determine config file path")
)

// ReadConfig reads config at the given path, or at the default location
// if the path is empty.
func ReadConfig(requestedPath string) (config.Provider, error) {
	var configPath string

	if requestedPath != "" {
		// If config at specific path is requested, the path has to be valid
		if _, err := os.Stat(requestedPath); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidConfigPath, err)
		}
		configPath = requestedPath
	} else {
		// If we are falling back to the default path, it doesn't have to exist
		defaultPath, err := getConfigPath()
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(defaultPath); err != nil {
			return config.NopProvider{}, nil
		}
		configPath = defaultPath
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
	return "", ErrUnknownConfigPath
}
