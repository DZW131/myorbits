package harness

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zack-nova/orbit/cmd/orbit/cli/testutil"
)

func seedHarnessTemplateInstallSourceRepo(t *testing.T) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	repo.Run(t, "branch", "-m", "main")
	repo.WriteFile(t, ".harness/template.yaml", ""+
		"schema_version: 1\n"+
		"kind: harness_template\n"+
		"template:\n"+
		"  harness_id: workspace\n"+
		"  default_template: false\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-04-03T00:00:00Z\n"+
		"  includes_root_agents: false\n"+
		"members:\n"+
		"  - orbit_id: workspace\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    required: true\n")
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: harness_template\n"+
		"template:\n"+
		"  harness_id: workspace\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-04-03T00:00:00Z\n"+
		"members:\n"+
		"  - orbit_id: workspace\n")
	repo.WriteFile(t, ".harness/orbits/workspace.yaml", ""+
		"id: workspace\n"+
		"description: Workspace orbit\n"+
		"include:\n"+
		"  - docs/**\n"+
		"  - schema/**\n")
	repo.WriteFile(t, "docs/guide.md", "$project_name guide\n")
	repo.WriteFile(t, "schema/example.schema.json", "{\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\n  \"$id\": \"workspace/example.schema.json\",\n  \"title\": \"$project_name\"\n}\n")
	repo.AddAndCommit(t, "seed harness template source")

	return repo
}

func TestResolveLocalTemplateInstallSourceIgnoresNonMarkdownVariableSyntax(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateInstallSourceRepo(t)

	source, err := ResolveLocalTemplateInstallSource(context.Background(), repo.Root, "HEAD")
	require.NoError(t, err)
	require.Equal(t, "HEAD", source.Ref)
	require.Equal(t, []string{"workspace"}, source.MemberIDs())
	require.Contains(t, source.FilePaths(), "schema/example.schema.json")
	require.Equal(t, time.Date(2026, time.April, 3, 0, 0, 0, 0, time.UTC), source.Manifest.Template.CreatedAt)
}

func TestResolveLocalTemplateInstallSourceIgnoresLegacyOrbitTemplateManifest(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateInstallSourceRepo(t)
	repo.WriteFile(t, ".orbit/template.yaml", ""+
		"schema_version: 1\n"+
		"kind: template\n"+
		"template:\n"+
		"  orbit_id: docs\n"+
		"  default_template: true\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: stray999\n"+
		"  created_at: 2026-04-03T00:00:00Z\n"+
		"variables: {}\n")
	repo.AddAndCommit(t, "add stray legacy orbit template manifest")

	source, err := ResolveLocalTemplateInstallSource(context.Background(), repo.Root, "HEAD")
	require.NoError(t, err)
	require.Equal(t, "HEAD", source.Ref)
	require.Equal(t, []string{"workspace"}, source.MemberIDs())
	require.Contains(t, source.FilePaths(), "docs/guide.md")
}

func TestResolveLocalTemplateInstallSourceRejectsMissingBranchManifest(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateInstallSourceRepo(t)
	repo.Run(t, "rm", ".harness/manifest.yaml")
	repo.AddAndCommit(t, "remove branch manifest")

	_, err := ResolveLocalTemplateInstallSource(context.Background(), repo.Root, "HEAD")
	require.Error(t, err)
	require.ErrorContains(t, err, ".harness/manifest.yaml")
	require.ErrorContains(t, err, "valid harness template branch")

	var notFoundErr *LocalTemplateInstallSourceNotFoundError
	require.NotErrorAs(t, err, &notFoundErr)
}

func TestResolveLocalTemplateInstallSourceRejectsBranchManifestTemplateMismatch(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateInstallSourceRepo(t)
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: harness_template\n"+
		"template:\n"+
		"  harness_id: another\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-04-03T00:00:00Z\n"+
		"members:\n"+
		"  - orbit_id: workspace\n")
	repo.AddAndCommit(t, "corrupt branch manifest harness id")

	_, err := ResolveLocalTemplateInstallSource(context.Background(), repo.Root, "HEAD")
	require.Error(t, err)
	require.ErrorContains(t, err, ".harness/manifest.yaml")
	require.ErrorContains(t, err, ".harness/template.yaml")
	require.ErrorContains(t, err, "harness_id")
}

func TestResolveLocalTemplateInstallSourceRejectsBranchManifestDefaultTemplateMismatch(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateInstallSourceRepo(t)
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: harness_template\n"+
		"template:\n"+
		"  harness_id: workspace\n"+
		"  default_template: true\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-04-03T00:00:00Z\n"+
		"members:\n"+
		"  - orbit_id: workspace\n")
	repo.AddAndCommit(t, "corrupt branch manifest default template")

	_, err := ResolveLocalTemplateInstallSource(context.Background(), repo.Root, "HEAD")
	require.Error(t, err)
	require.ErrorContains(t, err, ".harness/manifest.yaml")
	require.ErrorContains(t, err, ".harness/template.yaml")
	require.ErrorContains(t, err, "default_template")
}

func TestEnumerateRemoteTemplateInstallSourcesRejectsBranchesWithoutBranchManifest(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateInstallSourceRepo(t)
	sourceRepo.Run(t, "branch", "harness-template/legacy-only")
	sourceRepo.Run(t, "checkout", "harness-template/legacy-only")
	sourceRepo.Run(t, "rm", ".harness/manifest.yaml")
	sourceRepo.AddAndCommit(t, "remove branch manifest from legacy-only branch")
	sourceRepo.Run(t, "checkout", "main")

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := testutil.NewRepo(t)

	candidates, err := EnumerateRemoteTemplateInstallSources(context.Background(), runtimeRepo.Root, remoteURL)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.Equal(t, "main", candidates[0].Branch)
}

func TestEnumerateRemoteTemplateInstallSourcesIgnoresLegacyOrbitTemplateManifest(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateInstallSourceRepo(t)
	sourceRepo.Run(t, "branch", "harness-template/with-legacy-template")
	sourceRepo.Run(t, "checkout", "harness-template/with-legacy-template")
	sourceRepo.WriteFile(t, ".orbit/template.yaml", ""+
		"schema_version: 1\n"+
		"kind: template\n"+
		"template:\n"+
		"  orbit_id: docs\n"+
		"  default_template: true\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: stray999\n"+
		"  created_at: 2026-04-03T00:00:00Z\n"+
		"variables: {}\n")
	sourceRepo.AddAndCommit(t, "add stray legacy orbit template manifest")
	sourceRepo.Run(t, "checkout", "main")

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := testutil.NewRepo(t)

	candidates, err := EnumerateRemoteTemplateInstallSources(context.Background(), runtimeRepo.Root, remoteURL)
	require.NoError(t, err)
	require.Len(t, candidates, 2)
	require.Equal(t, "harness-template/with-legacy-template", candidates[0].Branch)
	require.Equal(t, "main", candidates[1].Branch)
}

func TestEnumerateRemoteTemplateInstallSourcesRejectsBranchesWithDefaultTemplateMismatch(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateInstallSourceRepo(t)
	sourceRepo.Run(t, "branch", "harness-template/mismatch")
	sourceRepo.Run(t, "checkout", "harness-template/mismatch")
	sourceRepo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: harness_template\n"+
		"template:\n"+
		"  harness_id: workspace\n"+
		"  default_template: true\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-04-03T00:00:00Z\n"+
		"members:\n"+
		"  - orbit_id: workspace\n")
	sourceRepo.AddAndCommit(t, "corrupt branch manifest default template")
	sourceRepo.Run(t, "checkout", "main")

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := testutil.NewRepo(t)

	candidates, err := EnumerateRemoteTemplateInstallSources(context.Background(), runtimeRepo.Root, remoteURL)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.Equal(t, "main", candidates[0].Branch)
}

func TestResolveRemoteTemplateInstallSourceExplicitRefRejectsMissingBranchManifest(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateInstallSourceRepo(t)
	sourceRepo.Run(t, "branch", "harness-template/legacy-only")
	sourceRepo.Run(t, "checkout", "harness-template/legacy-only")
	sourceRepo.Run(t, "rm", ".harness/manifest.yaml")
	sourceRepo.AddAndCommit(t, "remove branch manifest from legacy-only branch")
	sourceRepo.Run(t, "checkout", "main")

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := testutil.NewRepo(t)

	_, _, err := ResolveRemoteTemplateInstallSource(context.Background(), runtimeRepo.Root, remoteURL, "harness-template/legacy-only")
	require.Error(t, err)
	require.ErrorContains(t, err, ".harness/manifest.yaml")

	var notFoundErr *RemoteTemplateInstallNotFoundError
	require.NotErrorAs(t, err, &notFoundErr)
}
