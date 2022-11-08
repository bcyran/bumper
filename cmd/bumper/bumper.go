package bumper

import (
	"errors"
	"fmt"
	"log"

	"github.com/bcyran/bumper/bumper"
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
			for _, pack := range packages {
				fmt.Printf("- %s (%s)\n", pack.Pkgbase, pack.Pkgver)
			}

			return nil
		},
	}

	if err := app.Run(args); err != nil {
		log.Fatal(err)
	}
}
