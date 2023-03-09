package bumper

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
)

var (
	checkConfigProvider, _ = config.NewYAML(config.Source(strings.NewReader("{empty: {}, check: {providers: {version: 2.0.0}}}")))
	emptyCheckConfig       = checkConfigProvider.Get("empty")
	checkConfigWithVersion = checkConfigProvider.Get("check")
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
	verProvFactory := func(url string, providersConfig config.Value) upstream.VersionProvider {
		return &fakeVersionProvider{version: providersConfig.Get("version").String()}
	}
	action := NewCheckAction(verProvFactory, checkConfigWithVersion)
	pkg := pack.Package{
		Srcinfo: &pack.Srcinfo{
			URL: "foo",
			FullVersion: &pack.FullVersion{
				Pkgver: pack.Version("1.0.0"),
			},
		},
	}

	result := action.Execute(&pkg)

	// result assertions
	assert.Equal(t, ActionSuccessStatus, result.GetStatus())
	assert.Equal(t, "1.0.0 → 2.0.0", result.String())
	// package assertions
	assert.Equal(t, upstream.Version("2.0.0"), pkg.UpstreamVersion)
	assert.True(t, pkg.IsOutdated)
}

func TestCheckAction_Skip(t *testing.T) {
	verProvFactory := func(url string, providersConfig config.Value) upstream.VersionProvider { return nil }
	action := NewCheckAction(verProvFactory, emptyCheckConfig)
	pkg := pack.Package{
		Srcinfo: &pack.Srcinfo{
			URL: "foo", FullVersion: &pack.FullVersion{
				Pkgver: pack.Version("1.2.r3.45"),
			},
		},
		IsVCS: true,
	}

	result := action.Execute(&pkg)

	assert.Equal(t, ActionSkippedStatus, result.GetStatus())
	assert.Equal(t, "1.2.r3.45", result.String())
}

func TestCheckAction_FailNoProvider(t *testing.T) {
	verProvFactory := func(url string, providersConfig config.Value) upstream.VersionProvider { return nil }
	action := NewCheckAction(verProvFactory, emptyCheckConfig)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{URL: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), "no upstream provider found")
}

func TestCheckAction_FailProviderFailed(t *testing.T) {
	expectedErr := "some random error"
	verProvFactory := func(url string, providersConfig config.Value) upstream.VersionProvider {
		return &fakeVersionProvider{err: fmt.Errorf(expectedErr)}
	}
	action := NewCheckAction(verProvFactory, emptyCheckConfig)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{URL: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
}

func TestCheckAction_FailChecksMultipleURLs(t *testing.T) {
	expectedErr := "some random error"
	checkedURLs := []string{}
	verProvFactory := func(url string, providersConfig config.Value) upstream.VersionProvider {
		checkedURLs = append(checkedURLs, url)
		return &fakeVersionProvider{err: fmt.Errorf(expectedErr)}
	}
	action := NewCheckAction(verProvFactory, emptyCheckConfig)
	pkg := pack.Package{
		Srcinfo: &pack.Srcinfo{
			URL:    "first.url",
			Source: []string{"second.url", "file.name::third.url"},
		},
	}

	result := action.Execute(&pkg)

	// upstream provider assertions
	assert.ElementsMatch(t, []string{"first.url", "second.url", "third.url"}, checkedURLs)
	// result assertions
	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
}

func TestCheckActionResult_String(t *testing.T) {
	cases := map[checkActionResult]string{
		{
			BaseActionResult: BaseActionResult{Status: ActionSuccessStatus},
			currentVersion:   pack.Version("curr"),
			upstreamVersion:  upstream.Version("upstr"),
			cmpResult:        1,
		}: "curr → upstr",
		{
			BaseActionResult: BaseActionResult{Status: ActionSuccessStatus},
			currentVersion:   pack.Version("curr"),
			upstreamVersion:  upstream.Version("upstr"),
			cmpResult:        -1,
		}: "upstr < curr !",
		{
			BaseActionResult: BaseActionResult{Status: ActionSuccessStatus},
			currentVersion:   pack.Version("curr"),
			cmpResult:        0,
		}: "curr",
	}

	for result, expectedString := range cases {
		assert.Equal(t, expectedString, result.String())
	}
}
