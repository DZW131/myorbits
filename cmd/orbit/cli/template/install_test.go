package orbittemplate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWriteAndLoadInstallRecordRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	appliedAt := time.Date(2026, time.March, 21, 10, 30, 0, 0, time.UTC)
	input := InstallRecord{
		SchemaVersion: 1,
		OrbitID:       "docs",
		Template: Source{
			SourceKind:     InstallSourceKindLocalBranch,
			SourceRepo:     "",
			SourceRef:      "orbit-template/docs",
			TemplateCommit: "abc123",
		},
		AppliedAt: appliedAt,
	}

	filename, err := WriteInstallRecord(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(repoRoot, ".orbit", "installs", "docs.yaml"), filename)

	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, ""+
		"schema_version: 1\n"+
		"orbit_id: docs\n"+
		"template:\n"+
		"    source_kind: local_branch\n"+
		"    source_repo: \"\"\n"+
		"    source_ref: orbit-template/docs\n"+
		"    template_commit: abc123\n"+
		"applied_at: 2026-03-21T10:30:00Z\n", string(data))

	loaded, err := LoadInstallRecord(repoRoot, "docs")
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestLoadInstallRecordRejectsMismatchedOrbitID(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	filename := filepath.Join(repoRoot, ".orbit", "installs", "docs.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
	require.NoError(t, os.WriteFile(filename, []byte(""+
		"schema_version: 1\n"+
		"orbit_id: cmd\n"+
		"template:\n"+
		"  source_kind: local_branch\n"+
		"  source_repo: \"\"\n"+
		"  source_ref: orbit-template/docs\n"+
		"  template_commit: abc123\n"+
		"applied_at: 2026-03-21T10:30:00Z\n"), 0o600))

	_, err := LoadInstallRecord(repoRoot, "docs")
	require.Error(t, err)
	require.ErrorContains(t, err, "orbit_id must match install path")
}

func TestValidateInstallRecordRejectsInvalidContracts(t *testing.T) {
	t.Parallel()

	appliedAt := time.Date(2026, time.March, 21, 10, 30, 0, 0, time.UTC)
	testCases := []struct {
		name     string
		input    InstallRecord
		contains string
	}{
		{
			name: "schema version must be frozen",
			input: InstallRecord{
				SchemaVersion: 2,
				OrbitID:       "docs",
				Template: Source{
					SourceKind:     InstallSourceKindLocalBranch,
					SourceRef:      "orbit-template/docs",
					TemplateCommit: "abc123",
				},
				AppliedAt: appliedAt,
			},
			contains: "schema_version must be 1",
		},
		{
			name: "source kind enum is constrained",
			input: InstallRecord{
				SchemaVersion: 1,
				OrbitID:       "docs",
				Template: Source{
					SourceKind:     "manual_copy",
					SourceRef:      "orbit-template/docs",
					TemplateCommit: "abc123",
				},
				AppliedAt: appliedAt,
			},
			contains: "template.source_kind",
		},
		{
			name: "remote source requires repo URL",
			input: InstallRecord{
				SchemaVersion: 1,
				OrbitID:       "docs",
				Template: Source{
					SourceKind:     InstallSourceKindRemoteGit,
					SourceRef:      "refs/heads/template/docs",
					TemplateCommit: "abc123",
				},
				AppliedAt: appliedAt,
			},
			contains: "template.source_repo",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateInstallRecord(testCase.input)
			require.Error(t, err)
			require.ErrorContains(t, err, testCase.contains)
		})
	}
}
