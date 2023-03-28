package bumper

import (
	"errors"
	"fmt"

	"github.com/bcyran/bumper/pack"
	"go.uber.org/config"
)

const expectedGitStatus = " M .SRCINFO\x00 M PKGBUILD\x00"

var ErrCommitAction = errors.New("commit action error")

type commitActionResult struct {
	BaseActionResult
}

func (result *commitActionResult) String() string {
	if result.Status == ActionSkippedStatus {
		return ""
	}
	if result.Status == ActionFailedStatus {
		return "commit failed"
	}
	return "committed"
}

type CommitAction struct {
	commandRunner CommandRunner
	commitConfig  config.Value
}

func NewCommitAction(commandRunner CommandRunner, commitConfig config.Value) *CommitAction {
	return &CommitAction{commandRunner: commandRunner, commitConfig: commitConfig}
}

func (action *CommitAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &commitActionResult{}

	if !pkg.IsOutdated {
		actionResult.Status = ActionSkippedStatus
		return actionResult
	}

	isChanged, err := action.isChanged(pkg)
	if err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w %w", ErrCommitAction, err)
		return actionResult
	}
	if !isChanged {
		actionResult.Status = ActionSkippedStatus
		return actionResult
	}

	if err := action.commit(pkg); err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w %w", ErrCommitAction, err)
		return actionResult
	}

	actionResult.Status = ActionSuccessStatus
	return actionResult
}

func (action *CommitAction) isChanged(pkg *pack.Package) (bool, error) {
	gitStatus, err := action.commandRunner(pkg.Path, "git", "status", "--porcelain", "--null", "--untracked-files=no")
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
	commitArgs := []string{"commit", "--message", commitMessage}

	var commitAuthor string
	action.commitConfig.Get("author").Populate(&commitAuthor) // nolint:errcheck
	if commitAuthor != "" {
		commitArgs = append(commitArgs, "--author", commitAuthor)
	}

	_, err = action.commandRunner(pkg.Path, "git", commitArgs...)
	return err
}
