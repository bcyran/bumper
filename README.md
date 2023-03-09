# bumper
[![CI](https://github.com/bcyran/bumper/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/bcyran/bumper/actions/workflows/test.yml)
[![codecov](https://codecov.io/github/bcyran/bumper/branch/master/graph/badge.svg?token=GCRN2996B0)](https://codecov.io/github/bcyran/bumper)
[![GitHub release](https://img.shields.io/github/v/release/bcyran/bumper)](https://github.com/bcyran/bumper/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/bcyran/bumper.svg)](https://pkg.go.dev/github.com/bcyran/bumper)
[![license](https://img.shields.io/github/license/bcyran/bumper)](https://github.com/bcyran/bumper/blob/master/LICENSE)

Helper for [Arch Linux User Repository](https://aur.archlinux.org/) (AUR) package maintainers to easily bump `$pkgver` in their packages.

![bumper demo gif](./demo.gif)

## TL;DR
Imagine you have all your maintained AUR packages in a single directory.
Running `bumper` in this dir will find the packages in subdirectories and perform the following actions for each of them:
1. **check** - attempt to use URLs found in `.SRCINFO` to infer the latest released version number.
  Compare that with the `.SRCINFO` version.
2. **bump** - update `$pkgver` and `$pkgrel` in `PKGBUILD`, run `updpkgsums`, regenerate `.SRCINFO`.
3. **make** - build the package to make sure it's still valid.
4. **commit** - `git commit` the changes.
5. **push** - `git push` the changes.

Nothing will be committed or pushed in case of `make` or any other action failure.

## Installation
### AUR
AUR packages are available: [bumper](https://aur.archlinux.org/packages/bumper) and [bumper-bin](https://aur.archlinux.org/packages/bumper-bin).

### Binary
You can download tarball containing the latest prebuilt binary from the [releases page](https://github.com/bcyran/bumper/releases).
The binary named `bumper` has to be placed in directory in your `$PATH`, e.g. `/usr/local/bin`.

### Go
```
go install github.com/bcyran/bumper@latest
```

## Usage
### Basics
`bumper` accepts a single optional argument - path to a directory, if omitted `$PWD` is used.
The path can be either a package directory or a directory with many subdirectories containing packages.

Consider the following directory structure:
```
/home/user/workspace/aur
├── package1
├── package2
└── package3
```
Example `bumper` usage could look like this:
```bash
# bump all three packages
bumper /home/user/workspace/aur
# or alternatively
cd /home/user/workspace/aur && bumper

# bump a single package
bumper /home/user/workspace/aur/package1
# or alternatively
cd /home/user/workspace/aur/package1 && bumper
```

### Options
| CLI option | Default | Description |
| ---------- | ------- | ----------- |
| `--bump`/`-b` | `true` | Bump outdated packages. If disabled, `bumper` will only check for updates. |
| `--make`/`-m` | `true` | Build the package after bumping and before commiting. |
| `--commit`/`-c` | `true` | Commit made changes. Disabling commit disables push as well. |
| `--push`/`-p` | `false` | Push commited changes. |
| `--config` | `$XDG_CONFIG_HOME/bumper/config.yaml`, `$HOME/.config/bumper/config.yaml` | Configuration file path. See [configuration section](#configuration). |
| `--depth`/`-d` | `1` | Depth of directory tree recursion when looking for packages. By default checks given directory and its children. |
| `--override`/`-o` | - | Override version for specified packages, e.g.: `-o mypackage=1.2.3`. This skips upstream check completely. Can be used multiple times for multiple overrides. |
| `--completion` | - | Generate and print shell completion script. Available: bash, zsh, fish. |
| `--version`/`-v` | - | Print version and exit. |
| `--help`/`-h` | - | Print help and exit. |

### Configuration
APIs used to retrieve the upstream versions can have some limitations for unauthorized access.
GitHub and GitLab APIs in particular use rate limiting, so requests made by `bumper` could fail after a few usages or when bumping a lot of packages.
You can configure `bumper` to use your API keys to avoid those limits.

It's also possible to configure the value used as the commit author.

Configuration file is expected to be present at `$XDG_CONFIG_HOME/bumper/config.yaml` or `$HOME/.config/bumper/config.yaml`.
The format is as follows:
```yaml
check:
  providers:
    github:
      apiKey: github_api_key
    gitlab:
      apiKeys:
        gitlab.com: gitlab_com_api_key
        other.gitlab.instance: other_api_key
commit:
  author: John Doe <john.doe@example.com>
```
**Warning**: All configuration fields are optional and the file isn't checked for additional keys!
This means that `bumper` will not fail if you make a typo or other mistake.
It will just continue as usual without using your keys.

## Supported upstream services
- [github.com](https://github.com) - releases and tags API.
- [gitlab.com](https://gitlab.com) and other GitLab instances - releases and tags API.
  Instances other than gitlab.com need to have `git` in domain name to be considered.
- [pypi.org](https://pypi.org) - package metadata API.

## Credits / resources
- https://github.com/simon04/aur-out-of-date
- https://github.com/knqyf263/go-rpm-version
