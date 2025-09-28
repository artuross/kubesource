package commandexec_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/artuross/kubesource/pkg/commandexec"
)

func TestExecutor(t *testing.T) {
	tests := []struct {
		name           string
		cmd            cmdArgs
		expectedOutput string
		expectError    error
	}{
		{
			name:           "exec with args",
			cmd:            cmd("printf", "hello world"),
			expectedOutput: "hello world",
			expectError:    nil,
		},
		{
			name:           "exec without args",
			cmd:            cmd("true"),
			expectedOutput: "",
			expectError:    nil,
		},
		{
			name:           "command not found",
			cmd:            cmd("this-command-should-not-exist-anywhere"),
			expectedOutput: "",
			expectError:    exec.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := commandexec.NewExecutor()

			output, err := executor.Exec(tt.cmd.command, tt.cmd.args...)

			requireErrorIsOrNil(t, tt.expectError, err)
			assert.Equal(t, tt.expectedOutput, string(output))
		})
	}
}

type cmdArgs struct {
	command string
	args    []string
}

func cmd(command string, args ...string) cmdArgs {
	return cmdArgs{command: command, args: args}
}

func requireErrorIsOrNil(t *testing.T, expectedErr, err error) {
	if expectedErr == nil {
		require.NoError(t, err)
		return
	}

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}
