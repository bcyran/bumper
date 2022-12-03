package bumper

import (
	"log"
	"os"
	"sync"

	"github.com/bcyran/bumper/bumper"
	"github.com/bcyran/bumper/upstream"
	"github.com/gosuri/uilive"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

type DoActions struct {
	bump bool
	make bool
}

func Main(args []string) {
	doActions := DoActions{
		bump: true,
		make: true,
	}

	app := &cli.App{
		Name:      "bumper",
		Version:   "0.1.0",
		Usage:     "easily bump $pkgver in AUR packages",
		ArgsUsage: "[directory]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "bump",
				Aliases:     []string{"b"},
				Usage:       "bump outdated packages",
				Value:       true,
				Destination: &doActions.bump,
			},
			&cli.BoolFlag{
				Name:        "make",
				Aliases:     []string{"m"},
				Usage:       "make (build) bumped packages",
				Value:       true,
				Destination: &doActions.make,
			},
		},
		Action: func(ctx *cli.Context) error {
			var workDir string
			switch ctx.Args().Len() {
			case 0:
				workDir = "."
			case 1:
				workDir = ctx.Args().First()
			default:
				return cli.Exit("Too many arguments, only one path allowed!", 1)
			}

			actions := createActions(doActions)
			runBumper(workDir, actions)

			return nil
		},
	}
	cli.AppHelpTemplate = `{{.Name}} - {{.Usage}}

Usage:
  {{.HelpName}} {{if .VisibleFlags}}[options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
{{if .VisibleFlags}}
Options:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
`

	if err := app.Run(args); err != nil {
		log.Fatal(err)
	}
}

func createActions(doActions DoActions) []bumper.Action {
	actions := []bumper.Action{
		bumper.NewCheckAction(upstream.NewVersionProvider),
	}

	if doActions.bump == true {
		actions = append(actions, bumper.NewBumpAction(bumper.ExecCommand))
	} else {
		return actions
	}

	if doActions.make == true {
		actions = append(actions, bumper.NewBuildAction(bumper.ExecCommand))
	}

	return actions
}

func runBumper(workDir string, actions []bumper.Action) {
	packages, err := bumper.CollectPackages(workDir, 1)
	if err != nil {
		log.Fatal(err)
	}

	pkgListDisplay := NewPackageListDisplay()
	pkgDisplays := make([]*PackageDisplay, len(packages))
	for i, pkg := range packages {
		pkgDisplays[i] = pkgListDisplay.AddPackage(pkg.Pkgbase)
	}

	handleResult := func(pkgIndex int, result bumper.ActionResult) {
		pkgDisplays[pkgIndex].AddResult(result)
	}
	handleFinished := func(pkgIndex int) {
		pkgDisplays[pkgIndex].SetFinished()
	}

	ttyOutput := isatty.IsTerminal(os.Stdout.Fd())

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		bumper.Run(packages, actions, handleResult, handleFinished)
		wg.Done()
	}()

	if ttyOutput {
		wg.Add(1)
		go func() {
			pkgListDisplay.LiveDisplay(uilive.New())
			wg.Done()
		}()
	}

	wg.Wait()
	if !ttyOutput {
		pkgListDisplay.Display(os.Stdout)
	}
}
