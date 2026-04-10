package orbittemplate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zack-nova/orbit/cmd/orbit/cli/testutil"
)

func TestResolvePublishOrbitIDRejectsStrayLegacyDefinitionsAlongsideHostedSource(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	repo.Run(t, "branch", "-m", "main")
	repo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"meta:\n"+
		"  file: .harness/orbits/docs.yaml\n"+
		"  include_in_projection: true\n"+
		"  include_in_write: true\n"+
		"  include_in_export: true\n"+
		"  include_description_in_orchestration: true\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n")
	repo.WriteFile(t, ".orbit/orbits/api.yaml", ""+
		"id: api\n"+
		"description: API orbit\n"+
		"include:\n"+
		"  - api/**\n")
	repo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	repo.WriteFile(t, "api/spec.md", "API spec\n")
	_, err := WriteSourceManifest(repo.Root, SourceManifest{
		SchemaVersion: sourceSchemaVersion,
		Kind:          SourceKind,
		SourceBranch:  "main",
		Publish: &SourcePublishConfig{
			OrbitID: "docs",
		},
	})
	require.NoError(t, err)
	repo.AddAndCommit(t, "seed hosted source repo")

	_, err = resolvePublishOrbitID(context.Background(), repo.Root, "")
	require.Error(t, err)
	require.ErrorContains(t, err, ".orbit/orbits")
	require.ErrorContains(t, err, "init-source")
}
