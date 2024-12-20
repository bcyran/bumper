package bumper

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
)

var (
	versionOverride        = "4.2.0"
	invalidVersionOverride = "whatever"

	fakeVersionCheckConfigProvider, _ = config.NewYAML(config.Source(strings.NewReader("{empty: {}, check: {providers: {fakeVersionProvider: 2.0.0}}}")))
	emptyCheckConfig                  = fakeVersionCheckConfigProvider.Get("empty")
	fakeVersionCheckConfig            = fakeVersionCheckConfigProvider.Get("check")

	versionOverrideCheckConfigProvider, _ = config.NewYAML(config.Source(strings.NewReader(fmt.Sprintf("{check: {versionOverrides: {foopkg: %s}}}", versionOverride))))
	versionOverrideCheckConfig            = versionOverrideCheckConfigProvider.Get("check")

	invalidOverrideCheckConfigProvider, _ = config.NewYAML(config.Source(strings.NewReader(fmt.Sprintf("{check: {versionOverrides: {foopkg: %s}}}", invalidVersionOverride))))
	invalidVersionOverrideCheckConfig     = invalidOverrideCheckConfigProvider.Get("check")
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

func (provider *fakeVersionProvider) Equal(_ interface{}) bool {
	return false
}

func TestCheckAction_Success(t *testing.T) {
	verProvFactory := func(_url string, providersConfig config.Value) upstream.VersionProvider {
		return &fakeVersionProvider{version: providersConfig.Get("fakeVersionProvider").String()}
	}
	action := NewCheckAction(verProvFactory, fakeVersionCheckConfig)
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

func TestCheckAction_SuccessVersionOverride(t *testing.T) {
	verProvFactory := func(_url string, _providersConfig config.Value) upstream.VersionProvider {
		t.Error("provider should not be called when version override provided")
		return nil
	}
	action := NewCheckAction(verProvFactory, versionOverrideCheckConfig)
	pkg := pack.Package{
		Srcinfo: &pack.Srcinfo{
			Pkgbase: "foopkg",
			URL:     "foo",
			FullVersion: &pack.FullVersion{
				Pkgver: pack.Version("1.0.0"),
			},
		},
	}

	result := action.Execute(&pkg)

	// result assertions
	assert.Equal(t, ActionSuccessStatus, result.GetStatus())
	assert.Equal(t, fmt.Sprintf("1.0.0 → %s", versionOverride), result.String())
	// package assertions
	assert.Equal(t, upstream.Version(versionOverride), pkg.UpstreamVersion)
	assert.True(t, pkg.IsOutdated)
}

func TestCheckAction_Skip(t *testing.T) {
	verProvFactory := func(_url string, _providersConfig config.Value) upstream.VersionProvider { return nil }
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
	verProvFactory := func(_url string, _providersConfig config.Value) upstream.VersionProvider { return nil }
	action := NewCheckAction(verProvFactory, emptyCheckConfig)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{URL: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), "no upstream provider found")
}

func TestCheckAction_FailProviderFailed(t *testing.T) {
	const expectedErr = "some random error"
	verProvFactory := func(_url string, _providersConfig config.Value) upstream.VersionProvider {
		return &fakeVersionProvider{err: errors.New(expectedErr)}
	}
	action := NewCheckAction(verProvFactory, emptyCheckConfig)
	pkg := pack.Package{Srcinfo: &pack.Srcinfo{URL: "foo"}}

	result := action.Execute(&pkg)

	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
}

func TestCheckAction_FailChecksMultipleURLs(t *testing.T) {
	const expectedErr = "some random error"
	checkedURLs := []string{}
	verProvFactory := func(url string, _providersConfig config.Value) upstream.VersionProvider {
		checkedURLs = append(checkedURLs, url)
		return &fakeVersionProvider{err: errors.New(expectedErr)}
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

func TestCheckAction_FailInvalidVersionOverride(t *testing.T) {
	verProvFactory := func(_url string, _providersConfig config.Value) upstream.VersionProvider {
		t.Error("provider should not be called when version override provided")
		return nil
	}
	action := NewCheckAction(verProvFactory, invalidVersionOverrideCheckConfig)
	pkg := pack.Package{
		Srcinfo: &pack.Srcinfo{
			Pkgbase: "foopkg",
			URL:     "foo",
			FullVersion: &pack.FullVersion{
				Pkgver: pack.Version("1.0.0"),
			},
		},
	}

	result := action.Execute(&pkg)

	// result assertions
	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "?", result.String())
	assert.ErrorContains(t, result.GetError(), fmt.Sprintf("version override '%s' is not a valid version", invalidVersionOverride))
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
