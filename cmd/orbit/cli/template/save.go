package orbittemplate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	"github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
)

// TemplateSavePreviewInput describes the shared runtime-to-template preview pipeline.
type TemplateSavePreviewInput struct {
	RepoRoot      string
	OrbitID       string
	TargetBranch  string
	DefaultBranch bool
	Now           time.Time
	EditTemplate  bool
	Editor        Editor
}

// TemplateSavePreview contains the fully built template candidate plus branch-manifest metadata.
type TemplateSavePreview struct {
	RepoRoot             string
	OrbitID              string
	TargetBranch         string
	Files                []CandidateFile
	ReplacementSummaries []FileReplacementSummary
	Ambiguities          []FileReplacementAmbiguity
	Warnings             []string
	Manifest             Manifest
}

// FilePaths returns the stable preview file list without the generated manifest path.
func (preview TemplateSavePreview) FilePaths() []string {
	paths := make([]string, 0, len(preview.Files))
	for _, file := range preview.Files {
		paths = append(paths, file.Path)
	}

	return paths
}

// TemplateSaveInput describes the real branch-writing save path.
type TemplateSaveInput struct {
	Preview   TemplateSavePreviewInput
	Overwrite bool
}

// TemplateSaveResult contains the preview plus the written template branch result.
type TemplateSaveResult struct {
	Preview     TemplateSavePreview
	WriteResult gitpkg.WriteTemplateBranchResult
}

// BuildTemplateSavePreview runs the documented Phase 2A-1 runtime-to-template preview pipeline.
func BuildTemplateSavePreview(ctx context.Context, input TemplateSavePreviewInput) (TemplateSavePreview, error) {
	repoConfig, err := loadTemplateSaveRepositoryConfig(ctx, input.RepoRoot)
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("load repository config: %w", err)
	}
	if err := orbit.ValidateRepositoryConfig(repoConfig.Global, repoConfig.Orbits); err != nil {
		return TemplateSavePreview{}, fmt.Errorf("validate repository config: %w", err)
	}

	definition, found := repoConfig.OrbitByID(input.OrbitID)
	if !found {
		fallbackConfig, fallbackErr := orbit.LoadRepositoryConfig(ctx, input.RepoRoot)
		if fallbackErr != nil {
			return TemplateSavePreview{}, fmt.Errorf("orbit %q not found", input.OrbitID)
		}
		if err := orbit.ValidateRepositoryConfig(fallbackConfig.Global, fallbackConfig.Orbits); err != nil {
			return TemplateSavePreview{}, fmt.Errorf("validate repository config: %w", err)
		}
		fallbackDefinition, fallbackFound := fallbackConfig.OrbitByID(input.OrbitID)
		if !fallbackFound {
			return TemplateSavePreview{}, fmt.Errorf("orbit %q not found", input.OrbitID)
		}
		repoConfig = fallbackConfig
		definition = fallbackDefinition
	}

	trackedFiles, err := gitpkg.TrackedFiles(ctx, input.RepoRoot)
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("load tracked files: %w", err)
	}

	_, plan, err := orbit.LoadOrbitSpecAndProjectionPlan(ctx, input.RepoRoot, repoConfig, definition.ID, trackedFiles)
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("load orbit export plan: %w", err)
	}

	varsFile, _, err := loadOptionalRepoVarsFile(ctx, input.RepoRoot)
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("load runtime vars: %w", err)
	}

	buildResult, err := BuildTemplateContent(ctx, BuildInput{
		RepoRoot:  input.RepoRoot,
		OrbitID:   definition.ID,
		UserScope: plan.ExportPaths,
		Bindings:  varsFile.Variables,
	})
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("build template content: %w", err)
	}
	buildResult.Files = normalizeTemplateSaveFiles(buildResult.Files)
	sortFileReplacementSummaries(buildResult.ReplacementSummaries)
	sortFileReplacementAmbiguities(buildResult.Ambiguities)

	currentBranch, err := gitpkg.CurrentBranch(ctx, input.RepoRoot)
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("resolve current branch: %w", err)
	}
	headCommit, err := gitpkg.HeadCommit(ctx, input.RepoRoot)
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("resolve HEAD commit: %w", err)
	}

	files := buildResult.Files
	if input.EditTemplate {
		files, err = editTemplateFiles(ctx, definition.ID, buildResult.Files, input.Editor)
		if err != nil {
			return TemplateSavePreview{}, fmt.Errorf("edit template candidate: %w", err)
		}
		files = normalizeTemplateSaveFiles(files)
	}

	manifest, err := buildTemplateSaveManifest(
		definition.ID,
		input.DefaultBranch,
		currentBranch,
		headCommit,
		resolveSaveTime(input.Now),
		varsFile.Variables,
		files,
	)
	if err != nil {
		return TemplateSavePreview{}, fmt.Errorf("build template manifest: %w", err)
	}

	return TemplateSavePreview{
		RepoRoot:             input.RepoRoot,
		OrbitID:              definition.ID,
		TargetBranch:         input.TargetBranch,
		Files:                files,
		ReplacementSummaries: buildResult.ReplacementSummaries,
		Ambiguities:          buildResult.Ambiguities,
		Warnings:             nil,
		Manifest:             manifest,
	}, nil
}

func loadTemplateSaveRepositoryConfig(ctx context.Context, repoRoot string) (orbit.RepositoryConfig, error) {
	repoConfig, err := orbit.LoadRuntimeRepositoryConfig(ctx, repoRoot)
	if err == nil && len(repoConfig.Orbits) > 0 {
		return repoConfig, nil
	}

	discoveredConfig, discoveredErr := loadHostedDiscoveryRepositoryConfig(ctx, repoRoot)
	if discoveredErr == nil && len(discoveredConfig.Orbits) > 0 {
		return discoveredConfig, nil
	}

	if err == nil {
		return repoConfig, nil
	}
	if discoveredErr == nil {
		return discoveredConfig, nil
	}

	return orbit.RepositoryConfig{}, fmt.Errorf("load runtime repository config: %w", err)
}

func loadHostedDiscoveryRepositoryConfig(ctx context.Context, repoRoot string) (orbit.RepositoryConfig, error) {
	globalConfig, err := orbit.LoadGlobalConfig(ctx, repoRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			globalConfig = orbit.DefaultGlobalConfig()
		} else {
			return orbit.RepositoryConfig{}, fmt.Errorf("load global config: %w", err)
		}
	}

	definitions, err := orbit.DiscoverHostedDefinitions(ctx, repoRoot)
	if err != nil {
		return orbit.RepositoryConfig{}, fmt.Errorf("load hosted orbit definitions: %w", err)
	}

	return orbit.RepositoryConfig{
		Global: globalConfig,
		Orbits: definitions,
	}, nil
}

// SaveTemplateBranch writes the previewed template tree to a local branch through the Git writer.
func SaveTemplateBranch(ctx context.Context, input TemplateSaveInput) (TemplateSaveResult, error) {
	preview, err := BuildTemplateSavePreview(ctx, input.Preview)
	if err != nil {
		return TemplateSaveResult{}, err
	}
	if len(preview.Ambiguities) > 0 {
		return TemplateSaveResult{}, fmt.Errorf("replacement ambiguity detected; resolve the previewed ambiguities before saving")
	}

	files := make([]gitpkg.TemplateTreeFile, 0, len(preview.Files))
	for _, file := range preview.Files {
		files = append(files, gitpkg.TemplateTreeFile{
			Path:    file.Path,
			Content: file.Content,
			Mode:    file.Mode,
		})
	}

	branchManifest, err := branchManifestYAML(preview.Manifest)
	if err != nil {
		return TemplateSaveResult{}, fmt.Errorf("build branch manifest: %w", err)
	}

	writeResult, err := gitpkg.WriteTemplateBranch(ctx, preview.RepoRoot, gitpkg.WriteTemplateBranchInput{
		Branch:       preview.TargetBranch,
		Overwrite:    input.Overwrite,
		Message:      fmt.Sprintf("orbit template save %s", preview.OrbitID),
		ManifestPath: branchManifestPath,
		Manifest:     branchManifest,
		Files:        files,
	})
	if err != nil {
		return TemplateSaveResult{}, fmt.Errorf("write template branch: %w", err)
	}

	return TemplateSaveResult{
		Preview:     preview,
		WriteResult: writeResult,
	}, nil
}

func buildTemplateSaveManifest(
	orbitID string,
	defaultTemplate bool,
	createdFromBranch string,
	createdFromCommit string,
	createdAt time.Time,
	runtimeBindings map[string]bindings.VariableBinding,
	files []CandidateFile,
) (Manifest, error) {
	companionPath, _, err := templateCompanionPaths(orbitID)
	if err != nil {
		return Manifest{}, fmt.Errorf("build companion path: %w", err)
	}

	scanFiles := make([]CandidateFile, 0, len(files))
	for _, file := range files {
		if file.Path != companionPath {
			scanFiles = append(scanFiles, file)
		}
	}
	agentsCandidate, hasAgentsCandidate, err := companionAgentsCandidate(files, orbitID)
	if err != nil {
		return Manifest{}, fmt.Errorf("build companion agents candidate: %w", err)
	}
	if hasAgentsCandidate {
		scanFiles = append(scanFiles, agentsCandidate)
	}

	scanResult := ScanVariables(scanFiles, nil)
	variables := make(map[string]VariableSpec, len(scanResult.Referenced))
	for _, name := range scanResult.Referenced {
		variables[name] = VariableSpec{
			Description: runtimeBindings[name].Description,
			Required:    true,
		}
	}

	manifest := Manifest{
		SchemaVersion: manifestSchemaVersion,
		Kind:          TemplateKind,
		Template: Metadata{
			OrbitID:           orbitID,
			DefaultTemplate:   defaultTemplate,
			CreatedFromBranch: createdFromBranch,
			CreatedFromCommit: createdFromCommit,
			CreatedAt:         createdAt,
		},
		Variables: variables,
	}
	if err := ValidateManifest(manifest); err != nil {
		return Manifest{}, fmt.Errorf("validate manifest: %w", err)
	}

	return manifest, nil
}

func resolveSaveTime(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now().UTC()
	}

	return now.UTC()
}
