package git_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	"github.com/zack-nova/orbit/cmd/orbit/cli/testutil"
)

func TestDiscoverRepoAndTrackedFiles(t *testing.T) {
	t.Parallel()

	repo := testutil.NewRepo(t)
	repo.WriteFile(t, "README.md", "hello\n")
	repo.WriteFile(t, "docs/guide with space.md", "guide\n")
	repo.AddAndCommit(t, "initial commit")

	subdir := filepath.Join(repo.Root, "docs")
	require.NoError(t, os.MkdirAll(subdir, 0o750))

	ctx := context.Background()

	discoveredRepo, err := gitpkg.DiscoverRepo(ctx, subdir)
	require.NoError(t, err)
	require.Equal(t, repo.Root, discoveredRepo.Root)
	require.Equal(t, repo.GitDir(t), discoveredRepo.GitDir)

	trackedFiles, err := gitpkg.TrackedFiles(ctx, repo.Root)
	require.NoError(t, err)
	require.Equal(t, []string{
		"README.md",
		"docs/guide with space.md",
	}, trackedFiles)
}
