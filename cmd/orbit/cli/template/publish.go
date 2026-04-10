package orbittemplate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	"gopkg.in/yaml.v3"
)

type TemplatePublishMode string

const (
	TemplatePublishModeSource   TemplatePublishMode = "source"
	TemplatePublishModeTemplate TemplatePublishMode = "orbit_template"
)

type orbitTemplateBranchManifest struct {
	SchemaVersion int                               `yaml:"schema_version"`
	Kind          string                            `yaml:"kind"`
	Template      orbitTemplateBranchManifestSource `yaml:"template"`
	Variables     map[string]VariableSpec           `yaml:"variables"`
}

type orbitTemplateBranchManifestSource struct {
	OrbitID           string    `yaml:"orbit_id"`
	DefaultTemplate   *bool     `yaml:"default_template,omitempty"`
	CreatedFromBranch string    `yaml:"created_from_branch"`
	CreatedFromCommit string    `yaml:"created_from_commit"`
	CreatedAt         time.Time `yaml:"created_at"`
}

// TemplatePublishInput is the high-level author workflow input for local orbit template publish.
type TemplatePublishInput struct {
	RepoRoot           string
	OrbitID            string
	DefaultTemplate    bool
	DefaultTemplateSet bool
	Push               bool
	Remote             string
}

// TemplatePublishPreview contains the resolved local publish plan.
type TemplatePublishPreview struct {
	RepoRoot        string
	OrbitID         string
	SourceBranch    string
	PublishBranch   string
	DefaultTemplate bool
	Mode            TemplatePublishMode
	SavePreview     TemplateSavePreview
}

// TemplatePublishResult contains the preview plus the local publish outcome.
type TemplatePublishResult struct {
	Preview    TemplatePublishPreview
	Changed    bool
	Commit     string
	RemotePush TemplatePublishRemoteResult
}

// TemplatePublishRemoteResult reports the remote side of one publish execution.
type TemplatePublishRemoteResult struct {
	Attempted bool
	Success   bool
	Remote    string
	Reason    string
}

// PublishError reports a publish flow that produced a local result but failed later.
type PublishError struct {
	Result TemplatePublishResult
	Err    error
}

func (err *PublishError) Error() string {
	return err.Err.Error()
}

func (err *PublishError) Unwrap() error {
	return err.Err
}

// BuildTemplatePublishPreview validates source-branch preconditions and constructs the fixed local publish plan.
func BuildTemplatePublishPreview(ctx context.Context, input TemplatePublishInput) (TemplatePublishPreview, error) {
	currentBranch, err := gitpkg.CurrentBranch(ctx, input.RepoRoot)
	if err != nil {
		return TemplatePublishPreview{}, fmt.Errorf("resolve current branch: %w", err)
	}
	if currentBranch == "HEAD" {
		return TemplatePublishPreview{}, fmt.Errorf("publish requires a current branch; detached HEAD is not supported")
	}

	revisionKind, err := loadCurrentRevisionManifestKind(input.RepoRoot)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return TemplatePublishPreview{}, fmt.Errorf("load %s: %w", branchManifestPath, err)
	}
	switch revisionKind {
	case "orbit_template":
		return buildDirectTemplatePublishPreview(ctx, input, currentBranch)
	case "runtime", "harness_template":
		return TemplatePublishPreview{}, fmt.Errorf("publish requires a source or orbit_template revision; current revision kind is %q", revisionKind)
	}

	return buildSourceTemplatePublishPreview(ctx, input)
}

func buildSourceTemplatePublishPreview(ctx context.Context, input TemplatePublishInput) (TemplatePublishPreview, error) {
	sourceManifest, err := LoadSourceManifest(input.RepoRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return TemplatePublishPreview{}, fmt.Errorf("load %s: %w", sourceManifestRelativePath, err)
		}
		return TemplatePublishPreview{}, fmt.Errorf("load %s: %w", sourceManifestRelativePath, err)
	}

	if err := ensurePublishSourceContracts(ctx, input.RepoRoot, sourceManifest); err != nil {
		return TemplatePublishPreview{}, err
	}

	orbitID, err := resolvePublishOrbitID(ctx, input.RepoRoot, input.OrbitID)
	if err != nil {
		return TemplatePublishPreview{}, err
	}

	publishBranch := fmt.Sprintf("orbit-template/%s", orbitID)
	savePreview, err := BuildTemplateSavePreview(ctx, TemplateSavePreviewInput{
		RepoRoot:      input.RepoRoot,
		OrbitID:       orbitID,
		TargetBranch:  publishBranch,
		DefaultBranch: input.DefaultTemplate,
	})
	if err != nil {
		return TemplatePublishPreview{}, fmt.Errorf("build publish preview: %w", err)
	}

	return TemplatePublishPreview{
		RepoRoot:        input.RepoRoot,
		OrbitID:         orbitID,
		SourceBranch:    sourceManifest.SourceBranch,
		PublishBranch:   publishBranch,
		DefaultTemplate: input.DefaultTemplate,
		Mode:            TemplatePublishModeSource,
		SavePreview:     savePreview,
	}, nil
}

// PublishTemplate performs the local publish path only. Remote push is intentionally out of scope here.
func PublishTemplate(ctx context.Context, input TemplatePublishInput) (TemplatePublishResult, error) {
	preview, err := BuildTemplatePublishPreview(ctx, input)
	if err != nil {
		return TemplatePublishResult{}, err
	}
	result := TemplatePublishResult{
		Preview: preview,
	}

	if preview.Mode == TemplatePublishModeTemplate {
		result.Changed = false
	} else {
		noop, err := isTemplatePublishNoOp(ctx, preview)
		if err != nil {
			return TemplatePublishResult{}, fmt.Errorf("compare existing publish branch: %w", err)
		}
		if noop {
			result.Changed = false
		} else {
			saveResult, err := SaveTemplateBranch(ctx, TemplateSaveInput{
				Preview: TemplateSavePreviewInput{
					RepoRoot:      preview.RepoRoot,
					OrbitID:       preview.OrbitID,
					TargetBranch:  preview.PublishBranch,
					DefaultBranch: preview.DefaultTemplate,
				},
				Overwrite: true,
			})
			if err != nil {
				return TemplatePublishResult{}, fmt.Errorf("publish template branch: %w", err)
			}
			result.Changed = true
			result.Commit = saveResult.WriteResult.Commit
		}
	}

	if !input.Push {
		return result, nil
	}

	remote := strings.TrimSpace(input.Remote)
	if remote == "" {
		remote = "origin"
	}
	result.RemotePush.Remote = remote

	if preview.Mode == TemplatePublishModeSource {
		relation, err := gitpkg.CompareBranchToRemoteBranch(ctx, preview.RepoRoot, remote, preview.SourceBranch)
		if err != nil {
			result.RemotePush.Reason = "remote_source_branch_unavailable"
			return result, &PublishError{
				Result: result,
				Err:    fmt.Errorf("check remote source branch %q on %q: %w", preview.SourceBranch, remote, err),
			}
		}
		if relation == gitpkg.BranchRelationBehind || relation == gitpkg.BranchRelationDiverged {
			result.RemotePush.Reason = "source_branch_not_up_to_date"
			return result, &PublishError{
				Result: result,
				Err:    fmt.Errorf("local source branch %q is not up to date with %s/%s", preview.SourceBranch, remote, preview.SourceBranch),
			}
		}
	} else {
		remoteExists, err := remoteBranchExists(ctx, preview.RepoRoot, remote, preview.PublishBranch)
		if err != nil {
			result.RemotePush.Reason = "remote_publish_branch_unavailable"
			return result, &PublishError{
				Result: result,
				Err:    fmt.Errorf("check remote template branch %q on %q: %w", preview.PublishBranch, remote, err),
			}
		}
		if remoteExists {
			relation, err := gitpkg.CompareBranchToRemoteBranch(ctx, preview.RepoRoot, remote, preview.PublishBranch)
			if err != nil {
				result.RemotePush.Reason = "remote_publish_branch_unavailable"
				return result, &PublishError{
					Result: result,
					Err:    fmt.Errorf("check remote template branch %q on %q: %w", preview.PublishBranch, remote, err),
				}
			}
			if relation == gitpkg.BranchRelationBehind || relation == gitpkg.BranchRelationDiverged {
				result.RemotePush.Reason = "publish_branch_not_up_to_date"
				return result, &PublishError{
					Result: result,
					Err:    fmt.Errorf("local template branch %q is not up to date with %s/%s", preview.PublishBranch, remote, preview.PublishBranch),
				}
			}
		}
	}

	result.RemotePush.Attempted = true
	if err := gitpkg.PushBranch(ctx, preview.RepoRoot, remote, preview.PublishBranch); err != nil {
		result.RemotePush.Reason = "push_failed"
		return result, &PublishError{
			Result: result,
			Err:    fmt.Errorf("push published branch %q to %q: %w", preview.PublishBranch, remote, err),
		}
	}
	result.RemotePush.Success = true

	return result, nil
}

func buildDirectTemplatePublishPreview(ctx context.Context, input TemplatePublishInput, currentBranch string) (TemplatePublishPreview, error) {
	if err := ensureCleanTrackedWorktree(ctx, input.RepoRoot); err != nil {
		return TemplatePublishPreview{}, err
	}

	branchManifest, err := loadCurrentOrbitTemplateBranchManifest(input.RepoRoot)
	if err != nil {
		return TemplatePublishPreview{}, fmt.Errorf("load %s: %w", branchManifestPath, err)
	}
	orbitID := branchManifest.Template.OrbitID
	if err := diagnoseDirectTemplateAgentsArtifact(ctx, input.RepoRoot, orbitID); err != nil {
		return TemplatePublishPreview{}, err
	}

	source, err := ResolveLocalTemplateSource(ctx, input.RepoRoot, currentBranch)
	if err != nil {
		return TemplatePublishPreview{}, fmt.Errorf("validate current template branch %q: %w", currentBranch, err)
	}
	orbitID = source.Manifest.Template.OrbitID

	expectedBranch := fmt.Sprintf("orbit-template/%s", orbitID)
	if currentBranch != expectedBranch {
		return TemplatePublishPreview{}, fmt.Errorf("direct template publish requires current branch %q to match fixed template branch %q", currentBranch, expectedBranch)
	}
	if input.OrbitID != "" && input.OrbitID != orbitID {
		return TemplatePublishPreview{}, fmt.Errorf("requested orbit %q must match current template orbit %q", input.OrbitID, orbitID)
	}

	defaultTemplate := source.Manifest.Template.DefaultTemplate
	if branchManifest.Template.DefaultTemplate != nil {
		defaultTemplate = *branchManifest.Template.DefaultTemplate
	}
	if input.DefaultTemplateSet && input.DefaultTemplate != defaultTemplate {
		return TemplatePublishPreview{}, fmt.Errorf("current template branch default_template is %t; direct template publish does not rewrite it", defaultTemplate)
	}

	return TemplatePublishPreview{
		RepoRoot:        input.RepoRoot,
		OrbitID:         orbitID,
		SourceBranch:    currentBranch,
		PublishBranch:   currentBranch,
		DefaultTemplate: defaultTemplate,
		Mode:            TemplatePublishModeTemplate,
	}, nil
}

func ensurePublishSourceContracts(ctx context.Context, repoRoot string, sourceManifest SourceManifest) error {
	currentBranch, err := gitpkg.CurrentBranch(ctx, repoRoot)
	if err != nil {
		return fmt.Errorf("resolve current branch: %w", err)
	}
	if currentBranch != sourceManifest.SourceBranch {
		return fmt.Errorf("publish requires current branch %q to match source branch %q", currentBranch, sourceManifest.SourceBranch)
	}

	paths, err := gitpkg.ListAllFilesAtRev(ctx, repoRoot, "HEAD")
	if err != nil {
		return fmt.Errorf("list tracked files at HEAD: %w", err)
	}
	for _, path := range paths {
		if strings.HasPrefix(path, ".harness/") && !isAllowedSourceBranchHarnessPath(path) {
			return fmt.Errorf("source branch must not contain %s", path)
		}
	}

	return ensureCleanTrackedWorktree(ctx, repoRoot)
}

func resolvePublishOrbitID(ctx context.Context, repoRoot string, requestedOrbitID string) (string, error) {
	sourceManifest, err := LoadSourceManifest(repoRoot)
	if err != nil {
		return "", fmt.Errorf("load %s: %w", sourceManifestRelativePath, err)
	}

	definition, err := loadSingleHostedSourceOrbitDefinition(ctx, repoRoot)
	if err != nil {
		return "", err
	}
	orbitID := definition.ID
	if sourceManifest.Publish == nil || strings.TrimSpace(sourceManifest.Publish.OrbitID) == "" {
		return "", fmt.Errorf("source.orbit_id must be present")
	}
	if sourceManifest.Publish != nil && sourceManifest.Publish.OrbitID != orbitID {
		return "", fmt.Errorf("source.orbit_id %q must match single source orbit %q", sourceManifest.Publish.OrbitID, orbitID)
	}
	if requestedOrbitID != "" {
		if requestedOrbitID != orbitID {
			return "", fmt.Errorf("requested orbit %q must match single source orbit %q", requestedOrbitID, orbitID)
		}
		return requestedOrbitID, nil
	}
	return orbitID, nil
}

func isTemplatePublishNoOp(ctx context.Context, preview TemplatePublishPreview) (bool, error) {
	exists, err := gitpkg.LocalBranchExists(ctx, preview.RepoRoot, preview.PublishBranch)
	if err != nil {
		return false, fmt.Errorf("check target publish branch %q: %w", preview.PublishBranch, err)
	}
	if !exists {
		return false, nil
	}

	if !templateFilesMatchRevision(ctx, preview.RepoRoot, preview.PublishBranch, preview.SavePreview) {
		return false, nil
	}

	branchManifestValid := false
	branchManifestData, readBranchManifestErr := gitpkg.ReadFileAtRev(ctx, preview.RepoRoot, preview.PublishBranch, branchManifestPath)
	if readBranchManifestErr == nil {
		currentBranchManifest, parseBranchManifestErr := parseOrbitTemplateBranchManifestData(branchManifestData)
		if parseBranchManifestErr == nil {
			branchManifestValid = publishBranchManifestEquivalent(preview.SavePreview.Manifest, currentBranchManifest)
		}
	}
	if !branchManifestValid {
		return false, nil
	}

	return true, nil
}

func templateFilesMatchRevision(ctx context.Context, repoRoot string, rev string, preview TemplateSavePreview) bool {
	paths, err := gitpkg.ListAllFilesAtRev(ctx, repoRoot, rev)
	if err != nil {
		return false
	}

	expected := make(map[string]CandidateFile, len(preview.Files))
	for _, file := range preview.Files {
		expected[file.Path] = file
	}

	actualCount := 0
	for _, path := range paths {
		if path == manifestRelativePath || path == branchManifestPath {
			continue
		}
		actualCount++

		expectedFile, ok := expected[path]
		if !ok {
			return false
		}

		content, err := gitpkg.ReadFileAtRev(ctx, repoRoot, rev, path)
		if err != nil {
			return false
		}
		if !reflect.DeepEqual(expectedFile.Content, content) {
			return false
		}

		mode, err := gitpkg.FileModeAtRev(ctx, repoRoot, rev, path)
		if err != nil {
			return false
		}
		if expectedFile.Mode != mode {
			return false
		}
	}

	return actualCount == len(expected)
}

func parseOrbitTemplateBranchManifestData(data []byte) (orbitTemplateBranchManifest, error) {
	var manifest orbitTemplateBranchManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return orbitTemplateBranchManifest{}, fmt.Errorf("decode branch manifest: %w", err)
	}

	return manifest, nil
}

func publishBranchManifestEquivalent(expected Manifest, actual orbitTemplateBranchManifest) bool {
	return actual.SchemaVersion == manifestSchemaVersion &&
		actual.Kind == "orbit_template" &&
		actual.Template.OrbitID == expected.Template.OrbitID &&
		(actual.Template.DefaultTemplate == nil || *actual.Template.DefaultTemplate == expected.Template.DefaultTemplate) &&
		actual.Template.CreatedFromBranch == expected.Template.CreatedFromBranch &&
		actual.Template.CreatedFromCommit == expected.Template.CreatedFromCommit &&
		branchManifestVariablesEquivalent(expected.Variables, actual.Variables) &&
		!actual.Template.CreatedAt.IsZero()
}

func branchManifestVariablesEquivalent(expected map[string]VariableSpec, actual map[string]VariableSpec) bool {
	if len(expected) != len(actual) {
		return false
	}
	for name, spec := range expected {
		if actualSpec, ok := actual[name]; !ok || actualSpec != spec {
			return false
		}
	}

	return true
}

func ensureCleanTrackedWorktree(ctx context.Context, repoRoot string) error {
	statusEntries, err := gitpkg.WorktreeStatus(ctx, repoRoot)
	if err != nil {
		return fmt.Errorf("load worktree status: %w", err)
	}
	for _, entry := range statusEntries {
		if entry.Tracked {
			return fmt.Errorf("publish requires a clean tracked worktree; found %s %s", entry.Code, entry.Path)
		}
	}

	return nil
}

func loadCurrentRevisionManifestKind(repoRoot string) (string, error) {
	filename := filepath.Join(repoRoot, filepath.FromSlash(branchManifestPath))
	//nolint:gosec // The branch manifest path is fixed under the repo root.
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", filename, err)
	}

	var manifest struct {
		Kind string `yaml:"kind"`
	}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return "", fmt.Errorf("decode branch manifest kind: %w", err)
	}
	return strings.TrimSpace(manifest.Kind), nil
}

func loadCurrentOrbitTemplateBranchManifest(repoRoot string) (orbitTemplateBranchManifest, error) {
	filename := filepath.Join(repoRoot, filepath.FromSlash(branchManifestPath))
	//nolint:gosec // The branch manifest path is fixed under the repo root.
	data, err := os.ReadFile(filename)
	if err != nil {
		return orbitTemplateBranchManifest{}, fmt.Errorf("read %s: %w", filename, err)
	}

	manifest, err := parseOrbitTemplateBranchManifestData(data)
	if err != nil {
		return orbitTemplateBranchManifest{}, err
	}
	if manifest.Kind != "orbit_template" {
		return orbitTemplateBranchManifest{}, fmt.Errorf("kind must be %q", "orbit_template")
	}
	if strings.TrimSpace(manifest.Template.OrbitID) == "" {
		return orbitTemplateBranchManifest{}, fmt.Errorf("template.orbit_id must not be empty")
	}

	return manifest, nil
}

func remoteBranchExists(ctx context.Context, repoRoot string, remote string, branch string) (bool, error) {
	heads, err := gitpkg.ListRemoteHeads(ctx, repoRoot, remote)
	if err != nil {
		return false, fmt.Errorf("list remote heads for %q: %w", remote, err)
	}
	for _, head := range heads {
		if head.Name == branch || head.Ref == "refs/heads/"+branch {
			return true, nil
		}
	}

	return false, nil
}

func diagnoseDirectTemplateAgentsArtifact(ctx context.Context, repoRoot string, orbitID string) error {
	filename := filepath.Join(repoRoot, sharedFilePathAgents)
	//nolint:gosec // The root AGENTS path is fixed under the repo root.
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", sharedFilePathAgents, err)
	}

	document, err := ParseRuntimeAgentsDocument(data)
	if err != nil {
		return fmt.Errorf("current template revision contains invalid %s; fix or remove %s before publishing: %w", sharedFilePathAgents, sharedFilePathAgents, err)
	}

	block, extractErr := extractRuntimeAgentsBlock(document, orbitID)
	if extractErr != nil {
		return fmt.Errorf("current template revision contains root %s, but published orbit templates must not include it; remove %s before publishing", sharedFilePathAgents, sharedFilePathAgents)
	}

	spec, err := loadDirectTemplatePublishOrbitSpec(ctx, repoRoot, orbitID)
	if err != nil {
		return err
	}
	expectedBody := orbitAgentsBody(spec)
	if len(expectedBody) == 0 {
		return fmt.Errorf(
			"current template revision contains root %s for orbit %q, but structured brief truth is missing; run `orbit brief backfill --orbit %s` or remove %s before publishing",
			sharedFilePathAgents,
			orbitID,
			orbitID,
			sharedFilePathAgents,
		)
	}
	if bytes.Equal(ensureTrailingNewline(block), ensureTrailingNewline(expectedBody)) {
		return fmt.Errorf("current template revision contains materialized root %s for orbit %q; remove %s before publishing", sharedFilePathAgents, orbitID, sharedFilePathAgents)
	}

	return fmt.Errorf(
		"current template revision contains drifted materialized root %s for orbit %q; run `orbit brief backfill --orbit %s` or remove %s before publishing",
		sharedFilePathAgents,
		orbitID,
		orbitID,
		sharedFilePathAgents,
	)
}

func loadDirectTemplatePublishOrbitSpec(ctx context.Context, repoRoot string, orbitID string) (orbitpkg.OrbitSpec, error) {
	spec, err := orbitpkg.LoadHostedOrbitSpec(ctx, repoRoot, orbitID)
	if err == nil {
		return spec, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return orbitpkg.OrbitSpec{}, fmt.Errorf("load hosted orbit spec for %q: %w", orbitID, err)
	}

	spec, err = orbitpkg.LoadOrbitSpec(ctx, repoRoot, orbitID)
	if err != nil {
		return orbitpkg.OrbitSpec{}, fmt.Errorf("load template orbit spec for %q: %w", orbitID, err)
	}

	return spec, nil
}
