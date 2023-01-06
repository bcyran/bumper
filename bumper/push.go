package bumper

import (
	"fmt"
	"strings"

	"github.com/bcyran/bumper/pack"
)

const masterBranch = "master"

var diffTarget = fmt.Sprintf("%s...origin/%s", masterBranch, masterBranch)

type pushActionResult struct {
	BaseActionResult
}

func (result *pushActionResult) String() string {
	if result.Status == ACTION_SKIPPED {
		return ""
	}
	if result.Status == ACTION_FAILED {
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
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	isOnMaster, err := action.isOnMaster(pkg)
	if err != nil || isOnMaster == false {
		actionResult.Status = ACTION_FAILED
		return actionResult
	}

	isBehindOrigin, err := action.isBehindOrigin(pkg)
	if err != nil {
		actionResult.Status = ACTION_FAILED
		return actionResult
	}
	if isBehindOrigin == false {
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	if result := action.push(pkg); result == false {
		actionResult.Status = ACTION_FAILED
		return actionResult
	}

	actionResult.Status = ACTION_SUCCESS
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

func (action *PushAction) push(pkg *pack.Package) bool {
	_, err := action.commandRunner(pkg.Path, "git", "push")
	if err != nil {
		return false
	}

	return true
}