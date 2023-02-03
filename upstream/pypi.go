package upstream

import (
	"fmt"
	"regexp"
)

var (
	pypiPackageRegex = regexp.MustCompile(`(files\.pythonhosted\.org|pypi\.python.org|pypi\.org|pypi\.io)/packages/source/[a-z]{1}/([^/#?]+)/`)
)

type pypiProvider struct {
	packageName string
}

type pypiPackageResp struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

func newPypiProvider(url string) *pypiProvider {
	match := pypiPackageRegex.FindStringSubmatch(url)
	if len(match) == 0 {
		return nil
	}
	return &pypiProvider{packageName: match[2]}
}

func (pypi *pypiProvider) packageInfoURL() string {
	return fmt.Sprintf("https://pypi.org/pypi/%s/json", pypi.packageName)
}

func (pypi *pypiProvider) Equal(other interface{}) bool {
	switch other := other.(type) {
	case *pypiProvider:
		return pypi.packageName == other.packageName
	default:
		return false
	}
}

func (pypi *pypiProvider) LatestVersion() (Version, error) {
	var packageInfo pypiPackageResp
	if err := httpGetJSON(pypi.packageInfoURL(), &packageInfo, nil); err != nil {
		return "", err
	}
	if version, isValid := parseVersion(packageInfo.Info.Version); isValid == true {
		return version, nil
	}
	return "", ErrVersionNotFound
}
