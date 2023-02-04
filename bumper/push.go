package bumper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bcyran/bumper/pack"
)

const masterBranch = "master"

var (
	diffTarget    = fmt.Sprintf("%s...origin/%s", masterBranch, masterBranch)
	ErrPushAction = errors.New("push action error")
)

type pushActionResult struct {
	BaseActionResult
}

func (result *pushActionResult) String() string {
	if result.Status == ActionSkippedStatus {
		return ""
	}
	if result.Status == ActionFailedStatus {
		return "push failed"
	}
	return "pushed"
}

type PushAction struct {
	commandRunner CommandRunner
}

func NewPushAction(commandRunner CommandRunner) *PushAction {
	return &PushAction{commandRunner: commandRunner}
}

func (action *PushAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &pushActionResult{}

	if !pkg.IsOutdated {
		actionResult.Status = ActionSkippedStatus
		return actionResult
	}

	isOnMaster, err := action.isOnMaster(pkg)
	if err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: %w", ErrPushAction, err)
		return actionResult
	}
	if !isOnMaster {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: not on master branch", ErrPushAction)
		return actionResult
	}

	isBehindOrigin, err := action.isBehindOrigin(pkg)
	if err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: %w", ErrPushAction, err)
		return actionResult
	}
	if !isBehindOrigin {
		actionResult.Status = ActionSkippedStatus
		return actionResult
	}

	if err := action.push(pkg); err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: %w", ErrPushAction, err)
		return actionResult
	}

	actionResult.Status = ActionSuccessStatus
	return actionResult
}

func (action *PushAction) isOnMaster(pkg *pack.Package) (bool, error) {
	currentBranch, err := action.commandRunner(pkg.Path, "git", "branch", "--show-current")
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(string(currentBranch)) != masterBranch {
		return false, nil
	}

	return true, nil
}

func (action *PushAction) isBehindOrigin(pkg *pack.Package) (bool, error) {
	gitRevList, err := action.commandRunner(pkg.Path, "git", "rev-list", "--left-right", "--count", diffTarget)
	if err != nil {
		return false, err
	}

	counts := strings.Fields(string(gitRevList))
	if counts[0] == "0" {
		return false, nil
	}

	return true, nil
}

func (action *PushAction) push(pkg *pack.Package) error {
	_, err := action.commandRunner(pkg.Path, "git", "push")
	return err
}
