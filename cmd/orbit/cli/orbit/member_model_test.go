package orbit_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	"github.com/zack-nova/orbit/cmd/orbit/cli/testutil"
)

func TestOrbitMemberRoleParsing(t *testing.T) {
	t.Parallel()

	expectedRoles := []orbitpkg.OrbitMemberRole{
		orbitpkg.OrbitMemberMeta,
		orbitpkg.OrbitMemberSubject,
		orbitpkg.OrbitMemberRule,
		orbitpkg.OrbitMemberProcess,
	}

	require.Equal(t, expectedRoles, orbitpkg.AllOrbitMemberRoles())

	for _, role := range expectedRoles {
		require.True(t, role.IsValid())

		parsed, err := orbitpkg.ParseOrbitMemberRole(string(role))
		require.NoError(t, err)
		require.Equal(t, role, parsed)
	}

	require.False(t, orbitpkg.OrbitMemberRole("outside").IsValid())

	_, err := orbitpkg.ParseOrbitMemberRole("outside")
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid orbit member role")
}

func TestProjectionPlanPathsForRole(t *testing.T) {
	t.Parallel()

	plan := orbitpkg.ProjectionPlan{
		MetaPaths:    []string{".orbit/orbits/docs.yaml"},
		SubjectPaths: []string{"docs/guide.md"},
		RulePaths:    []string{".markdownlint.yaml"},
		ProcessPaths: []string{"docs/process/review.md"},
	}

	require.Equal(t, []string{".orbit/orbits/docs.yaml"}, plan.PathsForRole(orbitpkg.OrbitMemberMeta))
	require.Equal(t, []string{"docs/guide.md"}, plan.PathsForRole(orbitpkg.OrbitMemberSubject))
	require.Equal(t, []string{".markdownlint.yaml"}, plan.PathsForRole(orbitpkg.OrbitMemberRule))
	require.Equal(t, []string{"docs/process/review.md"}, plan.PathsForRole(orbitpkg.OrbitMemberProcess))
	require.Empty(t, plan.PathsForRole(orbitpkg.OrbitMemberRole("outside")))
}

func TestParseOrbitSpecDataRoundTripsMemberModelFields(t *testing.T) {
	t.Parallel()

	sourcePath := "/repo/.orbit/orbits/docs.yaml"
	data := []byte("" +
		"id: docs\n" +
		"name: Documentation\n" +
		"description: User-facing docs orbit.\n" +
		"meta:\n" +
		"  file: .orbit/orbits/docs.yaml\n" +
		"  agents_template: |\n" +
		"    You are the $project_name docs orbit.\n" +
		"    Keep release notes current.\n" +
		"  include_in_projection: true\n" +
		"  include_in_write: true\n" +
		"  include_in_export: true\n" +
		"  include_description_in_orchestration: true\n" +
		"members:\n" +
		"  - key: docs-content\n" +
		"    name: Docs Content\n" +
		"    role: subject\n" +
		"    paths:\n" +
		"      include:\n" +
		"        - docs/**\n" +
		"        - README.md\n" +
		"      exclude:\n" +
		"        - docs/generated/**\n" +
		"  - key: docs-process\n" +
		"    description: Writer and agent workflow.\n" +
		"    role: process\n" +
		"    lane: agents\n" +
		"    paths:\n" +
		"      include:\n" +
		"        - docs/process/**\n" +
		"    scopes:\n" +
		"      write: false\n" +
		"      orchestration: true\n" +
		"rules:\n" +
		"  scope:\n" +
		"    projection_roles: [meta, subject, rule, process]\n" +
		"    write_roles: [meta, rule]\n" +
		"    export_roles: [meta, rule]\n" +
		"    orchestration_roles: [meta, rule, process]\n" +
		"  orchestration:\n" +
		"    include_orbit_description: true\n" +
		"    materialize_agents_from_meta: true\n")

	spec, err := orbitpkg.ParseOrbitSpecData(data, sourcePath)
	require.NoError(t, err)

	require.Equal(t, "docs", spec.ID)
	require.Equal(t, "Documentation", spec.Name)
	require.Equal(t, sourcePath, spec.SourcePath)
	require.NotNil(t, spec.Meta)
	require.Equal(t, ".orbit/orbits/docs.yaml", spec.Meta.File)
	require.Equal(t, "You are the $project_name docs orbit.\nKeep release notes current.\n", spec.Meta.AgentsTemplate)
	require.Len(t, spec.Members, 2)
	require.Equal(t, orbitpkg.OrbitMemberSubject, spec.Members[0].Role)
	require.Equal(t, orbitpkg.OrbitMemberProcess, spec.Members[1].Role)
	require.NotNil(t, spec.Members[1].Scopes)
	require.NotNil(t, spec.Rules)
	require.True(t, spec.HasMemberSchema())

	marshaled, err := yaml.Marshal(spec)
	require.NoError(t, err)

	roundTripped, err := orbitpkg.ParseOrbitSpecData(marshaled, sourcePath)
	require.NoError(t, err)
	require.Equal(t, spec, roundTripped)
}

func TestParseHostedOrbitSpecDataAcceptsHarnessHostedMetaFile(t *testing.T) {
	t.Parallel()

	sourcePath := "/repo/.harness/orbits/docs.yaml"
	spec, err := orbitpkg.ParseHostedOrbitSpecData([]byte(""+
		"id: docs\n"+
		"meta:\n"+
		"  file: .harness/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"), sourcePath)
	require.NoError(t, err)
	require.Equal(t, sourcePath, spec.SourcePath)
	require.NotNil(t, spec.Meta)
	require.Equal(t, ".harness/orbits/docs.yaml", spec.Meta.File)
}

func TestParseHostedOrbitSpecDataRejectsLegacyMetaFile(t *testing.T) {
	t.Parallel()

	_, err := orbitpkg.ParseHostedOrbitSpecData([]byte(""+
		"id: docs\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"), "/repo/.harness/orbits/docs.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "meta.file")
	require.ErrorContains(t, err, ".harness/orbits/docs.yaml")
}

func TestOrbitSpecLegacyDefinitionCompatibility(t *testing.T) {
	t.Parallel()

	definition := orbitpkg.Definition{
		ID:          "docs",
		Description: "User-facing docs orbit.",
		Include:     []string{"docs/**", "README.md"},
		Exclude:     []string{"docs/generated/**"},
		SourcePath:  "/repo/.orbit/orbits/docs.yaml",
	}

	spec := orbitpkg.OrbitSpecFromDefinition(definition)
	require.False(t, spec.HasMemberSchema())
	require.Equal(t, definition.ID, spec.ID)
	require.Equal(t, definition.Description, spec.Description)
	require.Equal(t, definition.Include, spec.Include)
	require.Equal(t, definition.Exclude, spec.Exclude)
	require.Equal(t, definition.SourcePath, spec.SourcePath)

	require.Equal(t, definition, spec.LegacyDefinition())
}

func TestParseOrbitSpecDataAcceptsLegacySchema(t *testing.T) {
	t.Parallel()

	sourcePath := "/repo/.orbit/orbits/docs.yaml"
	data := []byte("" +
		"id: docs\n" +
		"description: User-facing docs orbit.\n" +
		"include:\n" +
		"  - docs/**\n" +
		"exclude:\n" +
		"  - docs/generated/**\n")

	spec, err := orbitpkg.ParseOrbitSpecData(data, sourcePath)
	require.NoError(t, err)
	require.False(t, spec.HasMemberSchema())
	require.Equal(t, orbitpkg.Definition{
		ID:          "docs",
		Description: "User-facing docs orbit.",
		Include:     []string{"docs/**"},
		Exclude:     []string{"docs/generated/**"},
		SourcePath:  sourcePath,
	}, spec.LegacyDefinition())
}

func TestParseOrbitSpecDataRejectsInvalidMemberRole(t *testing.T) {
	t.Parallel()

	_, err := orbitpkg.ParseOrbitSpecData([]byte(""+
		"id: docs\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: outside\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"), "/repo/.orbit/orbits/docs.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "members[0].role")
	require.ErrorContains(t, err, "invalid orbit member role")
}

func TestParseOrbitSpecDataRejectsInvalidRuleScopeRole(t *testing.T) {
	t.Parallel()

	_, err := orbitpkg.ParseOrbitSpecData([]byte(""+
		"id: docs\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"+
		"rules:\n"+
		"  scope:\n"+
		"    write_roles: [meta, outside]\n"), "/repo/.orbit/orbits/docs.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "rules.scope.write_roles[1]")
	require.ErrorContains(t, err, "invalid orbit member role")
}

func TestParseOrbitSpecDataRejectsDuplicateMemberKeys(t *testing.T) {
	t.Parallel()

	_, err := orbitpkg.ParseOrbitSpecData([]byte(""+
		"id: docs\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"+
		"  - key: docs-content\n"+
		"    role: rule\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - .markdownlint.yaml\n"), "/repo/.orbit/orbits/docs.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "members[1].key must be unique")
}

func TestParseOrbitSpecDataRejectsMismatchedMetaFile(t *testing.T) {
	t.Parallel()

	_, err := orbitpkg.ParseOrbitSpecData([]byte(""+
		"id: docs\n"+
		"meta:\n"+
		"  file: .orbit/orbits/guides.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"), "/repo/.orbit/orbits/docs.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "meta.file")
	require.ErrorContains(t, err, ".orbit/orbits/docs.yaml")
}

func TestParseOrbitSpecDataRejectsMixedLegacyAndMemberSchema(t *testing.T) {
	t.Parallel()

	_, err := orbitpkg.ParseOrbitSpecData([]byte(""+
		"id: docs\n"+
		"include:\n"+
		"  - docs/**\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"), "/repo/.orbit/orbits/docs.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "legacy include/exclude")
}

func TestParseOrbitSpecDataRejectsInvalidScopePatchField(t *testing.T) {
	t.Parallel()

	_, err := orbitpkg.ParseOrbitSpecData([]byte(""+
		"id: docs\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-process\n"+
		"    role: process\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/process/**\n"+
		"    scopes:\n"+
		"      projection: false\n"), "/repo/.orbit/orbits/docs.yaml")
	require.Error(t, err)
	require.ErrorContains(t, err, "field projection not found")
}

func TestParseDefinitionDataKeepsLegacySchemaUnchanged(t *testing.T) {
	t.Parallel()

	sourcePath := "/repo/.orbit/orbits/docs.yaml"
	data := []byte("" +
		"id: docs\n" +
		"description: User-facing docs orbit.\n" +
		"include:\n" +
		"  - docs/**\n" +
		"  - README.md\n" +
		"exclude:\n" +
		"  - docs/generated/**\n")

	definition, err := orbitpkg.ParseDefinitionData(data, sourcePath)
	require.NoError(t, err)
	require.Equal(t, orbitpkg.Definition{
		ID:          "docs",
		Description: "User-facing docs orbit.",
		Include:     []string{"docs/**", "README.md"},
		Exclude:     []string{"docs/generated/**"},
		SourcePath:  sourcePath,
	}, definition)
}

func TestParseDefinitionDataDefaultsLegacyExcludeToEmptySlice(t *testing.T) {
	t.Parallel()

	sourcePath := "/repo/.orbit/orbits/docs.yaml"
	definition, err := orbitpkg.ParseDefinitionData([]byte(""+
		"id: docs\n"+
		"description: User-facing docs orbit.\n"+
		"include:\n"+
		"  - docs/**\n"), sourcePath)
	require.NoError(t, err)
	require.Equal(t, orbitpkg.Definition{
		ID:          "docs",
		Description: "User-facing docs orbit.",
		Include:     []string{"docs/**"},
		Exclude:     []string{},
		SourcePath:  sourcePath,
	}, definition)
}

func TestParseDefinitionDataAcceptsMemberSchemaCompatibilityProjection(t *testing.T) {
	t.Parallel()

	sourcePath := "/repo/.orbit/orbits/docs.yaml"
	definition, err := orbitpkg.ParseDefinitionData([]byte(""+
		"id: docs\n"+
		"description: User-facing docs orbit.\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n"+
		"      exclude:\n"+
		"        - docs/generated/**\n"+
		"  - key: docs-rules\n"+
		"    role: rule\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - .markdownlint.yaml\n"), sourcePath)
	require.NoError(t, err)
	require.Equal(t, orbitpkg.Definition{
		ID:          "docs",
		Description: "User-facing docs orbit.",
		Include:     []string{"docs/**", ".markdownlint.yaml"},
		Exclude:     []string{"docs/generated/**"},
		SourcePath:  sourcePath,
	}, definition)
}

func TestLoadRepositoryConfigAcceptsMemberSchemaDefinitions(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	repo.WriteFile(t, ".orbit/config.yaml", "version: 1\n")
	repo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: User-facing docs orbit.\n"+
		"meta:\n"+
		"  file: .orbit/orbits/docs.yaml\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n")

	config, err := orbitpkg.LoadRepositoryConfig(context.Background(), repo.Root)
	require.NoError(t, err)
	require.Len(t, config.Orbits, 1)
	require.Equal(t, orbitpkg.Definition{
		ID:          "docs",
		Description: "User-facing docs orbit.",
		Include:     []string{"docs/**"},
		Exclude:     []string{},
		SourcePath:  repo.Root + "/.orbit/orbits/docs.yaml",
	}, config.Orbits[0])
}
