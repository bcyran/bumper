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
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString(versionBefore, pkgrelBefore)), 0o644)

	// mock return values for two command runs
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: nil},                // retval for updpkgsums
		{Stdout: []byte(expectedSrcinfo), Err: nil}, // retval for makepkg --printsrcinfo
	}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// result assertions
	assert.Equal(t, ACTION_SUCCESS, result.GetStatus())
	assert.Equal(t, "bumped", result.String())

	// PKGBUILD assertions
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

func TestBumpAction_Skip(t *testing.T) {
	// bump should be skipped because package is not outdated
	pkg := &pack.Package{IsOutdated: false}

	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&[]testutils.CommandRunnerRetval{})
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// result assertions
	assert.Equal(t, ACTION_SKIPPED, result.GetStatus())
	assert.Equal(t, "", result.String())
	// command assertions
	assert.Len(t, *commandRuns, 0) // no commands ran
}

func TestBumpAction_FailBump(t *testing.T) {
	// bump should fail because there's no PKGBUILD file
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")

	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&[]testutils.CommandRunnerRetval{})
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
	assert.Equal(t, "bump failed", result.String())
	assert.ErrorContains(t, result.GetError(), "PKGBUILD: no such file or directory")
	assert.ErrorContains(t, result.GetError(), "bump action error")
}

func TestBumpAction_FailUpdpkgsums(t *testing.T) {
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString("", "")), 0o644)

	expectedErr := "omg, updpkgsums failed"
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: fmt.Errorf(expectedErr)}, // retval for updpkgsums
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
	assert.Equal(t, "updpkgsums failed", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
	assert.ErrorContains(t, result.GetError(), "bump action error")
}

func TestBumpAction_FailMakepkg(t *testing.T) {
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString("", "")), 0o644)

	expectedErr := "oh no, poor makepkg error"
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: nil},                     // retval for updpkgsums
		{Stdout: []byte{}, Err: fmt.Errorf(expectedErr)}, // retval for makepkg
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
	assert.Equal(t, "makepkg failed", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
	assert.ErrorContains(t, result.GetError(), "bump action error")
}
