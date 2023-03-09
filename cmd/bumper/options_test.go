package bumper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
)

func TestConfigFromVersionOverrides_Success(t *testing.T) {
	overrideString := []string{"foopkg=1.2.3", "barpkg=6.6.6"}

	source, err := configFromVersionOverrides(overrideString)
	assert.Nil(t, err)

	actualConfig, err := config.NewYAML(source)
	assert.Nil(t, err)
	assert.Equal(t, "1.2.3", actualConfig.Get("check.versionOverrides.foopkg").String())
	assert.Equal(t, "6.6.6", actualConfig.Get("check.versionOverrides.barpkg").String())
}

func TestConfigFromVersionOverrides_Fail(t *testing.T) {
	overrideString := []string{"invalidstring"}

	_, err := configFromVersionOverrides(overrideString)
	assert.ErrorIs(t, err, ErrInvalidOverride)
	assert.ErrorContains(t, err, "'invalidstring'")
}
