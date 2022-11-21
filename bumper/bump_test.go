package bumper

import (
	"os"
	"strings"
	"testing"

	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"github.com/stretchr/testify/assert"
)

type commandRunnerParams struct {
	cwd     string
	command string
	args    []string
}

type commandRunnerRetval struct {
	stdout []byte
	err    error
}

func makeFakeCommandRunner(stdout []byte, err error) (CommandRunner, *[]commandRunnerParams) {
	var commandRuns []commandRunnerParams
	fakeExecCommand := func(cwd string, command string, args ...string) ([]byte, error) {
		opts := commandRunnerParams{
			cwd:     cwd,
			command: command,
			args:    args,
		}
		commandRuns = append(commandRuns, opts)
		return stdout, err
	}
	return fakeExecCommand, &commandRuns
}

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

func TestBumpAction_BumpsPkgverPkgrel(t *testing.T) {
	versionBefore := "1.0.0"
	pkgrelBefore := "2"
	expectedVersion := "2.0.0"
	expectedPkgrel := "1"
	pkg := makeOutdatedPackage(t.TempDir(), versionBefore, pkgrelBefore, expectedVersion)
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString(versionBefore, pkgrelBefore)), 0644)

	fakeCommandRunner, _ := makeFakeCommandRunner([]byte{}, nil)
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_SUCCESS, result.GetStatus())
	assert.True(t, result.bumpOk)
	pkgbuild, _ := os.ReadFile(pkg.PkgbuildPath())
	assert.Equal(t, pkgbuildString(expectedVersion, expectedPkgrel), string(pkgbuild))
}

func TestBumpAction_Updpkgsums(t *testing.T) {
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString("", "")), 0644)

	fakeCommandRunner, commandRuns := makeFakeCommandRunner([]byte{}, nil)
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_SUCCESS, result.GetStatus())
	assert.True(t, result.updpkgsumsOk)
	expectedCommandRun := commandRunnerParams{
		cwd: pkg.Path, command: "updpkgsums", args: nil,
	}
	assert.Equal(t, expectedCommandRun, (*commandRuns)[0])
}

func TestBumpAction_Makepkg(t *testing.T) {
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")
	os.WriteFile(pkg.PkgbuildPath(), []byte(pkgbuildString("", "")), 0644)
	expectedSrcinfo := "expected_srcinfo"

	fakeCommandRunner, commandRuns := makeFakeCommandRunner([]byte(expectedSrcinfo), nil)
	action := NewBumpAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_SUCCESS, result.GetStatus())
	assert.True(t, result.makepkgOk)
	expectedCommandRun := commandRunnerParams{
		cwd: pkg.Path, command: "makepkg", args: []string{"--printsrcinfo"},
	}
	assert.Equal(t, expectedCommandRun, (*commandRuns)[1])
	srcinfo, _ := os.ReadFile(pkg.SrcinfoPath())
	assert.Equal(t, expectedSrcinfo, string(srcinfo))
}
