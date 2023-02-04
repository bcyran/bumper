package bumper

import (
	"errors"
	"fmt"

	"github.com/bcyran/bumper/pack"
)

var ErrMakeAction = errors.New("make action error")

type makeActionResult struct {
	BaseActionResult
}

func (result *makeActionResult) String() string {
	if result.Status == ActionSkippedStatus {
		return ""
	}
	if result.Status == ActionFailedStatus {
		return "build failed"
	}
	return "built"
}

type MakeAction struct {
	commandRunner CommandRunner
}

func NewMakeAction(commandRunner CommandRunner) *MakeAction {
	return &MakeAction{commandRunner: commandRunner}
}

func (action *MakeAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &makeActionResult{}

	if !pkg.IsOutdated {
		actionResult.Status = ActionSkippedStatus
		return actionResult
	}

	_, err := action.commandRunner(pkg.Path, "makepkg", "--force", "--clean")
	if err != nil {
		actionResult.Status = ActionFailedStatus
		actionResult.Error = fmt.Errorf("%w: %w", ErrMakeAction, err)
	} else {
		actionResult.Status = ActionSuccessStatus
	}

	return actionResult
}
