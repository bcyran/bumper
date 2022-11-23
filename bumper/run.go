package bumper

import (
	"reflect"

	"github.com/bcyran/bumper/pack"
)

type ResultHandler func(pkgIndex int, result ActionResult, isFinal bool)

func Run(pkgs []pack.Package, actions []Action, resultHandler ResultHandler) {
	packageChans := make([]chan ActionResult, len(pkgs))
	for i := range pkgs {
		packageChans[i] = make(chan ActionResult)
	}

	cases := make([]reflect.SelectCase, len(packageChans))
	for i, pkgChan := range packageChans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(pkgChan)}
	}

	for i := range pkgs {
		go RunPackgeActions(&pkgs[i], actions, packageChans[i])
	}

	running := len(cases)
	for running > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			cases[chosen].Chan = reflect.ValueOf(nil)
			running -= 1
			continue
		}

		go resultHandler(chosen, value.Interface().(ActionResult), !ok)
	}

}

func RunPackgeActions(pkg *pack.Package, actions []Action, resultChan chan ActionResult) {
	for _, action := range actions {
		actionResult := action.Execute(pkg)

		resultChan <- actionResult

		if actionResult.GetStatus() == ACTION_FAILED {
			break
		}
	}

	close(resultChan)
}
