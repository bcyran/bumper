package bumper

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bcyran/bumper/bumper"
	"github.com/fatih/color"
)

const updateIntervalMs = 100

var (
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	successColor  = color.New(color.FgGreen).SprintFunc()
	failureColor  = color.New(color.FgRed).SprintFunc()
	progressColor = color.New(color.FgYellow).SprintFunc()
)

type Flusher interface {
	Flush() error
}

type WriteFlusher interface {
	io.Writer
	Flusher
}

type PackageDisplay struct {
	name          string
	actionResults []bumper.ActionResult
	finished      bool
	failed        bool
	spinnerFrame  int
	mtx           *sync.RWMutex
}

func newPackageDisplay(name string) *PackageDisplay {
	return &PackageDisplay{
		name:          name,
		actionResults: []bumper.ActionResult{},
		finished:      false,
		failed:        false,
		spinnerFrame:  0,
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
	lastResult := pkgDisplay.actionResults[len(pkgDisplay.actionResults)-1]
	if lastResult.GetStatus() == bumper.ACTION_FAILED {
		pkgDisplay.failed = true
	}
	pkgDisplay.mtx.Unlock()
}

func (pkgDisplay *PackageDisplay) AnimationTick() {
	pkgDisplay.mtx.Lock()
	pkgDisplay.spinnerFrame = (pkgDisplay.spinnerFrame + 1) % len(spinnerFrames)
	pkgDisplay.mtx.Unlock()
}

func (pkgDisplay *PackageDisplay) String() string {
	pkgDisplay.mtx.RLock()
	resultsStrings := make([]string, 0)
	var bullet string
	for _, result := range pkgDisplay.actionResults {
		if resStr := result.String(); resStr != "" {
			resultsStrings = append(resultsStrings, result.String())
		}
	}
	if pkgDisplay.finished == true {
		if pkgDisplay.failed == true {
			bullet = failureColor("✗")
		} else {
			bullet = successColor("✓")
		}
	} else {
		bullet = progressColor(spinnerFrames[pkgDisplay.spinnerFrame])
		resultsStrings = append(resultsStrings, "...")
	}
	pkgDisplay.mtx.RUnlock()

	return fmt.Sprintf("%s %s: %s", bullet, pkgDisplay.name, strings.Join(resultsStrings, ", "))
}

type PackageListDisplay struct {
	packages []*PackageDisplay
}

func NewPackageListDisplay() *PackageListDisplay {
	return &PackageListDisplay{
		packages: []*PackageDisplay{},
	}
}

func (pkgListDisplay *PackageListDisplay) AddPackage(name string) *PackageDisplay {
	newPkgDisplay := newPackageDisplay(name)
	pkgListDisplay.packages = append(pkgListDisplay.packages, newPkgDisplay)
	return newPkgDisplay
}

func (pkgListDisplay *PackageListDisplay) Display(out io.Writer) {
	for _, pkgDisplay := range pkgListDisplay.packages {
		fmt.Fprintln(out, pkgDisplay.String())
	}
}

func (pkgListDisplay *PackageListDisplay) LiveDisplay(out WriteFlusher) {
	for range time.Tick(updateIntervalMs * time.Millisecond) {
		pkgListDisplay.Display(out)
		out.Flush()

		finishedCount := 0
		for _, pkgDisplay := range pkgListDisplay.packages {
			if pkgDisplay.finished == true {
				finishedCount += 1
			}
			pkgDisplay.AnimationTick()
		}

		if finishedCount == len(pkgListDisplay.packages) {
			break
		}
	}
}
