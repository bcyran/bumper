package bumper

import (
	"fmt"
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

// makeFakeCommandRunner creates fake CommandRunner which doesn't use exec.Command.
// Instead it appends each call params to a slice for later assertions.
// Each call returns stdout and err values from given retvals slice.
func makeFakeCommandRunner(retvals *[]commandRunnerRetval) (CommandRunner, *[]commandRunnerParams) {
	var commandRuns []commandRunnerParams
	fakeExecCommand := func(cwd string, command string, args ...string) ([]byte, error) {
		opts := commandRunnerParams{
			cwd:     cwd,
			command: command,
			args:    args,
		}
		commandRuns = append(commandRuns, opts)
		retval := (*retvals)[0]
		*retvals = (*retvals)[1:]
		return retval.stdout, retval.err
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
	commandRetvals := []commandRunnerRetval{
		{stdout: []byte{}, err: nil},                // retval for updpkgsums
		{stdout: []byte(expectedSrcinfo), err: nil}, // retval for makepkg --printsrcinfo
	}
	fakeCommandRunner, commandRuns := makeFakeCommandRunner(&commandRetvals)

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
	expectedUpdpkgsumsCommand := commandRunnerParams{
		cwd: pkg.Path, command: "updpkgsums", args: nil,
	}
	assert.Equal(t, expectedUpdpkgsumsCommand, (*commandRuns)[0])

	// makepkg --printsrcinfo has been ran and result written to .SRCINFO
	expectedMakepkgCommand := commandRunnerParams{
		cwd: pkg.Path, command: "makepkg", args: []string{"--printsrcinfo"},
	}
	assert.Equal(t, expectedMakepkgCommand, (*commandRuns)[1])
	srcinfo, _ := os.ReadFile(pkg.SrcinfoPath())
	assert.Equal(t, expectedSrcinfo, string(srcinfo))
}

func TestBumpAction_FailBump(t *testing.T) {
	// bump should fail because there's no PKGBUILD file
	pkg := makeOutdatedPackage(t.TempDir(), "", "", "")

	fakeCommandRunner, _ := makeFakeCommandRunner(&[]commandRunnerRetval{})
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

	commandRetvals := []commandRunnerRetval{
		{stdout: []byte{}, err: fmt.Errorf("foo bar")}, // retval for updpkgsums
	}
	fakeCommandRunner, _ := makeFakeCommandRunner(&commandRetvals)

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

	commandRetvals := []commandRunnerRetval{
		{stdout: []byte{}, err: nil},                   // retval for updpkgsums
		{stdout: []byte{}, err: fmt.Errorf("foo bar")}, // retval for makepkg
	}
	fakeCommandRunner, _ := makeFakeCommandRunner(&commandRetvals)

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
		}: "bump ✓",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			bumpOk:           false,
		}: "bump ✗",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			bumpOk:           true,
			updpkgsumsOk:     false,
		}: "updpkgsums ✗",
		{
			BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
			bumpOk:           true,
			updpkgsumsOk:     true,
			makepkgOk:        false,
		}: "makepkg ✗",
	}

	for result, expectedString := range cases {
		assert.Equal(t, expectedString, result.String())
	}
}
