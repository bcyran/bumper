package testutils

type CommandRunner = func(cwd string, command string, args ...string) ([]byte, error)

type CommandRunnerParams struct {
	Cwd     string
	Command string
	Args    []string
}

type CommandRunnerRetval struct {
	Stdout []byte
	Err    error
}

// MakeFakeCommandRunner creates fake CommandRunner which doesn't use exec.Command.
// Instead it appends each call params to a slice for later assertions.
// Each call returns stdout and err values from given retvals slice.
func MakeFakeCommandRunner(retvals *[]CommandRunnerRetval) (CommandRunner, *[]CommandRunnerParams) {
	var commandRuns []CommandRunnerParams
	fakeExecCommand := func(cwd string, command string, args ...string) ([]byte, error) {
		opts := CommandRunnerParams{
			Cwd:     cwd,
			Command: command,
			Args:    args,
		}
		commandRuns = append(commandRuns, opts)
		retval := (*retvals)[0]
		*retvals = (*retvals)[1:]
		return retval.Stdout, retval.Err
	}
	return fakeExecCommand, &commandRuns
}
