package bumper

import (
	"github.com/bcyran/bumper/pack"
)

type buildActionResult struct {
	BaseActionResult
}

func (result *buildActionResult) String() string {
	if result.Status == ACTION_SKIPPED {
		return ""
	}
	if result.Status == ACTION_FAILED {
		return "build failed"
	}
	return "built"
}

type BuildAction struct {
	commandRunner CommandRunner
}

func NewBuildAction(commandRunner CommandRunner) *BuildAction {
	return &BuildAction{commandRunner: commandRunner}
}

func (action *BuildAction) Execute(pkg *pack.Package) ActionResult {
	actionResult := &buildActionResult{}

	if !pkg.IsOutdated {
		actionResult.Status = ACTION_SKIPPED
		return actionResult
	}

	_, err := action.commandRunner(pkg.Path, "makepkg", "--force", "--clean")
	if err != nil {
		actionResult.Status = ACTION_FAILED
	} else {
		actionResult.Status = ACTION_SUCCESS
	}

	return actionResult
}
