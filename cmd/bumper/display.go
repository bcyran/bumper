package bumper

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bcyran/bumper/bumper"
	"github.com/gosuri/uilive"
)

type PackageDisplay struct {
	name          string
	actionResults []bumper.ActionResult
	finished      bool
	mtx           *sync.RWMutex
}

func newPackageDisplay(name string) *PackageDisplay {
	return &PackageDisplay{
		name:          name,
		actionResults: []bumper.ActionResult{},
		finished:      false,
		mtx:           &sync.RWMutex{},
	}
}

func (pkgDisplay *PackageDisplay) AddResult(actionResult bumper.ActionResult) {
	pkgDisplay.mtx.Lock()
	pkgDisplay.actionResults = append(pkgDisplay.actionResults, actionResult)
	pkgDisplay.mtx.Unlock()
}

func (pkgDisplay *PackageDisplay) SetFinished() {
	pkgDisplay.mtx.Lock()
	pkgDisplay.finished = true
	pkgDisplay.mtx.Unlock()
}

func (pkgDisplay *PackageDisplay) String() string {
	pkgDisplay.mtx.RLock()
	resultsStrings := make([]string, 0)
	for _, result := range pkgDisplay.actionResults {
		if resStr := result.String(); resStr != "" {
			resultsStrings = append(resultsStrings, result.String())
		}
	}
	if pkgDisplay.finished == false {
		resultsStrings = append(resultsStrings, "...")
	}
	pkgDisplay.mtx.RUnlock()

	return fmt.Sprintf("- %s: %s", pkgDisplay.name, strings.Join(resultsStrings, ", "))
}

type PackageListDisplay struct {
	packages   []*PackageDisplay
	liveWriter *uilive.Writer
}

func NewPackageListDisplay() *PackageListDisplay {
	return &PackageListDisplay{
		packages:   []*PackageDisplay{},
		liveWriter: uilive.New(),
	}
}

func (pkgListDisplay *PackageListDisplay) AddPackage(name string) *PackageDisplay {
	newPkgDisplay := newPackageDisplay(name)
	pkgListDisplay.packages = append(pkgListDisplay.packages, newPkgDisplay)
	return newPkgDisplay
}

func (pkgListDisplay *PackageListDisplay) String() string {
	outString := ""
	for _, pkgDisplay := range pkgListDisplay.packages {
		outString += pkgDisplay.String() + "\n"
	}
	return outString
}

func (pkgListDisplay *PackageListDisplay) Display() {
	fmt.Fprintf(pkgListDisplay.liveWriter, pkgListDisplay.String())
	pkgListDisplay.liveWriter.Flush()
}
