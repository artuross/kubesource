package commandexec

import "os/exec"

var _ CommandExecutor = (*Executor)(nil)

// CommandExecutor interface for executing external commands and resolving binaries
type CommandExecutor interface {
	Exec(name string, args ...string) ([]byte, error)
	LookPath(file string) (string, error)
}

// Executor implements CommandExecutor for real command execution
type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Exec(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func (e *Executor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
