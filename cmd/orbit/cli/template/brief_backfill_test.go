package orbittemplate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
)

func TestReverseVariableizeBriefReplacesUniqueRuntimeValues(t *testing.T) {
	t.Parallel()

	result, err := ReverseVariableizeBrief([]byte("Welcome to Acme.\nOpen https://docs.acme.test.\n"), map[string]bindings.VariableBinding{
		"project_name": {
			Value: "Acme",
		},
		"docs_url": {
			Value: "https://docs.acme.test",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "Welcome to $project_name.\nOpen $docs_url.\n", string(result.Content))
	require.ElementsMatch(t, []ReplacementSummary{
		{
			Variable: "docs_url",
			Literal:  "https://docs.acme.test",
			Count:    1,
		},
		{
			Variable: "project_name",
			Literal:  "Acme",
			Count:    1,
		},
	}, result.Replacements)
}

func TestReverseVariableizeBriefRejectsDuplicateLiteralBindings(t *testing.T) {
	t.Parallel()

	_, err := ReverseVariableizeBrief([]byte("Welcome to Acme.\n"), map[string]bindings.VariableBinding{
		"project_name": {
			Value: "Acme",
		},
		"product_name": {
			Value: "Acme",
		},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "reverse replacement is ambiguous")
	require.ErrorContains(t, err, "Acme")
	require.ErrorContains(t, err, "product_name")
	require.ErrorContains(t, err, "project_name")
}

func TestReverseVariableizeBriefRejectsOverlappingLiteralMatches(t *testing.T) {
	t.Parallel()

	_, err := ReverseVariableizeBrief([]byte("Follow the Acme Docs workflow.\n"), map[string]bindings.VariableBinding{
		"project_name": {
			Value: "Acme",
		},
		"workflow_name": {
			Value: "Acme Docs",
		},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "reverse replacement is ambiguous")
	require.ErrorContains(t, err, "Acme")
	require.ErrorContains(t, err, "Acme Docs")
}
