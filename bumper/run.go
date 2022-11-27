package bumper

import (
	"reflect"
	"sync"

	"github.com/bcyran/bumper/pack"
)

type ResultHandler func(pkgIndex int, result ActionResult)
type FinishedHandler func(pkgIndex int)

func Run(pkgs []pack.Package, actions []Action, resultHandler ResultHandler, finishedHandler FinishedHandler) {
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

	handlersWg := sync.WaitGroup{}
	running := len(cases)
	for running > 0 {
		chosen, value, ok := reflect.Select(cases)
		if !ok {
			cases[chosen].Chan = reflect.ValueOf(nil)
			running -= 1

			handlersWg.Add(1)
			go func() {
				finishedHandler(chosen)
				handlersWg.Done()
			}()
			continue
		}

		handlersWg.Add(1)
		go func() {
			resultHandler(chosen, value.Interface().(ActionResult))
			handlersWg.Done()
		}()
	}

	handlersWg.Wait()
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
