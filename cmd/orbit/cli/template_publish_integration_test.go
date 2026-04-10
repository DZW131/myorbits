package cli_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
	"github.com/zack-nova/orbit/cmd/orbit/cli/testutil"
)

func TestTemplatePublishCreatesFixedTemplateBranchFromSourceBranch(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--default", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		OrbitID         string `json:"orbit_id"`
		PublishRef      string `json:"publish_ref"`
		Branch          string `json:"branch"`
		SourceBranch    string `json:"source_branch"`
		DefaultTemplate bool   `json:"default_template"`
		LocalPublish    struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool `json:"attempted"`
			Success   bool `json:"success"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "docs", payload.OrbitID)
	require.Equal(t, "refs/heads/orbit-template/docs", payload.PublishRef)
	require.Equal(t, "orbit-template/docs", payload.Branch)
	require.Equal(t, "main", payload.SourceBranch)
	require.True(t, payload.DefaultTemplate)
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)
	require.False(t, payload.RemotePush.Attempted)
	require.False(t, payload.RemotePush.Success)

	files := splitLines(strings.TrimSpace(repo.Run(t, "ls-tree", "-r", "--name-only", "orbit-template/docs")))
	require.Equal(t, []string{
		".harness/manifest.yaml",
		".harness/orbits/docs.yaml",
		"docs/guide.md",
	}, files)
	_, err = gitpkg.ReadFileAtRev(context.Background(), repo.Root, "orbit-template/docs", ".orbit/source.yaml")
	require.Error(t, err)
}

func TestTemplatePublishValidatesCurrentTemplateBranchDirectly(t *testing.T) {
	t.Parallel()

	repo := seedDirectTemplatePublishRepo(t)

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		OrbitID         string `json:"orbit_id"`
		PublishRef      string `json:"publish_ref"`
		Branch          string `json:"branch"`
		SourceBranch    string `json:"source_branch"`
		DefaultTemplate bool   `json:"default_template"`
		LocalPublish    struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool `json:"attempted"`
			Success   bool `json:"success"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "docs", payload.OrbitID)
	require.Equal(t, "refs/heads/orbit-template/docs", payload.PublishRef)
	require.Equal(t, "orbit-template/docs", payload.Branch)
	require.Equal(t, "orbit-template/docs", payload.SourceBranch)
	require.False(t, payload.DefaultTemplate)
	require.True(t, payload.LocalPublish.Success)
	require.False(t, payload.LocalPublish.Changed)
	require.Empty(t, payload.LocalPublish.Commit)
	require.False(t, payload.RemotePush.Attempted)
	require.False(t, payload.RemotePush.Success)
}

func TestTemplatePublishFailsWhenCurrentTemplateBranchContainsRootAgentsPayload(t *testing.T) {
	t.Parallel()

	repo := seedDirectTemplatePublishRepo(t)
	repo.WriteFile(t, "AGENTS.md", "temporary author entry\n")
	repo.AddAndCommit(t, "add temporary root agents")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, "AGENTS.md")
	require.ErrorContains(t, err, "remove AGENTS.md")
	require.NotErrorIs(t, err, os.ErrNotExist)
}

func TestTemplatePublishFailsWhenCurrentTemplateBranchContainsDriftedBriefArtifact(t *testing.T) {
	t.Parallel()

	repo := seedDirectTemplatePublishRepoWithStructuredBrief(t)
	agentsData, err := orbittemplate.WrapRuntimeAgentsBlock("docs", []byte("Drifted guidance for $project_name\n"))
	require.NoError(t, err)
	repo.WriteFile(t, "AGENTS.md", string(agentsData))
	repo.AddAndCommit(t, "add drifted brief artifact")

	_, _, err = executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, "AGENTS.md")
	require.ErrorContains(t, err, "orbit brief backfill --orbit docs")
	require.ErrorContains(t, err, "drifted")
}

func TestTemplatePublishPrefersHostedBriefDiagnosticsWhenHostedAndLegacySpecsDisagree(t *testing.T) {
	t.Parallel()

	repo := seedDirectTemplatePublishRepoWithStructuredBrief(t)

	legacySpec, err := orbitpkg.DefaultHostedMemberSchemaSpec("docs")
	require.NoError(t, err)
	legacySpec.Description = "Legacy docs orbit"
	require.NotNil(t, legacySpec.Meta)
	legacySpec.Meta.File = ".orbit/orbits/docs.yaml"
	legacySpec.Meta.AgentsTemplate = "Legacy docs guidance\n"
	_, err = orbitpkg.WriteOrbitSpec(repo.Root, legacySpec)
	require.NoError(t, err)

	hostedSpec, err := orbitpkg.DefaultHostedMemberSchemaSpec("docs")
	require.NoError(t, err)
	hostedSpec.Description = "Docs orbit"
	require.NotNil(t, hostedSpec.Meta)
	hostedSpec.Meta.AgentsTemplate = "Docs orbit for $project_name\n"
	_, err = orbitpkg.WriteHostedOrbitSpec(repo.Root, hostedSpec)
	require.NoError(t, err)

	agentsData, err := orbittemplate.WrapRuntimeAgentsBlock("docs", []byte("Docs orbit for $project_name\n"))
	require.NoError(t, err)
	repo.WriteFile(t, "AGENTS.md", string(agentsData))
	repo.AddAndCommit(t, "add hosted-first brief artifact")

	_, _, err = executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, "materialized root AGENTS.md")
	require.ErrorContains(t, err, "remove AGENTS.md")
	require.NotContains(t, err.Error(), "drifted")
	require.NotContains(t, err.Error(), "orbit brief backfill --orbit docs")
}

func TestTemplatePublishPushesCurrentTemplateBranchDirectly(t *testing.T) {
	t.Parallel()

	repo := seedDirectTemplatePublishRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, repo)
	repo.Run(t, "remote", "add", "origin", remoteURL)

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--push", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Branch       string `json:"branch"`
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "orbit-template/docs", payload.Branch)
	require.True(t, payload.LocalPublish.Success)
	require.False(t, payload.LocalPublish.Changed)
	require.Empty(t, payload.LocalPublish.Commit)
	require.True(t, payload.RemotePush.Attempted)
	require.True(t, payload.RemotePush.Success)
	require.Equal(t, "origin", payload.RemotePush.Remote)

	headCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "HEAD"))
	remoteRef := strings.TrimSpace(repo.Run(t, "ls-remote", remoteURL, "refs/heads/orbit-template/docs"))
	require.Contains(t, remoteRef, headCommit)
}

func TestTemplatePublishIgnoresLegacyTemplateManifestWhenBranchManifestIsValid(t *testing.T) {
	t.Parallel()

	repo := seedDirectTemplatePublishRepo(t)
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: orbit_template\n"+
		"template:\n"+
		"  orbit_id: docs\n"+
		"  default_template: true\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: "+strings.TrimSpace(repo.Run(t, "rev-parse", "main"))+"\n"+
		"  created_at: 2026-04-07T12:00:00Z\n")
	repo.WriteFile(t, ".orbit/template.yaml", ""+
		"schema_version: 1\n"+
		"kind: template\n"+
		"template:\n"+
		"  orbit_id: docs\n"+
		"  default_template: false\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: "+strings.TrimSpace(repo.Run(t, "rev-parse", "main"))+"\n"+
		"  created_at: 2026-04-07T12:00:00Z\n"+
		"variables: {}\n")
	repo.AddAndCommit(t, "corrupt branch manifest default template")

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Branch       string `json:"branch"`
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "orbit-template/docs", payload.Branch)
	require.True(t, payload.LocalPublish.Success)
	require.False(t, payload.LocalPublish.Changed)
	require.Empty(t, payload.LocalPublish.Commit)
}

func TestTemplatePublishFailsWithoutSourceManifest(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.Run(t, "rm", ".harness/manifest.yaml")
	repo.AddAndCommit(t, "remove source manifest")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, ".harness/manifest.yaml")
}

func TestTemplatePublishFailsWhenCurrentBranchDiffersFromSourceBranch(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.Run(t, "checkout", "-b", "feature")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, "source branch")
	require.ErrorContains(t, err, "main")
}

func TestTemplatePublishFailsWhenSourceBranchUsesLegacyDefinitionHost(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	repo.Run(t, "branch", "-m", "main")
	repo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: source\n"+
		"source:\n"+
		"  orbit_id: docs\n"+
		"  source_branch: main\n")
	repo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	repo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	repo.AddAndCommit(t, "seed legacy-hosted source repo")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, ".harness/orbits")
	require.ErrorContains(t, err, "init-source")
}

func TestTemplatePublishFailsWhenSourceBranchStillContainsStrayLegacyDefinitions(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.WriteFile(t, ".orbit/orbits/api.yaml", ""+
		"id: api\n"+
		"description: API orbit\n"+
		"include:\n"+
		"  - api/**\n")
	repo.WriteFile(t, "api/spec.md", "API spec\n")
	repo.AddAndCommit(t, "add stray legacy orbit definition")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, ".orbit/orbits")
	require.ErrorContains(t, err, "init-source")
}

func TestTemplatePublishNoOpWhenPublishedTemplateTreeIsUnchanged(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)

	firstStdout, firstStderr, firstErr := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, firstErr)
	require.Empty(t, firstStderr)

	var firstPayload struct {
		LocalPublish struct {
			Commit string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(firstStdout), &firstPayload))
	require.NotEmpty(t, firstPayload.LocalPublish.Commit)

	secondStdout, secondStderr, secondErr := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, secondErr)
	require.Empty(t, secondStderr)

	var secondPayload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(secondStdout), &secondPayload))
	require.True(t, secondPayload.LocalPublish.Success)
	require.False(t, secondPayload.LocalPublish.Changed)
	require.Empty(t, secondPayload.LocalPublish.Commit)

	headCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "orbit-template/docs"))
	require.Equal(t, firstPayload.LocalPublish.Commit, headCommit)
}

func TestTemplatePublishNoOpWhenPublishedBranchOnlyHasLegacyTemplateManifestDrift(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)

	firstStdout, firstStderr, firstErr := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, firstErr)
	require.Empty(t, firstStderr)

	var firstPayload struct {
		LocalPublish struct {
			Commit string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(firstStdout), &firstPayload))
	require.NotEmpty(t, firstPayload.LocalPublish.Commit)

	repo.Run(t, "checkout", "orbit-template/docs")
	repo.WriteFile(t, ".orbit/template.yaml", ""+
		"schema_version: 1\n"+
		"kind: template\n"+
		"template:\n"+
		"  orbit_id: docs\n"+
		"  default_template: true\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-03-31T00:00:00Z\n"+
		"variables: {}\n")
	repo.AddAndCommit(t, "drift legacy template manifest only")
	repo.Run(t, "checkout", "main")

	secondStdout, secondStderr, secondErr := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, secondErr)
	require.Empty(t, secondStderr)

	var secondPayload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(secondStdout), &secondPayload))
	require.True(t, secondPayload.LocalPublish.Success)
	require.False(t, secondPayload.LocalPublish.Changed)
	require.Empty(t, secondPayload.LocalPublish.Commit)

	headCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "orbit-template/docs"))
	require.NotEqual(t, firstPayload.LocalPublish.Commit, headCommit)
}

func TestTemplatePublishRepairsStaleBranchManifest(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)

	_, _, err := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, err)

	repo.Run(t, "checkout", "orbit-template/docs")
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: runtime\n"+
		"runtime:\n"+
		"  id: workspace\n"+
		"  created_at: 2026-04-06T00:00:00Z\n"+
		"  updated_at: 2026-04-06T00:00:00Z\n"+
		"members: []\n")
	repo.AddAndCommit(t, "corrupt branch manifest")
	repo.Run(t, "checkout", "main")

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)

	manifestData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "orbit-template/docs", ".harness/manifest.yaml")
	require.NoError(t, err)
	require.Contains(t, string(manifestData), "kind: orbit_template")
	require.Contains(t, string(manifestData), "orbit_id: docs")
}

func TestTemplatePublishFailsWhenSourceBranchContainsMultipleDefinitions(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.WriteFile(t, ".harness/orbits/api.yaml", ""+
		"id: api\n"+
		"description: API orbit\n"+
		"include:\n"+
		"  - api/**\n")
	repo.WriteFile(t, "api/spec.md", "API spec\n")
	repo.AddAndCommit(t, "add second orbit definition")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, "exactly one")
	require.ErrorContains(t, err, "orbit definition")

	_, _, err = executeCLI(t, repo.Root, "template", "publish", "--orbit", "api")
	require.Error(t, err)
	require.ErrorContains(t, err, "exactly one")
}

func TestTemplatePublishFailsWhenExplicitOrbitDiffersFromSingleSourceOrbit(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)

	_, _, err := executeCLI(t, repo.Root, "template", "publish", "--orbit", "api")
	require.Error(t, err)
	require.ErrorContains(t, err, "single source orbit")
	require.ErrorContains(t, err, "docs")
}

func TestTemplatePublishFailsWhenSourceOrbitIDDoesNotMatchSingleSourceOrbit(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: source\n"+
		"source:\n"+
		"  orbit_id: api\n"+
		"  source_branch: main\n")
	repo.AddAndCommit(t, "mismatch source orbit id")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, "source.orbit_id")
	require.ErrorContains(t, err, "docs")
}

func TestTemplatePublishFailsWhenSourceOrbitIDIsMissing(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: source\n"+
		"source:\n"+
		"  source_branch: main\n")
	repo.AddAndCommit(t, "remove source orbit id")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, "source.orbit_id")
}

func TestTemplatePublishFailsWhenSourceBranchContainsHarnessMetadata(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.WriteFile(t, ".harness/vars.yaml", ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Orbit\n")
	repo.AddAndCommit(t, "add forbidden harness metadata")

	_, _, err := executeCLI(t, repo.Root, "template", "publish")
	require.Error(t, err)
	require.ErrorContains(t, err, ".harness/")
}

func TestTemplatePublishIgnoresLegacyTemplateManifestOnSourceBranch(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	repo.WriteFile(t, ".orbit/template.yaml", ""+
		"schema_version: 1\n"+
		"kind: template\n"+
		"template:\n"+
		"  orbit_id: docs\n"+
		"  default_template: false\n"+
		"  created_from_branch: main\n"+
		"  created_from_commit: abc123\n"+
		"  created_at: 2026-03-31T00:00:00Z\n"+
		"variables: {}\n")
	repo.AddAndCommit(t, "introduce conflicting template marker")

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)

	_, err = gitpkg.ReadFileAtRev(context.Background(), repo.Root, "orbit-template/docs", ".orbit/template.yaml")
	require.Error(t, err)
}

func TestTemplatePublishRejectsRemoteWithoutPush(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)

	_, _, err := executeCLI(t, repo.Root, "template", "publish", "--remote", "origin")
	require.Error(t, err)
	require.ErrorContains(t, err, "--remote")
	require.ErrorContains(t, err, "--push")

	exists, existsErr := gitpkg.LocalBranchExists(context.Background(), repo.Root, "orbit-template/docs")
	require.NoError(t, existsErr)
	require.False(t, exists)
}

func TestTemplatePublishPushesToDefaultOriginWhenSourceBranchEqualsRemote(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, repo)
	repo.Run(t, "remote", "add", "origin", remoteURL)

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--push", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Branch       string `json:"branch"`
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "orbit-template/docs", payload.Branch)
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)
	require.True(t, payload.RemotePush.Attempted)
	require.True(t, payload.RemotePush.Success)
	require.Equal(t, "origin", payload.RemotePush.Remote)

	remoteRef := strings.TrimSpace(repo.Run(t, "ls-remote", remoteURL, "refs/heads/orbit-template/docs"))
	require.Contains(t, remoteRef, payload.LocalPublish.Commit)
}

func TestTemplatePublishPushesToExplicitRemoteWhenRequested(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, repo)
	repo.Run(t, "remote", "add", "upstream", remoteURL)

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--push", "--remote", "upstream", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)
	require.True(t, payload.RemotePush.Attempted)
	require.True(t, payload.RemotePush.Success)
	require.Equal(t, "upstream", payload.RemotePush.Remote)

	remoteRef := strings.TrimSpace(repo.Run(t, "ls-remote", remoteURL, "refs/heads/orbit-template/docs"))
	require.Contains(t, remoteRef, payload.LocalPublish.Commit)
}

func TestTemplatePublishPushesWhenLocalSourceBranchIsAheadOfRemote(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, repo)
	repo.Run(t, "remote", "add", "origin", remoteURL)

	repo.WriteFile(t, "README.md", "local ahead change\n")
	repo.AddAndCommit(t, "advance local main only")

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--push", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)
	require.True(t, payload.RemotePush.Attempted)
	require.True(t, payload.RemotePush.Success)
	require.Equal(t, "origin", payload.RemotePush.Remote)
}

func TestTemplatePublishNoOpStillPushesWhenRequested(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, repo)
	repo.Run(t, "remote", "add", "origin", remoteURL)

	firstStdout, firstStderr, firstErr := executeCLI(t, repo.Root, "template", "publish", "--push", "--json")
	require.NoError(t, firstErr)
	require.Empty(t, firstStderr)

	var firstPayload struct {
		LocalPublish struct {
			Commit string `json:"commit"`
		} `json:"local_publish"`
	}
	require.NoError(t, json.Unmarshal([]byte(firstStdout), &firstPayload))
	require.NotEmpty(t, firstPayload.LocalPublish.Commit)

	secondStdout, secondStderr, secondErr := executeCLI(t, repo.Root, "template", "publish", "--push", "--json")
	require.NoError(t, secondErr)
	require.Empty(t, secondStderr)

	var secondPayload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(secondStdout), &secondPayload))
	require.True(t, secondPayload.LocalPublish.Success)
	require.False(t, secondPayload.LocalPublish.Changed)
	require.Empty(t, secondPayload.LocalPublish.Commit)
	require.True(t, secondPayload.RemotePush.Attempted)
	require.True(t, secondPayload.RemotePush.Success)
	require.Equal(t, "origin", secondPayload.RemotePush.Remote)

	remoteRef := strings.TrimSpace(repo.Run(t, "ls-remote", remoteURL, "refs/heads/orbit-template/docs"))
	require.Contains(t, remoteRef, firstPayload.LocalPublish.Commit)
}

func TestTemplatePublishBlocksPushWhenLocalSourceBranchIsBehindRemote(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, repo)
	repo.Run(t, "remote", "add", "origin", remoteURL)
	advanceRemoteBranch(t, remoteURL, "main", "advance remote main", "README.md", "remote change\n")

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--push", "--json")
	require.Error(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Branch       string `json:"branch"`
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
			Reason    string `json:"reason"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "orbit-template/docs", payload.Branch)
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)
	require.False(t, payload.RemotePush.Attempted)
	require.False(t, payload.RemotePush.Success)
	require.Equal(t, "origin", payload.RemotePush.Remote)
	require.Equal(t, "source_branch_not_up_to_date", payload.RemotePush.Reason)

	headCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "orbit-template/docs"))
	require.Equal(t, payload.LocalPublish.Commit, headCommit)
}

func TestTemplatePublishBlocksPushWhenLocalSourceBranchDivergesFromRemote(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, repo)
	repo.Run(t, "remote", "add", "origin", remoteURL)

	repo.WriteFile(t, "README.md", "local change\n")
	repo.AddAndCommit(t, "local main diverges")
	advanceRemoteBranch(t, remoteURL, "main", "remote main diverges", "docs/guide.md", "remote guide\n")

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--push", "--json")
	require.Error(t, err)
	require.Empty(t, stderr)

	var payload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
			Reason    string `json:"reason"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)
	require.False(t, payload.RemotePush.Attempted)
	require.False(t, payload.RemotePush.Success)
	require.Equal(t, "origin", payload.RemotePush.Remote)
	require.Equal(t, "source_branch_not_up_to_date", payload.RemotePush.Reason)
}

func TestTemplatePublishPreservesLocalPublishWhenRemoteIsMissing(t *testing.T) {
	t.Parallel()

	repo := seedTemplatePublishRepo(t)

	stdout, stderr, err := executeCLI(t, repo.Root, "template", "publish", "--push", "--remote", "missing", "--json")
	require.Error(t, err)
	require.Empty(t, stderr)

	var payload struct {
		LocalPublish struct {
			Success bool   `json:"success"`
			Changed bool   `json:"changed"`
			Commit  string `json:"commit"`
		} `json:"local_publish"`
		RemotePush struct {
			Attempted bool   `json:"attempted"`
			Success   bool   `json:"success"`
			Remote    string `json:"remote"`
			Reason    string `json:"reason"`
		} `json:"remote_push"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.LocalPublish.Success)
	require.True(t, payload.LocalPublish.Changed)
	require.NotEmpty(t, payload.LocalPublish.Commit)
	require.False(t, payload.RemotePush.Success)
	require.Equal(t, "missing", payload.RemotePush.Remote)
	require.NotEmpty(t, payload.RemotePush.Reason)

	headCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "orbit-template/docs"))
	require.Equal(t, payload.LocalPublish.Commit, headCommit)
}

func seedTemplatePublishRepo(t *testing.T) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	repo.Run(t, "branch", "-m", "main")
	repo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: source\n"+
		"source:\n"+
		"  orbit_id: docs\n"+
		"  source_branch: main\n")
	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	repo.WriteFile(t, "README.md", "author docs\n")
	repo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	repo.AddAndCommit(t, "seed template source repo")

	return repo
}

func seedDirectTemplatePublishRepo(t *testing.T) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	repo.Run(t, "branch", "-m", "main")
	repo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	repo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	repo.AddAndCommit(t, "seed runtime repo for direct template publish")

	_, err := orbittemplate.SaveTemplateBranch(context.Background(), orbittemplate.TemplateSaveInput{
		Preview: orbittemplate.TemplateSavePreviewInput{
			RepoRoot:     repo.Root,
			OrbitID:      "docs",
			TargetBranch: "orbit-template/docs",
			Now:          time.Date(2026, time.April, 7, 12, 0, 0, 0, time.UTC),
		},
		Overwrite: true,
	})
	require.NoError(t, err)

	repo.Run(t, "checkout", "orbit-template/docs")

	return repo
}

func seedDirectTemplatePublishRepoWithStructuredBrief(t *testing.T) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	repo.Run(t, "branch", "-m", "main")
	repo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"meta:\n"+
		"  file: .harness/orbits/docs.yaml\n"+
		"  agents_template: |\n"+
		"    Docs orbit for $project_name\n"+
		"  include_in_projection: true\n"+
		"  include_in_write: true\n"+
		"  include_in_export: true\n"+
		"  include_description_in_orchestration: true\n"+
		"members:\n"+
		"  - key: docs-content\n"+
		"    role: subject\n"+
		"    paths:\n"+
		"      include:\n"+
		"        - docs/**\n")
	repo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	repo.AddAndCommit(t, "seed runtime repo for structured direct template publish")

	_, err := orbittemplate.SaveTemplateBranch(context.Background(), orbittemplate.TemplateSaveInput{
		Preview: orbittemplate.TemplateSavePreviewInput{
			RepoRoot:     repo.Root,
			OrbitID:      "docs",
			TargetBranch: "orbit-template/docs",
			Now:          time.Date(2026, time.April, 7, 12, 0, 0, 0, time.UTC),
		},
		Overwrite: true,
	})
	require.NoError(t, err)

	repo.Run(t, "checkout", "orbit-template/docs")

	return repo
}

func advanceRemoteBranch(t *testing.T, remoteURL string, branch string, message string, path string, contents string) {
	t.Helper()

	cloneRoot := filepath.Join(t.TempDir(), "clone")
	command := exec.Command("git", "clone", remoteURL, cloneRoot)
	output, err := command.CombinedOutput()
	require.NoError(t, err, "git clone failed:\n%s", string(output))

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = cloneRoot
		result, runErr := cmd.CombinedOutput()
		require.NoError(t, runErr, "git %s failed:\n%s", strings.Join(args, " "), string(result))
	}

	run("config", "user.name", "Orbit Test")
	run("config", "user.email", "orbit@example.com")
	run("checkout", branch)

	target := filepath.Join(cloneRoot, filepath.FromSlash(path))
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
	require.NoError(t, os.WriteFile(target, []byte(contents), 0o600))
	run("add", "--", path)
	run("commit", "-m", message)
	run("push", "origin", branch)
}
