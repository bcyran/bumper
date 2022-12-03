package bumper

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"github.com/stretchr/testify/assert"
)

func makeOutdatedPackage(dir string, pkgver string, pkgrel string, upstreamVersion string) *pack.Package {
	return &pack.Package{
		Path: dir,
		Srcinfo: &pack.Srcinfo{
			FullVersion: &pack.FullVersion{
				Pkgver: pack.Version(pkgver),
				Pkgrel: pkgrel,
			},
		},
		UpstreamVersion: upstream.Version(upstreamVersion),
		IsOutdated:      true,
	}
}

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

func TestBumpAction_Success(t *testing.T) {
	versionBefore := "1.0.0"
	pkgrelBefore := "2"
	expectedVersion := "2.0.0"
	expectedPkgrel := "1"
	expectedSrcinfo := "expected_srcinfo"
	expectedPkgbuild := pkgbuildString(expectedVersion, expectedPkgrel)

	// build our Package struct and write PKGBUILD
	pkg := makeOutdatedPackage(t.TempDir(), versionBefore, pkgrelBefore, expectedVersion)
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString(versionBefore, pkgrelBefore)), 0644)

	// mock return values for two command runs
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: nil},                // retval for updpkgsums
		{Stdout: []byte(expectedSrcinfo), Err: nil}, // retval for makepkg --printsrcinfo
	}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// returned result is correct
	expectedResult := &bumpActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
		bumpOk:           true,
		updpkgsumsOk:     true,
		makepkgOk:        true,
	}
	assert.Equal(t, expectedResult, result)

	// pkgver and pkgrel are updated in pkgbuild
	pkgbuild, _ := os.ReadFile(pkg.PkgbuildPath())
	assert.Equal(t, expectedPkgbuild, string(pkgbuild))

	// updpkgsums command has been ran
	expectedUpdpkgsumsCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "updpkgsums", Args: nil,
	}
	assert.Equal(t, expectedUpdpkgsumsCommand, (*commandRuns)[0])

	// makepkg --printsrcinfo has been ran and result written to .SRCINFO
	expectedMakepkgCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "makepkg", Args: []string{"--printsrcinfo"},
	}
	assert.Equal(t, expectedMakepkgCommand, (*commandRuns)[1])
	srcinfo, _ := os.ReadFile(pkg.SrcinfoPath())
	assert.Equal(t, expectedSrcinfo, string(srcinfo))
}

func TestBumpAction_FailBump(t *testing.T) {
	// bump should fail because there's no PKGBUILD file
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")

	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&[]testutils.CommandRunnerRetval{})
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	expectedResult := &bumpActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
		bumpOk:           false,
	}
	assert.Equal(t, expectedResult, result)
}

func TestBumpAction_FailUpdpkgsums(t *testing.T) {
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString("", "")), 0644)

	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: fmt.Errorf("foo bar")}, // retval for updpkgsums
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	expectedResult := &bumpActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
		bumpOk:           true,
		updpkgsumsOk:     false,
	}
	assert.Equal(t, expectedResult, result)
}

func TestBumpAction_FailMakepkg(t *testing.T) {
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString("", "")), 0644)

	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: nil},                   // retval for updpkgsums
		{Stdout: []byte{}, Err: fmt.Errorf("foo bar")}, // retval for makepkg
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	expectedResult := &bumpActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
		bumpOk:           true,
		updpkgsumsOk:     true,
		makepkgOk:        false,
	}
	assert.Equal(t, expectedResult, result)
}

func TestBumpActionResult_String(t *testing.T) {
	cases := map[bumpActionResult]string{
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
		}: "bumped",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			bumpOk:           false,
		}: "bump failed",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			bumpOk:           true,
			updpkgsumsOk:     false,
		}: "updpkgsums failed",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			bumpOk:           true,
			updpkgsumsOk:     true,
			makepkgOk:        false,
		}: "makepkg failed",
	}

	for result, expectedString := range cases {
		assert.Equal(t, expectedString, result.String())
	}
}
