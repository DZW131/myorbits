package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	"github.com/zack-nova/orbit/cmd/orbit/cli/ids"
	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	statepkg "github.com/zack-nova/orbit/cmd/orbit/cli/state"
	viewpkg "github.com/zack-nova/orbit/cmd/orbit/cli/view"
)

type contextKey string

const workingDirContextKey contextKey = "working_dir"

var errOrbitNotInitialized = errors.New("orbit is not initialized; run `orbit init` first")

// WithWorkingDir injects the working directory used by command tests.
func WithWorkingDir(ctx context.Context, workingDir string) context.Context {
	return context.WithValue(ctx, workingDirContextKey, workingDir)
}

func addJSONFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output machine-readable JSON")
}

func wantJSON(cmd *cobra.Command) (bool, error) {
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return false, fmt.Errorf("read json flag: %w", err)
	}

	return jsonOutput, nil
}

func emitJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode json output: %w", err)
	}

	return nil
}

func workingDirFromCommand(cmd *cobra.Command) (string, error) {
	if cmd.Context() != nil {
		if workingDir, ok := cmd.Context().Value(workingDirContextKey).(string); ok && strings.TrimSpace(workingDir) != "" {
			return workingDir, nil
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return workingDir, nil
}

func repoFromCommand(cmd *cobra.Command) (gitpkg.Repo, error) {
	workingDir, err := workingDirFromCommand(cmd)
	if err != nil {
		return gitpkg.Repo{}, err
	}

	repo, err := gitpkg.DiscoverRepo(cmd.Context(), workingDir)
	if err != nil {
		return gitpkg.Repo{}, fmt.Errorf("discover git repository: %w", err)
	}

	return repo, nil
}

func loadValidatedRepositoryConfig(ctx context.Context, repoRoot string) (orbitpkg.RepositoryConfig, error) {
	config, err := orbitpkg.LoadRuntimeRepositoryConfig(ctx, repoRoot)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return orbitpkg.RepositoryConfig{}, fmt.Errorf("load repository config: %w", err)
		}

		hostedConfig, hostedErr := orbitpkg.LoadHostedRepositoryConfig(ctx, repoRoot)
		if hostedErr != nil {
			return orbitpkg.RepositoryConfig{}, errOrbitNotInitialized
		}

		return validateLoadedRepositoryConfig(hostedConfig)
	}

	return validateLoadedRepositoryConfig(config)
}

func validateLoadedRepositoryConfig(config orbitpkg.RepositoryConfig) (orbitpkg.RepositoryConfig, error) {
	orbitpkg.SortDefinitions(config.Orbits)

	if err := orbitpkg.ValidateRepositoryConfig(config.Global, config.Orbits); err != nil {
		return orbitpkg.RepositoryConfig{}, fmt.Errorf("validate repository config: %w", err)
	}

	return config, nil
}

func loadValidatedAuthoringRepositoryConfig(ctx context.Context, repoRoot string) (orbitpkg.RepositoryConfig, error) {
	config, err := orbitpkg.LoadHostedRepositoryConfig(ctx, repoRoot)
	if err != nil {
		return orbitpkg.RepositoryConfig{}, fmt.Errorf("load repository config: %w", err)
	}

	orbitpkg.SortDefinitions(config.Orbits)

	if err := orbitpkg.ValidateRepositoryConfig(config.Global, config.Orbits); err != nil {
		return orbitpkg.RepositoryConfig{}, fmt.Errorf("validate repository config: %w", err)
	}

	return config, nil
}

func definitionByID(config orbitpkg.RepositoryConfig, orbitID string) (orbitpkg.Definition, error) {
	if err := ids.ValidateOrbitID(orbitID); err != nil {
		return orbitpkg.Definition{}, fmt.Errorf("validate orbit id: %w", err)
	}

	definition, found := config.OrbitByID(orbitID)
	if !found {
		return orbitpkg.Definition{}, fmt.Errorf("orbit %q not found", orbitID)
	}

	return definition, nil
}

type currentOrbitCommandContext struct {
	Repo       gitpkg.Repo
	Store      statepkg.FSStore
	Config     orbitpkg.RepositoryConfig
	Current    statepkg.CurrentOrbitState
	Definition orbitpkg.Definition
}

func loadCurrentOrbitCommandContext(cmd *cobra.Command) (currentOrbitCommandContext, error) {
	repo, err := repoFromCommand(cmd)
	if err != nil {
		return currentOrbitCommandContext{}, err
	}

	store, err := statepkg.NewFSStore(repo.GitDir)
	if err != nil {
		return currentOrbitCommandContext{}, fmt.Errorf("create state store: %w", err)
	}

	current, err := store.ReadCurrentOrbit()
	if err != nil {
		if errors.Is(err, statepkg.ErrCurrentOrbitNotFound) {
			return currentOrbitCommandContext{}, errors.New("current orbit is not set; run `orbit enter <orbit-id>` first")
		}
		return currentOrbitCommandContext{}, fmt.Errorf("read current orbit state: %w", err)
	}

	config, err := loadValidatedRepositoryConfig(cmd.Context(), repo.Root)
	if err != nil {
		return currentOrbitCommandContext{}, err
	}

	definition, err := viewpkg.CurrentDefinition(config, current)
	if err != nil {
		return currentOrbitCommandContext{}, fmt.Errorf("resolve current orbit definition: %w", err)
	}

	return currentOrbitCommandContext{
		Repo:       repo,
		Store:      store,
		Config:     config,
		Current:    current,
		Definition: definition,
	}, nil
}
