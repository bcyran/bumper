package bumper

import (
	"errors"
	"fmt"
	"log"

	"github.com/bcyran/bumper/bumper"
	"github.com/urfave/cli/v2"
	"github.com/bcyran/bumper/upstream"
)

const name = "bumper"
const version = "0.0.0"

func Main(args []string) {
	app := &cli.App{
		Name:  name,
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
			loadedPackage, err := bumper.LoadPackage(path)
			if err != nil {
				return fmt.Errorf("Not a valid package: %s\n", err)
			}
			fmt.Printf("Package:\n")
			fmt.Printf("path: %s\n", loadedPackage.Path)
			fmt.Printf("pkgname: %s\n", loadedPackage.Pkgname)
			fmt.Printf("url: %s\n", loadedPackage.Url)
			fmt.Printf("pkgver: %s\n", loadedPackage.Pkgver)
			fmt.Printf("pkgrel: %s\n", loadedPackage.Pkgrel)

			upstreamGithub := upstream.NewGitHub(loadedPackage.Url)
			latestUpstreamVersion, err := upstreamGithub.LatestVersion()

			if err != nil {
				log.Fatal(err)
				return nil
			}

			fmt.Printf("upstream version: %s\n", latestUpstreamVersion)
			
			return nil
		},
	}

	if err := app.Run(args); err != nil {
		log.Fatal(err)
	}
}
