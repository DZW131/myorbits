package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHarnessRootHelpIncludesV03Examples(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeHarnessCLI(t, t.TempDir(), "--help")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "Examples:")
	require.Contains(t, stdout, "harness create demo-repo")
	require.Contains(t, stdout, "harness bindings plan orbit-template/docs orbit-template/cmd")
	require.Contains(t, stdout, "harness init")
	require.Contains(t, stdout, "harness install orbit-template/docs --bindings .harness/vars.yaml")
	require.Contains(t, stdout, "harness check")
	require.Contains(t, stdout, "harness template save --to harness-template/workspace")
}

func TestHarnessBindingsPlanHelpIncludesExamples(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeHarnessCLI(t, t.TempDir(), "bindings", "plan", "--help")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "Examples:")
	require.Contains(t, stdout, "harness bindings plan orbit-template/docs orbit-template/cmd")
	require.Contains(t, stdout, "harness bindings plan orbit-template/docs orbit-template/cmd --out .harness/vars.yaml")
	require.Contains(t, stdout, "harness bindings plan orbit-template/docs orbit-template/cmd --json")
	require.Contains(t, stdout, "shared bindings skeleton")
}

func TestHarnessInstallHelpIncludesFormalExamples(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeHarnessCLI(t, t.TempDir(), "install", "--help")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "Examples:")
	require.Contains(t, stdout, "harness install orbit-template/docs --bindings .harness/vars.yaml")
	require.Contains(t, stdout, "harness install https://example.com/acme/templates.git --ref orbit-template/docs --bindings .harness/vars.yaml")
	require.Contains(t, stdout, "harness install orbit-template/docs --overwrite-existing --bindings .harness/vars.yaml --json")
	require.Contains(t, stdout, "--progress")
}

func TestHarnessTemplateSaveHelpIncludesDefaultExample(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeHarnessCLI(t, t.TempDir(), "template", "save", "--help")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "Examples:")
	require.Contains(t, stdout, "harness template save --to harness-template/workspace --dry-run")
	require.Contains(t, stdout, "--edit-template")
	require.Contains(t, stdout, "harness template save --to harness-template/workspace --default --json")
	require.Contains(t, stdout, "--dry-run")
	require.Contains(t, stdout, "--overwrite")
	require.Contains(t, stdout, "--default")
}
