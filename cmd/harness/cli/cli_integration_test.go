package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	harnesscli "github.com/zack-nova/orbit/cmd/harness/cli"
	harnesscommands "github.com/zack-nova/orbit/cmd/harness/cli/commands"
	orbitcli "github.com/zack-nova/orbit/cmd/orbit/cli"
	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
	orbitcommands "github.com/zack-nova/orbit/cmd/orbit/cli/commands"
	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	harnesspkg "github.com/zack-nova/orbit/cmd/orbit/cli/harness"
	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
	"github.com/zack-nova/orbit/cmd/orbit/cli/testutil"
)

func TestHarnessInitCreatesManifestAndHostedOrbitsDir(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "initialized harness in "+repo.Root+"\n", stdout)

	manifestFile, err := harnesspkg.LoadManifestFile(repo.Root)
	require.NoError(t, err)
	require.Equal(t, harnesspkg.ManifestKindRuntime, manifestFile.Kind)
	require.Equal(t, []harnesspkg.ManifestMember{}, manifestFile.Members)

	orbitSpecsDirInfo, err := os.Stat(filepath.Join(repo.Root, ".harness", "orbits"))
	require.NoError(t, err)
	require.True(t, orbitSpecsDirInfo.IsDir())

	_, err = os.Stat(filepath.Join(repo.Root, ".harness", "runtime.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = os.Stat(filepath.Join(repo.Root, ".orbit", "config.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestHarnessCreateInitializesNewRepository(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	resolvedTargetPath, err := filepath.EvalSymlinks(baseDir)
	require.NoError(t, err)
	resolvedTargetPath = filepath.Join(resolvedTargetPath, "Project A")

	stdout, stderr, err := executeHarnessCLI(t, baseDir, "create", "Project A", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessRoot    string `json:"harness_root"`
		ManifestPath   string `json:"manifest_path"`
		OrbitsDir      string `json:"orbits_dir"`
		GitInitialized bool   `json:"git_initialized"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, resolvedTargetPath, payload.HarnessRoot)
	require.Equal(t, filepath.Join(resolvedTargetPath, ".harness", "manifest.yaml"), payload.ManifestPath)
	require.Equal(t, filepath.Join(resolvedTargetPath, ".harness", "orbits"), payload.OrbitsDir)
	require.True(t, payload.GitInitialized)

	_, err = os.Stat(filepath.Join(resolvedTargetPath, ".git"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(resolvedTargetPath, ".harness", "runtime.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestHarnessRootPrintsResolvedRootFromSubdirectory(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	subdir := filepath.Join(repo.Root, "nested", "path")
	require.NoError(t, os.MkdirAll(subdir, 0o755))

	stdout, stderr, err := executeHarnessCLI(t, subdir, "root")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, repo.Root+"\n", stdout)
}

func TestHarnessInspectReportsZeroMemberRuntime(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "inspect", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessRoot       string   `json:"harness_root"`
		HarnessID         string   `json:"harness_id"`
		MemberCount       int      `json:"member_count"`
		Members           []string `json:"members"`
		VarsCount         int      `json:"vars_count"`
		InstallCount      int      `json:"install_count"`
		BundleCount       int      `json:"bundle_count"`
		CurrentProjection string   `json:"current_projection"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), payload.HarnessID)
	require.Equal(t, 0, payload.MemberCount)
	require.Empty(t, payload.Members)
	require.Equal(t, 0, payload.VarsCount)
	require.Equal(t, 0, payload.InstallCount)
	require.Equal(t, 0, payload.BundleCount)
	require.Empty(t, payload.CurrentProjection)
}

func TestHarnessInspectTextOutputForZeroMemberRuntime(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "inspect")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, ""+
		"harness_root: "+repo.Root+"\n"+
		"harness_id: "+harnesspkg.DefaultHarnessIDForPath(repo.Root)+"\n"+
		"harness_name: "+filepath.Base(repo.Root)+"\n"+
		"member_count: 0\n"+
		"members: none\n"+
		"vars_count: 0\n"+
		"install_count: 0\n"+
		"bundle_count: 0\n"+
		"current_projection: none\n", stdout)
}

func TestHarnessInspectReportsBundleCount(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	_, err = harnesspkg.WriteBundleRecord(repo.Root, harnesspkg.BundleRecord{
		SchemaVersion:      1,
		HarnessID:          "workspace",
		Template:           orbittemplate.Source{SourceKind: orbittemplate.InstallSourceKindLocalBranch, SourceRepo: "", SourceRef: "harness-template/workspace", TemplateCommit: "abc123"},
		MemberIDs:          []string{"docs"},
		AppliedAt:          time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
		IncludesRootAgents: false,
		OwnedPaths:         []string{"docs/guide.md"},
	})
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "inspect", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		InstallCount int `json:"install_count"`
		BundleCount  int `json:"bundle_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, 0, payload.InstallCount)
	require.Equal(t, 1, payload.BundleCount)
}

func TestHarnessCheckSucceedsForZeroMemberRuntime(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessRoot  string `json:"harness_root"`
		HarnessID    string `json:"harness_id"`
		OK           bool   `json:"ok"`
		FindingCount int    `json:"finding_count"`
		Findings     []struct {
			Kind    string `json:"kind"`
			OrbitID string `json:"orbit_id"`
			Path    string `json:"path"`
			Message string `json:"message"`
		} `json:"findings"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), payload.HarnessID)
	require.True(t, payload.OK)
	require.Zero(t, payload.FindingCount)
	require.Empty(t, payload.Findings)
}

func TestHarnessCheckTextOutputForZeroMemberRuntime(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, ""+
		"harness_root: "+repo.Root+"\n"+
		"harness_id: "+harnesspkg.DefaultHarnessIDForPath(repo.Root)+"\n"+
		"ok: true\n"+
		"finding_count: 0\n"+
		"findings: none\n", stdout)
}

func TestHarnessCheckIgnoresLegacyRuntimeFileWhenManifestIsValid(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/runtime.yaml", ""+
		"schema_version: nope\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.True(t, payload.OK)
	require.Zero(t, payload.FindingCount)
	require.Empty(t, payload.Findings)
}

func TestHarnessCheckReportsManifestSchemaInvalidForDuplicateMembers(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/manifest.yaml", ""+
		"schema_version: 1\n"+
		"kind: runtime\n"+
		"runtime:\n"+
		"  id: workspace\n"+
		"  created_at: 2026-03-25T10:00:00Z\n"+
		"  updated_at: 2026-03-25T10:00:00Z\n"+
		"members:\n"+
		"  - orbit_id: docs\n"+
		"    source: manual\n"+
		"    added_at: 2026-03-25T10:00:00Z\n"+
		"  - orbit_id: docs\n"+
		"    source: manual\n"+
		"    added_at: 2026-03-25T10:05:00Z\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	require.Len(t, payload.Findings, 1)
	require.Equal(t, "manifest_schema_invalid", payload.Findings[0].Kind)
	require.Equal(t, ".harness/manifest.yaml", payload.Findings[0].Path)
	require.Contains(t, payload.Findings[0].Message, "members[1].orbit_id must be unique")
}

func TestHarnessCheckReportsMissingDefinitionForRuntimeMember(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	runtimeFile.Members = append(runtimeFile.Members, harnesspkg.RuntimeMember{
		OrbitID: "docs",
		Source:  harnesspkg.MemberSourceManual,
		AddedAt: time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC),
	})
	runtimeFile.Harness.UpdatedAt = time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC)
	_, err = harnesspkg.WriteRuntimeFile(repo.Root, runtimeFile)
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	finding := requireHarnessCheckFinding(t, payload, "missing_definition")
	require.Equal(t, "docs", finding.OrbitID)
	require.Equal(t, ".harness/orbits/docs.yaml", finding.Path)
	require.Contains(t, finding.Message, "definition")
}

func TestHarnessCheckReportsMissingBundleRecordForBundleBackedMember(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	runtimeFile.Members = append(runtimeFile.Members, harnesspkg.RuntimeMember{
		OrbitID: "docs",
		Source:  harnesspkg.MemberSourceInstallBundle,
		AddedAt: time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
	})
	_, err = harnesspkg.WriteRuntimeFile(repo.Root, runtimeFile)
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	finding := requireHarnessCheckFinding(t, payload, "bundle_member_mismatch")
	require.Equal(t, "docs", finding.OrbitID)
	require.Contains(t, finding.Message, "bundle-backed member")
}

func TestHarnessCheckSucceedsForBundleBackedMemberWithMatchingBundleRecord(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	runtimeFile.Members = append(runtimeFile.Members, harnesspkg.RuntimeMember{
		OrbitID: "docs",
		Source:  harnesspkg.MemberSourceInstallBundle,
		AddedAt: time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
	})
	_, err = harnesspkg.WriteRuntimeFile(repo.Root, runtimeFile)
	require.NoError(t, err)

	_, err = harnesspkg.WriteBundleRecord(repo.Root, harnesspkg.BundleRecord{
		SchemaVersion:      1,
		HarnessID:          "workspace",
		Template:           orbittemplate.Source{SourceKind: orbittemplate.InstallSourceKindLocalBranch, SourceRepo: "", SourceRef: "harness-template/workspace", TemplateCommit: "abc123"},
		MemberIDs:          []string{"docs"},
		AppliedAt:          time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
		IncludesRootAgents: false,
		OwnedPaths:         []string{"docs/guide.md"},
	})
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.True(t, payload.OK)
	require.Zero(t, payload.FindingCount)
}

func TestHarnessCheckReportsInstallMemberMismatchWithoutInstallRecord(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"include:\n"+
		"  - docs/**\n")

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	runtimeFile.Members = append(runtimeFile.Members, harnesspkg.RuntimeMember{
		OrbitID: "docs",
		Source:  harnesspkg.MemberSourceInstallOrbit,
		AddedAt: time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC),
	})
	runtimeFile.Harness.UpdatedAt = time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC)
	_, err = harnesspkg.WriteRuntimeFile(repo.Root, runtimeFile)
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	finding := requireHarnessCheckFinding(t, payload, "install_member_mismatch")
	require.Equal(t, "docs", finding.OrbitID)
	require.Equal(t, ".harness/installs/docs.yaml", finding.Path)
	require.Contains(t, finding.Message, "missing install record")
}

func TestHarnessCheckReportsInstallPathMismatch(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/installs/wrong.yaml", ""+
		"schema_version: 1\n"+
		"orbit_id: docs\n"+
		"template:\n"+
		"  source_kind: local_branch\n"+
		"  source_repo: \"\"\n"+
		"  source_ref: orbit-template/docs\n"+
		"  template_commit: deadbeef\n"+
		"applied_at: 2026-03-21T12:00:00Z\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	finding := requireHarnessCheckFinding(t, payload, "install_path_mismatch")
	require.Equal(t, "docs", finding.OrbitID)
	require.Equal(t, ".harness/installs/wrong.yaml", finding.Path)
	require.Contains(t, finding.Message, "path")
}

func TestHarnessCheckReportsInstallRecordInvalidSeparatelyFromPathMismatch(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/installs/broken.yaml", ""+
		"schema_version: 1\n"+
		"orbit_id: docs\n"+
		"template:\n"+
		"  source_kind: local_branch\n"+
		"  source_ref: [broken\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	finding := requireHarnessCheckFinding(t, payload, "install_record_invalid")
	require.Empty(t, finding.OrbitID)
	require.Equal(t, ".harness/installs/broken.yaml", finding.Path)
	require.Contains(t, finding.Message, "invalid")

	stdout, stderr, err = executeHarnessCLI(t, repo.Root, "check")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "finding: install_record_invalid orbit_id=- path=.harness/installs/broken.yaml")
}

func TestHarnessCheckReportsDefinitionAndRuntimeFileDrift(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, _, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Drifted docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	repo.WriteFile(t, "docs/guide.md", "Locally drifted guide\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	definitionFinding := requireHarnessCheckFinding(t, payload, "definition_drift")
	require.Equal(t, "docs", definitionFinding.OrbitID)
	require.Equal(t, ".harness/orbits/docs.yaml", definitionFinding.Path)
	runtimeFinding := requireHarnessCheckFinding(t, payload, "runtime_file_drift")
	require.Equal(t, "docs", runtimeFinding.OrbitID)
	require.Equal(t, "docs/guide.md", runtimeFinding.Path)
}

func TestHarnessCheckReportsProvenanceUnresolvable(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, _, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/installs/docs.yaml", ""+
		"schema_version: 1\n"+
		"orbit_id: docs\n"+
		"template:\n"+
		"  source_kind: local_branch\n"+
		"  source_repo: \"\"\n"+
		"  source_ref: orbit-template/docs\n"+
		"  template_commit: deadbeefdeadbeefdeadbeefdeadbeefdeadbeef\n"+
		"applied_at: 2026-03-21T12:40:00Z\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	payload := decodeHarnessCheckPayload(t, stdout)
	require.False(t, payload.OK)
	finding := requireHarnessCheckFinding(t, payload, "provenance_unresolvable")
	require.Equal(t, "docs", finding.OrbitID)
	require.Equal(t, ".harness/installs/docs.yaml", finding.Path)
}

func executeHarnessCLI(t *testing.T, workingDir string, args ...string) (string, string, error) {
	t.Helper()

	rootCmd := harnesscli.NewRootCommand()
	rootCmd.SetArgs(args)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err := rootCmd.ExecuteContext(harnesscommands.WithWorkingDir(context.Background(), workingDir))

	return stdout.String(), stderr.String(), err
}

func executeOrbitCLI(t *testing.T, workingDir string, args ...string) (string, string, error) {
	t.Helper()

	rootCmd := orbitcli.NewRootCommand()
	rootCmd.SetArgs(args)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	err := rootCmd.ExecuteContext(orbitcommands.WithWorkingDir(context.Background(), workingDir))

	return stdout.String(), stderr.String(), err
}

func TestHarnessHelpIncludesBootstrapCommands(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeHarnessCLI(t, t.TempDir(), "--help")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "add")
	require.Contains(t, stdout, "bindings")
	require.Contains(t, stdout, "check")
	require.Contains(t, stdout, "create")
	require.Contains(t, stdout, "init")
	require.Contains(t, stdout, "inspect")
	require.Contains(t, stdout, "install")
	require.Contains(t, stdout, "remove")
	require.Contains(t, stdout, "root")
	require.Contains(t, stdout, "template")
}

func TestHarnessBindingsPlanJSONMergesLocalTemplateVariablesAndPrefillsRepoValues(t *testing.T) {
	t.Parallel()

	repo := seedHarnessBindingsPlanRepo(t, []bindingsPlanTemplateSpec{
		{
			OrbitID: "docs",
			VarsYAML: "" +
				"schema_version: 1\n" +
				"variables:\n" +
				"  project_name:\n" +
				"    value: Orbit\n" +
				"    description: Product title\n",
			Files: map[string]string{
				"docs/guide.md": "Orbit guide\n",
			},
		},
		{
			OrbitID: "cmd",
			VarsYAML: "" +
				"schema_version: 1\n" +
				"variables:\n" +
				"  project_name:\n" +
				"    value: Orbit\n" +
				"    description: Product title\n" +
				"  binary_name:\n" +
				"    value: orbit\n" +
				"    description: CLI binary\n",
			Files: map[string]string{
				"cmd/README.md": "Run Orbit as `orbit`.\n",
			},
		},
	}, ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n")

	stdout, stderr, err := executeHarnessCLI(
		t,
		repo.Root,
		"bindings",
		"plan",
		"orbit-template/docs",
		"orbit-template/cmd",
		"--json",
	)
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		RepoRoot    string `json:"repo_root"`
		SourceCount int    `json:"source_count"`
		Sources     []struct {
			Kind    string `json:"kind"`
			Ref     string `json:"ref"`
			OrbitID string `json:"orbit_id"`
			Commit  string `json:"commit"`
		} `json:"sources"`
		ReusedValues    []string `json:"reused_values"`
		MissingRequired []string `json:"missing_required"`
		Bindings        struct {
			SchemaVersion int `json:"schema_version"`
			Variables     map[string]struct {
				Value       string `json:"value"`
				Description string `json:"description"`
			} `json:"variables"`
		} `json:"bindings"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, repo.Root, payload.RepoRoot)
	require.Equal(t, 2, payload.SourceCount)
	require.Len(t, payload.Sources, 2)
	require.Equal(t, []string{"project_name"}, payload.ReusedValues)
	require.Equal(t, []string{"binary_name"}, payload.MissingRequired)
	require.Equal(t, 1, payload.Bindings.SchemaVersion)
	require.Equal(t, "Orbit", payload.Bindings.Variables["project_name"].Value)
	require.Equal(t, "Product title", payload.Bindings.Variables["project_name"].Description)
	require.Equal(t, "", payload.Bindings.Variables["binary_name"].Value)
	require.Equal(t, "CLI binary", payload.Bindings.Variables["binary_name"].Description)
}

func TestHarnessBindingsPlanFailsOnVariableDescriptionConflict(t *testing.T) {
	t.Parallel()

	repo := seedHarnessBindingsPlanRepo(t, []bindingsPlanTemplateSpec{
		{
			OrbitID: "docs",
			VarsYAML: "" +
				"schema_version: 1\n" +
				"variables:\n" +
				"  project_name:\n" +
				"    value: Orbit\n" +
				"    description: Product title\n",
			Files: map[string]string{
				"docs/guide.md": "Orbit guide\n",
			},
		},
		{
			OrbitID: "cmd",
			VarsYAML: "" +
				"schema_version: 1\n" +
				"variables:\n" +
				"  project_name:\n" +
				"    value: Orbit\n" +
				"    description: CLI title\n",
			Files: map[string]string{
				"cmd/README.md": "Orbit command guide\n",
			},
		},
	}, ""+
		"schema_version: 1\n"+
		"variables: {}\n")

	_, _, err := executeHarnessCLI(
		t,
		repo.Root,
		"bindings",
		"plan",
		"orbit-template/docs",
		"orbit-template/cmd",
		"--json",
	)
	require.Error(t, err)
	require.ErrorContains(t, err, `variable conflict for "project_name"`)
	require.ErrorContains(t, err, `sources: orbit-template/cmd, orbit-template/docs`)
}

func TestHarnessAddAndRemoveManageManualMembers(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(repo.Root, ".harness", "orbits"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo.Root, ".harness", "orbits", "docs.yaml"), []byte(""+
		"id: docs\n"+
		"description: docs orbit\n"+
		"include:\n"+
		"  - docs/**\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "add", "docs")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "added orbit docs to harness "+repo.Root+"\n", stdout)

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	require.Len(t, runtimeFile.Members, 1)
	require.Equal(t, "docs", runtimeFile.Members[0].OrbitID)
	require.Equal(t, harnesspkg.MemberSourceManual, runtimeFile.Members[0].Source)

	stdout, stderr, err = executeHarnessCLI(t, repo.Root, "remove", "docs")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "removed orbit docs from harness "+repo.Root+"\n", stdout)

	runtimeFile, err = harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	require.Empty(t, runtimeFile.Members)
}

func TestHarnessAddAndRemoveJSONOutputUsesManifestPath(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(repo.Root, ".harness", "orbits"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo.Root, ".harness", "orbits", "docs.yaml"), []byte(""+
		"id: docs\n"+
		"description: docs orbit\n"+
		"include:\n"+
		"  - docs/**\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "add", "docs", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var addPayload struct {
		HarnessRoot  string `json:"harness_root"`
		OrbitID      string `json:"orbit_id"`
		ManifestPath string `json:"manifest_path"`
		MemberCount  int    `json:"member_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &addPayload))
	require.Equal(t, repo.Root, addPayload.HarnessRoot)
	require.Equal(t, "docs", addPayload.OrbitID)
	require.Equal(t, filepath.Join(repo.Root, ".harness", "manifest.yaml"), addPayload.ManifestPath)
	require.Equal(t, 1, addPayload.MemberCount)

	stdout, stderr, err = executeHarnessCLI(t, repo.Root, "remove", "docs", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var removePayload struct {
		HarnessRoot  string `json:"harness_root"`
		OrbitID      string `json:"orbit_id"`
		ManifestPath string `json:"manifest_path"`
		MemberCount  int    `json:"member_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &removePayload))
	require.Equal(t, repo.Root, removePayload.HarnessRoot)
	require.Equal(t, "docs", removePayload.OrbitID)
	require.Equal(t, filepath.Join(repo.Root, ".harness", "manifest.yaml"), removePayload.ManifestPath)
	require.Zero(t, removePayload.MemberCount)
}

func TestHarnessAddIgnoresLegacyRuntimeFileWhenManifestIsValid(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(repo.Root, ".harness", "orbits"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo.Root, ".harness", "orbits", "docs.yaml"), []byte(""+
		"id: docs\n"+
		"description: docs orbit\n"+
		"include:\n"+
		"  - docs/**\n"), 0o600))
	repo.WriteFile(t, ".harness/runtime.yaml", "schema_version: nope\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "add", "docs")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "added orbit docs to harness "+repo.Root+"\n", stdout)

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	require.Len(t, runtimeFile.Members, 1)
	require.Equal(t, "docs", runtimeFile.Members[0].OrbitID)
}

func TestHarnessInstallLocalTemplateWritesInstallRecordVarsAndMember(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessRoot  string   `json:"harness_root"`
		OrbitID      string   `json:"orbit_id"`
		WrittenPaths []string `json:"written_paths"`
		MemberCount  int      `json:"member_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, "docs", payload.OrbitID)
	require.Contains(t, payload.WrittenPaths, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/installs/docs.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/manifest.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/vars.yaml")
	require.Equal(t, 1, payload.MemberCount)

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	require.Len(t, runtimeFile.Members, 1)
	require.Equal(t, "docs", runtimeFile.Members[0].OrbitID)
	require.Equal(t, harnesspkg.MemberSourceInstallOrbit, runtimeFile.Members[0].Source)

	manifestFile, err := harnesspkg.LoadManifestFile(repo.Root)
	require.NoError(t, err)
	require.Len(t, manifestFile.Members, 1)
	require.Equal(t, "docs", manifestFile.Members[0].OrbitID)
	require.Equal(t, harnesspkg.ManifestMemberSourceInstallOrbit, manifestFile.Members[0].Source)
}

func TestHarnessInstallTextOutputContract(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, ""+
		"installed orbit docs into harness "+repo.Root+"\n"+
		"source_ref: orbit-template/docs\n"+
		"member_count: 1\n"+
		"files: 5\n"+
		"warnings: none\n", stdout)
}

func TestHarnessInstallPlainProgressWritesStagesToStderr(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--progress", "plain")
	require.NoError(t, err)
	require.Equal(t, ""+
		"installed orbit docs into harness "+repo.Root+"\n"+
		"source_ref: orbit-template/docs\n"+
		"member_count: 1\n"+
		"files: 5\n"+
		"warnings: none\n", stdout)
	require.Contains(t, stderr, "progress: resolving install source\n")
	require.Contains(t, stderr, "progress: resolving bindings\n")
	require.Contains(t, stderr, "progress: checking conflicts\n")
	require.Contains(t, stderr, "progress: writing files\n")
	require.Contains(t, stderr, "progress: updating runtime metadata\n")
	require.Contains(t, stderr, "progress: install complete\n")
}

func TestHarnessInstallPlainProgressPreservesJSONStdout(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--progress", "plain", "--json")
	require.NoError(t, err)
	require.Contains(t, stderr, "progress: resolving install source\n")
	require.Contains(t, stderr, "progress: install complete\n")

	var payload struct {
		OrbitID string `json:"orbit_id"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "docs", payload.OrbitID)
}

func TestHarnessInstallQuietProgressSuppressesStages(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--progress", "quiet")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, ""+
		"installed orbit docs into harness "+repo.Root+"\n"+
		"source_ref: orbit-template/docs\n"+
		"member_count: 1\n"+
		"files: 5\n"+
		"warnings: none\n", stdout)
}

func TestHarnessInstallDryRunJSONIncludesRuntimeWriteAndDoesNotMutateRuntime(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Preview Orbit\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun      bool     `json:"dry_run"`
		HarnessRoot string   `json:"harness_root"`
		OrbitID     string   `json:"orbit_id"`
		Files       []string `json:"files"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.DryRun)
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, "docs", payload.OrbitID)
	require.Contains(t, payload.Files, "docs/guide.md")
	require.Contains(t, payload.Files, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.Files, ".harness/installs/docs.yaml")
	require.Contains(t, payload.Files, ".harness/manifest.yaml")
	require.Contains(t, payload.Files, ".harness/vars.yaml")

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	require.Empty(t, runtimeFile.Members)

	_, err = os.Stat(filepath.Join(repo.Root, ".harness", "orbits", "docs.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(repo.Root, ".harness", "installs", "docs.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestHarnessInstallFailsWhenOrbitAlreadyInstalled(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, _, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)
	repo.AddAndCommit(t, "commit installed runtime before template update")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.Error(t, err)
	require.Empty(t, stderr)
	require.Empty(t, stdout)
	require.ErrorContains(t, err, "already installed")
}

func TestHarnessInstallRemoteTemplateWritesInstallRecordVarsAndMember(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessInstallRepo(t)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	bindingsPath := filepath.Join(runtimeRepo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Remote Installed Orbit\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL, "--ref", "orbit-template/docs", "--bindings", bindingsPath, "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessRoot  string   `json:"harness_root"`
		OrbitID      string   `json:"orbit_id"`
		MemberCount  int      `json:"member_count"`
		WrittenPaths []string `json:"written_paths"`
		Source       struct {
			Kind string `json:"kind"`
			Repo string `json:"repo"`
			Ref  string `json:"ref"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, runtimeRepo.Root, payload.HarnessRoot)
	require.Equal(t, "docs", payload.OrbitID)
	require.Equal(t, 1, payload.MemberCount)
	require.Equal(t, "remote_git", payload.Source.Kind)
	require.Equal(t, remoteURL, payload.Source.Repo)
	require.Equal(t, "orbit-template/docs", payload.Source.Ref)
	require.Contains(t, payload.WrittenPaths, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/installs/docs.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/manifest.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/vars.yaml")

	runtimeFile, err := harnesspkg.LoadRuntimeFile(runtimeRepo.Root)
	require.NoError(t, err)
	require.Len(t, runtimeFile.Members, 1)
	require.Equal(t, "docs", runtimeFile.Members[0].OrbitID)
	require.Equal(t, harnesspkg.MemberSourceInstallOrbit, runtimeFile.Members[0].Source)

	installRecord, err := harnesspkg.LoadInstallRecord(runtimeRepo.Root, "docs")
	require.NoError(t, err)
	require.Equal(t, remoteURL, installRecord.Template.SourceRepo)
	require.Equal(t, "orbit-template/docs", installRecord.Template.SourceRef)
}

func TestHarnessInstallAcceptanceSmokeFromPublishedSourceBranch(t *testing.T) {
	t.Parallel()

	sourceRepo := testutil.NewRepo(t)
	sourceRepo.Run(t, "branch", "-m", "main")
	sourceRepo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"projection_visible: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	sourceRepo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	sourceRepo.WriteFile(t, "README.md", "author docs\n")
	sourceRepo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	sourceRepo.AddAndCommit(t, "seed source authoring repo")

	_, _, err := executeOrbitCLI(t, sourceRepo.Root, "template", "init-source")
	require.NoError(t, err)
	sourceRepo.AddAndCommit(t, "initialize source branch")

	_, _, err = executeOrbitCLI(t, sourceRepo.Root, "template", "publish", "--default")
	require.NoError(t, err)

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL, "--ref", "orbit-template/docs", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var installPayload struct {
		OrbitID string `json:"orbit_id"`
		Source  struct {
			Kind string `json:"kind"`
			Repo string `json:"repo"`
			Ref  string `json:"ref"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &installPayload))
	require.Equal(t, "docs", installPayload.OrbitID)
	require.Equal(t, "remote_git", installPayload.Source.Kind)
	require.Equal(t, remoteURL, installPayload.Source.Repo)
	require.Equal(t, "orbit-template/docs", installPayload.Source.Ref)

	templateDefinition, err := gitpkg.ReadFileAtRev(context.Background(), sourceRepo.Root, "orbit-template/docs", ".harness/orbits/docs.yaml")
	require.NoError(t, err)
	runtimeDefinition, readErr := os.ReadFile(filepath.Join(runtimeRepo.Root, ".harness", "orbits", "docs.yaml"))
	require.NoError(t, readErr)
	templateDefinitionParsed, parseErr := orbitpkg.ParseHostedOrbitSpecData(templateDefinition, ".harness/orbits/docs.yaml")
	require.NoError(t, parseErr)
	runtimeDefinitionParsed, parseErr := orbitpkg.ParseHostedOrbitSpecData(runtimeDefinition, ".harness/orbits/docs.yaml")
	require.NoError(t, parseErr)
	if templateDefinitionParsed.Exclude == nil {
		templateDefinitionParsed.Exclude = []string{}
	}
	if runtimeDefinitionParsed.Exclude == nil {
		runtimeDefinitionParsed.Exclude = []string{}
	}
	require.Equal(t, ".harness/orbits/docs.yaml", templateDefinitionParsed.SourcePath)
	require.Equal(t, ".harness/orbits/docs.yaml", runtimeDefinitionParsed.SourcePath)
	templateDefinitionParsed.SourcePath = ""
	runtimeDefinitionParsed.SourcePath = ""
	require.Equal(t, templateDefinitionParsed, runtimeDefinitionParsed)

	checkStdout, checkStderr, err := executeHarnessCLI(t, runtimeRepo.Root, "check", "--json")
	require.NoError(t, err)
	require.Empty(t, checkStderr)

	checkPayload := decodeHarnessCheckPayload(t, checkStdout)
	require.Truef(t, checkPayload.OK, "unexpected harness check payload: %s", checkStdout)
	require.Zero(t, checkPayload.FindingCount)

	guideData, readErr := os.ReadFile(filepath.Join(runtimeRepo.Root, "docs", "guide.md"))
	require.NoError(t, readErr)
	require.Equal(t, "Orbit guide\n", string(guideData))
}

func TestOrbitRuntimeCommandsWorkAfterHarnessInstallWithoutLegacyOrbitConfig(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessInstallRepo(t)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	bindingsPath := filepath.Join(t.TempDir(), "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, err := os.Stat(filepath.Join(runtimeRepo.Root, ".orbit", "config.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)

	_, _, err = executeHarnessCLI(t, runtimeRepo.Root, "install", sourceRepo.Root, "--ref", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)
	runtimeRepo.AddAndCommit(t, "commit installed runtime")

	_, err = os.Stat(filepath.Join(runtimeRepo.Root, ".orbit", "config.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)

	stdout, stderr, err := executeOrbitCLI(t, runtimeRepo.Root, "enter", "docs")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "entered orbit docs (2 file(s))\n", stdout)

	stdout, stderr, err = executeOrbitCLI(t, runtimeRepo.Root, "current")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "docs\n", stdout)

	stdout, stderr, err = executeOrbitCLI(t, runtimeRepo.Root, "status")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "current: docs\n")
	require.Contains(t, stdout, "in-scope:\nnone\n")

	stdout, stderr, err = executeOrbitCLI(t, runtimeRepo.Root, "diff")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Empty(t, stdout)

	stdout, stderr, err = executeOrbitCLI(t, runtimeRepo.Root, "leave")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "left orbit docs\n", stdout)
}

func TestOrbitRuntimeCommandsWorkInCreatedRuntimeAfterFirstCommit(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessInstallRepo(t)
	baseDir := t.TempDir()
	resolvedBaseDir, err := filepath.EvalSymlinks(baseDir)
	require.NoError(t, err)
	runtimeRoot := filepath.Join(resolvedBaseDir, "runtime-repo")
	bindingsPath := filepath.Join(baseDir, "install-bindings.yaml")

	_, _, err = executeHarnessCLI(t, baseDir, "create", "runtime-repo")
	require.NoError(t, err)
	runGitInDir(t, runtimeRoot, "config", "user.name", "Orbit Test")
	runGitInDir(t, runtimeRoot, "config", "user.email", "orbit@example.com")

	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, err = os.Stat(filepath.Join(runtimeRoot, ".orbit", "config.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)

	_, _, err = executeHarnessCLI(t, runtimeRoot, "install", sourceRepo.Root, "--ref", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)
	runGitInDir(t, runtimeRoot, "add", "-A")
	runGitInDir(t, runtimeRoot, "commit", "-m", "commit installed runtime")

	stdout, stderr, err := executeOrbitCLI(t, runtimeRoot, "enter", "docs")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Equal(t, "entered orbit docs (2 file(s))\n", stdout)
}

func TestHarnessInstallAcceptanceSmokeFromSourceRepoURL(t *testing.T) {
	t.Parallel()

	sourceRepo := testutil.NewRepo(t)
	sourceRepo.Run(t, "branch", "-m", "main")
	sourceRepo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"projection_visible: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	sourceRepo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	sourceRepo.WriteFile(t, "README.md", "author docs\n")
	sourceRepo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	sourceRepo.AddAndCommit(t, "seed source authoring repo")

	_, _, err := executeOrbitCLI(t, sourceRepo.Root, "template", "init-source")
	require.NoError(t, err)
	sourceRepo.AddAndCommit(t, "initialize source branch")

	_, _, err = executeOrbitCLI(t, sourceRepo.Root, "template", "publish", "--default")
	require.NoError(t, err)

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL, "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var installPayload struct {
		OrbitID string `json:"orbit_id"`
		Source  struct {
			Kind           string `json:"kind"`
			Repo           string `json:"repo"`
			Ref            string `json:"ref"`
			RequestedRef   string `json:"requested_ref"`
			ResolvedRef    string `json:"resolved_ref"`
			ResolutionKind string `json:"resolution_kind"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &installPayload))
	require.Equal(t, "docs", installPayload.OrbitID)
	require.Equal(t, "remote_git", installPayload.Source.Kind)
	require.Equal(t, remoteURL, installPayload.Source.Repo)
	require.Equal(t, "orbit-template/docs", installPayload.Source.Ref)
	require.Equal(t, "main", installPayload.Source.RequestedRef)
	require.Equal(t, "orbit-template/docs", installPayload.Source.ResolvedRef)
	require.Equal(t, "source_alias", installPayload.Source.ResolutionKind)
}

func TestHarnessInstallSourceRepoPlainProgressShowsAliasStages(t *testing.T) {
	t.Parallel()

	sourceRepo := testutil.NewRepo(t)
	sourceRepo.Run(t, "branch", "-m", "main")
	sourceRepo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"projection_visible: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	sourceRepo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	sourceRepo.WriteFile(t, "README.md", "author docs\n")
	sourceRepo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	sourceRepo.AddAndCommit(t, "seed source authoring repo")

	_, _, err := executeOrbitCLI(t, sourceRepo.Root, "template", "init-source")
	require.NoError(t, err)
	sourceRepo.AddAndCommit(t, "initialize source branch")

	_, _, err = executeOrbitCLI(t, sourceRepo.Root, "template", "publish", "--default")
	require.NoError(t, err)

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL, "--progress", "plain", "--json")
	require.NoError(t, err)
	require.Contains(t, stderr, "progress: resolving install source\n")
	require.Contains(t, stderr, "progress: resolving remote template candidates\n")
	require.Contains(t, stderr, "progress: source branch detected; resolving published template\n")
	require.Contains(t, stderr, "progress: fetching selected template\n")
	require.Contains(t, stderr, "progress: install complete\n")

	var payload struct {
		OrbitID string `json:"orbit_id"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "docs", payload.OrbitID)
}

func TestHarnessInstallAcceptanceSmokeFromSourceRepoAliasRef(t *testing.T) {
	t.Parallel()

	sourceRepo := testutil.NewRepo(t)
	sourceRepo.Run(t, "branch", "-m", "main")
	sourceRepo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"projection_visible: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	sourceRepo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	sourceRepo.WriteFile(t, "README.md", "author docs\n")
	sourceRepo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	sourceRepo.AddAndCommit(t, "seed source authoring repo")

	_, _, err := executeOrbitCLI(t, sourceRepo.Root, "template", "init-source")
	require.NoError(t, err)
	sourceRepo.AddAndCommit(t, "initialize source branch")

	_, _, err = executeOrbitCLI(t, sourceRepo.Root, "template", "publish", "--default")
	require.NoError(t, err)

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL, "--ref", "main", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var installPayload struct {
		OrbitID string `json:"orbit_id"`
		Source  struct {
			Kind           string `json:"kind"`
			Repo           string `json:"repo"`
			Ref            string `json:"ref"`
			RequestedRef   string `json:"requested_ref"`
			ResolvedRef    string `json:"resolved_ref"`
			ResolutionKind string `json:"resolution_kind"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &installPayload))
	require.Equal(t, "docs", installPayload.OrbitID)
	require.Equal(t, "remote_git", installPayload.Source.Kind)
	require.Equal(t, remoteURL, installPayload.Source.Repo)
	require.Equal(t, "orbit-template/docs", installPayload.Source.Ref)
	require.Equal(t, "main", installPayload.Source.RequestedRef)
	require.Equal(t, "orbit-template/docs", installPayload.Source.ResolvedRef)
	require.Equal(t, "source_alias", installPayload.Source.ResolutionKind)

	installRecord, err := harnesspkg.LoadInstallRecord(runtimeRepo.Root, "docs")
	require.NoError(t, err)
	require.Equal(t, "orbit-template/docs", installRecord.Template.SourceRef)
}

func TestHarnessInstallFailsClosedWhenSourceRepoPublishedTemplateIsMissing(t *testing.T) {
	t.Parallel()

	sourceRepo := testutil.NewRepo(t)
	sourceRepo.Run(t, "branch", "-m", "main")
	sourceRepo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"projection_visible: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	sourceRepo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	sourceRepo.WriteFile(t, "README.md", "author docs\n")
	sourceRepo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	sourceRepo.AddAndCommit(t, "seed source authoring repo")

	_, _, err := executeOrbitCLI(t, sourceRepo.Root, "template", "init-source")
	require.NoError(t, err)
	sourceRepo.AddAndCommit(t, "initialize source branch")

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL, "--ref", "main")
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	require.ErrorContains(t, err, "orbit template publish")
	require.ErrorContains(t, err, "orbit-template/docs")
}

func TestHarnessInstallFailsClosedWhenSourceRepoURLPublishedTemplateIsMissing(t *testing.T) {
	t.Parallel()

	sourceRepo := testutil.NewRepo(t)
	sourceRepo.Run(t, "branch", "-m", "main")
	sourceRepo.WriteFile(t, ".orbit/config.yaml", ""+
		"version: 1\n"+
		"shared_scope: []\n"+
		"projection_visible: []\n"+
		"behavior:\n"+
		"  outside_changes_mode: warn\n"+
		"  block_switch_if_hidden_dirty: true\n"+
		"  commit_append_trailer: true\n"+
		"  sparse_checkout_mode: no-cone\n")
	sourceRepo.WriteFile(t, ".orbit/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	sourceRepo.WriteFile(t, "README.md", "author docs\n")
	sourceRepo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	sourceRepo.AddAndCommit(t, "seed source authoring repo")

	_, _, err := executeOrbitCLI(t, sourceRepo.Root, "template", "init-source")
	require.NoError(t, err)
	sourceRepo.AddAndCommit(t, "initialize source branch")

	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL)
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	require.ErrorContains(t, err, "orbit template publish")
	require.ErrorContains(t, err, "orbit-template/docs")
}

func TestHarnessInstallOverwriteExistingReplacesOwnedFilesAndUpdatesInstallRecord(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, _, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)
	repo.AddAndCommit(t, "commit installed runtime before reinstall checks")

	originalRecord, err := harnesspkg.LoadInstallRecord(repo.Root, "docs")
	require.NoError(t, err)

	runtimeBranch := strings.TrimSpace(repo.Run(t, "branch", "--show-current"))
	repo.Run(t, "checkout", "orbit-template/docs")
	repo.WriteFile(t, "docs/reference.md", "$project_name reference\n")
	repo.Run(t, "rm", "-f", "docs/guide.md")
	repo.AddAndCommit(t, "update template branch contents")
	updatedCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "HEAD"))
	repo.Run(t, "checkout", runtimeBranch)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--overwrite-existing", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessRoot  string   `json:"harness_root"`
		OrbitID      string   `json:"orbit_id"`
		WrittenPaths []string `json:"written_paths"`
		MemberCount  int      `json:"member_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, "docs", payload.OrbitID)
	require.Equal(t, 1, payload.MemberCount)
	require.Contains(t, payload.WrittenPaths, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/installs/docs.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/manifest.yaml")
	require.Contains(t, payload.WrittenPaths, "docs/reference.md")

	_, err = os.Stat(filepath.Join(repo.Root, "docs", "guide.md"))
	require.ErrorIs(t, err, os.ErrNotExist)

	referenceData, err := os.ReadFile(filepath.Join(repo.Root, "docs", "reference.md"))
	require.NoError(t, err)
	require.Equal(t, "Installed Orbit reference\n", string(referenceData))

	updatedRecord, err := harnesspkg.LoadInstallRecord(repo.Root, "docs")
	require.NoError(t, err)
	require.NotEqual(t, originalRecord.Template.TemplateCommit, updatedRecord.Template.TemplateCommit)
	require.Equal(t, updatedCommit, updatedRecord.Template.TemplateCommit)

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	require.Len(t, runtimeFile.Members, 1)
	require.Equal(t, "docs", runtimeFile.Members[0].OrbitID)
	require.Equal(t, harnesspkg.MemberSourceInstallOrbit, runtimeFile.Members[0].Source)
}

func TestHarnessInstallReinstallAfterRemoveRequiresOverwriteExisting(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, _, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)
	repo.AddAndCommit(t, "commit installed runtime before corrupting install record")

	_, _, err = executeHarnessCLI(t, repo.Root, "remove", "docs")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	require.ErrorContains(t, err, "requires --overwrite-existing")

	stdout, stderr, err = executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--overwrite-existing", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		MemberCount int `json:"member_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, 1, payload.MemberCount)

	runtimeFile, err := harnesspkg.LoadRuntimeFile(repo.Root)
	require.NoError(t, err)
	require.Len(t, runtimeFile.Members, 1)
	require.Equal(t, "docs", runtimeFile.Members[0].OrbitID)
	require.Equal(t, harnesspkg.MemberSourceInstallOrbit, runtimeFile.Members[0].Source)
}

func TestHarnessInstallOverwriteFailsWhenExistingOwnedFilesCannotBeSafelyReconstructed(t *testing.T) {
	t.Parallel()

	repo := seedHarnessInstallRepo(t)
	bindingsPath := filepath.Join(repo.Root, "install-bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"), 0o600))

	_, _, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath)
	require.NoError(t, err)
	repo.AddAndCommit(t, "commit installed runtime before corrupting install record")

	runtimeBranch := strings.TrimSpace(repo.Run(t, "branch", "--show-current"))
	repo.Run(t, "checkout", "orbit-template/docs")
	repo.WriteFile(t, "docs/reference.md", "$project_name reference\n")
	repo.Run(t, "rm", "-f", "docs/guide.md")
	repo.AddAndCommit(t, "update template branch contents")
	repo.Run(t, "checkout", runtimeBranch)

	installRecordPath, err := harnesspkg.InstallRecordPath(repo.Root, "docs")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(installRecordPath, []byte(""+
		"schema_version: 1\n"+
		"orbit_id: docs\n"+
		"template:\n"+
		"  source_kind: local_branch\n"+
		"  source_repo: \"\"\n"+
		"  source_ref: orbit-template/docs\n"+
		"  template_commit: deadbeefdeadbeefdeadbeefdeadbeefdeadbeef\n"+
		"applied_at: 2026-03-26T10:00:00Z\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "orbit-template/docs", "--bindings", bindingsPath, "--overwrite-existing")
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	require.ErrorContains(t, err, "reconstruct existing install ownership")

	guideData, err := os.ReadFile(filepath.Join(repo.Root, "docs", "guide.md"))
	require.NoError(t, err)
	require.Equal(t, "Installed Orbit guide\n", string(guideData))

	_, err = os.Stat(filepath.Join(repo.Root, "docs", "reference.md"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestHarnessInstallDryRunHarnessTemplateLocalPreviewJSON(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	repo := seedEmptyHarnessRuntimeRepo(t)
	repo.Run(t, "remote", "add", "source", remoteURL)
	repo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun       bool                          `json:"dry_run"`
		TemplateKind string                        `json:"template_kind"`
		HarnessRoot  string                        `json:"harness_root"`
		HarnessID    string                        `json:"harness_id"`
		MemberIDs    []string                      `json:"member_ids"`
		Files        []string                      `json:"files"`
		Conflicts    []orbittemplate.ApplyConflict `json:"conflicts"`
		Source       struct {
			Kind string `json:"kind"`
			Ref  string `json:"ref"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.DryRun)
	require.Equal(t, "harness_template", payload.TemplateKind)
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(sourceRepo.Root), payload.HarnessID)
	require.Equal(t, []string{"cmd", "docs"}, payload.MemberIDs)
	require.Equal(t, "local_branch", payload.Source.Kind)
	require.Equal(t, "harness-template/workspace", payload.Source.Ref)
	require.Contains(t, payload.Files, ".harness/orbits/cmd.yaml")
	require.Contains(t, payload.Files, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.Files, "AGENTS.md")
	require.Contains(t, payload.Files, "cmd/main.go")
	require.Contains(t, payload.Files, "docs/guide.md")
	require.NotContains(t, payload.Files, ".harness/template.yaml")
	require.Empty(t, payload.Conflicts)
}

func TestHarnessInstallDryRunHarnessTemplateAllowsDisjointWithExistingBundleMember(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, ".harness/orbits/qa.yaml", ""+
		"id: qa\n"+
		"description: QA orbit\n"+
		"include:\n"+
		"  - qa/**\n")
	runtimeFile, err := harnesspkg.LoadRuntimeFile(runtimeRepo.Root)
	require.NoError(t, err)
	runtimeFile.Members = append(runtimeFile.Members, harnesspkg.RuntimeMember{
		OrbitID: "qa",
		Source:  harnesspkg.MemberSourceInstallBundle,
		AddedAt: time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
	})
	_, err = harnesspkg.WriteRuntimeFile(runtimeRepo.Root, runtimeFile)
	require.NoError(t, err)
	_, err = harnesspkg.WriteBundleRecord(runtimeRepo.Root, harnesspkg.BundleRecord{
		SchemaVersion:      1,
		HarnessID:          "qa-bundle",
		Template:           orbittemplate.Source{SourceKind: orbittemplate.InstallSourceKindLocalBranch, SourceRepo: "", SourceRef: "harness-template/qa", TemplateCommit: "abc123"},
		MemberIDs:          []string{"qa"},
		AppliedAt:          time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
		IncludesRootAgents: false,
		OwnedPaths:         []string{"qa/checklist.md"},
	})
	require.NoError(t, err)
	runtimeRepo.AddAndCommit(t, "seed disjoint bundle member")

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Empty(t, payload.Conflicts)
}

func TestHarnessInstallDryRunHarnessTemplateAllowsDisjointWithExistingOrbitMember(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, ".harness/orbits/qa.yaml", ""+
		"id: qa\n"+
		"description: QA orbit\n"+
		"include:\n"+
		"  - qa/**\n")
	_, _, err = executeHarnessCLI(t, runtimeRepo.Root, "add", "qa")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Empty(t, payload.Conflicts)
}

func TestHarnessInstallDryRunHarnessTemplateConflictsOnExistingMember(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	_, _, err = executeHarnessCLI(t, runtimeRepo.Root, "add", "docs")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Contains(t, payload.Conflicts, orbittemplate.ApplyConflict{
		Path:    ".harness/manifest.yaml",
		Message: `member "docs" already exists in harness runtime`,
	})
}

func TestHarnessInstallDryRunHarnessTemplateTextOutputReportsConflicts(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	_, _, err = executeHarnessCLI(t, runtimeRepo.Root, "add", "docs")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--dry-run")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "conflicts:\n")
	require.Contains(t, stdout, ".harness/manifest.yaml: member \"docs\" already exists in harness runtime\n")
}

func TestHarnessInstallDryRunHarnessTemplateConflictsOnPathContent(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, "docs/guide.md", "conflicting docs\n")
	runtimeRepo.AddAndCommit(t, "seed conflicting docs path")

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Contains(t, payload.Conflicts, orbittemplate.ApplyConflict{
		Path:    "docs/guide.md",
		Message: "target path already exists with different content",
	})
}

func TestHarnessInstallDryRunHarnessTemplateConflictsOnVariableDescription(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	_, err = harnesspkg.WriteVarsFile(runtimeRepo.Root, bindings.VarsFile{
		SchemaVersion: 1,
		Variables: map[string]bindings.VariableBinding{
			"project_name": {
				Value:       "Orbit",
				Description: "A different meaning",
			},
		},
	})
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Contains(t, payload.Conflicts, orbittemplate.ApplyConflict{
		Path:    ".harness/vars.yaml",
		Message: `variable conflict for "project_name"`,
	})
}

func TestHarnessInstallDryRunHarnessTemplateConflictsOnInvalidAgentsLane(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, "AGENTS.md", "<<broken>>\n<!-- orbit:block:docs -->\n")
	runtimeRepo.AddAndCommit(t, "seed invalid agents")

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Contains(t, payload.Conflicts, orbittemplate.ApplyConflict{
		Path:    "AGENTS.md",
		Message: "runtime AGENTS.md is invalid for harness block merge (" + harnesspkg.DefaultHarnessIDForPath(sourceRepo.Root) + ")",
	})
}

func TestHarnessInstallDryRunHarnessTemplateRemotePreviewJSON(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(
		t,
		runtimeRepo.Root,
		"install",
		remoteURL,
		"--ref",
		"harness-template/workspace",
		"--dry-run",
		"--json",
	)
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun       bool     `json:"dry_run"`
		TemplateKind string   `json:"template_kind"`
		HarnessRoot  string   `json:"harness_root"`
		HarnessID    string   `json:"harness_id"`
		MemberIDs    []string `json:"member_ids"`
		Source       struct {
			Kind string `json:"kind"`
			Repo string `json:"repo"`
			Ref  string `json:"ref"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.DryRun)
	require.Equal(t, "harness_template", payload.TemplateKind)
	require.Equal(t, runtimeRepo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(sourceRepo.Root), payload.HarnessID)
	require.Equal(t, []string{"cmd", "docs"}, payload.MemberIDs)
	require.Equal(t, "remote_git", payload.Source.Kind)
	require.Equal(t, remoteURL, payload.Source.Repo)
	require.Equal(t, "harness-template/workspace", payload.Source.Ref)
}

func TestHarnessInstallDryRunSelectsRemoteHarnessTemplateWhenItIsTheOnlyInstallableCandidate(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(
		t,
		runtimeRepo.Root,
		"install",
		remoteURL,
		"--dry-run",
		"--json",
	)
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		TemplateKind string `json:"template_kind"`
		HarnessID    string `json:"harness_id"`
		Source       struct {
			Kind string `json:"kind"`
			Repo string `json:"repo"`
			Ref  string `json:"ref"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "harness_template", payload.TemplateKind)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(sourceRepo.Root), payload.HarnessID)
	require.Equal(t, "remote_git", payload.Source.Kind)
	require.Equal(t, remoteURL, payload.Source.Repo)
	require.Equal(t, "harness-template/workspace", payload.Source.Ref)
}

func TestHarnessInstallDryRunPrefersUniqueDefaultRemoteHarnessTemplate(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	_, _, err = executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/qa", "--default")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(
		t,
		runtimeRepo.Root,
		"install",
		remoteURL,
		"--dry-run",
		"--json",
	)
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		TemplateKind string `json:"template_kind"`
		HarnessID    string `json:"harness_id"`
		Source       struct {
			Kind string `json:"kind"`
			Repo string `json:"repo"`
			Ref  string `json:"ref"`
		} `json:"source"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "harness_template", payload.TemplateKind)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(sourceRepo.Root), payload.HarnessID)
	require.Equal(t, "remote_git", payload.Source.Kind)
	require.Equal(t, remoteURL, payload.Source.Repo)
	require.Equal(t, "harness-template/qa", payload.Source.Ref)
}

func TestHarnessInstallHarnessTemplatePlainProgressShowsStages(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", remoteURL, "--ref", "harness-template/workspace", "--dry-run", "--progress", "plain", "--json")
	require.NoError(t, err)
	require.Contains(t, stderr, "progress: resolving install source\n")
	require.Contains(t, stderr, "progress: fetching selected template\n")
	require.Contains(t, stderr, "progress: resolving bindings\n")
	require.Contains(t, stderr, "progress: checking conflicts\n")
	require.Contains(t, stderr, "progress: install complete\n")

	var payload struct {
		TemplateKind string `json:"template_kind"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "harness_template", payload.TemplateKind)
}

func TestHarnessInstallHarnessTemplateFailsClosedOnNonDisjointLocalRuntime(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "install", "harness-template/workspace", "--json")
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	require.ErrorContains(t, err, "conflicts detected; mixed harness template install requires disjoint targets")
}

func TestHarnessInstallDryRunHarnessTemplateUsesRenderedContentForPathConflictAnalysis(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, "cmd/main.go", "package main\n\nconst name = \"orbitctl\"\n")
	runtimeRepo.AddAndCommit(t, "seed rendered command path")

	bindingsPath := filepath.Join(runtimeRepo.Root, "bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbitctl\n"+
		"    description: CLI binary\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--bindings", bindingsPath, "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	for _, conflict := range payload.Conflicts {
		require.NotEqual(t, "cmd/main.go", conflict.Path)
	}
}

func TestHarnessInstallHarnessTemplateLocalWriteJSON(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	bindingsPath := filepath.Join(runtimeRepo.Root, "bindings.yaml")
	require.NoError(t, os.WriteFile(bindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbit-installed\n"+
		"    description: CLI binary\n"), 0o600))

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--bindings", bindingsPath, "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun       bool     `json:"dry_run"`
		HarnessRoot  string   `json:"harness_root"`
		TemplateKind string   `json:"template_kind"`
		HarnessID    string   `json:"harness_id"`
		MemberIDs    []string `json:"member_ids"`
		WrittenPaths []string `json:"written_paths"`
		MemberCount  int      `json:"member_count"`
		BundleCount  int      `json:"bundle_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.False(t, payload.DryRun)
	require.Equal(t, runtimeRepo.Root, payload.HarnessRoot)
	require.Equal(t, "harness_template", payload.TemplateKind)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(sourceRepo.Root), payload.HarnessID)
	require.Equal(t, []string{"cmd", "docs"}, payload.MemberIDs)
	require.Equal(t, 2, payload.MemberCount)
	require.Equal(t, 1, payload.BundleCount)
	require.Contains(t, payload.WrittenPaths, ".harness/orbits/cmd.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/bundles/"+payload.HarnessID+".yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/manifest.yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/vars.yaml")
	require.Contains(t, payload.WrittenPaths, "AGENTS.md")
	require.Contains(t, payload.WrittenPaths, "cmd/main.go")
	require.Contains(t, payload.WrittenPaths, "docs/guide.md")

	cmdData, err := os.ReadFile(filepath.Join(runtimeRepo.Root, "cmd", "main.go"))
	require.NoError(t, err)
	require.Equal(t, "package main\n\nconst name = \"orbitctl\"\n", string(cmdData))

	runtimeFile, err := harnesspkg.LoadRuntimeFile(runtimeRepo.Root)
	require.NoError(t, err)
	require.Len(t, runtimeFile.Members, 2)
	require.ElementsMatch(t, []harnesspkg.RuntimeMember{
		{
			OrbitID: "cmd",
			Source:  harnesspkg.MemberSourceInstallBundle,
			AddedAt: runtimeFile.Members[0].AddedAt,
		},
		{
			OrbitID: "docs",
			Source:  harnesspkg.MemberSourceInstallBundle,
			AddedAt: runtimeFile.Members[1].AddedAt,
		},
	}, runtimeFile.Members)

	manifestFile, err := harnesspkg.LoadManifestFile(runtimeRepo.Root)
	require.NoError(t, err)
	require.ElementsMatch(t, []harnesspkg.ManifestMember{
		{
			OrbitID: "cmd",
			Source:  harnesspkg.ManifestMemberSourceInstallBundle,
			AddedAt: manifestFile.Members[0].AddedAt,
		},
		{
			OrbitID: "docs",
			Source:  harnesspkg.ManifestMemberSourceInstallBundle,
			AddedAt: manifestFile.Members[1].AddedAt,
		},
	}, manifestFile.Members)

	bundleRecord, err := harnesspkg.LoadBundleRecord(runtimeRepo.Root, payload.HarnessID)
	require.NoError(t, err)
	require.Equal(t, []string{"cmd", "docs"}, bundleRecord.MemberIDs)
	require.True(t, bundleRecord.IncludesRootAgents)
	require.Contains(t, bundleRecord.OwnedPaths, "AGENTS.md")
	require.Contains(t, bundleRecord.OwnedPaths, "cmd/main.go")

	agentsData, err := os.ReadFile(filepath.Join(runtimeRepo.Root, "AGENTS.md"))
	require.NoError(t, err)
	document, err := orbittemplate.ParseRuntimeAgentsDocument(agentsData)
	require.NoError(t, err)
	require.Len(t, document.Segments, 1)
	require.Equal(t, payload.HarnessID, document.Segments[0].OrbitID)
	require.Contains(t, string(document.Segments[0].Content), "Installed Orbit")
}

func TestHarnessInstallHarnessTemplateOverwriteExistingReplacesSameBundle(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	initialBindingsPath := filepath.Join(runtimeRepo.Root, "bindings.yaml")
	require.NoError(t, os.WriteFile(initialBindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbit-installed\n"+
		"    description: CLI binary\n"), 0o600))

	_, _, err = executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--bindings", initialBindingsPath)
	require.NoError(t, err)

	_, _, err = executeHarnessCLI(t, sourceRepo.Root, "remove", "docs")
	require.NoError(t, err)
	sourceRepo.WriteFile(t, "cmd/main.go", "package main\n\nconst name = \"orbit-next\"\n")
	sourceRepo.WriteFile(t, "AGENTS.md", "Workspace guide for $project_name v2\n")
	sourceRepo.AddAndCommit(t, "update harness template source")

	_, err = harnesspkg.SaveTemplateBranch(context.Background(), harnesspkg.TemplateSaveInput{
		Preview: harnesspkg.TemplateSavePreviewInput{
			RepoRoot:     sourceRepo.Root,
			TargetBranch: "harness-template/workspace",
			Now:          time.Date(2026, time.April, 1, 12, 0, 0, 0, time.UTC),
		},
		Overwrite: true,
	})
	require.NoError(t, err)
	sourceRepo.Run(t, "push", remoteURL, "harness-template/workspace")
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	overwriteBindingsPath := filepath.Join(runtimeRepo.Root, "overwrite-bindings.yaml")
	require.NoError(t, os.WriteFile(overwriteBindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbit-next\n"+
		"    description: CLI binary\n"), 0o600))

	dryRunStdout, dryRunStderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--bindings", overwriteBindingsPath, "--overwrite-existing", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, dryRunStderr)

	var dryRunPayload struct {
		Conflicts []orbittemplate.ApplyConflict `json:"conflicts"`
	}
	require.NoError(t, json.Unmarshal([]byte(dryRunStdout), &dryRunPayload))
	require.Empty(t, dryRunPayload.Conflicts)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--bindings", overwriteBindingsPath, "--overwrite-existing", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessID    string   `json:"harness_id"`
		MemberIDs    []string `json:"member_ids"`
		WrittenPaths []string `json:"written_paths"`
		MemberCount  int      `json:"member_count"`
		BundleCount  int      `json:"bundle_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(sourceRepo.Root), payload.HarnessID)
	require.Equal(t, []string{"cmd"}, payload.MemberIDs)
	require.Equal(t, 1, payload.MemberCount)
	require.Equal(t, 1, payload.BundleCount)
	require.Contains(t, payload.WrittenPaths, ".harness/bundles/"+payload.HarnessID+".yaml")
	require.Contains(t, payload.WrittenPaths, ".harness/manifest.yaml")
	require.Contains(t, payload.WrittenPaths, "cmd/main.go")
	require.Contains(t, payload.WrittenPaths, "AGENTS.md")

	cmdData, err := os.ReadFile(filepath.Join(runtimeRepo.Root, "cmd", "main.go"))
	require.NoError(t, err)
	require.Equal(t, "package main\n\nconst name = \"orbit-next\"\n", string(cmdData))

	_, err = os.Stat(filepath.Join(runtimeRepo.Root, ".harness", "orbits", "docs.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(filepath.Join(runtimeRepo.Root, "docs", "guide.md"))
	require.ErrorIs(t, err, os.ErrNotExist)

	runtimeFile, err := harnesspkg.LoadRuntimeFile(runtimeRepo.Root)
	require.NoError(t, err)
	require.Equal(t, []harnesspkg.RuntimeMember{{
		OrbitID: "cmd",
		Source:  harnesspkg.MemberSourceInstallBundle,
		AddedAt: runtimeFile.Members[0].AddedAt,
	}}, runtimeFile.Members)

	bundleRecord, err := harnesspkg.LoadBundleRecord(runtimeRepo.Root, payload.HarnessID)
	require.NoError(t, err)
	require.Equal(t, []string{"cmd"}, bundleRecord.MemberIDs)
	require.NotContains(t, bundleRecord.OwnedPaths, ".harness/orbits/docs.yaml")
	require.NotContains(t, bundleRecord.OwnedPaths, "docs/guide.md")

	agentsData, err := os.ReadFile(filepath.Join(runtimeRepo.Root, "AGENTS.md"))
	require.NoError(t, err)
	document, err := orbittemplate.ParseRuntimeAgentsDocument(agentsData)
	require.NoError(t, err)
	require.Len(t, document.Segments, 1)
	require.Equal(t, payload.HarnessID, document.Segments[0].OrbitID)
	require.Contains(t, string(document.Segments[0].Content), "Installed Orbit v2")
}

func TestHarnessInstallHarnessTemplateOverwriteExistingFailsAcrossInstallUnits(t *testing.T) {
	t.Parallel()

	sourceRepo := seedHarnessTemplateSaveRepo(t)
	_, _, err := executeHarnessCLI(t, sourceRepo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	remoteURL := testutil.NewBareRemoteFromRepo(t, sourceRepo)
	runtimeRepo := seedEmptyHarnessRuntimeRepo(t)
	runtimeRepo.Run(t, "remote", "add", "source", remoteURL)
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	initialBindingsPath := filepath.Join(runtimeRepo.Root, "bindings.yaml")
	require.NoError(t, os.WriteFile(initialBindingsPath, []byte(""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Installed Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbit-installed\n"+
		"    description: CLI binary\n"), 0o600))

	_, _, err = executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--bindings", initialBindingsPath)
	require.NoError(t, err)

	sourceRepo.WriteFile(t, ".harness/orbits/qa.yaml", ""+
		"id: qa\n"+
		"description: QA orbit\n"+
		"include:\n"+
		"  - qa/**\n")
	sourceRepo.WriteFile(t, "qa/checklist.md", "QA checklist\n")
	_, _, err = executeHarnessCLI(t, sourceRepo.Root, "add", "qa")
	require.NoError(t, err)
	sourceRepo.AddAndCommit(t, "add qa member to bundle source")
	_, err = harnesspkg.SaveTemplateBranch(context.Background(), harnesspkg.TemplateSaveInput{
		Preview: harnesspkg.TemplateSavePreviewInput{
			RepoRoot:     sourceRepo.Root,
			TargetBranch: "harness-template/workspace",
			Now:          time.Date(2026, time.April, 1, 13, 0, 0, 0, time.UTC),
		},
		Overwrite: true,
	})
	require.NoError(t, err)
	sourceRepo.Run(t, "push", remoteURL, "harness-template/workspace")
	runtimeRepo.Run(t, "fetch", "source", "harness-template/workspace:harness-template/workspace")

	runtimeRepo.WriteFile(t, ".harness/orbits/qa.yaml", ""+
		"id: qa\n"+
		"description: QA orbit\n"+
		"include:\n"+
		"  - qa/**\n")
	_, _, err = executeHarnessCLI(t, runtimeRepo.Root, "add", "qa")
	require.NoError(t, err)

	stdout, stderr, err := executeHarnessCLI(t, runtimeRepo.Root, "install", "harness-template/workspace", "--bindings", initialBindingsPath, "--overwrite-existing")
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	require.ErrorContains(t, err, `member "qa" already exists in harness runtime`)
}

func TestHarnessTemplateSaveCreatesHarnessTemplateBranch(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)
	currentBranch := strings.TrimSpace(repo.Run(t, "rev-parse", "--abbrev-ref", "HEAD"))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		HarnessRoot        string   `json:"harness_root"`
		HarnessID          string   `json:"harness_id"`
		TargetBranch       string   `json:"target_branch"`
		Commit             string   `json:"commit"`
		Files              []string `json:"files"`
		MemberCount        int      `json:"member_count"`
		DefaultTemplate    bool     `json:"default_template"`
		IncludesRootAgents bool     `json:"includes_root_agents"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), payload.HarnessID)
	require.Equal(t, "harness-template/workspace", payload.TargetBranch)
	require.NotEmpty(t, payload.Commit)
	require.Equal(t, 2, payload.MemberCount)
	require.False(t, payload.DefaultTemplate)
	require.True(t, payload.IncludesRootAgents)
	require.Contains(t, payload.Files, ".harness/template.yaml")
	require.Contains(t, payload.Files, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.Files, ".harness/orbits/cmd.yaml")
	require.Contains(t, payload.Files, "docs/guide.md")
	require.Contains(t, payload.Files, "cmd/main.go")
	require.Contains(t, payload.Files, "AGENTS.md")

	require.Equal(t, currentBranch, strings.TrimSpace(repo.Run(t, "rev-parse", "--abbrev-ref", "HEAD")))

	files := splitLines(strings.TrimSpace(repo.Run(t, "ls-tree", "-r", "--name-only", "harness-template/workspace")))
	require.Equal(t, []string{
		".harness/manifest.yaml",
		".harness/orbits/cmd.yaml",
		".harness/orbits/docs.yaml",
		".harness/template.yaml",
		"AGENTS.md",
		"cmd/main.go",
		"docs/guide.md",
	}, files)

	_, err = gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", ".orbit/template.yaml")
	require.Error(t, err)

	manifestData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", ".harness/template.yaml")
	require.NoError(t, err)
	require.Contains(t, string(manifestData), "kind: harness_template")
	require.Contains(t, string(manifestData), "default_template: false")
	require.Contains(t, string(manifestData), "includes_root_agents: true")

	branchManifestData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", ".harness/manifest.yaml")
	require.NoError(t, err)
	require.Contains(t, string(branchManifestData), "kind: harness_template")
	require.Contains(t, string(branchManifestData), "default_template: false")

	agentsData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", "AGENTS.md")
	require.NoError(t, err)
	require.Contains(t, string(agentsData), "$project_name")
	require.Contains(t, string(agentsData), "$command_name")

	inspectStdout, inspectStderr, err := executeOrbitCLI(t, repo.Root, "branch", "inspect", "harness-template/workspace", "--json")
	require.NoError(t, err)
	require.Empty(t, inspectStderr)

	var inspectPayload struct {
		Inspection struct {
			Classification struct {
				Kind         string `json:"kind"`
				TemplateKind string `json:"template_kind"`
			} `json:"classification"`
			HarnessID          string `json:"harness_id"`
			MemberCount        int    `json:"member_count"`
			DefinitionCount    int    `json:"definition_count"`
			IncludesRootAgents bool   `json:"includes_root_agents"`
		} `json:"inspection"`
	}
	require.NoError(t, json.Unmarshal([]byte(inspectStdout), &inspectPayload))
	require.Equal(t, "template", inspectPayload.Inspection.Classification.Kind)
	require.Equal(t, "harness", inspectPayload.Inspection.Classification.TemplateKind)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), inspectPayload.Inspection.HarnessID)
	require.Equal(t, 2, inspectPayload.Inspection.MemberCount)
	require.Equal(t, 2, inspectPayload.Inspection.DefinitionCount)
	require.True(t, inspectPayload.Inspection.IncludesRootAgents)
}

func TestHarnessTemplateSaveIgnoresLegacyRuntimeFileWhenManifestIsValid(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)
	repo.WriteFile(t, ".harness/runtime.yaml", "schema_version: nope\n")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun       bool   `json:"dry_run"`
		HarnessID    string `json:"harness_id"`
		TargetBranch string `json:"target_branch"`
		MemberCount  int    `json:"member_count"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.DryRun)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), payload.HarnessID)
	require.Equal(t, "harness-template/workspace", payload.TargetBranch)
	require.Equal(t, 2, payload.MemberCount)
}

func TestHarnessTemplateSaveMarksTemplateAsDefaultWhenRequested(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--default", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		TargetBranch    string `json:"target_branch"`
		DefaultTemplate bool   `json:"default_template"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "harness-template/workspace", payload.TargetBranch)
	require.True(t, payload.DefaultTemplate)

	manifestData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", ".harness/template.yaml")
	require.NoError(t, err)
	require.Contains(t, string(manifestData), "default_template: true")

	branchManifestData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", ".harness/manifest.yaml")
	require.NoError(t, err)
	require.Contains(t, string(branchManifestData), "default_template: true")
}

func TestHarnessTemplateSaveDryRunDoesNotWriteBranch(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)
	currentBranch := strings.TrimSpace(repo.Run(t, "rev-parse", "--abbrev-ref", "HEAD"))
	currentCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "HEAD"))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--dry-run")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "harness template save dry-run -> harness-template/workspace")
	require.Contains(t, stdout, "files:")
	require.Contains(t, stdout, ".harness/template.yaml")
	require.Contains(t, stdout, ".harness/orbits/docs.yaml")
	require.Contains(t, stdout, "AGENTS.md")
	require.Contains(t, stdout, "replacements:")
	require.Contains(t, stdout, "AGENTS.md: project_name <- Orbit (1)")
	require.Contains(t, stdout, "AGENTS.md: command_name <- orbitctl (1)")
	require.Contains(t, stdout, "ambiguities: none")
	require.Contains(t, stdout, "manifest:")
	require.Contains(t, stdout, "harness_id: "+harnesspkg.DefaultHarnessIDForPath(repo.Root))
	require.Contains(t, stdout, "default_template: false")
	require.Contains(t, stdout, "created_from_branch: "+currentBranch)
	require.Contains(t, stdout, "created_from_commit: "+currentCommit)
	require.Contains(t, stdout, "includes_root_agents: true")
	require.Contains(t, stdout, "members:")
	require.Contains(t, stdout, "cmd")
	require.Contains(t, stdout, "docs")
	require.Contains(t, stdout, "variables:")
	require.Contains(t, stdout, "command_name [required] CLI binary")
	require.Contains(t, stdout, "project_name [required] Product title")

	exists, err := gitpkg.LocalBranchExists(context.Background(), repo.Root, "harness-template/workspace")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestHarnessTemplateSaveDryRunJSONContract(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--dry-run", "--default", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun             bool     `json:"dry_run"`
		HarnessRoot        string   `json:"harness_root"`
		HarnessID          string   `json:"harness_id"`
		TargetBranch       string   `json:"target_branch"`
		Files              []string `json:"files"`
		MemberCount        int      `json:"member_count"`
		DefaultTemplate    bool     `json:"default_template"`
		IncludesRootAgents bool     `json:"includes_root_agents"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.DryRun)
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), payload.HarnessID)
	require.Equal(t, "harness-template/workspace", payload.TargetBranch)
	require.Equal(t, 2, payload.MemberCount)
	require.True(t, payload.DefaultTemplate)
	require.True(t, payload.IncludesRootAgents)
	require.Contains(t, payload.Files, ".harness/template.yaml")
	require.Contains(t, payload.Files, ".harness/orbits/docs.yaml")
	require.Contains(t, payload.Files, "AGENTS.md")

	exists, err := gitpkg.LocalBranchExists(context.Background(), repo.Root, "harness-template/workspace")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestHarnessTemplateSaveTextOutputContract(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, stdout, "saved harness template "+harnesspkg.DefaultHarnessIDForPath(repo.Root)+" to branch harness-template/workspace\n")
	require.Contains(t, stdout, "commit: ")
	require.Contains(t, stdout, "files: 6\n")
	require.Contains(t, stdout, "member_count: 2\n")
	require.Contains(t, stdout, "default_template: false\n")
	require.Contains(t, stdout, "includes_root_agents: true\n")
}

func TestHarnessTemplateSaveRequiresOverwriteForExistingBranch(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)

	_, _, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)

	_, _, err = executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace")
	require.Error(t, err)
	require.ErrorContains(t, err, "already exists; re-run with --overwrite to replace it")
}

func TestHarnessTemplateSaveJSONFailureRequiresOverwriteForExistingBranch(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)

	_, _, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	firstCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "harness-template/workspace"))

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun            bool   `json:"dry_run"`
		Saved             bool   `json:"saved"`
		Stage             string `json:"stage"`
		Reason            string `json:"reason"`
		HarnessRoot       string `json:"harness_root"`
		HarnessID         string `json:"harness_id"`
		TargetBranch      string `json:"target_branch"`
		OverwriteRequired bool   `json:"overwrite_required"`
		Message           string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.False(t, payload.DryRun)
	require.False(t, payload.Saved)
	require.Equal(t, "write", payload.Stage)
	require.Equal(t, "target_branch_exists", payload.Reason)
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), payload.HarnessID)
	require.Equal(t, "harness-template/workspace", payload.TargetBranch)
	require.True(t, payload.OverwriteRequired)
	require.Contains(t, payload.Message, "already exists; re-run with --overwrite to replace it")
	require.Equal(t, firstCommit, strings.TrimSpace(repo.Run(t, "rev-parse", "harness-template/workspace")))
}

func TestHarnessTemplateSaveOverwriteRewritesExistingBranch(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)

	_, _, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace")
	require.NoError(t, err)
	firstCommit := strings.TrimSpace(repo.Run(t, "rev-parse", "harness-template/workspace"))

	repo.WriteFile(t, "docs/guide.md", "Orbit guide v2 for $project_name\n")
	repo.AddAndCommit(t, "update runtime docs")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--overwrite", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		Commit       string `json:"commit"`
		TargetBranch string `json:"target_branch"`
		DryRun       bool   `json:"dry_run"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.Equal(t, "harness-template/workspace", payload.TargetBranch)
	require.False(t, payload.DryRun)
	require.NotEqual(t, firstCommit, payload.Commit)

	savedData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", "docs/guide.md")
	require.NoError(t, err)
	require.Contains(t, string(savedData), "$project_name")
	require.Contains(t, string(savedData), "v2")
}

func TestHarnessTemplateSaveEditTemplateWritesEditedTemplateWithoutMutatingRuntimeWorktree(t *testing.T) {
	repo := seedHarnessTemplateSaveRepo(t)
	repo.WriteFile(t, ".harness/vars.yaml", ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbitctl\n"+
		"    description: CLI binary\n"+
		"  service_url:\n"+
		"    value: http://localhost:3000\n"+
		"    description: Service URL\n")
	repo.AddAndCommit(t, "add service url binding")

	editorScript := filepath.Join(repo.Root, "edit-template.sh")
	require.NoError(t, os.WriteFile(editorScript, []byte(""+
		"#!/bin/sh\n"+
		"printf '%s\\n' '$project_name guide at $service_url' > \"$1/docs/guide.md\"\n"), 0o755))
	t.Setenv("EDITOR", editorScript)

	_, _, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--edit-template")
	require.NoError(t, err)

	runtimeData, err := os.ReadFile(filepath.Join(repo.Root, "docs", "guide.md"))
	require.NoError(t, err)
	require.Equal(t, "Orbit guide\n", string(runtimeData))

	templateData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", "docs/guide.md")
	require.NoError(t, err)
	require.Equal(t, "$project_name guide at $service_url\n", string(templateData))

	manifestData, err := gitpkg.ReadFileAtRev(context.Background(), repo.Root, "harness-template/workspace", ".harness/template.yaml")
	require.NoError(t, err)
	require.Contains(t, string(manifestData), "service_url")
}

func TestHarnessTemplateSaveEditTemplateFailsWhenDefinitionIsRemoved(t *testing.T) {
	repo := seedHarnessTemplateSaveRepo(t)

	editorScript := filepath.Join(repo.Root, "edit-template.sh")
	require.NoError(t, os.WriteFile(editorScript, []byte(""+
		"#!/bin/sh\n"+
		"rm -f \"$1/.harness/orbits/docs.yaml\"\n"), 0o755))
	t.Setenv("EDITOR", editorScript)

	_, _, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--edit-template")
	require.Error(t, err)
	require.ErrorContains(t, err, "edited harness template must keep member definition")

	exists, err := gitpkg.LocalBranchExists(context.Background(), repo.Root, "harness-template/workspace")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestHarnessTemplateSaveFailsClosedOnReplacementAmbiguity(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)
	repo.WriteFile(t, ".harness/vars.yaml", ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  product_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbitctl\n"+
		"    description: CLI binary\n")
	repo.AddAndCommit(t, "introduce template save ambiguity")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace")
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	require.ErrorContains(t, err, "replacement ambiguity detected")
	require.ErrorContains(t, err, `AGENTS.md [root_agents]`)
	require.ErrorContains(t, err, `docs/guide.md [docs]`)

	exists, err := gitpkg.LocalBranchExists(context.Background(), repo.Root, "harness-template/workspace")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestHarnessTemplateSaveJSONFailureIncludesAmbiguityContributors(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)
	repo.WriteFile(t, ".harness/vars.yaml", ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  product_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbitctl\n"+
		"    description: CLI binary\n")
	repo.AddAndCommit(t, "introduce template save ambiguity")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun             bool   `json:"dry_run"`
		Saved              bool   `json:"saved"`
		HarnessRoot        string `json:"harness_root"`
		HarnessID          string `json:"harness_id"`
		TargetBranch       string `json:"target_branch"`
		DefaultTemplate    bool   `json:"default_template"`
		IncludesRootAgents bool   `json:"includes_root_agents"`
		Message            string `json:"message"`
		Ambiguities        []struct {
			Path         string   `json:"path"`
			Contributors []string `json:"contributors"`
			Ambiguities  []struct {
				Literal   string   `json:"literal"`
				Variables []string `json:"variables"`
			} `json:"ambiguities"`
		} `json:"ambiguities"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.False(t, payload.DryRun)
	require.False(t, payload.Saved)
	require.Equal(t, repo.Root, payload.HarnessRoot)
	require.Equal(t, harnesspkg.DefaultHarnessIDForPath(repo.Root), payload.HarnessID)
	require.Equal(t, "harness-template/workspace", payload.TargetBranch)
	require.False(t, payload.DefaultTemplate)
	require.True(t, payload.IncludesRootAgents)
	require.Contains(t, payload.Message, "replacement ambiguity detected")
	require.Contains(t, payload.Message, `AGENTS.md [root_agents]`)
	require.Contains(t, payload.Message, `docs/guide.md [docs]`)
	require.Contains(t, payload.Ambiguities, struct {
		Path         string   `json:"path"`
		Contributors []string `json:"contributors"`
		Ambiguities  []struct {
			Literal   string   `json:"literal"`
			Variables []string `json:"variables"`
		} `json:"ambiguities"`
	}{
		Path:         "AGENTS.md",
		Contributors: []string{"root_agents"},
		Ambiguities: []struct {
			Literal   string   `json:"literal"`
			Variables []string `json:"variables"`
		}{{
			Literal:   "Orbit",
			Variables: []string{"product_name", "project_name"},
		}},
	})

	exists, err := gitpkg.LocalBranchExists(context.Background(), repo.Root, "harness-template/workspace")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestHarnessTemplateSaveDryRunJSONIncludesAmbiguityContributors(t *testing.T) {
	t.Parallel()

	repo := seedHarnessTemplateSaveRepo(t)
	repo.WriteFile(t, ".harness/vars.yaml", ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  product_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbitctl\n"+
		"    description: CLI binary\n")
	repo.AddAndCommit(t, "introduce template save ambiguity")

	stdout, stderr, err := executeHarnessCLI(t, repo.Root, "template", "save", "--to", "harness-template/workspace", "--dry-run", "--json")
	require.NoError(t, err)
	require.Empty(t, stderr)

	var payload struct {
		DryRun      bool `json:"dry_run"`
		Ambiguities []struct {
			Path         string   `json:"path"`
			Contributors []string `json:"contributors"`
			Ambiguities  []struct {
				Literal   string   `json:"literal"`
				Variables []string `json:"variables"`
			} `json:"ambiguities"`
		} `json:"ambiguities"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))
	require.True(t, payload.DryRun)
	require.Contains(t, payload.Ambiguities, struct {
		Path         string   `json:"path"`
		Contributors []string `json:"contributors"`
		Ambiguities  []struct {
			Literal   string   `json:"literal"`
			Variables []string `json:"variables"`
		} `json:"ambiguities"`
	}{
		Path:         "AGENTS.md",
		Contributors: []string{"root_agents"},
		Ambiguities: []struct {
			Literal   string   `json:"literal"`
			Variables []string `json:"variables"`
		}{
			{Literal: "Orbit", Variables: []string{"product_name", "project_name"}},
		},
	})
	require.Contains(t, payload.Ambiguities, struct {
		Path         string   `json:"path"`
		Contributors []string `json:"contributors"`
		Ambiguities  []struct {
			Literal   string   `json:"literal"`
			Variables []string `json:"variables"`
		} `json:"ambiguities"`
	}{
		Path:         "docs/guide.md",
		Contributors: []string{"docs"},
		Ambiguities: []struct {
			Literal   string   `json:"literal"`
			Variables []string `json:"variables"`
		}{
			{Literal: "Orbit", Variables: []string{"product_name", "project_name"}},
		},
	})
}

func seedHarnessInstallRepo(t *testing.T) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	repo.WriteFile(t, ".harness/vars.yaml", ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n")
	repo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	repo.AddAndCommit(t, "seed runtime repo")

	_, _, err = executeOrbitCLI(t, repo.Root, "template", "save", "docs", "--to", "orbit-template/docs")
	require.NoError(t, err)

	repo.Run(t, "rm", "-f", ".harness/orbits/docs.yaml", ".harness/vars.yaml", "docs/guide.md")
	repo.AddAndCommit(t, "clear runtime branch")

	return repo
}

func seedEmptyHarnessRuntimeRepo(t *testing.T) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)
	repo.WriteFile(t, "README.md", "runtime repo\n")
	repo.AddAndCommit(t, "seed runtime repo")

	return repo
}

func seedHarnessTemplateSaveRepo(t *testing.T) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	repo.WriteFile(t, ".harness/orbits/docs.yaml", ""+
		"id: docs\n"+
		"description: Docs orbit\n"+
		"include:\n"+
		"  - docs/**\n")
	repo.WriteFile(t, ".harness/orbits/cmd.yaml", ""+
		"id: cmd\n"+
		"description: Cmd orbit\n"+
		"include:\n"+
		"  - cmd/**\n")
	repo.WriteFile(t, ".harness/vars.yaml", ""+
		"schema_version: 1\n"+
		"variables:\n"+
		"  project_name:\n"+
		"    value: Orbit\n"+
		"    description: Product title\n"+
		"  command_name:\n"+
		"    value: orbitctl\n"+
		"    description: CLI binary\n")
	repo.WriteFile(t, "docs/guide.md", "Orbit guide\n")
	repo.WriteFile(t, "cmd/main.go", "package main\n\nconst name = \"orbitctl\"\n")
	repo.WriteFile(t, "AGENTS.md", ""+
		"Workspace guide for Orbit\n"+
		"<!-- keep -->\n"+
		"Use orbitctl consistently.\n")

	_, _, err = executeHarnessCLI(t, repo.Root, "add", "docs")
	require.NoError(t, err)
	_, _, err = executeHarnessCLI(t, repo.Root, "add", "cmd")
	require.NoError(t, err)

	repo.AddAndCommit(t, "seed harness template save runtime")

	return repo
}

type bindingsPlanTemplateSpec struct {
	OrbitID  string
	VarsYAML string
	Files    map[string]string
}

func seedHarnessBindingsPlanRepo(t *testing.T, templates []bindingsPlanTemplateSpec, currentVarsYAML string) *testutil.Repo {
	t.Helper()

	repo := testutil.NewRepo(t)
	_, _, err := executeHarnessCLI(t, repo.Root, "init")
	require.NoError(t, err)

	for _, template := range templates {
		repo.WriteFile(t, filepath.Join(".harness", "orbits", template.OrbitID+".yaml"), ""+
			"id: "+template.OrbitID+"\n"+
			"description: "+template.OrbitID+" orbit\n"+
			"include:\n"+
			"  - "+template.OrbitID+"/**\n")
		repo.WriteFile(t, ".harness/vars.yaml", template.VarsYAML)
		for path, content := range template.Files {
			repo.WriteFile(t, path, content)
		}
		repo.AddAndCommit(t, "seed "+template.OrbitID+" runtime content")

		_, err = orbittemplate.SaveTemplateBranch(context.Background(), orbittemplate.TemplateSaveInput{
			Preview: orbittemplate.TemplateSavePreviewInput{
				RepoRoot:     repo.Root,
				OrbitID:      template.OrbitID,
				TargetBranch: "orbit-template/" + template.OrbitID,
				Now:          time.Date(2026, time.April, 9, 12, 0, 0, 0, time.UTC),
			},
		})
		require.NoError(t, err)
	}

	repo.WriteFile(t, ".harness/vars.yaml", currentVarsYAML)
	repo.AddAndCommit(t, "seed runtime shared vars")

	return repo
}

type harnessCheckPayload struct {
	HarnessRoot  string `json:"harness_root"`
	HarnessID    string `json:"harness_id"`
	OK           bool   `json:"ok"`
	FindingCount int    `json:"finding_count"`
	Findings     []struct {
		Kind    string `json:"kind"`
		OrbitID string `json:"orbit_id"`
		Path    string `json:"path"`
		Message string `json:"message"`
	} `json:"findings"`
}

func decodeHarnessCheckPayload(t *testing.T, stdout string) harnessCheckPayload {
	t.Helper()

	var payload harnessCheckPayload
	require.NoError(t, json.Unmarshal([]byte(stdout), &payload))

	return payload
}

func requireHarnessCheckFinding(t *testing.T, payload harnessCheckPayload, kind string) struct {
	Kind    string `json:"kind"`
	OrbitID string `json:"orbit_id"`
	Path    string `json:"path"`
	Message string `json:"message"`
} {
	t.Helper()

	for _, finding := range payload.Findings {
		if finding.Kind == kind {
			return finding
		}
	}

	t.Fatalf("finding %q not found in %+v", kind, payload.Findings)
	return struct {
		Kind    string `json:"kind"`
		OrbitID string `json:"orbit_id"`
		Path    string `json:"path"`
		Message string `json:"message"`
	}{}
}

func splitLines(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return strings.Split(strings.TrimSpace(value), "\n")
}

func runGitInDir(t *testing.T, dir string, args ...string) {
	t.Helper()

	command := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := command.CombinedOutput()
	require.NoError(t, err, "git %s failed:\n%s", strings.Join(args, " "), string(output))
}
