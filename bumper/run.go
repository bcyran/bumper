package bumper

import (
	"sync"

	"github.com/bcyran/bumper/pack"
)

type (
	ResultHandler   func(pkgIndex int, result ActionResult)
	FinishedHandler func(pkgIndex int)
)

// Run runs actions for the packages, blocks until all results are handled.
// For each result, and on finished processing, appropriate handlers are called.
// Actions and handlers for a single package are run sequentially, but the packages are handled concurrently.
func Run(pkgs []pack.Package, actions []Action, resultHandler ResultHandler, finishedHandler FinishedHandler) {
	pkgWorkersWg := sync.WaitGroup{}
	for i := range pkgs {
		pkgWorkersWg.Add(1)
		go func(pkgIndex int) {
			packageResultHandler := func(result ActionResult) { resultHandler(pkgIndex, result) }
			packageFinishedHandler := func() { finishedHandler(pkgIndex) }
			packageWorker(&pkgs[pkgIndex], actions, packageResultHandler, packageFinishedHandler)
			pkgWorkersWg.Done()
		}(i)
	}
	pkgWorkersWg.Wait()
}

// packageWorker runs RunPackageActions, listens for results, and runs the handlers sequentially.
func packageWorker(pkg *pack.Package, actions []Action, resultHandler func(ActionResult), finishedHandler func()) {
	resultChan := make(chan ActionResult)
	go runPackageActions(pkg, actions, resultChan)
	for result := range resultChan {
		resultHandler(result)
	}
	finishedHandler()
}

// runPackageActions runs actions for a package sequentially and writes the results to the channel.
func runPackageActions(pkg *pack.Package, actions []Action, resultChan chan ActionResult) {
	for _, action := range actions {
		actionResult := action.Execute(pkg)

		resultChan <- actionResult

		if actionResult.GetStatus() != ActionSuccessStatus {
			break
		}
	}

	close(resultChan)
}
