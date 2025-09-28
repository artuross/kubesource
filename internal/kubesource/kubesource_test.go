package kubesource_test

import (
	"testing"

	"github.com/artuross/kubesource/internal/kubesource"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFindDirectories exercises kubesource.FindDirectories with a table-driven approach
// using an in‑memory (afero) filesystem. The goal is to reach 100% coverage by
// verifying: normal discovery (multiple + nested), root (".") handling, empty
// result set, and error propagation (non-existent root).
func TestFindDirectories(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		type testCase struct {
			name          string
			root          string
			inputFiles    []string
			expectDirs    []string
			expectedError error
		}

		tests := []testCase{
			{
				name: "multiple nested matches sorted",
				root: ".",
				inputFiles: []string{
					"srv/app1/kubesource.yaml",
					"srv/app2/sub/kubesource.yaml",
					"srv/app2/sub/other.txt",
					"srv/app3/readme.md",
				},
				expectDirs: []string{
					"srv/app1",
					"srv/app2/sub",
				},
			},
			{
				name: "nested matches with non-dot root",
				root: "srv",
				inputFiles: []string{
					"srv/app1/kubesource.yaml",
					"srv/app2/sub/kubesource.yaml",
					"srv/app2/sub/other.txt",
				},
				expectDirs: []string{
					"srv/app1",
					"srv/app2/sub",
				},
			},
			{
				name: "deep root single match",
				root: "srv/app2/sub",
				inputFiles: []string{
					"srv/app2/sub/kubesource.yaml",
					"srv/app2/sub/ignored.txt",
				},
				expectDirs: []string{
					"srv/app2/sub",
				},
			},
			{
				name: "single match at root directory",
				root: ".",
				inputFiles: []string{
					"kubesource.yaml",
					"unrelated.txt",
				},
				expectDirs: []string{
					".",
				},
			},
			{
				name: "no matches returns empty slice",
				root: ".",
				inputFiles: []string{
					"something.txt",
					"nested/else.conf",
				},
				expectDirs: nil,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				fs := afero.NewMemMapFs()

				for _, file := range tc.inputFiles {
					err := afero.WriteFile(fs, file, []byte("content"), 0o644)
					require.NoError(t, err)
				}

				dirs, err := kubesource.FindDirectories(fs, tc.root)
				require.NoError(t, err)
				assert.Equal(t, tc.expectDirs, dirs)
			})
		}
	})
}
