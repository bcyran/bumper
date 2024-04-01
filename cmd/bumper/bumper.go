package bumper

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/bcyran/bumper/bumper"
	"github.com/bcyran/bumper/upstream"
	"github.com/gosuri/uilive"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"go.uber.org/config"
)

type DoActions struct {
	bump   bool
	make   bool
	commit bool
	push   bool
}

var (
	doActions = DoActions{
		bump:   true,
		make:   true,
		commit: true,
		push:   false,
	}
	collectDepth     = 1
	configPath       = ""
	completion       = ""
	versionOverrides = []string{}
)

var bumperCmd = &cobra.Command{
	Use:   "bumper [dir]",
	Short: "Easily bump $pkgver in AUR packages",
	Long: `Tool for AUR package maintainers to easily align packages with upstream releases.

Uses URLs found in PKGBUILD to check whether the packaged software has a new
version available.

For each package with update available it can perform the following actions:
  * bump: update PKGBUILD and .SRCINFO
  * make: build package to make sure it's still valid after the bump
  * commit: commit the changes
  * push: push committed changes

Actions can be selected using CLI flags. By default push action is disabled.

Packages are searched recursively starting in the given dir (current working
directory by default if no dir is given). Default recursion depth is 1 which
enables you to run bumper in a dir containing multiple package dirs.`,
	Example: `  bumper                                find and bump packages in $PWD
  bumper --override my-package=1.2.3    override my-package version to 1.2.3
  bumper --bump=false                   find packages, check updates in $PWD
  bumper ~/workspace/aur                find and bump packages in given dir
  bumper ~/workspace/aur/my-package     bump single package`,
	Version: "1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		if completion != "" {
			generateCompletion(cmd, completion)
			os.Exit(0)
		}

		var workDir string
		switch len(args) {
		case 0:
			workDir = "."
		case 1:
			workDir = args[0]
		default:
			cmd.PrintErr("Too many arguments, only one path allowed!")
			os.Exit(1)
		}

		workDir, err := filepath.Abs(workDir)
		if err != nil {
			cmd.PrintErrf("Fatal error, invalid path: %v.\n", err)
			os.Exit(1)
		}

		bumperCLIConfig, err := configFromVersionOverrides(versionOverrides)
		if err != nil {
			fmt.Printf("Fatal error, invalid CLI option: %v.\n", err)
			os.Exit(1)
		}

		bumperConfig, err := bumper.ReadConfig(configPath, bumperCLIConfig)
		if err != nil {
			fmt.Printf("Fatal error, invalid config: %v.\n", err)
			os.Exit(1)
		}

		actions := createActions(doActions, bumperConfig)
		runBumper(workDir, actions)
	},
	ValidArgsFunction: func(_cmd *cobra.Command, _args []string, _toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	},
}

func init() {
	bumperCmd.Flags().BoolVarP(&doActions.bump, "bump", "b", true, "bump outdated packages")
	bumperCmd.Flags().BoolVarP(&doActions.make, "make", "m", true, "build bumped packages")
	bumperCmd.Flags().BoolVarP(&doActions.commit, "commit", "c", true, "commit changes")
	bumperCmd.Flags().BoolVarP(&doActions.push, "push", "p", false, "push committed changes")
	bumperCmd.Flags().IntVarP(&collectDepth, "depth", "d", 1, "depth of dir recursion in search for packages")
	bumperCmd.Flags().StringVarP(&configPath, "config", "", "", "path to configuration file")
	bumperCmd.Flags().StringVarP(&completion, "completion", "", "", "generate completion for shell: bash, zsh, fish")
	bumperCmd.Flags().StringArrayVarP(&versionOverrides, "override", "o", []string{}, "override upstream version, format: package=version")
	bumperCmd.RegisterFlagCompletionFunc("completion", func(_cmd *cobra.Command, _args []string, _toComplete string) ([]string, cobra.ShellCompDirective) { //nolint:errcheck
		return []string{"bash", "zsh", "fish"}, cobra.ShellCompDirectiveDefault
	})
}

func createActions(doActions DoActions, bumperConfig config.Provider) []bumper.Action {
	actions := []bumper.Action{
		bumper.NewCheckAction(upstream.NewVersionProvider, bumperConfig.Get("check")),
	}

	if doActions.bump {
		actions = append(actions, bumper.NewBumpAction(bumper.ExecCommand))
	} else {
		return actions
	}

	if doActions.make {
		actions = append(actions, bumper.NewMakeAction(bumper.ExecCommand))
	}

	if doActions.commit {
		actions = append(actions, bumper.NewCommitAction(bumper.ExecCommand, bumperConfig.Get("commit")))
	} else {
		return actions
	}

	if doActions.push {
		actions = append(actions, bumper.NewPushAction(bumper.ExecCommand))
	}

	return actions
}

func runBumper(workDir string, actions []bumper.Action) {
	packages, err := bumper.CollectPackages(workDir, collectDepth)
	if err != nil {
		fmt.Printf("Fatal error, could not collect packages: %v.\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Printf("No AUR packages found.\n")
		os.Exit(1)
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

func generateCompletion(cmd *cobra.Command, shell string) {
	var err error
	switch shell {
	case "bash":
		err = cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		err = cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		err = cmd.Root().GenFishCompletion(os.Stdout, true)
	default:
		fmt.Printf("invalid shell name '%s', allowed: bash, zsh, fish.\n", shell)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("completion generation failed: %v\n", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := bumperCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
