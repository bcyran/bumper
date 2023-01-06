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

	// expect result to be correct
	expectedResult := &commitActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS},
	}
	assert.Equal(t, expectedResult, result)

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

	// expect result to be correct
	expectedResult := &commitActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_FAILED},
	}
	assert.Equal(t, expectedResult, result)
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

	// expect result to be correct
	expectedResult := &commitActionResult{
		BaseActionResult: BaseActionResult{Status: ACTION_SKIPPED},
	}
	assert.Equal(t, expectedResult, result)
}

func TestCommitActionResult_String(t *testing.T) {
	cases := map[commitActionResult]string{
		{BaseActionResult: BaseActionResult{Status: ACTION_SUCCESS}}: "committed",
		{BaseActionResult: BaseActionResult{Status: ACTION_FAILED}}:  "commit failed",
		{BaseActionResult: BaseActionResult{Status: ACTION_SKIPPED}}: "",
	}

	for result, expectedString := range cases {
		assert.Equal(t, expectedString, result.String())
	}
}
