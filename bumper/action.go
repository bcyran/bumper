package bumper

import (
	"github.com/bcyran/bumper/pack"
)

type ActionStatus int

const (
	ACTION_SUCCESS ActionStatus = iota
	ACTION_SKIPPED
	ACTION_FAILED
)

type ActionResult interface {
	GetStatus() ActionStatus
	String() string
}

type Action interface {
	Execute(pack *pack.Package) ActionResult
}

type BaseActionResult struct {
	Status ActionStatus
}

func (result BaseActionResult) GetStatus() ActionStatus {
	return result.Status
}
