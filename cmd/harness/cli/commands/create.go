package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	harnesspkg "github.com/zack-nova/orbit/cmd/orbit/cli/harness"
)

type createOutput struct {
	HarnessRoot      string `json:"harness_root"`
	ManifestPath     string `json:"manifest_path"`
	OrbitsDir        string `json:"orbits_dir"`
	GitInitialized   bool   `json:"git_initialized"`
	ManifestCreated  bool   `json:"manifest_created"`
	OrbitsDirCreated bool   `json:"orbits_dir_created"`
}

// NewCreateCommand creates the harness create command.
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <path>",
		Short: "Create a new harness runtime repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath, err := absolutePathFromArg(cmd, args[0])
			if err != nil {
				return err
			}
			if err := os.MkdirAll(targetPath, 0o750); err != nil {
				return fmt.Errorf("create target directory %s: %w", targetPath, err)
			}

			gitInitialized, err := ensureGitRepoRoot(cmd.Context(), targetPath)
			if err != nil {
				return err
			}

			repo, err := gitpkg.DiscoverRepo(cmd.Context(), targetPath)
			if err != nil {
				return fmt.Errorf("discover git repository: %w", err)
			}
			if comparablePath(repo.Root) != comparablePath(targetPath) {
				return fmt.Errorf("expected harness root %s to be a git repo root, got %s", targetPath, repo.Root)
			}

			manifestPath := harnesspkg.ManifestPath(repo.Root)
			if _, err := os.Stat(manifestPath); err == nil {
				manifest, loadErr := harnesspkg.LoadManifestFile(repo.Root)
				if loadErr == nil {
					if manifest.Kind != harnesspkg.ManifestKindRuntime {
						return fmt.Errorf("harness already initialized with non-runtime manifest at %s", manifestPath)
					}
					return fmt.Errorf("harness runtime already initialized at %s", manifestPath)
				}
				return fmt.Errorf("load existing harness manifest: %w", loadErr)
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("stat %s: %w", manifestPath, err)
			}

			bootstrap, err := harnesspkg.BootstrapRuntimeControlPlane(repo.Root, time.Now().UTC())
			if err != nil {
				return fmt.Errorf("bootstrap harness control plane: %w", err)
			}

			output := createOutput{
				HarnessRoot:      repo.Root,
				ManifestPath:     bootstrap.ManifestPath,
				OrbitsDir:        bootstrap.OrbitsDir,
				GitInitialized:   gitInitialized,
				ManifestCreated:  bootstrap.ManifestCreated,
				OrbitsDirCreated: bootstrap.OrbitsDirCreated,
			}

			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}
			if jsonOutput {
				return emitJSON(cmd.OutOrStdout(), output)
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "created harness in %s\n", repo.Root); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}

			return nil
		},
	}
	addJSONFlag(cmd)

	return cmd
}

func ensureGitRepoRoot(ctx context.Context, targetPath string) (bool, error) {
	repoRoot, err := gitpkg.RepoRoot(ctx, targetPath)
	switch {
	case err == nil && repoRoot == targetPath:
		return false, nil
	case err == nil:
		// target exists inside another repository; create an independent repo here.
	case err != nil:
		// not a repo yet; initialize below.
	}

	//nolint:gosec // Git is invoked with explicit argument lists from an internal command implementation.
	cmd := exec.CommandContext(ctx, "git", "init", targetPath)
	if output, runErr := cmd.CombinedOutput(); runErr != nil {
		return false, fmt.Errorf("git init %s: %w: %s", targetPath, runErr, string(output))
	}

	return true, nil
}

func comparablePath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return filepath.Clean(resolved)
	}

	return filepath.Clean(path)
}
