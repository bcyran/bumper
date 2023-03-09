package upstream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion_Valid(t *testing.T) {
	cases := map[string]Version{
		"v1.2.3": Version("1.2.3"),
		"1.2.3":  Version("1.2.3"),
		"1.2_b3": Version("1.2_b3"),
		"10":     Version("10"),
	}

	for rawVersion, expectedResult := range cases {
		result, valid := ParseVersion(rawVersion)

		assert.True(t, valid)
		assert.Equal(t, expectedResult, result)
	}
}

func TestParseVersion_Invalid(t *testing.T) {
	cases := []string{
		"abc",
		"",
		"1.2-b3",
		"1.2.%3",
		"1:2:3",
	}

	for _, rawVersion := range cases {
		_, valid := ParseVersion(rawVersion)

		assert.False(t, valid)
	}
}
