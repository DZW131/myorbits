package harness

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

func TestBuildBindingsPlanMergesPreviewsAndPrefillsRepoValues(t *testing.T) {
	t.Parallel()

	result, err := BuildBindingsPlan([]orbittemplate.BindingsInitPreview{
		{
			Source: orbittemplate.Source{
				SourceKind:     orbittemplate.InstallSourceKindLocalBranch,
				SourceRef:      "orbit-template/docs",
				TemplateCommit: "abc123",
			},
			Manifest: orbittemplate.Manifest{
				SchemaVersion: 1,
				Kind:          orbittemplate.TemplateKind,
				Template: orbittemplate.Metadata{
					OrbitID:           "docs",
					CreatedFromBranch: "main",
					CreatedFromCommit: "abc123",
					CreatedAt:         time.Date(2026, time.April, 9, 10, 0, 0, 0, time.UTC),
				},
				Variables: map[string]orbittemplate.VariableSpec{
					"project_name": {Description: "Product title", Required: true},
				},
			},
		},
		{
			Source: orbittemplate.Source{
				SourceKind:     orbittemplate.InstallSourceKindLocalBranch,
				SourceRef:      "orbit-template/cmd",
				TemplateCommit: "def456",
			},
			Manifest: orbittemplate.Manifest{
				SchemaVersion: 1,
				Kind:          orbittemplate.TemplateKind,
				Template: orbittemplate.Metadata{
					OrbitID:           "cmd",
					CreatedFromBranch: "main",
					CreatedFromCommit: "def456",
					CreatedAt:         time.Date(2026, time.April, 9, 10, 0, 0, 0, time.UTC),
				},
				Variables: map[string]orbittemplate.VariableSpec{
					"project_name": {Description: "Product title", Required: true},
					"binary_name":  {Description: "CLI binary", Required: true},
				},
			},
		},
	}, bindings.VarsFile{
		SchemaVersion: 1,
		Variables: map[string]bindings.VariableBinding{
			"project_name": {
				Value:       "Orbit",
				Description: "Product title",
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, []BindingsPlanSource{
		{
			Kind:    orbittemplate.InstallSourceKindLocalBranch,
			Ref:     "orbit-template/docs",
			Commit:  "abc123",
			OrbitID: "docs",
		},
		{
			Kind:    orbittemplate.InstallSourceKindLocalBranch,
			Ref:     "orbit-template/cmd",
			Commit:  "def456",
			OrbitID: "cmd",
		},
	}, result.Sources)
	require.Equal(t, []string{"binary_name"}, result.MissingRequired)
	require.Equal(t, []string{"project_name"}, result.ReusedValues)
	require.Equal(t, bindings.VarsFile{
		SchemaVersion: 1,
		Variables: map[string]bindings.VariableBinding{
			"binary_name": {
				Value:       "",
				Description: "CLI binary",
			},
			"project_name": {
				Value:       "Orbit",
				Description: "Product title",
			},
		},
	}, result.Bindings)
}

func TestBuildBindingsPlanFailsOnVariableDescriptionConflict(t *testing.T) {
	t.Parallel()

	_, err := BuildBindingsPlan([]orbittemplate.BindingsInitPreview{
		{
			Source: orbittemplate.Source{SourceKind: orbittemplate.InstallSourceKindLocalBranch, SourceRef: "orbit-template/docs"},
			Manifest: orbittemplate.Manifest{
				SchemaVersion: 1,
				Kind:          orbittemplate.TemplateKind,
				Template: orbittemplate.Metadata{
					OrbitID:           "docs",
					CreatedFromBranch: "main",
					CreatedFromCommit: "abc123",
					CreatedAt:         time.Date(2026, time.April, 9, 10, 0, 0, 0, time.UTC),
				},
				Variables: map[string]orbittemplate.VariableSpec{
					"project_name": {Description: "Product title", Required: true},
				},
			},
		},
		{
			Source: orbittemplate.Source{SourceKind: orbittemplate.InstallSourceKindLocalBranch, SourceRef: "orbit-template/cmd"},
			Manifest: orbittemplate.Manifest{
				SchemaVersion: 1,
				Kind:          orbittemplate.TemplateKind,
				Template: orbittemplate.Metadata{
					OrbitID:           "cmd",
					CreatedFromBranch: "main",
					CreatedFromCommit: "def456",
					CreatedAt:         time.Date(2026, time.April, 9, 10, 0, 0, 0, time.UTC),
				},
				Variables: map[string]orbittemplate.VariableSpec{
					"project_name": {Description: "CLI title", Required: true},
				},
			},
		},
	}, bindings.VarsFile{SchemaVersion: 1, Variables: map[string]bindings.VariableBinding{}})
	require.Error(t, err)
	require.ErrorContains(t, err, `variable conflict for "project_name"`)
	require.ErrorContains(t, err, `sources: orbit-template/cmd, orbit-template/docs`)
}
