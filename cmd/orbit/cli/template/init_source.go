package orbittemplate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
)

// InitSourceResult summarizes one source-branch initialization run.
type InitSourceResult struct {
	RepoRoot           string
	SourceManifestPath string
	SourceBranch       string
	PublishOrbitID     string
	Changed            bool
}

// InitSourceBranch initializes one single-orbit source branch marker in the current branch.
func InitSourceBranch(ctx context.Context, repoRoot string) (InitSourceResult, error) {
	currentBranch, err := gitpkg.CurrentBranch(ctx, repoRoot)
	if err != nil {
		return InitSourceResult{}, fmt.Errorf("resolve current branch: %w", err)
	}
	if currentBranch == "HEAD" {
		return InitSourceResult{}, fmt.Errorf("init-source requires a current branch; detached HEAD is not supported")
	}

	templateManifestExists, err := gitpkg.PathExistsAtRev(ctx, repoRoot, "HEAD", manifestRelativePath)
	if err != nil {
		return InitSourceResult{}, fmt.Errorf("check %s at HEAD: %w", manifestRelativePath, err)
	}

	paths, err := gitpkg.ListAllFilesAtRev(ctx, repoRoot, "HEAD")
	if err != nil {
		return InitSourceResult{}, fmt.Errorf("list tracked files at HEAD: %w", err)
	}
	for _, path := range paths {
		if strings.HasPrefix(path, ".harness/") && !isAllowedSourceBranchHarnessPath(path) {
			return InitSourceResult{}, fmt.Errorf("source branch must not contain %s", path)
		}
	}

	definition, host, err := loadSingleSourceOrbitDefinitionSelection(ctx, repoRoot)
	if err != nil {
		return InitSourceResult{}, err
	}
	migratedLegacyDefinition := false
	if host == sourceBranchDefinitionHostLegacy {
		if err := migrateLegacySourceOrbitDefinition(repoRoot, definition); err != nil {
			return InitSourceResult{}, err
		}
		migratedLegacyDefinition = true
	}
	removedLegacyTemplateManifest := false
	if templateManifestExists {
		if err := removeLegacyTemplateManifest(repoRoot); err != nil {
			return InitSourceResult{}, err
		}
		removedLegacyTemplateManifest = true
	}

	manifest := SourceManifest{
		SchemaVersion: sourceSchemaVersion,
		Kind:          SourceKind,
		SourceBranch:  currentBranch,
		Publish: &SourcePublishConfig{
			OrbitID: definition.ID,
		},
	}
	result := InitSourceResult{
		RepoRoot:           repoRoot,
		SourceManifestPath: SourceManifestPath(repoRoot),
		SourceBranch:       currentBranch,
		PublishOrbitID:     definition.ID,
	}

	existing, err := LoadSourceManifest(repoRoot)
	switch {
	case err == nil:
		if reflect.DeepEqual(existing, manifest) {
			result.Changed = migratedLegacyDefinition || removedLegacyTemplateManifest
			return result, nil
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return InitSourceResult{}, fmt.Errorf("load %s: %w", sourceManifestRelativePath, err)
	}

	writtenPath, err := WriteSourceManifest(repoRoot, manifest)
	if err != nil {
		return InitSourceResult{}, fmt.Errorf("write %s: %w", sourceManifestRelativePath, err)
	}
	result.SourceManifestPath = writtenPath
	result.Changed = true

	return result, nil
}

func removeLegacyTemplateManifest(repoRoot string) error {
	filename := filepath.Join(repoRoot, filepath.FromSlash(manifestRelativePath))
	if err := os.Remove(filename); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", manifestRelativePath, err)
	}

	return nil
}
