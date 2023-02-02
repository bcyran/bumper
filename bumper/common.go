package bumper

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CommandRunner = func(cwd string, command string, args ...string) ([]byte, error)

func ExecCommand(cwd string, command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = cwd
	stdoutBuf := bytes.Buffer{}
	stderrBuf := strings.Builder{}
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		return []byte{}, fmt.Errorf("%s %s error (%w): %s", command, strings.Join(args, " "), err, stderrBuf.String())
	}
	return stdoutBuf.Bytes(), nil
}
