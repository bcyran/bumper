package bumper

import (
	"errors"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/bcyran/bumper/pack"
	"github.com/stretchr/testify/assert"
)

func TestMakeAction_Success(t *testing.T) {
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
	action := NewMakeAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// result assertions
	assert.Equal(t, ActionSuccessStatus, result.GetStatus())
	assert.Equal(t, "built", result.String())
	// command assertions
	expectedBuildCommand := testutils.CommandRunnerParams{
		Cwd: pkgPath, Command: "makepkg", Args: []string{"--force", "--clean"},
	}
	assert.Equal(t, expectedBuildCommand, (*commandRuns)[0])
}

func TestMakeAction_Skip(t *testing.T) {
	// our Package struct
	pkg := &pack.Package{
		Path:       "/foo/bar/baz",
		IsOutdated: false,
	}

	// mock return value for run command
	commandRetvals := []testutils.CommandRunnerRetval{}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewMakeAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// result assertions
	assert.Equal(t, ActionSkippedStatus, result.GetStatus())
	assert.Equal(t, "", result.String())
	// command assertions
	assert.Len(t, *commandRuns, 0) // no commands ran
}

func TestMakeAction_Fail(t *testing.T) {
	// our Package struct
	pkgPath := "/foo/bar/baz"
	pkg := &pack.Package{
		Path:       pkgPath,
		IsOutdated: true,
	}

	// mock return value for run command
	const expectedErr = "oooh makepkg crashed"
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: errors.New(expectedErr)},
	}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewMakeAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// result assertions
	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "build failed", result.String())
	// command assertions
	expectedBuildCommand := testutils.CommandRunnerParams{
		Cwd: pkgPath, Command: "makepkg", Args: []string{"--force", "--clean"},
	}
	assert.Equal(t, expectedBuildCommand, (*commandRuns)[0])
	assert.ErrorContains(t, result.GetError(), expectedErr)
	assert.ErrorContains(t, result.GetError(), "make action error")
}
