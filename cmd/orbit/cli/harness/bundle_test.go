package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

func TestWriteAndLoadBundleRecordRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	appliedAt := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)
	input := BundleRecord{
		SchemaVersion:      1,
		HarnessID:          "workspace",
		Template:           orbittemplate.Source{SourceKind: orbittemplate.InstallSourceKindLocalBranch, SourceRepo: "", SourceRef: "harness-template/workspace", TemplateCommit: "abc123"},
		MemberIDs:          []string{"cmd", "docs"},
		AppliedAt:          appliedAt,
		IncludesRootAgents: true,
		OwnedPaths:         []string{"AGENTS.md", "cmd/main.go", "docs/guide.md"},
	}

	filename, err := WriteBundleRecord(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(repoRoot, ".harness", "bundles", "workspace.yaml"), filename)

	loaded, err := LoadBundleRecord(repoRoot, "workspace")
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestLoadBundleRecordRejectsMismatchedHarnessID(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	filename := filepath.Join(repoRoot, ".harness", "bundles", "workspace.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
	require.NoError(t, os.WriteFile(filename, []byte(""+
		"schema_version: 1\n"+
		"harness_id: other\n"+
		"template:\n"+
		"  source_kind: local_branch\n"+
		"  source_repo: \"\"\n"+
		"  source_ref: harness-template/workspace\n"+
		"  template_commit: abc123\n"+
		"member_ids:\n"+
		"  - docs\n"+
		"applied_at: 2026-04-01T09:00:00Z\n"+
		"includes_root_agents: false\n"+
		"owned_paths: []\n"), 0o600))

	_, err := LoadBundleRecord(repoRoot, "workspace")
	require.Error(t, err)
	require.ErrorContains(t, err, "harness_id must match bundle path")
}
