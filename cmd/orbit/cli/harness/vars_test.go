package harness

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
)

func TestWriteAndLoadHarnessVarsFileRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	input := bindings.VarsFile{
		SchemaVersion: 1,
		Variables: map[string]bindings.VariableBinding{
			"project_name": {
				Value:       "HarnessOS",
				Description: "项目名称",
			},
		},
	}

	filename, err := WriteVarsFile(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, VarsPath(repoRoot), filename)

	loaded, err := LoadVarsFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}
