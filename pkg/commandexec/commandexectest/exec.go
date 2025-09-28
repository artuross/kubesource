package commandexectest

import (
	"os/exec"
	"strings"

	"github.com/artuross/kubesource/pkg/commandexec"
)

var _ commandexec.CommandExecutor = (*Executor)(nil)

// CommandHandler represents a function that handles a specific command
type CommandHandler func(name string, args ...string) ([]byte, error)

// Executor implements commandexec.Executor for testing
type Executor struct {
	handlers map[string]CommandHandler
	binaries map[string]string
}

func NewExecutor() *Executor {
	return &Executor{
		handlers: make(map[string]CommandHandler),
		binaries: make(map[string]string),
	}
}

// AddHandler registers a handler for a specific full command string
func (e *Executor) AddHandler(fullCommand string, handler CommandHandler) {
	e.handlers[fullCommand] = handler
}

// AddBinary registers a binary name that LookPath should resolve.
func (e *Executor) AddBinary(name, path string) {
	e.binaries[name] = path
}

// Exec executes the command by looking up the registered handler.
func (e *Executor) Exec(name string, args ...string) ([]byte, error) {
	fullCommand := buildCommandString(name, args...)

	handler, exists := e.handlers[fullCommand]
	if !exists {
		return nil, exec.ErrNotFound
	}

	return handler(name, args...)
}

// LookPath simulates exec.LookPath; it returns the registered binary path.
func (e *Executor) LookPath(file string) (string, error) {
	path, ok := e.binaries[file]
	if !ok {
		return "", exec.ErrNotFound
	}

	return path, nil
}

func buildCommandString(name string, args ...string) string {
	if len(args) == 0 {
		return name
	}

	return name + " " + strings.Join(args, " ")
}
