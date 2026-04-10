package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWriteAndLoadRuntimeManifestFileRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	createdAt := time.Date(2026, time.April, 5, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.April, 5, 10, 30, 0, 0, time.UTC)
	input := ManifestFile{
		SchemaVersion: 1,
		Kind:          ManifestKindRuntime,
		Runtime: &ManifestRuntimeMetadata{
			ID:        "project_a",
			Name:      "Project A",
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		Members: []ManifestMember{
			{
				OrbitID: "docs",
				Source:  ManifestMemberSourceInstallOrbit,
				AddedAt: time.Date(2026, time.April, 5, 10, 10, 0, 0, time.UTC),
			},
			{
				OrbitID: "cli",
				Source:  ManifestMemberSourceManual,
				AddedAt: time.Date(2026, time.April, 5, 10, 5, 0, 0, time.UTC),
			},
			{
				OrbitID: "workspace",
				Source:  ManifestMemberSourceInstallBundle,
				AddedAt: time.Date(2026, time.April, 5, 10, 15, 0, 0, time.UTC),
			},
		},
	}

	filename, err := WriteManifestFile(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, ManifestPath(repoRoot), filename)

	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, ""+
		"schema_version: 1\n"+
		"kind: runtime\n"+
		"runtime:\n"+
		"    id: project_a\n"+
		"    name: Project A\n"+
		"    created_at: 2026-04-05T10:00:00Z\n"+
		"    updated_at: 2026-04-05T10:30:00Z\n"+
		"members:\n"+
		"    - orbit_id: cli\n"+
		"      source: manual\n"+
		"      added_at: 2026-04-05T10:05:00Z\n"+
		"    - orbit_id: docs\n"+
		"      source: install_orbit\n"+
		"      added_at: 2026-04-05T10:10:00Z\n"+
		"    - orbit_id: workspace\n"+
		"      source: install_bundle\n"+
		"      added_at: 2026-04-05T10:15:00Z\n", string(data))

	expected := input
	expected.Members = []ManifestMember{
		{
			OrbitID: "cli",
			Source:  ManifestMemberSourceManual,
			AddedAt: time.Date(2026, time.April, 5, 10, 5, 0, 0, time.UTC),
		},
		{
			OrbitID: "docs",
			Source:  ManifestMemberSourceInstallOrbit,
			AddedAt: time.Date(2026, time.April, 5, 10, 10, 0, 0, time.UTC),
		},
		{
			OrbitID: "workspace",
			Source:  ManifestMemberSourceInstallBundle,
			AddedAt: time.Date(2026, time.April, 5, 10, 15, 0, 0, time.UTC),
		},
	}

	loaded, err := LoadManifestFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, expected, loaded)
}

func TestWriteAndLoadOrbitTemplateManifestFileRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	input := ManifestFile{
		SchemaVersion: 1,
		Kind:          ManifestKindOrbitTemplate,
		Template: &ManifestTemplateMetadata{
			OrbitID:           "docs",
			DefaultTemplate:   true,
			CreatedFromBranch: "main",
			CreatedFromCommit: "abc123",
			CreatedAt:         time.Date(2026, time.April, 5, 11, 0, 0, 0, time.UTC),
		},
		Variables: map[string]ManifestVariableSpec{
			"project_name": {
				Description: "Product title",
				Required:    true,
			},
		},
	}

	filename, err := WriteManifestFile(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, ManifestPath(repoRoot), filename)

	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, ""+
		"schema_version: 1\n"+
		"kind: orbit_template\n"+
		"template:\n"+
		"    orbit_id: docs\n"+
		"    default_template: true\n"+
		"    created_from_branch: main\n"+
		"    created_from_commit: abc123\n"+
		"    created_at: 2026-04-05T11:00:00Z\n"+
		"variables:\n"+
		"    project_name:\n"+
		"        description: Product title\n"+
		"        required: true\n", string(data))

	loaded, err := LoadManifestFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestWriteAndLoadHarnessTemplateManifestFileRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	input := ManifestFile{
		SchemaVersion:      1,
		Kind:               ManifestKindHarnessTemplate,
		IncludesRootAgents: true,
		Template: &ManifestTemplateMetadata{
			HarnessID:         "workspace",
			DefaultTemplate:   true,
			CreatedFromBranch: "main",
			CreatedFromCommit: "abc123",
			CreatedAt:         time.Date(2026, time.April, 5, 12, 0, 0, 0, time.UTC),
		},
		Members: []ManifestMember{
			{OrbitID: "docs"},
			{OrbitID: "cmd"},
		},
	}

	filename, err := WriteManifestFile(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, ManifestPath(repoRoot), filename)

	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, ""+
		"schema_version: 1\n"+
		"kind: harness_template\n"+
		"template:\n"+
		"    harness_id: workspace\n"+
		"    default_template: true\n"+
		"    created_from_branch: main\n"+
		"    created_from_commit: abc123\n"+
		"    created_at: 2026-04-05T12:00:00Z\n"+
		"members:\n"+
		"    - orbit_id: cmd\n"+
		"    - orbit_id: docs\n"+
		"includes_root_agents: true\n", string(data))

	loaded, err := LoadManifestFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, ManifestFile{
		SchemaVersion:      1,
		Kind:               ManifestKindHarnessTemplate,
		IncludesRootAgents: true,
		Template: &ManifestTemplateMetadata{
			HarnessID:         "workspace",
			DefaultTemplate:   true,
			CreatedFromBranch: "main",
			CreatedFromCommit: "abc123",
			CreatedAt:         time.Date(2026, time.April, 5, 12, 0, 0, 0, time.UTC),
		},
		Members: []ManifestMember{
			{OrbitID: "cmd"},
			{OrbitID: "docs"},
		},
	}, loaded)
}

func TestWriteAndLoadSourceManifestFileRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	input := ManifestFile{
		SchemaVersion: 1,
		Kind:          ManifestKindSource,
		Source: &ManifestSourceMetadata{
			OrbitID:      "docs",
			SourceBranch: "main",
		},
	}

	filename, err := WriteManifestFile(repoRoot, input)
	require.NoError(t, err)
	require.Equal(t, ManifestPath(repoRoot), filename)

	data, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, ""+
		"schema_version: 1\n"+
		"kind: source\n"+
		"source:\n"+
		"    orbit_id: docs\n"+
		"    source_branch: main\n", string(data))

	loaded, err := LoadManifestFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, input, loaded)
}

func TestLoadHarnessTemplateManifestFileDefaultsIncludesRootAgentsToFalse(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	filename := ManifestPath(repoRoot)

	require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
	require.NoError(t, os.WriteFile(filename, []byte(""+
		"schema_version: 1\n"+
		"kind: harness_template\n"+
		"template:\n"+
		"  harness_id: project_a\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-04-05T12:00:00Z\n"+
		"members:\n"+
		"  - orbit_id: docs\n"), 0o600))

	loaded, err := LoadManifestFile(repoRoot)
	require.NoError(t, err)
	require.Equal(t, ManifestFile{
		SchemaVersion: 1,
		Kind:          ManifestKindHarnessTemplate,
		Template: &ManifestTemplateMetadata{
			HarnessID:         "project_a",
			DefaultTemplate:   false,
			CreatedFromBranch: "main",
			CreatedFromCommit: "abc123",
			CreatedAt:         time.Date(2026, time.April, 5, 12, 0, 0, 0, time.UTC),
		},
		Members: []ManifestMember{
			{OrbitID: "docs"},
		},
		IncludesRootAgents: false,
	}, loaded)
}

func TestValidateRuntimeManifestFileRejectsMixedFieldsAndInvalidMembers(t *testing.T) {
	t.Parallel()

	input := ManifestFile{
		SchemaVersion: 1,
		Kind:          ManifestKindRuntime,
		Runtime: &ManifestRuntimeMetadata{
			ID:        "project_a",
			CreatedAt: time.Date(2026, time.April, 5, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, time.April, 5, 10, 30, 0, 0, time.UTC),
		},
		Template: &ManifestTemplateMetadata{
			OrbitID:           "docs",
			CreatedFromBranch: "main",
			CreatedFromCommit: "abc123",
			CreatedAt:         time.Date(2026, time.April, 5, 11, 0, 0, 0, time.UTC),
		},
		Members: []ManifestMember{
			{
				OrbitID: "docs",
				Source:  ManifestMemberSourceManual,
				AddedAt: time.Date(2026, time.April, 5, 10, 5, 0, 0, time.UTC),
			},
			{
				OrbitID: "docs",
				Source:  "bundle",
				AddedAt: time.Date(2026, time.April, 5, 10, 10, 0, 0, time.UTC),
			},
		},
	}

	err := ValidateRuntimeManifestFile(input)
	require.Error(t, err)
	require.ErrorContains(t, err, "template must not be present")
}

func TestValidateOrbitTemplateManifestFileRejectsHarnessOnlyFields(t *testing.T) {
	t.Parallel()

	input := ManifestFile{
		SchemaVersion: 1,
		Kind:          ManifestKindOrbitTemplate,
		Template: &ManifestTemplateMetadata{
			OrbitID:           "docs",
			HarnessID:         "project_a",
			DefaultTemplate:   true,
			CreatedFromBranch: "main",
			CreatedFromCommit: "abc123",
			CreatedAt:         time.Date(2026, time.April, 5, 11, 0, 0, 0, time.UTC),
		},
	}

	err := ValidateOrbitTemplateManifestFile(input)
	require.Error(t, err)
	require.ErrorContains(t, err, "template.harness_id must not be present")
}

func TestValidateHarnessTemplateManifestFileRejectsInvalidMembers(t *testing.T) {
	t.Parallel()

	input := ManifestFile{
		SchemaVersion: 1,
		Kind:          ManifestKindHarnessTemplate,
		Template: &ManifestTemplateMetadata{
			HarnessID:         "project_a",
			CreatedFromBranch: "main",
			CreatedFromCommit: "abc123",
			CreatedAt:         time.Date(2026, time.April, 5, 12, 0, 0, 0, time.UTC),
		},
		Members: []ManifestMember{
			{OrbitID: "docs"},
			{
				OrbitID: "docs",
				Source:  ManifestMemberSourceManual,
			},
		},
	}

	err := ValidateHarnessTemplateManifestFile(input)
	require.Error(t, err)
	require.ErrorContains(t, err, "members[1].orbit_id must be unique")
}

func TestParseManifestFileDataRejectsInvalidKindsAndMixedBranchFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		data        string
		messagePart string
	}{
		{
			name: "unsupported kind",
			data: "" +
				"schema_version: 1\n" +
				"kind: unsupported\n",
			messagePart: "kind must be one of",
		},
		{
			name: "source requires source metadata",
			data: "" +
				"schema_version: 1\n" +
				"kind: source\n",
			messagePart: "source must be present",
		},
		{
			name: "source must not carry runtime fields",
			data: "" +
				"schema_version: 1\n" +
				"kind: source\n" +
				"source:\n" +
				"  orbit_id: docs\n" +
				"  source_branch: main\n" +
				"runtime:\n" +
				"  id: project_a\n" +
				"  created_at: 2026-04-05T10:00:00Z\n" +
				"  updated_at: 2026-04-05T10:30:00Z\n",
			messagePart: "runtime must not be present",
		},
		{
			name: "runtime mixed with template fields",
			data: "" +
				"schema_version: 1\n" +
				"kind: runtime\n" +
				"runtime:\n" +
				"  id: project_a\n" +
				"  created_at: 2026-04-05T10:00:00Z\n" +
				"  updated_at: 2026-04-05T10:30:00Z\n" +
				"template:\n" +
				"  orbit_id: docs\n" +
				"  created_from_branch: main\n" +
				"  created_from_commit: abc123\n" +
				"  created_at: 2026-04-05T11:00:00Z\n" +
				"members: []\n",
			messagePart: "template must not be present",
		},
		{
			name: "orbit template must not declare members",
			data: "" +
				"schema_version: 1\n" +
				"kind: orbit_template\n" +
				"template:\n" +
				"  orbit_id: docs\n" +
				"  default_template: false\n" +
				"  created_from_branch: main\n" +
				"  created_from_commit: abc123\n" +
				"  created_at: 2026-04-05T11:00:00Z\n" +
				"members: []\n",
			messagePart: "members must not be present",
		},
		{
			name: "harness template requires at least one member",
			data: "" +
				"schema_version: 1\n" +
				"kind: harness_template\n" +
				"template:\n" +
				"  harness_id: project_a\n" +
				"  created_from_branch: main\n" +
				"  created_from_commit: abc123\n" +
				"  created_at: 2026-04-05T12:00:00Z\n" +
				"members: []\n",
			messagePart: "members must contain at least one member",
		},
		{
			name: "harness template member must not carry runtime fields",
			data: "" +
				"schema_version: 1\n" +
				"kind: harness_template\n" +
				"template:\n" +
				"  harness_id: project_a\n" +
				"  created_from_branch: main\n" +
				"  created_from_commit: abc123\n" +
				"  created_at: 2026-04-05T12:00:00Z\n" +
				"members:\n" +
				"  - orbit_id: docs\n" +
				"    source: manual\n",
			messagePart: "members[0].source must not be present",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseManifestFileData([]byte(tc.data))
			require.Error(t, err)
			require.ErrorContains(t, err, tc.messagePart)
		})
	}
}
