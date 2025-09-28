package commandexectest_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/artuross/kubesource/pkg/commandexec/commandexectest"
)

func TestExecutor_Exec(t *testing.T) {
	tests := []struct {
		name           string
		handlers       map[string]commandexectest.CommandHandler
		cmd            cmdArgs
		expectedOutput string
		expectedError  error
	}{
		{
			name: "exec with args",
			handlers: map[string]commandexectest.CommandHandler{
				"echo hello world": successHandler("hello world"),
			},
			cmd:            cmd("echo", "hello", "world"),
			expectedOutput: "hello world",
			expectedError:  nil,
		},
		{
			name: "exec without args",
			handlers: map[string]commandexectest.CommandHandler{
				"pwd": successHandler("pwd output"),
			},
			cmd:            cmd("pwd"),
			expectedOutput: "pwd output",
			expectedError:  nil,
		},
		{
			name: "handler error propagation",
			handlers: map[string]commandexectest.CommandHandler{
				"failing-command": func(name string, args ...string) ([]byte, error) {
					return nil, assert.AnError
				},
			},
			cmd:            cmd("failing-command"),
			expectedOutput: "",
			expectedError:  assert.AnError,
		},
		{
			name:           "command not found",
			handlers:       map[string]commandexectest.CommandHandler{},
			cmd:            cmd("nonexistent-command", "arg1", "arg2"),
			expectedOutput: "",
			expectedError:  exec.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := commandexectest.NewExecutor()

			for command, handler := range tt.handlers {
				executor.AddHandler(command, handler)
			}

			output, err := executor.Exec(tt.cmd.command, tt.cmd.args...)
			requireErrorIsOrNil(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedOutput, string(output))
		})
	}
}

func requireErrorIsOrNil(t *testing.T, expectedErr, err error) {
	if expectedErr == nil {
		require.NoError(t, err)
		return
	}

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func successHandler(output string) commandexectest.CommandHandler {
	return func(name string, args ...string) ([]byte, error) {
		return []byte(output), nil
	}
}

type cmdArgs struct {
	command string
	args    []string
}

func cmd(command string, args ...string) cmdArgs {
	return cmdArgs{command: command, args: args}
}
