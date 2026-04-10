package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWriteAndLoadRuntimeFileRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	createdAt := time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.March, 25, 11, 0, 0, 0, time.UTC)
	input := RuntimeFile{
		SchemaVersion: 1,
		Kind:          RuntimeKind,
		Harness: RuntimeMetadata{
			ID:        "project_a",
			Name:      "Project A",
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		Members: []RuntimeMember{
			{
				OrbitID: "docs",
				Source:  MemberSourceInstallOrbit,
				AddedAt: time.Date(2026, time.March, 25, 10, 20, 0, 0, time.UTC),
			},
			{
				OrbitID: "cli",
				Source:  MemberSourceManual,
				AddedAt: time.Date(2026, time.March, 25, 10, 10, 0, 0, time.UTC),
			},
			{
				OrbitID: "bundle",
				Source:  MemberSourceInstallBundle,
				AddedAt: time.Date(2026, time.March, 25, 10, 30, 0, 0, time.UTC),
			},
		},
	}

	filename, err := WriteRuntimeFile(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, ManifestPath(repoRoot), filename)
	_, err = os.Stat(filepath.Join(repoRoot, ".harness", "runtime.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)

	expected := input
	expected.Members = []RuntimeMember{
		{
			OrbitID: "bundle",
			Source:  MemberSourceInstallBundle,
			AddedAt: time.Date(2026, time.March, 25, 10, 30, 0, 0, time.UTC),
		},
		{
			OrbitID: "cli",
			Source:  MemberSourceManual,
			AddedAt: time.Date(2026, time.March, 25, 10, 10, 0, 0, time.UTC),
		},
		{
			OrbitID: "docs",
			Source:  MemberSourceInstallOrbit,
			AddedAt: time.Date(2026, time.March, 25, 10, 20, 0, 0, time.UTC),
		},
	}

	loaded, err := LoadRuntimeFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, expected, loaded)

	manifest, err := LoadManifestFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, ManifestFileFromRuntimeFile(expected), manifest)
}

func TestWriteAndLoadRuntimeFileAllowsZeroMembers(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	input := RuntimeFile{
		SchemaVersion: 1,
		Kind:          RuntimeKind,
		Harness: RuntimeMetadata{
			ID:        "project_a",
			CreatedAt: time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, time.March, 25, 11, 0, 0, 0, time.UTC),
		},
		Members: []RuntimeMember{},
	}

	_, err := WriteRuntimeFile(repoRoot, input)
	require.NoError(t, err)

	loaded, err := LoadRuntimeFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestWriteAndLoadRuntimeFileAtPathRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	filename := filepath.Join(repoRoot, ".harness", "runtime.yaml")
	input := RuntimeFile{
		SchemaVersion: 1,
		Kind:          RuntimeKind,
		Harness: RuntimeMetadata{
			ID:        "project_a",
			CreatedAt: time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, time.March, 25, 11, 0, 0, 0, time.UTC),
		},
		Members: []RuntimeMember{
			{OrbitID: "docs", Source: MemberSourceInstallOrbit, AddedAt: time.Date(2026, time.March, 25, 10, 20, 0, 0, time.UTC)},
		},
	}

	writtenPath, err := WriteRuntimeFileAtPath(filename, input)
	require.NoError(t, err)
	require.Equal(t, filename, writtenPath)

	loaded, err := LoadRuntimeFileAtPath(filename)
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestLoadRuntimeFileIgnoresLegacyRuntimeFileWhenManifestIsValid(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	input := RuntimeFile{
		SchemaVersion: 1,
		Kind:          RuntimeKind,
		Harness: RuntimeMetadata{
			ID:        "project_a",
			CreatedAt: time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, time.March, 25, 11, 0, 0, 0, time.UTC),
		},
		Members: []RuntimeMember{
			{OrbitID: "docs", Source: MemberSourceManual, AddedAt: time.Date(2026, time.March, 25, 10, 10, 0, 0, time.UTC)},
		},
	}

	_, err := WriteManifestFile(repoRoot, ManifestFileFromRuntimeFile(input))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, ".harness", "runtime.yaml"), []byte("schema_version: nope\n"), 0o600))

	loaded, err := LoadRuntimeFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestValidateRuntimeFileRejectsDuplicateMembers(t *testing.T) {
	t.Parallel()

	input := RuntimeFile{
		SchemaVersion: 1,
		Kind:          RuntimeKind,
		Harness: RuntimeMetadata{
			ID:        "project_a",
			CreatedAt: time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, time.March, 25, 11, 0, 0, 0, time.UTC),
		},
		Members: []RuntimeMember{
			{OrbitID: "docs", Source: MemberSourceManual, AddedAt: time.Date(2026, time.March, 25, 10, 10, 0, 0, time.UTC)},
			{OrbitID: "docs", Source: MemberSourceInstallOrbit, AddedAt: time.Date(2026, time.March, 25, 10, 20, 0, 0, time.UTC)},
		},
	}

	err := ValidateRuntimeFile(input)
	require.Error(t, err)
	require.ErrorContains(t, err, "must be unique")
}
