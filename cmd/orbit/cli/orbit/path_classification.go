package orbit

import (
	"fmt"

	"github.com/zack-nova/orbit/cmd/orbit/cli/ids"
)

const (
	PathRoleMeta    = string(OrbitMemberMeta)
	PathRoleSubject = string(OrbitMemberSubject)
	PathRoleRule    = string(OrbitMemberRule)
	PathRoleProcess = string(OrbitMemberProcess)
	PathRoleOutside = "outside"
)

// PathClassification captures the role-aware scope flags for one repo path.
type PathClassification struct {
	Role          string
	Projection    bool
	OrbitWrite    bool
	Export        bool
	Orchestration bool
}

// ClassifyOrbitPath classifies one repo-relative path against an already-resolved
// orbit spec and projection plan.
func ClassifyOrbitPath(
	config RepositoryConfig,
	spec OrbitSpec,
	plan ProjectionPlan,
	candidate string,
	tracked bool,
) (PathClassification, error) {
	normalizedPath, err := ids.NormalizeRepoRelativePath(candidate)
	if err != nil {
		return PathClassification{}, fmt.Errorf("normalize candidate path %q: %w", candidate, err)
	}

	if tracked {
		if classification, ok := trackedPathClassification(plan, normalizedPath); ok {
			return classification, nil
		}

		return PathClassification{Role: PathRoleOutside}, nil
	}

	return classifyUntrackedOrbitPath(config, spec, normalizedPath)
}

func trackedPathClassification(plan ProjectionPlan, normalizedPath string) (PathClassification, bool) {
	switch {
	case pathInGroup(plan.MetaPaths, normalizedPath):
		return classificationFromPlan(plan, PathRoleMeta, normalizedPath), true
	case pathInGroup(plan.SubjectPaths, normalizedPath):
		return classificationFromPlan(plan, PathRoleSubject, normalizedPath), true
	case pathInGroup(plan.RulePaths, normalizedPath):
		return classificationFromPlan(plan, PathRoleRule, normalizedPath), true
	case pathInGroup(plan.ProcessPaths, normalizedPath):
		return classificationFromPlan(plan, PathRoleProcess, normalizedPath), true
	default:
		return PathClassification{}, false
	}
}

func classifyUntrackedOrbitPath(
	config RepositoryConfig,
	spec OrbitSpec,
	normalizedPath string,
) (PathClassification, error) {
	companionPath, err := specControlPath(spec)
	if err != nil {
		return PathClassification{}, fmt.Errorf("resolve companion path: %w", err)
	}
	if normalizedPath == companionPath {
		return metaPathClassification(spec), nil
	}
	if isControlPlanePath(normalizedPath) {
		return PathClassification{Role: PathRoleOutside}, nil
	}

	if !spec.HasMemberSchema() {
		return classifyLegacyUntrackedOrbitPath(config.Global, spec.LegacyDefinition(), normalizedPath)
	}

	matchedMemberIndex := -1
	for index, member := range spec.Members {
		matches, matchErr := pathMatchesMember(member, normalizedPath)
		if matchErr != nil {
			return PathClassification{}, fmt.Errorf("match member %q for %q: %w", member.Key, normalizedPath, matchErr)
		}
		if !matches {
			continue
		}
		if matchedMemberIndex >= 0 {
			return PathClassification{}, fmt.Errorf(
				"path %q matches multiple orbit members %q and %q",
				normalizedPath,
				spec.Members[matchedMemberIndex].Key,
				member.Key,
			)
		}
		matchedMemberIndex = index
	}

	roleDefaults := defaultProjectionPlanRoleScopes(spec)
	if matchedMemberIndex >= 0 {
		member := spec.Members[matchedMemberIndex]
		flags := roleDefaults.forRole(member.Role)
		flags.applyPatch(member.Scopes)

		return classificationForRole(member.Role, flags), nil
	}

	sharedScopePatterns, err := normalizePatterns(config.Global.SharedScope)
	if err != nil {
		return PathClassification{}, fmt.Errorf("normalize shared scope patterns: %w", err)
	}
	inSharedScope, err := matchNormalizedAny(sharedScopePatterns, normalizedPath)
	if err != nil {
		return PathClassification{}, fmt.Errorf("match shared scope patterns for %q: %w", normalizedPath, err)
	}
	if inSharedScope {
		return classificationForRole(OrbitMemberSubject, roleDefaults.forRole(OrbitMemberSubject)), nil
	}

	projectionVisiblePatterns, err := normalizePatterns(config.Global.ProjectionVisible)
	if err != nil {
		return PathClassification{}, fmt.Errorf("normalize projection-visible patterns: %w", err)
	}
	projectionVisible, err := matchNormalizedAny(projectionVisiblePatterns, normalizedPath)
	if err != nil {
		return PathClassification{}, fmt.Errorf("match projection-visible patterns for %q: %w", normalizedPath, err)
	}
	if projectionVisible {
		return classificationForRole(OrbitMemberProcess, roleDefaults.forRole(OrbitMemberProcess)), nil
	}

	return PathClassification{Role: PathRoleOutside}, nil
}

func classifyLegacyUntrackedOrbitPath(
	config GlobalConfig,
	definition Definition,
	normalizedPath string,
) (PathClassification, error) {
	includePatterns, err := normalizePatterns(definition.Include)
	if err != nil {
		return PathClassification{}, fmt.Errorf("normalize include patterns: %w", err)
	}
	excludePatterns, err := normalizePatterns(definition.Exclude)
	if err != nil {
		return PathClassification{}, fmt.Errorf("normalize exclude patterns: %w", err)
	}
	sharedScopePatterns, err := normalizePatterns(config.SharedScope)
	if err != nil {
		return PathClassification{}, fmt.Errorf("normalize shared scope patterns: %w", err)
	}
	projectionVisiblePatterns, err := normalizePatterns(config.ProjectionVisible)
	if err != nil {
		return PathClassification{}, fmt.Errorf("normalize projection-visible patterns: %w", err)
	}

	excluded, err := matchNormalizedAny(excludePatterns, normalizedPath)
	if err != nil {
		return PathClassification{}, fmt.Errorf("match exclude patterns for %q: %w", normalizedPath, err)
	}
	if excluded {
		return PathClassification{Role: PathRoleOutside}, nil
	}

	included, err := matchNormalizedAny(includePatterns, normalizedPath)
	if err != nil {
		return PathClassification{}, fmt.Errorf("match include patterns for %q: %w", normalizedPath, err)
	}
	if included {
		return legacySubjectClassification(), nil
	}

	inSharedScope, err := matchNormalizedAny(sharedScopePatterns, normalizedPath)
	if err != nil {
		return PathClassification{}, fmt.Errorf("match shared scope patterns for %q: %w", normalizedPath, err)
	}
	if inSharedScope {
		return legacySubjectClassification(), nil
	}

	projectionVisible, err := matchNormalizedAny(projectionVisiblePatterns, normalizedPath)
	if err != nil {
		return PathClassification{}, fmt.Errorf("match projection-visible patterns for %q: %w", normalizedPath, err)
	}
	if projectionVisible {
		return legacyProcessClassification(), nil
	}

	return PathClassification{Role: PathRoleOutside}, nil
}

func classificationFromPlan(plan ProjectionPlan, role string, normalizedPath string) PathClassification {
	return PathClassification{
		Role:          role,
		Projection:    pathInGroup(plan.ProjectionPaths, normalizedPath),
		OrbitWrite:    pathInGroup(plan.OrbitWritePaths, normalizedPath),
		Export:        pathInGroup(plan.ExportPaths, normalizedPath),
		Orchestration: pathInGroup(plan.OrchestrationPaths, normalizedPath),
	}
}

func metaPathClassification(spec OrbitSpec) PathClassification {
	if spec.HasMemberSchema() && spec.Meta != nil {
		return PathClassification{
			Role:          PathRoleMeta,
			Projection:    spec.Meta.IncludeInProjection,
			OrbitWrite:    spec.Meta.IncludeInWrite,
			Export:        spec.Meta.IncludeInExport,
			Orchestration: spec.Meta.IncludeDescriptionInOrchestration,
		}
	}

	return PathClassification{
		Role:          PathRoleMeta,
		Projection:    true,
		OrbitWrite:    true,
		Export:        true,
		Orchestration: true,
	}
}

func classificationForRole(role OrbitMemberRole, flags planScopeFlags) PathClassification {
	return PathClassification{
		Role:          string(role),
		Projection:    flags.Projection,
		OrbitWrite:    flags.Write,
		Export:        flags.Export,
		Orchestration: flags.Orchestration,
	}
}

func legacySubjectClassification() PathClassification {
	return PathClassification{
		Role:          PathRoleSubject,
		Projection:    true,
		OrbitWrite:    true,
		Export:        true,
		Orchestration: false,
	}
}

func legacyProcessClassification() PathClassification {
	return PathClassification{
		Role:          PathRoleProcess,
		Projection:    true,
		OrbitWrite:    false,
		Export:        false,
		Orchestration: false,
	}
}

func pathInGroup(group []string, target string) bool {
	for _, candidate := range group {
		if candidate == target {
			return true
		}
	}

	return false
}
