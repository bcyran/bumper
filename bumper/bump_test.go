package bumper

import (
	"os"
	"strings"
	"testing"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"github.com/stretchr/testify/assert"
)

func pkgbuildString(pkgver string, pkgrel string) string {
	pkgBuildTemplate := `
pkgname=foo
pkgver={pkgver}
pkgrel={pkgrel}
url='https://foo.bar/{version}/baz'
`
	pkgbuild := strings.ReplaceAll(pkgBuildTemplate, "{pkgver}", pkgver)
	return strings.ReplaceAll(pkgbuild, "{pkgrel}", pkgrel)
}

func TestBumpAction_BumpsPkgverPkgrel(t *testing.T) {
	versionBefore := "1.0.0"
	pkgrelBefore := "2"
	expectedVersion := "2.0.0"
	expectedPkgrel := "1"
	pkg := pack.Package{
		Path: t.TempDir(),
		Srcinfo: &pack.Srcinfo{
			FullVersion: &pack.FullVersion{
				Pkgver: pack.Version(versionBefore),
				Pkgrel: pkgrelBefore,
			},
		},
		UpstreamVersion: upstream.Version(expectedVersion),
		IsOutdated:      true,
	}
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString(versionBefore, pkgrelBefore)), 0644)

	action := BumpAction{}
	result := action.Execute(&pkg)

	assert.Equal(t, ACTION_SUCCESS, result.GetStatus())
	assert.True(t, result.bumpOk)
	pkgbuild, _ := os.ReadFile(pkg.PkgbuildPath())
	assert.Equal(t, pkgbuildString(expectedVersion, expectedPkgrel), string(pkgbuild))
}
