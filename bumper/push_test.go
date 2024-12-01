package bumper

import (
	"errors"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/bcyran/bumper/pack"
	"github.com/stretchr/testify/assert"
)

func TestPushAction_Success(t *testing.T) {
	pkg := &pack.Package{
		Path:       "/foo/bar/baz",
		IsOutdated: true,
	}

	// mock return values for commands
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte("master\n"), Err: nil}, // checking branch
		{Stdout: []byte("1\t0\n"), Err: nil},   // checking if up to date with origin
		{Stdout: []byte{}, Err: nil},           // git push
	}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewPushAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// result assertions
	assert.Equal(t, ActionSuccessStatus, result.GetStatus())
	assert.Equal(t, "pushed", result.String())

	// expect valid git branch command
	expectedBranchCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "git", Args: []string{"branch", "--show-current"},
	}
	assert.Equal(t, expectedBranchCommand, (*commandRuns)[0])

	// expect valid git rev-list command
	expectedRevListCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "git", Args: []string{"rev-list", "--left-right", "--count", "master...origin/master"},
	}
	assert.Equal(t, expectedRevListCommand, (*commandRuns)[1])

	// expect valid git push command
	expectedCommitCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "git", Args: []string{"push"},
	}
	assert.Equal(t, expectedCommitCommand, (*commandRuns)[2])
}

func TestPushAction_Skip(t *testing.T) {
	pkg := &pack.Package{
		Path:       "/foo/bar/baz",
		IsOutdated: true,
	}

	// mock return values for commands
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte("master\n"), Err: nil}, // checking branch
		{Stdout: []byte("0\t0\n"), Err: nil},   // checking if up to date with origin
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewPushAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ActionSkippedStatus, result.GetStatus())
	assert.Equal(t, "", result.String())
}

func TestPushAction_FailWrongBranch(t *testing.T) {
	pkg := &pack.Package{
		Path:       "/foo/bar/baz",
		IsOutdated: true,
	}

	// mock return values for commands
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte("other\n"), Err: nil}, // checking branch
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewPushAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "push failed", result.String())
	assert.ErrorContains(t, result.GetError(), "not on master branch")
	assert.ErrorContains(t, result.GetError(), "push action error")
}

func TestPushAction_FailGitError(t *testing.T) {
	pkg := &pack.Package{
		Path:       "/foo/bar/baz",
		IsOutdated: true,
	}

	// mock return values for commands
	const expectedErr = "uh oh, git push failed"
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: errors.New(expectedErr)}, // checking branch
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewPushAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ActionFailedStatus, result.GetStatus())
	assert.Equal(t, "push failed", result.String())
	assert.ErrorContains(t, result.GetError(), expectedErr)
	assert.ErrorContains(t, result.GetError(), "push action error")
}
