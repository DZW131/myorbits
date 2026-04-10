package git_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	"github.com/zack-nova/orbit/cmd/orbit/cli/testutil"
)

func TestWithFetchedRemoteRefMakesRevisionReadableAndCleansUpOnSuccess(t *testing.T) {
	t.Parallel()

	source := testutil.NewRepo(t)
	source.WriteFile(t, "docs/guide.md", "remote template\n")
	source.AddAndCommit(t, "seed remote repo")
	source.Run(t, "branch", "-M", "template/docs")

	remoteURL := testutil.NewBareRemoteFromRepo(t, source)
	runtimeRepo := testutil.NewRepo(t)

	err := gitpkg.WithFetchedRemoteRef(context.Background(), runtimeRepo.Root, remoteURL, "refs/heads/template/docs", func(tempRef string) error {
		exists, err := gitpkg.RevisionExists(context.Background(), runtimeRepo.Root, tempRef)
		require.NoError(t, err)
		require.True(t, exists)

		data, err := gitpkg.ReadFileAtRev(context.Background(), runtimeRepo.Root, tempRef, "docs/guide.md")
		require.NoError(t, err)
		require.Equal(t, "remote template\n", string(data))

		return nil
	})
	require.NoError(t, err)
	require.Empty(t, runtimeRepo.Run(t, "for-each-ref", "--format=%(refname)", "refs/orbits/tmp/remote-source"))
}

func TestWithFetchedRemoteRefCleansUpOnCallbackError(t *testing.T) {
	t.Parallel()

	source := testutil.NewRepo(t)
	source.WriteFile(t, "docs/guide.md", "remote template\n")
	source.AddAndCommit(t, "seed remote repo")
	source.Run(t, "branch", "-M", "template/docs")

	remoteURL := testutil.NewBareRemoteFromRepo(t, source)
	runtimeRepo := testutil.NewRepo(t)
	callbackErr := errors.New("callback failed")

	err := gitpkg.WithFetchedRemoteRef(context.Background(), runtimeRepo.Root, remoteURL, "refs/heads/template/docs", func(_ string) error {
		return callbackErr
	})
	require.ErrorIs(t, err, callbackErr)
	require.Empty(t, runtimeRepo.Run(t, "for-each-ref", "--format=%(refname)", "refs/orbits/tmp/remote-source"))
}
