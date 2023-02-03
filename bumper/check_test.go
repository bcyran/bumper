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

func (provider *fakeVersionProvider) Equal(other interface{}) bool {
	return false
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

	// result assertions
	assert.Equal(t, ACTION_SUCCESS, result.GetStatus())
	assert.Equal(t, "1.0.0 -> 2.0.0", result.String())
	// package assertions
	assert.Equal(t, upstream.Version("2.0.0"), pkg.UpstreamVersion)
	assert.True(t, pkg.IsOutdated)
}

func TestCheckAction_Skip(t *testing.T) {
	verProvFactory := func(url string) upstream.VersionProvider { return nil }
	action := NewCheckAction(verProvFactory)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{Url: "foo"}, IsVCS: true}

	result := action.Execute(&pkg)

	assert.Equal(t, ACTION_SKIPPED, result.GetStatus())
	assert.Equal(t, "-", result.String())
}

func TestCheckAction_FailNoProvider(t *testing.T) {
	verProvFactory := func(url string) upstream.VersionProvider { return nil }
	action := NewCheckAction(verProvFactory)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{Url: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), "no upstream provider found")
}

func TestCheckAction_FailProviderFailed(t *testing.T) {
	expectedErr := "some random error"
	verProvFactory := func(url string) upstream.VersionProvider {
		return &fakeVersionProvider{err: fmt.Errorf(expectedErr)}
	}
	action := NewCheckAction(verProvFactory)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{Url: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
}

func TestCheckAction_FailChecksMultipleUrls(t *testing.T) {
	expectedErr := "some random error"
	checkedUrls := []string{}
	verProvFactory := func(url string) upstream.VersionProvider {
		checkedUrls = append(checkedUrls, url)
		return &fakeVersionProvider{err: fmt.Errorf(expectedErr)}
	}
	action := NewCheckAction(verProvFactory)
	pkg := pack.Package{
		Srcinfo: &pack.Srcinfo{
			Url:    "first.url",
			Source: []string{"second.url", "file.name::third.url"},
		},
	}

	result := action.Execute(&pkg)

	// upstream provider assertions
	assert.ElementsMatch(t, []string{"first.url", "second.url", "third.url"}, checkedUrls)
	// result assertions
	assert.Equal(t, ACTION_FAILED, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
}

func TestCheckActionResult_String(t *testing.T) {
	cases := map[checkActionResult]string{
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
		}: "upstr < curr !",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			currentVersion:   pack.Version("curr"),
			cmpResult:        0,
		}: "curr",
	}

	for result, expectedString := range cases {
		assert.Equal(t, expectedString, result.String())
	}
}
