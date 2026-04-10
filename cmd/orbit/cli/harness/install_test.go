package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

func TestWriteAndLoadHarnessInstallRecordRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	appliedAt := time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	input := orbittemplate.InstallRecord{
		SchemaVersion: 1,
		OrbitID:       "docs",
		Template: orbittemplate.Source{
			SourceKind:     orbittemplate.InstallSourceKindLocalBranch,
			SourceRepo:     "",
			SourceRef:      "orbit-template/docs",
			TemplateCommit: "abc123",
		},
		AppliedAt: appliedAt,
	}

	filename, err := WriteInstallRecord(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(repoRoot, ".harness", "installs", "docs.yaml"), filename)

	loaded, err := LoadInstallRecord(repoRoot, "docs")
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestLoadHarnessInstallRecordRejectsMismatchedOrbitID(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	filename := filepath.Join(repoRoot, ".harness", "installs", "docs.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
	require.NoError(t, os.WriteFile(filename, []byte(""+
		"schema_version: 1\n"+
		"orbit_id: cmd\n"+
		"template:\n"+
		"  source_kind: local_branch\n"+
		"  source_repo: \"\"\n"+
		"  source_ref: orbit-template/docs\n"+
		"  template_commit: abc123\n"+
		"applied_at: 2026-03-25T12:00:00Z\n"), 0o600))

	_, err := LoadInstallRecord(repoRoot, "docs")
	require.Error(t, err)
	require.ErrorContains(t, err, "orbit_id must match install path")
}
