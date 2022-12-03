package bumper

import (
	"os/exec"
)

type CommandRunner = func(cwd string, command string, args ...string) ([]byte, error)

func ExecCommand(cwd string, command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd
	stdout, err := cmd.Output()
	if err != nil {
		return []byte{}, err
	}
	return stdout, nil
}
