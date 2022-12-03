package bumper

import (
	"fmt"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/bcyran/bumper/pack"
	"github.com/stretchr/testify/assert"
)

func TestBuildAction_Success(t *testing.T) {
	// our Package struct
	pkgPath := "/foo/bar/baz"
	pkg := &pack.Package{
		Path:       pkgPath,
		IsOutdated: true,
	}

	// mock return value for run command
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: nil},
	}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewBuildAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// expect result to be correct
	expectedResult := &buildActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
	}
	assert.Equal(t, expectedResult, result)

	// expected valid makepkg command has been run
	expectedBuildCommand := testutils.CommandRunnerParams{
		Cwd: pkgPath, Command: "makepkg", Args: []string{"--force", "--clean"},
	}
	assert.Equal(t, expectedBuildCommand, (*commandRuns)[0])
}

func TestBuildAction_Fail(t *testing.T) {
	// our Package struct
	pkgPath := "/foo/bar/baz"
	pkg := &pack.Package{
		Path:       pkgPath,
		IsOutdated: true,
	}

	// mock return value for run command
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: fmt.Errorf("foo bar")},
	}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewBuildAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// expect result to be correct
	expectedResult := &buildActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
	}
	assert.Equal(t, expectedResult, result)

	// expected valid makepkg command has been run
	expectedBuildCommand := testutils.CommandRunnerParams{
		Cwd: pkgPath, Command: "makepkg", Args: []string{"--force", "--clean"},
	}
	assert.Equal(t, expectedBuildCommand, (*commandRuns)[0])
}

func TestBuildAction_Skip(t *testing.T) {
	// our Package struct
	pkg := &pack.Package{
		Path:       "/foo/bar/baz",
		IsOutdated: false,
	}

	// mock return value for run command
	commandRetvals := []testutils.CommandRunnerRetval{}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewBuildAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// expect result to be correct
	expectedResult := &buildActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SKIPPED},
	}
	assert.Equal(t, expectedResult, result)

	// expected no command has been run
	assert.Len(t, *commandRuns, 0)
}

func TestBuildActionResult_String(t *testing.T) {
	cases := map[buildActionResult]string{
		{BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS}}: "built",
		{BaseActionResult: BaseActionResult{Status: ACTION_FAILED}}:  "build failed",
		{BaseActionResult: BaseActionResult{Status: ACTION_SKIPPED}}: "",
	}

	for result, expectedString := range cases {
		assert.Equal(t, expectedString, result.String())
	}
}
