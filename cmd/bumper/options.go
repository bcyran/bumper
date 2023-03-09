package bumper

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/config"
)

const overrideSeparator = "="

var ErrInvalidOverride = errors.New("invalid version override")

func configFromVersionOverrides(versionOverrides []string) (config.YAMLOption, error) {
	overridesMap, err := parseVersionOverrides(versionOverrides)
	if err != nil {
		return nil, err
	}

	var overrideValue interface{}
	if len(overridesMap) != 0 {
		overrideValue = overridesMap
	} else {
		overrideValue = nil
	}

	checkConfig := map[string]map[string]interface{}{
		"check": {"versionOverrides": overrideValue},
	}
	return config.Static(checkConfig), nil
}

func parseVersionOverrides(versionOverrides []string) (map[string]string, error) {
	overridesMap := map[string]string{}

	for _, overrideString := range versionOverrides {
		pkgname, override, separatorFound := strings.Cut(overrideString, overrideSeparator)
		if !separatorFound {
			return nil, fmt.Errorf("%w: '%s'", ErrInvalidOverride, overrideString)
		}
		overridesMap[pkgname] = override
	}

	return overridesMap, nil
}
