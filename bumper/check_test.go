package bumper

import (
	"fmt"
	"testing"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"github.com/stretchr/testify/assert"
)

type fakeVersionProvider struct {
	version string
	err     error
}

func (provider *fakeVersionProvider) LatestVersion() (upstream.Version, error) {
	if provider.err != nil {
		return upstream.Version(""), provider.err
	}
	return upstream.Version(provider.version), nil
}

func TestCheckAction_Success(t *testing.T) {
	verProvFactory := func(url string) upstream.VersionProvider {
		return &fakeVersionProvider{version: "2.0.0"}
	}
	action := NewCheckAction(verProvFactory)
	pkg := pack.Package{
		Srcinfo: &pack.Srcinfo{
			Url: "foo",
			FullVersion: &pack.FullVersion{
				Pkgver: pack.Version("1.0.0"),
			},
		},
	}

	result := action.Execute(&pkg)

	expectedUpstreamVersion := upstream.Version("2.0.0")
	expectedResult := &checkActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
		currentVersion:   pack.Version("1.0.0"),
		upstreamVersion:  expectedUpstreamVersion,
		cmpResult:        1,
	}
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, expectedUpstreamVersion, pkg.UpstreamVersion)
	assert.True(t, pkg.IsOutdated)
}

func TestCheckAction_FailNoProvider(t *testing.T) {
	verProvFactory := func(url string) upstream.VersionProvider { return nil }
	action := NewCheckAction(verProvFactory)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{Url: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
}

func TestCheckAction_FailProviderFailed(t *testing.T) {
	verProvFactory := func(url string) upstream.VersionProvider {
		return &fakeVersionProvider{err: fmt.Errorf("err")}
	}
	action := NewCheckAction(verProvFactory)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{Url: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
}

func TestCheckActionResult_String(t *testing.T) {
	cases := map[checkActionResult]string{
		{
			BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
		}: "x",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			currentVersion:   pack.Version("curr"),
			upstreamVersion:  upstream.Version("upstr"),
			cmpResult:        1,
		}: "curr -> upstr",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			currentVersion:   pack.Version("curr"),
			upstreamVersion:  upstream.Version("upstr"),
			cmpResult:        -1,
		}: "curr < upstr !",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			currentVersion:   pack.Version("curr"),
			cmpResult:        0,
		}: "curr ✓",
	}

	for result, expectedString := range cases {
		assert.Equal(t, expectedString, result.String())
	}
}
