package bindings

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergePrefersBindingsFileThenRepoVarsThenInteractive(t *testing.T) {
	t.Parallel()

	result, err := Merge(MergeInput{
		Declared: map[string]VariableDeclaration{
			"owner_name": {
				Description: "负责人姓名",
				Required:    true,
			},
			"project_name": {
				Description: "项目名称",
				Required:    true,
			},
			"service_url": {
				Description: "服务地址",
				Required:    true,
			},
			"support_email": {
				Description: "支持邮箱",
				Required:    false,
			},
		},
		BindingsFile: map[string]VariableBinding{
			"project_name": {
				Value:       "From File",
				Description: "file description should not override declaration",
			},
		},
		RepoVars: map[string]VariableBinding{
			"project_name": {
				Value: "From Repo",
			},
			"service_url": {
				Value: "https://repo.example.com",
			},
			"support_email": {
				Value: "support@example.com",
			},
		},
		FillIn: map[string]VariableBinding{
			"owner_name": {
				Value: "Yixin",
			},
			"service_url": {
				Value: "https://interactive.example.com",
			},
		},
		FillSource: SourceInteractive,
	})
	require.NoError(t, err)

	require.Equal(t, map[string]ResolvedBinding{
		"owner_name": {
			Value:       "Yixin",
			Description: "负责人姓名",
			Required:    true,
			Source:      SourceInteractive,
		},
		"project_name": {
			Value:       "From File",
			Description: "项目名称",
			Required:    true,
			Source:      SourceBindingsFile,
		},
		"service_url": {
			Value:       "https://repo.example.com",
			Description: "服务地址",
			Required:    true,
			Source:      SourceRepoVars,
		},
		"support_email": {
			Value:       "support@example.com",
			Description: "支持邮箱",
			Required:    false,
			Source:      SourceRepoVars,
		},
	}, result.Resolved)
	require.Empty(t, result.Unresolved)
}

func TestMergeReturnsStableUnresolvedForMissingRequiredVariablesOnly(t *testing.T) {
	t.Parallel()

	result, err := Merge(MergeInput{
		Declared: map[string]VariableDeclaration{
			"api_key": {
				Description: "API 密钥",
				Required:    true,
			},
			"optional_note": {
				Description: "附加说明",
				Required:    false,
			},
			"zone": {
				Description: "可用区",
				Required:    true,
			},
		},
	})
	require.NoError(t, err)

	require.Empty(t, result.Resolved)
	require.Equal(t, []UnresolvedBinding{
		{
			Name:        "api_key",
			Description: "API 密钥",
			Required:    true,
		},
		{
			Name:        "zone",
			Description: "可用区",
			Required:    true,
		},
	}, result.Unresolved)
}

func TestMergeIgnoresUndeclaredInputs(t *testing.T) {
	t.Parallel()

	result, err := Merge(MergeInput{
		Declared: map[string]VariableDeclaration{
			"project_name": {
				Description: "项目名称",
				Required:    true,
			},
		},
		BindingsFile: map[string]VariableBinding{
			"project_name": {Value: "Orbit"},
			"ignored_file": {Value: "ignored"},
		},
		RepoVars: map[string]VariableBinding{
			"ignored_repo": {Value: "ignored"},
		},
		FillIn: map[string]VariableBinding{
			"ignored_fill": {Value: "ignored"},
		},
		FillSource: SourceEditor,
	})
	require.NoError(t, err)

	require.Equal(t, map[string]ResolvedBinding{
		"project_name": {
			Value:       "Orbit",
			Description: "项目名称",
			Required:    true,
			Source:      SourceBindingsFile,
		},
	}, result.Resolved)
	require.Empty(t, result.Unresolved)
}

func TestMergePreservesDescriptionFromDeclarationOrFallbackSource(t *testing.T) {
	t.Parallel()

	result, err := Merge(MergeInput{
		Declared: map[string]VariableDeclaration{
			"owner_name": {
				Required: true,
			},
			"project_name": {
				Description: "项目名称",
				Required:    true,
			},
		},
		BindingsFile: map[string]VariableBinding{
			"project_name": {
				Value:       "Orbit",
				Description: "should not replace declared description",
			},
		},
		RepoVars: map[string]VariableBinding{
			"owner_name": {
				Value:       "Miles",
				Description: "负责人姓名",
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, "项目名称", result.Resolved["project_name"].Description)
	require.Equal(t, "负责人姓名", result.Resolved["owner_name"].Description)
}

func TestMergeRejectsUnknownFillSource(t *testing.T) {
	t.Parallel()

	_, err := Merge(MergeInput{
		Declared: map[string]VariableDeclaration{
			"project_name": {
				Required: true,
			},
		},
		FillIn: map[string]VariableBinding{
			"project_name": {
				Value: "Orbit",
			},
		},
		FillSource: MergeSource("clipboard"),
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "fill source")
}

func TestMergeAllowsEmptyDeclaredVariables(t *testing.T) {
	t.Parallel()

	result, err := Merge(MergeInput{})
	require.NoError(t, err)
	require.Empty(t, result.Resolved)
	require.Empty(t, result.Unresolved)
}
