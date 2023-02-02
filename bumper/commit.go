package bumper

import (
	"errors"
	"fmt"

	"github.com/bcyran/bumper/pack"
)

const expectedGitStatus = " M .SRCINFO\x00 M PKGBUILD\x00"

var commitError = errors.New("commit action error")

type commitActionResult struct {
	BaseActionResult
}

func (result *commitActionResult) String() string {
	if result.Status == ACTION_SKIPPED {
		return ""
	}
	if result.Status == ACTION_FAILED {
		return "commit failed"
	}
	return "committed"
}

type CommitAction struct {
	commandRunner CommandRunner
}

func NewCommitAction(commandRunner CommandRunner) *CommitAction {
	return &CommitAction{commandRunner: commandRunner}
}

func (action *CommitAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &commitActionResult{}

	if !pkg.IsOutdated {
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	isChanged, err := action.isChanged(pkg)
	if err != nil {
		actionResult.Status = ACTION_FAILED
		actionResult.Error = fmt.Errorf("%w %w", commitError, err)
		return actionResult
	}
	if !isChanged {
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	if err := action.commit(pkg); err != nil {
		actionResult.Status = ACTION_FAILED
		actionResult.Error = fmt.Errorf("%w %w", commitError, err)
		return actionResult
	}

	actionResult.Status = ACTION_SUCCESS
	return actionResult
}

func (action *CommitAction) isChanged(pkg *pack.Package) (bool, error) {
	gitStatus, err := action.commandRunner(pkg.Path, "git", "status", "--porcelain", "--null")
	if err != nil {
		return false, err
	}

	if len(gitStatus) == 0 {
		return false, nil
	}

	if string(gitStatus) != expectedGitStatus {
		return false, errors.New("unexpected changes in the repository")
	}

	return true, nil
}

func (action *CommitAction) commit(pkg *pack.Package) error {
	_, err := action.commandRunner(pkg.Path, "git", "add", "PKGBUILD", ".SRCINFO")
	if err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("Bump version to %s", pkg.UpstreamVersion)
	_, err = action.commandRunner(pkg.Path, "git", "commit", "--message", commitMessage)
	return err
}
