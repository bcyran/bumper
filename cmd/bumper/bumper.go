package bumper

import (
	"errors"
	"log"

	"github.com/bcyran/bumper/bumper"
	"github.com/bcyran/bumper/upstream"
	"github.com/urfave/cli/v2"
)

const appName = "bumper"
const appVersion = "0.0.0"

func Main(args []string) {
	app := &cli.App{
		Name:  appName,
		Usage: "easily bump AUR package pkgver",
		Action: func(ctx *cli.Context) error {
			var path string
			switch ctx.Args().Len() {
			case 0:
				path = "."
			case 1:
				path = ctx.Args().First()
			default:
				return errors.New("Too many arguments, only one path allowed")
			}

			packages, err := bumper.CollectPackages(path, 1)
			if err != nil {
				log.Fatal(err)
			}

			actions := []bumper.Action{
				bumper.NewCheckAction(upstream.NewVersionProvider),
				bumper.NewBumpAction(bumper.ExecCommand),
			}

			pkgListDisplay := NewPackageListDisplay()
			pkgDisplays := make([]*PackageDisplay, len(packages))
			for i, pkg := range packages {
				pkgDisplays[i] = pkgListDisplay.AddPackage(pkg.Pkgbase)
			}

			handleResult := func(pkgIndex int, result bumper.ActionResult) {
				pkgDisplays[pkgIndex].AddResult(result)
				pkgListDisplay.Display()
			}
			handleFinished := func(pkgIndex int) {
				pkgDisplays[pkgIndex].SetFinished()
				pkgListDisplay.Display()
			}
			bumper.Run(packages, actions, handleResult, handleFinished)

			return nil
		},
	}

	if err := app.Run(args); err != nil {
		log.Fatal(err)
	}
}
