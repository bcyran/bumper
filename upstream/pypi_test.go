package upstream

import (
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestNewPypi_Valid(t *testing.T) {
	cases := map[string]pypiProvider{
		"https://files.pythonhosted.org/packages/source/f/foo/foo-1.1.0.tar.gz": {packageName: "foo"},
		"https://pypi.python.org/packages/source/b/bar/bar-1.1.0.tar.gz":        {packageName: "bar"},
		"https://pypi.org/packages/source/b/baz/baz-1.1.0.tar.gz":               {packageName: "baz"},
		"https://pypi.io/packages/source/f/foo/foo-1.1.0.tar.gz":                {packageName: "foo"},
	}

	for validURL, expectedResult := range cases {
		result := newPypiProvider(validURL)
		assert.Equal(t, &expectedResult, result)
	}
}

func TestNewPypi_Invalid(t *testing.T) {
	invalidURL := "https://whatever.url/packages/source/f/foo/foo-1.1.0.tar.gz"

	result := newPypiProvider(invalidURL)
	assert.Nil(t, result)
}

func TestPypiLatestVersion(t *testing.T) {
	defer gock.Off()
	gock.New("https://pypi.org").
		Get("/pypi/some-package/json").
		Reply(200).
		JSON(map[string]map[string]string{
			"info": {
				"name":    "some-package",
				"author":  "John Snow",
				"license": "MIT",
				"version": "1.2.3",
			},
		})

	pypi := pypiProvider{packageName: "some-package"}

	result, err := pypi.LatestVersion()

	assert.NoError(t, err)
	assert.Equal(t, Version("1.2.3"), result)
}

func TestPypiLatestVersion_4xx(t *testing.T) {
	defer gock.Off()
	gock.New("https://pypi.org").
		Get("/pypi/some-package/json").
		Reply(404).
		JSON(map[string]string{"message": "Not Found"})

	pypi := pypiProvider{packageName: "some-package"}

	_, err := pypi.LatestVersion()

	assert.ErrorIs(t, err, ErrVersionNotFound)
}

func TestPypiLatestVersion_5xx(t *testing.T) {
	defer gock.Off()
	gock.New("https://pypi.org").
		Get("/pypi/some-package/json").
		Reply(501).
		JSON(map[string]string{"message": "Not Found"})

	pypi := pypiProvider{packageName: "some-package"}

	_, err := pypi.LatestVersion()

	assert.ErrorIs(t, err, ErrProviderError)
}
