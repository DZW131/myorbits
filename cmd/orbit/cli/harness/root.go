package harness

import (
	"context"
	"errors"
	"fmt"
	"os"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
)

// ResolvedRoot captures the harness repo root plus the validated control-plane documents loaded from it.
type ResolvedRoot struct {
	Repo     gitpkg.Repo
	Manifest ManifestFile
	Runtime  RuntimeFile
}

// ResolveRoot discovers the git repo root from any working directory and validates .harness/manifest.yaml at that root.
func ResolveRoot(ctx context.Context, workingDir string) (ResolvedRoot, error) {
	repo, err := gitpkg.DiscoverRepo(ctx, workingDir)
	if err != nil {
		return ResolvedRoot{}, fmt.Errorf("discover git repository: %w", err)
	}

	manifest, err := LoadManifestFile(repo.Root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ResolvedRoot{}, fmt.Errorf("harness manifest is not initialized at %s", repo.Root)
		}
		return ResolvedRoot{}, fmt.Errorf("load harness manifest: %w", err)
	}
	if manifest.Kind != ManifestKindRuntime {
		return ResolvedRoot{}, fmt.Errorf("harness root must contain a runtime manifest at %s", repo.Root)
	}

	runtime, err := RuntimeFileFromManifestFile(manifest)
	if err != nil {
		return ResolvedRoot{}, fmt.Errorf("convert harness manifest to runtime view: %w", err)
	}

	return ResolvedRoot{
		Repo:     repo,
		Manifest: manifest,
		Runtime:  runtime,
	}, nil
}
