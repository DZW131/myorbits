package orbittemplate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
)

type sourceBranchDefinitionHost string

const (
	sourceBranchDefinitionHostHosted sourceBranchDefinitionHost = "hosted"
	sourceBranchDefinitionHostLegacy sourceBranchDefinitionHost = "legacy"
)

func isAllowedSourceBranchHarnessPath(path string) bool {
	return path == sourceManifestRelativePath || strings.HasPrefix(path, ".harness/orbits/")
}

func loadSourceBranchRepositoryConfig(ctx context.Context, repoRoot string) (orbitpkg.RepositoryConfig, error) {
	config, err := orbitpkg.LoadRuntimeRepositoryConfig(ctx, repoRoot)
	if err != nil {
		return orbitpkg.RepositoryConfig{}, fmt.Errorf("load hosted repository config: %w", err)
	}
	if len(config.Orbits) > 0 {
		return config, nil
	}

	config, err = orbitpkg.LoadRepositoryConfig(ctx, repoRoot)
	if err != nil {
		return orbitpkg.RepositoryConfig{}, fmt.Errorf("load repository config: %w", err)
	}

	return config, nil
}

func loadSingleSourceOrbitDefinitionSelection(
	ctx context.Context,
	repoRoot string,
) (orbitpkg.Definition, sourceBranchDefinitionHost, error) {
	config, err := orbitpkg.LoadRuntimeRepositoryConfig(ctx, repoRoot)
	if err == nil && len(config.Orbits) > 0 {
		if err := orbitpkg.ValidateRepositoryConfig(config.Global, config.Orbits); err != nil {
			return orbitpkg.Definition{}, "", fmt.Errorf("validate repository config: %w", err)
		}
		orbitpkg.SortDefinitions(config.Orbits)
		if len(config.Orbits) != 1 {
			return orbitpkg.Definition{}, "", fmt.Errorf("source branch must contain exactly one orbit definition")
		}

		return config.Orbits[0], sourceBranchDefinitionHostHosted, nil
	}

	config, err = loadSourceBranchRepositoryConfig(ctx, repoRoot)
	if err != nil {
		return orbitpkg.Definition{}, "", err
	}
	if err := orbitpkg.ValidateRepositoryConfig(config.Global, config.Orbits); err != nil {
		return orbitpkg.Definition{}, "", fmt.Errorf("validate repository config: %w", err)
	}
	orbitpkg.SortDefinitions(config.Orbits)
	if len(config.Orbits) != 1 {
		return orbitpkg.Definition{}, "", fmt.Errorf("source branch must contain exactly one orbit definition")
	}

	return config.Orbits[0], sourceBranchDefinitionHostLegacy, nil
}

func loadSingleHostedSourceOrbitDefinition(ctx context.Context, repoRoot string) (orbitpkg.Definition, error) {
	config, err := orbitpkg.LoadRuntimeRepositoryConfig(ctx, repoRoot)
	if err != nil {
		return orbitpkg.Definition{}, fmt.Errorf("load hosted repository config: %w", err)
	}
	if err := orbitpkg.ValidateRepositoryConfig(config.Global, config.Orbits); err != nil {
		return orbitpkg.Definition{}, fmt.Errorf("validate repository config: %w", err)
	}
	orbitpkg.SortDefinitions(config.Orbits)

	switch len(config.Orbits) {
	case 1:
		legacyDefinitions, legacyErr := orbitpkg.DiscoverDefinitions(ctx, repoRoot)
		if legacyErr != nil {
			return orbitpkg.Definition{}, fmt.Errorf("discover legacy orbit definitions: %w", legacyErr)
		}
		if len(legacyDefinitions) > 0 {
			return orbitpkg.Definition{}, fmt.Errorf(
				"source publish requires hosted-only orbit definitions; remove legacy definitions from .orbit/orbits/ or run `orbit template init-source` to reconcile the source branch",
			)
		}
		return config.Orbits[0], nil
	case 0:
		legacyDefinitions, legacyErr := orbitpkg.DiscoverDefinitions(ctx, repoRoot)
		if legacyErr == nil && len(legacyDefinitions) > 0 {
			return orbitpkg.Definition{}, fmt.Errorf(
				"source publish requires hosted orbit definitions under .harness/orbits/; run `orbit template init-source` to migrate legacy definitions",
			)
		}
		return orbitpkg.Definition{}, fmt.Errorf("source branch must contain exactly one orbit definition")
	default:
		return orbitpkg.Definition{}, fmt.Errorf("source branch must contain exactly one orbit definition")
	}
}

func migrateLegacySourceOrbitDefinition(repoRoot string, definition orbitpkg.Definition) error {
	spec := orbitpkg.OrbitSpecFromDefinition(definition)
	spec.SourcePath = ""

	if _, err := orbitpkg.WriteHostedOrbitSpec(repoRoot, spec); err != nil {
		return fmt.Errorf("write hosted orbit definition for %q: %w", definition.ID, err)
	}

	legacyPath, err := orbitpkg.DefinitionPath(repoRoot, definition.ID)
	if err != nil {
		return fmt.Errorf("build legacy orbit definition path for %q: %w", definition.ID, err)
	}
	if err := os.Remove(legacyPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove legacy orbit definition %s: %w", legacyPath, err)
	}

	legacyDir := filepath.Dir(legacyPath)
	if err := os.Remove(legacyDir); err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("remove legacy orbit definitions dir %s: %w", legacyDir, err)
	}

	return nil
}
