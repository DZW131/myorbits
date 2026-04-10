package harness

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
	"github.com/zack-nova/orbit/cmd/orbit/cli/git"
	"github.com/zack-nova/orbit/cmd/orbit/cli/internal/contractutil"
	"github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

// TemplateInstallPreviewInput describes one harness template install preview or apply analysis.
type TemplateInstallPreviewInput struct {
	RepoRoot                string
	Source                  LocalTemplateInstallSource
	InstallSource           orbittemplate.Source
	BindingsFilePath        string
	OverwriteExisting       bool
	Interactive             bool
	Prompter                orbittemplate.BindingPrompter
	EditorMode              bool
	Editor                  orbittemplate.Editor
	RequireResolvedBindings bool
	Now                     time.Time
}

// TemplateInstallPreview captures one harness template install preview plus mixed-install diagnostics.
type TemplateInstallPreview struct {
	Source                  LocalTemplateInstallSource
	InstallSource           orbittemplate.Source
	ResolvedBindings        map[string]bindings.ResolvedBinding
	RenderedDefinitionFiles []orbittemplate.CandidateFile
	RenderedFiles           []orbittemplate.CandidateFile
	RenderedRootAgentsFile  *orbittemplate.CandidateFile
	VarsFile                *bindings.VarsFile
	BundleRecord            BundleRecord
	Conflicts               []orbittemplate.ApplyConflict
	Warnings                []string
}

type templateInstallMaterialization struct {
	ResolvedBindings        map[string]bindings.ResolvedBinding
	RenderedDefinitionFiles []orbittemplate.CandidateFile
	RenderedFiles           []orbittemplate.CandidateFile
	RenderedRootAgentsFile  *orbittemplate.CandidateFile
	VarsFile                *bindings.VarsFile
	BundleRecord            BundleRecord
	Warnings                []string
}

// BuildTemplateInstallPreview analyzes one harness template install against the current runtime.
func BuildTemplateInstallPreview(
	ctx context.Context,
	input TemplateInstallPreviewInput,
) (TemplateInstallPreview, error) {
	runtimeFile, err := LoadRuntimeFile(input.RepoRoot)
	if err != nil {
		return TemplateInstallPreview{}, fmt.Errorf("load harness runtime: %w", err)
	}

	statusEntries, err := git.WorktreeStatus(ctx, input.RepoRoot)
	if err != nil {
		return TemplateInstallPreview{}, fmt.Errorf("load runtime worktree status: %w", err)
	}
	statusByPath := make(map[string]git.StatusEntry, len(statusEntries))
	for _, entry := range statusEntries {
		statusByPath[entry.Path] = entry
	}

	materialization, err := buildTemplateInstallMaterialization(ctx, input)
	if err != nil {
		return TemplateInstallPreview{}, err
	}

	conflicts := make([]orbittemplate.ApplyConflict, 0)
	var existingBundleRecord BundleRecord
	hasExistingBundleRecord := false
	if record, err := LoadBundleRecord(input.RepoRoot, input.Source.Manifest.Template.HarnessID); err == nil {
		existingBundleRecord = record
		hasExistingBundleRecord = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return TemplateInstallPreview{}, fmt.Errorf("load existing bundle record: %w", err)
	}
	sameUnitReplace := input.OverwriteExisting && hasExistingBundleRecord
	allowedOwnedPaths := make(map[string]struct{})
	allowedBundleMembers := make(map[string]struct{})
	if sameUnitReplace {
		for _, path := range existingBundleRecord.OwnedPaths {
			allowedOwnedPaths[path] = struct{}{}
		}
		for _, memberID := range existingBundleRecord.MemberIDs {
			allowedBundleMembers[memberID] = struct{}{}
		}
	}

	for _, member := range input.Source.Manifest.Members {
		for _, existing := range runtimeFile.Members {
			if existing.OrbitID != member.OrbitID {
				continue
			}
			if sameUnitReplace && existing.Source == MemberSourceInstallBundle {
				if _, ok := allowedBundleMembers[member.OrbitID]; ok {
					continue
				}
			}
			conflicts = append(conflicts, orbittemplate.ApplyConflict{
				Path:    ManifestRepoPath(),
				Message: fmt.Sprintf("member %q already exists in harness runtime", member.OrbitID),
			})
		}
	}

	if hasExistingBundleRecord && !sameUnitReplace {
		conflicts = append(conflicts, orbittemplate.ApplyConflict{
			Path:    BundleRecordsDirRepoPath(),
			Message: fmt.Sprintf("bundle %q already exists in harness runtime", input.Source.Manifest.Template.HarnessID),
		})
	}

	for _, file := range materialization.RenderedDefinitionFiles {
		conflictsForFile, err := analyzeTemplateInstallPathConflict(input.RepoRoot, file, statusByPath, allowedOwnedPaths)
		if err != nil {
			return TemplateInstallPreview{}, err
		}
		conflicts = append(conflicts, conflictsForFile...)
	}
	for _, file := range materialization.RenderedFiles {
		conflictsForFile, err := analyzeTemplateInstallPathConflict(input.RepoRoot, file, statusByPath, allowedOwnedPaths)
		if err != nil {
			return TemplateInstallPreview{}, err
		}
		conflicts = append(conflicts, conflictsForFile...)
	}
	if materialization.RenderedRootAgentsFile != nil {
		_, allowAgentsStatusConflict := allowedOwnedPaths[rootAgentsPath]
		agentsConflicts, err := analyzeBundleAgentsInstallPreview(
			input.RepoRoot,
			input.Source.Manifest.Template.HarnessID,
			statusByPath,
			allowAgentsStatusConflict,
		)
		if err != nil {
			return TemplateInstallPreview{}, err
		}
		conflicts = append(conflicts, agentsConflicts...)
	}

	currentVarsFile, err := LoadVarsFile(input.RepoRoot)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return TemplateInstallPreview{}, fmt.Errorf("load harness vars file: %w", err)
		}
	} else {
		for name, next := range input.Source.Manifest.Variables {
			existing, ok := currentVarsFile.Variables[name]
			if !ok {
				continue
			}
			if _, err := mergeTemplateVariableSpec(name, TemplateVariableSpec{
				Description: existing.Description,
				Required:    false,
			}, next); err != nil {
				conflicts = append(conflicts, orbittemplate.ApplyConflict{
					Path:    VarsRepoPath(),
					Message: err.Error(),
				})
			}
		}
	}

	sort.Slice(conflicts, func(left, right int) bool {
		if conflicts[left].Path == conflicts[right].Path {
			return conflicts[left].Message < conflicts[right].Message
		}
		return conflicts[left].Path < conflicts[right].Path
	})
	sort.Strings(materialization.Warnings)

	return TemplateInstallPreview{
		Source:                  input.Source,
		InstallSource:           input.InstallSource,
		ResolvedBindings:        materialization.ResolvedBindings,
		RenderedDefinitionFiles: materialization.RenderedDefinitionFiles,
		RenderedFiles:           materialization.RenderedFiles,
		RenderedRootAgentsFile:  materialization.RenderedRootAgentsFile,
		VarsFile:                materialization.VarsFile,
		BundleRecord:            materialization.BundleRecord,
		Conflicts:               conflicts,
		Warnings:                materialization.Warnings,
	}, nil
}

// ApplyTemplateInstallPreview writes one harness template install into the runtime repository.
func ApplyTemplateInstallPreview(
	ctx context.Context,
	repoRoot string,
	preview TemplateInstallPreview,
	overwriteExisting bool,
) (TemplateInstallResult, error) {
	if len(preview.Conflicts) > 0 {
		return TemplateInstallResult{}, fmt.Errorf(
			"conflicts detected; mixed harness template install requires disjoint targets: %s",
			preview.Conflicts[0].Message,
		)
	}

	var cleanupPlan bundleOwnedCleanupPlan
	var hasCleanupPlan bool
	if overwriteExisting {
		existingRecord, err := LoadBundleRecord(repoRoot, preview.Source.Manifest.Template.HarnessID)
		if err == nil {
			cleanupPlan = buildBundleOwnedCleanupPlan(existingRecord, preview)
			hasCleanupPlan = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return TemplateInstallResult{}, fmt.Errorf("load existing bundle record: %w", err)
		}
	}

	writtenPaths := make([]string, 0, len(preview.RenderedDefinitionFiles)+len(preview.RenderedFiles)+4)
	for _, file := range preview.RenderedDefinitionFiles {
		if err := writeTemplateCandidateFile(repoRoot, file); err != nil {
			return TemplateInstallResult{}, err
		}
		writtenPaths = append(writtenPaths, file.Path)
	}
	for _, file := range preview.RenderedFiles {
		if err := writeTemplateCandidateFile(repoRoot, file); err != nil {
			return TemplateInstallResult{}, err
		}
		writtenPaths = append(writtenPaths, file.Path)
	}
	if preview.RenderedRootAgentsFile != nil {
		if err := ApplyBundleAgentsPayload(repoRoot, preview.Source.Manifest.Template.HarnessID, preview.RenderedRootAgentsFile.Content); err != nil {
			return TemplateInstallResult{}, fmt.Errorf("write runtime AGENTS.md: %w", err)
		}
		writtenPaths = append(writtenPaths, rootAgentsPath)
	}
	if preview.VarsFile != nil {
		varsPath, err := WriteVarsFile(repoRoot, *preview.VarsFile)
		if err != nil {
			return TemplateInstallResult{}, fmt.Errorf("write harness vars file: %w", err)
		}
		writtenPaths = append(writtenPaths, mustRepoRelativePath(repoRoot, varsPath))
	}

	bundlePath, err := WriteBundleRecord(repoRoot, preview.BundleRecord)
	if err != nil {
		return TemplateInstallResult{}, fmt.Errorf("write bundle record: %w", err)
	}
	writtenPaths = append(writtenPaths, mustRepoRelativePath(repoRoot, bundlePath))

	var memberResult MutateMembersResult
	if hasCleanupPlan {
		memberResult, err = ReplaceBundleMembers(ctx, repoRoot, cleanupPlan.PreviousMemberIDs, preview.Source.MemberIDs(), preview.BundleRecord.AppliedAt)
		if err != nil {
			return TemplateInstallResult{}, fmt.Errorf("replace bundle-backed members: %w", err)
		}
	} else {
		memberResult, err = AddBundleMembers(ctx, repoRoot, preview.Source.MemberIDs(), preview.BundleRecord.AppliedAt)
		if err != nil {
			return TemplateInstallResult{}, fmt.Errorf("record bundle-backed members: %w", err)
		}
	}
	writtenPaths = append(writtenPaths, mustRepoRelativePath(repoRoot, memberResult.ManifestPath))

	if hasCleanupPlan {
		removedPaths, err := applyBundleOwnedCleanup(repoRoot, preview.Source.Manifest.Template.HarnessID, cleanupPlan)
		if err != nil {
			return TemplateInstallResult{}, fmt.Errorf("remove stale bundle-owned paths: %w", err)
		}
		writtenPaths = append(writtenPaths, removedPaths...)
	}

	sort.Strings(writtenPaths)
	writtenPaths = slicesCompactStrings(writtenPaths)

	return TemplateInstallResult{
		Preview:      preview,
		Runtime:      memberResult.Runtime,
		ManifestPath: memberResult.ManifestPath,
		BundlePath:   bundlePath,
		WrittenPaths: writtenPaths,
	}, nil
}

type bundleOwnedCleanupPlan struct {
	DeletePaths           []string
	RemoveRootAgentsBlock bool
	PreviousMemberIDs     []string
}

// TemplateInstallResult contains one successful harness template install write result.
type TemplateInstallResult struct {
	Preview      TemplateInstallPreview
	Runtime      RuntimeFile
	ManifestPath string
	BundlePath   string
	WrittenPaths []string
}

func buildRenderedTemplateInstallPayload(
	ctx context.Context,
	input TemplateInstallPreviewInput,
) (map[string]bindings.ResolvedBinding, *bindings.VarsFile, []orbittemplate.CandidateFile, *orbittemplate.CandidateFile, []string, error) {
	ordinaryFiles, rootAgentsFile := splitRootAgentsTemplateFiles(input.Source.Files)
	renderedFiles := cloneCandidateFiles(ordinaryFiles)
	var renderedRootAgentsFile *orbittemplate.CandidateFile
	if rootAgentsFile != nil {
		cloned := cloneInstallCandidateFile(*rootAgentsFile)
		renderedRootAgentsFile = &cloned
	}

	declared := make(map[string]bindings.VariableDeclaration, len(input.Source.Manifest.Variables))
	for name, spec := range input.Source.Manifest.Variables {
		declared[name] = bindings.VariableDeclaration{
			Description: spec.Description,
			Required:    spec.Required,
		}
	}
	if len(declared) == 0 {
		return map[string]bindings.ResolvedBinding{}, nil, renderedFiles, renderedRootAgentsFile, []string{}, nil
	}

	bindingsFile, err := loadOptionalTemplateInstallBindingsFile(input.BindingsFilePath)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("load --bindings file: %w", err)
	}
	repoVarsFile, hasRepoVarsFile, err := loadOptionalTemplateInstallRepoVarsFile(ctx, input.RepoRoot)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("load harness vars: %w", err)
	}

	mergeResult, err := resolveTemplateInstallBindings(
		ctx,
		declared,
		bindingsFile,
		repoVarsFile.Variables,
		input.Interactive,
		input.Prompter,
		input.EditorMode,
		input.Editor,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("merge bindings: %w", err)
	}
	if len(mergeResult.Unresolved) > 0 {
		names := make([]string, 0, len(mergeResult.Unresolved))
		for _, unresolved := range mergeResult.Unresolved {
			names = append(names, unresolved.Name)
		}
		sort.Strings(names)
		if input.RequireResolvedBindings {
			return nil, nil, nil, nil, nil, fmt.Errorf("missing required bindings: %s", strings.Join(names, ", "))
		}

		return mergeResult.Resolved, nil, renderedFiles, renderedRootAgentsFile, []string{
			fmt.Sprintf("preview kept template variables unresolved: %s", strings.Join(names, ", ")),
		}, nil
	}

	renderValues := make(map[string]string, len(mergeResult.Resolved))
	for name, resolved := range mergeResult.Resolved {
		renderValues[name] = resolved.Value
	}

	renderedFiles, err = orbittemplate.RenderTemplateFiles(ordinaryFiles, renderValues)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("render harness template files: %w", err)
	}
	if rootAgentsFile != nil {
		renderedAgentsFiles, err := orbittemplate.RenderTemplateFiles([]orbittemplate.CandidateFile{*rootAgentsFile}, renderValues)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("render harness template AGENTS.md: %w", err)
		}
		renderedRootAgentsFile = &renderedAgentsFiles[0]
	}

	varsFile := planTemplateInstallBindingsWrite(repoVarsFile, hasRepoVarsFile, mergeResult.Resolved)

	return mergeResult.Resolved, varsFile, renderedFiles, renderedRootAgentsFile, []string{}, nil
}

func buildTemplateInstallMaterialization(
	ctx context.Context,
	input TemplateInstallPreviewInput,
) (templateInstallMaterialization, error) {
	resolvedBindings, varsFile, renderedFiles, renderedRootAgentsFile, warnings, err := buildRenderedTemplateInstallPayload(ctx, input)
	if err != nil {
		return templateInstallMaterialization{}, err
	}
	renderedDefinitionFiles, err := materializeRuntimeDefinitionFiles(input.Source.DefinitionFiles)
	if err != nil {
		return templateInstallMaterialization{}, err
	}

	return templateInstallMaterialization{
		ResolvedBindings:        resolvedBindings,
		RenderedDefinitionFiles: renderedDefinitionFiles,
		RenderedFiles:           renderedFiles,
		RenderedRootAgentsFile:  renderedRootAgentsFile,
		VarsFile:                varsFile,
		BundleRecord:            buildBundleInstallRecord(input.Source, input.InstallSource, renderedDefinitionFiles, renderedFiles, renderedRootAgentsFile, input.Now),
		Warnings:                warnings,
	}, nil
}

func materializeRuntimeDefinitionFiles(files []orbittemplate.CandidateFile) ([]orbittemplate.CandidateFile, error) {
	rendered := cloneCandidateFiles(files)
	for index, file := range rendered {
		name := strings.TrimSuffix(filepath.Base(file.Path), filepath.Ext(file.Path))
		runtimePath, err := orbit.HostedDefinitionRelativePath(name)
		if err != nil {
			return nil, fmt.Errorf("build runtime definition path for %q: %w", file.Path, err)
		}
		rendered[index].Path = runtimePath
	}

	return rendered, nil
}

func analyzeTemplateInstallPathConflict(
	repoRoot string,
	file orbittemplate.CandidateFile,
	statusByPath map[string]git.StatusEntry,
	allowedOwnedPaths map[string]struct{},
) ([]orbittemplate.ApplyConflict, error) {
	conflicts := make([]orbittemplate.ApplyConflict, 0, 2)
	filename := filepath.Join(repoRoot, filepath.FromSlash(file.Path))

	//nolint:gosec // The target path is repo-local and derived from one validated template candidate path.
	if data, err := os.ReadFile(filename); err == nil {
		if !bytes.Equal(data, file.Content) {
			if _, ok := allowedOwnedPaths[file.Path]; !ok {
				conflicts = append(conflicts, orbittemplate.ApplyConflict{
					Path:    file.Path,
					Message: "target path already exists with different content",
				})
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read runtime path %s: %w", file.Path, err)
	}

	if status, ok := statusByPath[file.Path]; ok {
		if _, ok := allowedOwnedPaths[file.Path]; !ok {
			conflicts = append(conflicts, orbittemplate.ApplyConflict{
				Path:    file.Path,
				Message: fmt.Sprintf("target path has uncommitted worktree status %s", status.Code),
			})
		}
	}

	return conflicts, nil
}

func analyzeBundleAgentsInstallPreview(
	repoRoot string,
	harnessID string,
	statusByPath map[string]git.StatusEntry,
	allowExistingStatus bool,
) ([]orbittemplate.ApplyConflict, error) {
	conflicts := make([]orbittemplate.ApplyConflict, 0, 1)
	if status, ok := statusByPath[rootAgentsPath]; ok {
		if !allowExistingStatus {
			conflicts = append(conflicts, orbittemplate.ApplyConflict{
				Path:    rootAgentsPath,
				Message: fmt.Sprintf("AGENTS lane has uncommitted worktree status %s", status.Code),
			})
		}
	}

	filename := filepath.Join(repoRoot, rootAgentsPath)
	//nolint:gosec // The runtime AGENTS path is fixed under the repo root.
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return conflicts, nil
		}
		return nil, fmt.Errorf("read runtime AGENTS.md: %w", err)
	}

	if _, err := orbittemplate.ParseRuntimeAgentsDocument(data); err != nil {
		conflicts = append(conflicts, orbittemplate.ApplyConflict{
			Path:    rootAgentsPath,
			Message: fmt.Sprintf("runtime AGENTS.md is invalid for harness block merge (%s)", harnessID),
		})
	}

	return conflicts, nil
}

func buildBundleInstallRecord(
	source LocalTemplateInstallSource,
	installSource orbittemplate.Source,
	definitionFiles []orbittemplate.CandidateFile,
	renderedFiles []orbittemplate.CandidateFile,
	renderedRootAgentsFile *orbittemplate.CandidateFile,
	now time.Time,
) BundleRecord {
	ownedPaths := make([]string, 0, len(definitionFiles)+len(renderedFiles)+1)
	for _, file := range definitionFiles {
		ownedPaths = append(ownedPaths, file.Path)
	}
	for _, file := range renderedFiles {
		ownedPaths = append(ownedPaths, file.Path)
	}
	if renderedRootAgentsFile != nil {
		ownedPaths = append(ownedPaths, renderedRootAgentsFile.Path)
	}
	sort.Strings(ownedPaths)
	ownedPaths = slicesCompactStrings(ownedPaths)

	return BundleRecord{
		SchemaVersion:      bundleRecordSchemaVersion,
		HarnessID:          source.Manifest.Template.HarnessID,
		Template:           installSource,
		MemberIDs:          source.MemberIDs(),
		AppliedAt:          resolveMutationTime(now),
		IncludesRootAgents: renderedRootAgentsFile != nil,
		OwnedPaths:         ownedPaths,
	}
}

func buildBundleOwnedCleanupPlan(existingRecord BundleRecord, nextPreview TemplateInstallPreview) bundleOwnedCleanupPlan {
	nextOwnedPaths := make(map[string]struct{}, len(nextPreview.BundleRecord.OwnedPaths))
	for _, path := range nextPreview.BundleRecord.OwnedPaths {
		nextOwnedPaths[path] = struct{}{}
	}

	plan := bundleOwnedCleanupPlan{
		DeletePaths:       make([]string, 0),
		PreviousMemberIDs: append([]string(nil), existingRecord.MemberIDs...),
	}
	for _, oldPath := range existingRecord.OwnedPaths {
		if _, stillOwned := nextOwnedPaths[oldPath]; stillOwned {
			continue
		}
		if oldPath == rootAgentsPath {
			plan.RemoveRootAgentsBlock = true
			continue
		}
		plan.DeletePaths = append(plan.DeletePaths, oldPath)
	}
	sort.Strings(plan.DeletePaths)

	return plan
}

func applyBundleOwnedCleanup(repoRoot string, harnessID string, plan bundleOwnedCleanupPlan) ([]string, error) {
	removed := make([]string, 0, len(plan.DeletePaths)+1)
	for _, path := range plan.DeletePaths {
		filename := filepath.Join(repoRoot, filepath.FromSlash(path))
		if err := os.Remove(filename); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("remove stale bundle-owned path %s: %w", path, err)
		}
		removed = append(removed, path)
	}

	if plan.RemoveRootAgentsBlock {
		if err := RemoveBundleAgentsPayload(repoRoot, harnessID); err != nil {
			return nil, fmt.Errorf("remove stale bundle AGENTS block: %w", err)
		}
		removed = append(removed, rootAgentsPath)
	}

	sort.Strings(removed)
	return removed, nil
}

func writeTemplateCandidateFile(repoRoot string, file orbittemplate.CandidateFile) error {
	filename := filepath.Join(repoRoot, filepath.FromSlash(file.Path))
	perm, err := git.FilePermForMode(file.Mode)
	if err != nil {
		return fmt.Errorf("resolve rendered file mode %s: %w", file.Path, err)
	}
	if err := contractutil.AtomicWriteFileMode(filename, file.Content, perm); err != nil {
		return fmt.Errorf("write rendered file %s: %w", file.Path, err)
	}
	return nil
}

func splitRootAgentsTemplateFiles(files []orbittemplate.CandidateFile) ([]orbittemplate.CandidateFile, *orbittemplate.CandidateFile) {
	ordinary := make([]orbittemplate.CandidateFile, 0, len(files))
	var rootAgentsFile *orbittemplate.CandidateFile
	for _, file := range files {
		if file.Path == rootAgentsPath {
			cloned := cloneInstallCandidateFile(file)
			rootAgentsFile = &cloned
			continue
		}
		ordinary = append(ordinary, cloneInstallCandidateFile(file))
	}
	return ordinary, rootAgentsFile
}

func cloneCandidateFiles(files []orbittemplate.CandidateFile) []orbittemplate.CandidateFile {
	cloned := make([]orbittemplate.CandidateFile, 0, len(files))
	for _, file := range files {
		cloned = append(cloned, cloneInstallCandidateFile(file))
	}
	return cloned
}

func cloneInstallCandidateFile(file orbittemplate.CandidateFile) orbittemplate.CandidateFile {
	return orbittemplate.CandidateFile{
		Path:    file.Path,
		Content: append([]byte(nil), file.Content...),
		Mode:    file.Mode,
	}
}

func loadOptionalTemplateInstallBindingsFile(filename string) (map[string]bindings.VariableBinding, error) {
	if strings.TrimSpace(filename) == "" {
		return map[string]bindings.VariableBinding{}, nil
	}

	//nolint:gosec // The bindings file path is an explicit user-provided local file path.
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filename, err)
	}
	file, err := bindings.ParseVarsData(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}
	return file.Variables, nil
}

func loadOptionalTemplateInstallRepoVarsFile(ctx context.Context, repoRoot string) (bindings.VarsFile, bool, error) {
	empty := bindings.VarsFile{
		SchemaVersion: 1,
		Variables:     map[string]bindings.VariableBinding{},
	}
	if _, err := os.Stat(VarsPath(repoRoot)); err == nil {
		file, err := LoadVarsFile(repoRoot)
		if err != nil {
			return bindings.VarsFile{}, false, err
		}
		return file, true, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return bindings.VarsFile{}, false, fmt.Errorf("stat harness vars in worktree: %w", err)
	}

	existsAtHead, err := git.PathExistsAtRev(ctx, repoRoot, "HEAD", VarsRepoPath())
	if err != nil {
		return bindings.VarsFile{}, false, fmt.Errorf("check harness vars at HEAD: %w", err)
	}
	if !existsAtHead {
		return empty, false, nil
	}

	file, err := LoadVarsFileWorktreeOrHEAD(ctx, repoRoot)
	if err != nil {
		return bindings.VarsFile{}, false, err
	}
	return file, true, nil
}

func planTemplateInstallBindingsWrite(
	existing bindings.VarsFile,
	hasExisting bool,
	resolved map[string]bindings.ResolvedBinding,
) *bindings.VarsFile {
	if len(resolved) == 0 {
		return nil
	}

	merged := make(map[string]bindings.VariableBinding, len(existing.Variables)+len(resolved))
	for name, binding := range existing.Variables {
		merged[name] = binding
	}

	changed := false
	for name, binding := range resolved {
		if binding.Source == bindings.SourceRepoVars {
			continue
		}

		next := bindings.VariableBinding{
			Value:       binding.Value,
			Description: binding.Description,
		}
		if current, ok := merged[name]; ok && current == next {
			continue
		}
		merged[name] = next
		changed = true
	}

	if !changed {
		return nil
	}

	planned := bindings.VarsFile{
		SchemaVersion: 1,
		Variables:     merged,
	}
	if hasExisting && existing.SchemaVersion == planned.SchemaVersion && mapsEqualBindings(existing.Variables, planned.Variables) {
		return nil
	}
	return &planned
}

func resolveTemplateInstallBindings(
	ctx context.Context,
	declared map[string]bindings.VariableDeclaration,
	bindingsFile map[string]bindings.VariableBinding,
	repoVars map[string]bindings.VariableBinding,
	interactive bool,
	prompter orbittemplate.BindingPrompter,
	editorMode bool,
	editor orbittemplate.Editor,
) (bindings.MergeResult, error) {
	mergeInput := bindings.MergeInput{
		Declared:     declared,
		BindingsFile: bindingsFile,
		RepoVars:     repoVars,
	}
	mergeResult, err := bindings.Merge(mergeInput)
	if err != nil {
		return bindings.MergeResult{}, fmt.Errorf("merge declared bindings before interactive fill: %w", err)
	}
	if len(mergeResult.Unresolved) == 0 {
		return mergeResult, nil
	}

	switch {
	case interactive:
		if prompter == nil {
			return bindings.MergeResult{}, fmt.Errorf("interactive apply requires a binding prompter")
		}
		fillIn, err := prompter.PromptBindings(ctx, mergeResult.Unresolved)
		if err != nil {
			return bindings.MergeResult{}, fmt.Errorf("prompt for missing bindings: %w", err)
		}
		mergeInput.FillIn = fillIn
		mergeInput.FillSource = bindings.SourceInteractive
	case editorMode:
		fillIn, err := editMissingTemplateInstallBindings(ctx, mergeResult.Unresolved, editor)
		if err != nil {
			return bindings.MergeResult{}, fmt.Errorf("edit missing bindings: %w", err)
		}
		mergeInput.FillIn = fillIn
		mergeInput.FillSource = bindings.SourceEditor
	default:
		return mergeResult, nil
	}

	mergeResult, err = bindings.Merge(mergeInput)
	if err != nil {
		return bindings.MergeResult{}, fmt.Errorf("merge bindings after fill: %w", err)
	}
	return mergeResult, nil
}

func editMissingTemplateInstallBindings(
	ctx context.Context,
	unresolved []bindings.UnresolvedBinding,
	editor orbittemplate.Editor,
) (map[string]bindings.VariableBinding, error) {
	if editor == nil {
		return nil, fmt.Errorf("editor apply requires an editor")
	}

	declared := make(map[string]bindings.VariableDeclaration, len(unresolved))
	for _, missing := range unresolved {
		declared[missing.Name] = bindings.VariableDeclaration{
			Description: missing.Description,
			Required:    missing.Required,
		}
	}
	skeleton := bindings.SkeletonFromDeclarations(declared)
	data, err := bindings.MarshalVarsFile(skeleton)
	if err != nil {
		return nil, fmt.Errorf("encode editor bindings skeleton: %w", err)
	}

	tempFile, err := os.CreateTemp("", "orbit-bindings-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("create editor bindings temp file: %w", err)
	}
	tempName := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("close editor bindings temp file: %w", err)
	}
	defer func() { _ = os.Remove(tempName) }()

	//nolint:gosec // tempName comes from os.CreateTemp in this function and stays confined to the local temp dir.
	if err := os.WriteFile(tempName, data, 0o600); err != nil {
		return nil, fmt.Errorf("write editor bindings skeleton: %w", err)
	}
	if err := editor.Edit(ctx, tempName); err != nil {
		return nil, fmt.Errorf("run bindings editor: %w", err)
	}

	//nolint:gosec // tempName comes from os.CreateTemp in this function and is only rewritten by the configured editor process.
	editedData, err := os.ReadFile(tempName)
	if err != nil {
		return nil, fmt.Errorf("read edited bindings skeleton: %w", err)
	}
	editedFile, err := bindings.ParseVarsData(editedData)
	if err != nil {
		return nil, fmt.Errorf("parse edited bindings skeleton: %w", err)
	}

	filled := make(map[string]bindings.VariableBinding, len(editedFile.Variables))
	for name, binding := range editedFile.Variables {
		if strings.TrimSpace(binding.Value) == "" {
			continue
		}
		filled[name] = binding
	}
	return filled, nil
}

func mapsEqualBindings(left map[string]bindings.VariableBinding, right map[string]bindings.VariableBinding) bool {
	if len(left) != len(right) {
		return false
	}
	for name, leftBinding := range left {
		if rightBinding, ok := right[name]; !ok || rightBinding != leftBinding {
			return false
		}
	}
	return true
}

func mustRepoRelativePath(repoRoot string, absolutePath string) string {
	relativePath, err := filepath.Rel(repoRoot, absolutePath)
	if err != nil {
		return filepath.ToSlash(absolutePath)
	}
	return filepath.ToSlash(relativePath)
}

func slicesCompactStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	sort.Strings(values)
	deduped := values[:0]
	for _, value := range values {
		if len(deduped) > 0 && deduped[len(deduped)-1] == value {
			continue
		}
		deduped = append(deduped, value)
	}
	return deduped
}
