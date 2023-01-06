package bumper

import (
	"fmt"
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

	// expect result to be correct
	expectedResult := &pushActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
	}
	assert.Equal(t, expectedResult, result)

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

	// expect result to be correct
	expectedResult := &pushActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
	}
	assert.Equal(t, expectedResult, result)
}

func TestPushAction_FailGitError(t *testing.T) {
	pkg := &pack.Package{
		Path:       "/foo/bar/baz",
		IsOutdated: true,
	}

	// mock return values for commands
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte{}, Err: fmt.Errorf("some error")}, // checking branch
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewPushAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// expect result to be correct
	expectedResult := &pushActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
	}
	assert.Equal(t, expectedResult, result)
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

	// expect result to be correct
	expectedResult := &pushActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SKIPPED},
	}
	assert.Equal(t, expectedResult, result)
}
