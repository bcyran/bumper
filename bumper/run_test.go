package bumper

import (
	"fmt"
	"testing"

	"github.com/bcyran/bumper/pack"
	"github.com/stretchr/testify/assert"
)

type testActionResult struct {
	BaseActionResult
	retString string
}

func newTestActionResult(retStatus ActionStatus, retString string) *testActionResult {
	return &testActionResult{
		BaseActionResult: BaseActionResult{Status: retStatus},
		retString:        retString,
	}
}

func (result *testActionResult) String() string {
	return result.retString
}

type testAction struct {
	retStatus ActionStatus
	retString string
}

func newTestAction(retStatus ActionStatus, retString string) *testAction {
	return &testAction{retStatus: retStatus, retString: retString}
}

func (action *testAction) Execute(pkg *pack.Package) ActionResult {
	return &testActionResult{
		BaseActionResult: BaseActionResult{Status: action.retStatus},
		retString:        fmt.Sprintf("%s: %s", pkg.Pkgbase, action.retString),
	}
}

func TestRun_Success(t *testing.T) {
	// packages and actions we will run on them
	packages := []pack.Package{
		{Srcinfo: &pack.Srcinfo{Pkgbase: "pkgA"}},
		{Srcinfo: &pack.Srcinfo{Pkgbase: "pkgB"}},
	}
	actions := []Action{
		newTestAction(ActionSuccessStatus, "first result"),
		newTestAction(ActionFailedStatus, "second result"),
		newTestAction(ActionSuccessStatus, "this shouldn't be executed"),
	}

	// expected values
	expectedResults := [][]ActionResult{
		{
			newTestActionResult(ActionSuccessStatus, "pkgA: first result"),
			newTestActionResult(ActionFailedStatus, "pkgA: second result"),
		},
		{
			newTestActionResult(ActionSuccessStatus, "pkgB: first result"),
			newTestActionResult(ActionFailedStatus, "pkgB: second result"),
		},
	}
	expectedFinished := []bool{true, true}

	// actual values
	actualResults := make([][]ActionResult, len(packages))
	for i := range actualResults {
		actualResults[i] = []ActionResult{}
	}
	actualFinished := make([]bool, len(packages))

	// callbacks definition and running SUT
	handleResult := func(pkgIndex int, result ActionResult) {
		actualResults[pkgIndex] = append(actualResults[pkgIndex], result)
	}
	handleFinished := func(pkgIndex int) {
		actualFinished[pkgIndex] = true
	}
	Run(packages, actions, handleResult, handleFinished)

	// check if everything matches
	assert.ElementsMatch(t, actualResults, expectedResults)
	assert.ElementsMatch(t, actualFinished, expectedFinished)
}

func TestRun_Skip(t *testing.T) {
	// packages and actions we will run on them
	packages := []pack.Package{
		{Srcinfo: &pack.Srcinfo{Pkgbase: "pkgA"}},
	}
	actions := []Action{
		newTestAction(ActionSkippedStatus, "skipped, should stop now"),
		newTestAction(ActionSuccessStatus, "this shouldn't be executed"),
	}

	// expected values
	expectedResults := [][]ActionResult{
		{
			newTestActionResult(ActionSkippedStatus, "pkgA: skipped, should stop now"),
		},
	}
	expectedFinished := []bool{true}

	// actual values
	actualResults := make([][]ActionResult, len(packages))
	for i := range actualResults {
		actualResults[i] = []ActionResult{}
	}
	actualFinished := make([]bool, len(packages))

	// callbacks definition and running SUT
	handleResult := func(pkgIndex int, result ActionResult) {
		actualResults[pkgIndex] = append(actualResults[pkgIndex], result)
	}
	handleFinished := func(pkgIndex int) {
		actualFinished[pkgIndex] = true
	}
	Run(packages, actions, handleResult, handleFinished)

	// check if everything matches
	assert.ElementsMatch(t, actualResults, expectedResults)
	assert.ElementsMatch(t, actualFinished, expectedFinished)
}
