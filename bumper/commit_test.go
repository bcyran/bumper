package bumper

import (
	"fmt"
	"testing"

	"github.com/bcyran/bumper/internal/testutils"
	"github.com/bcyran/bumper/pack"
	"github.com/bcyran/bumper/upstream"
	"github.com/stretchr/testify/assert"
)

func TestCommitAction_Success(t *testing.T) {
	// our Package struct
	pkg := &pack.Package{
		Path:            "/foo/bar/baz",
		UpstreamVersion: upstream.Version("1.2.3"),
		IsOutdated:      true,
	}

	// mock return values for commands
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte(" M .SRCINFO\x00 M PKGBUILD\x00"), Err: nil}, // git status
		{Stdout: []byte{}, Err: nil},                                 // git add
		{Stdout: []byte{}, Err: nil},                                 // git commit
	}
	fakeCommandRunner, commandRuns := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewCommitAction(fakeCommandRunner)
	result := action.Execute(pkg)

	// result assertions
	assert.Equal(t, ACTION_SUCCESS, result.GetStatus())
	assert.Equal(t, "committed", result.String())

	// expect valid git status command
	expectedStatusCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "git", Args: []string{"status", "--porcelain", "--null"},
	}
	assert.Equal(t, expectedStatusCommand, (*commandRuns)[0])

	// expect valid git add command
	expectedAddCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "git", Args: []string{"add", "PKGBUILD", ".SRCINFO"},
	}
	assert.Equal(t, expectedAddCommand, (*commandRuns)[1])

	// expect valid git commit command
	expectedCommitMessage := fmt.Sprintf("Bump version to %s", pkg.UpstreamVersion)
	expectedCommitCommand := testutils.CommandRunnerParams{
		Cwd: pkg.Path, Command: "git", Args: []string{"commit", "--message", expectedCommitMessage},
	}
	assert.Equal(t, expectedCommitCommand, (*commandRuns)[2])
}

func TestCommitAction_Skip(t *testing.T) {
	// our Package struct
	pkg := &pack.Package{
		Path:            "/foo/bar/baz",
		UpstreamVersion: upstream.Version("1.2.3"),
		IsOutdated:      true,
	}

	// mock return values for commands
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte(""), Err: nil}, // git status reports no changes
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewCommitAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_SKIPPED, result.GetStatus())
	assert.Equal(t, "", result.String())
}

func TestCommitAction_Fail(t *testing.T) {
	// our Package struct
	pkg := &pack.Package{
		Path:            "/foo/bar/baz",
		UpstreamVersion: upstream.Version("1.2.3"),
		IsOutdated:      true,
	}

	// mock return values for commands
	commandRetvals := []testutils.CommandRunnerRetval{
		{Stdout: []byte(" M .SRCINFO\x00 M PKGBUILD\x00 A foo.txt\x00"), Err: nil}, // unexpected git status
	}
	fakeCommandRunner, _ := testutils.MakeFakeCommandRunner(&commandRetvals)

	// execute the action with our mocked command runner
	action := NewCommitAction(fakeCommandRunner)
	result := action.Execute(pkg)

	assert.Equal(t, ACTION_FAILED, result.GetStatus())
	assert.Equal(t, "commit failed", result.String())
}
