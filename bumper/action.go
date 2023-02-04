package bumper

import (
	"github.com/bcyran/bumper/pack"
)

type ActionStatus int

const (
	ActionSuccessStatus ActionStatus = iota
	ActionSkippedStatus
	ActionFailedStatus
)

type ActionResult interface {
	GetStatus() ActionStatus
	GetError() error
	String() string
}

type Action interface {
	Execute(pack *pack.Package) ActionResult
}

type BaseActionResult struct {
	Status ActionStatus
	Error  error
}

func (result *BaseActionResult) GetStatus() ActionStatus {
	return result.Status
}

func (result *BaseActionResult) GetError() error {
	return result.Error
}
