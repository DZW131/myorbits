package git

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/zack-nova/orbit/cmd/orbit/cli/ids"
)

// TrackedFiles returns the repository-relative tracked files from git ls-files -z.
func TrackedFiles(ctx context.Context, repoRoot string) ([]string, error) {
	output, err := runGit(ctx, repoRoot, "ls-files", "-z")
	if err != nil {
		return nil, fmt.Errorf("git ls-files -z: %w", err)
	}

	parts := parseNULTerminated(output)
	trackedFiles := make([]string, 0, len(parts))

	for _, part := range parts {
		normalized, err := ids.NormalizeRepoRelativePath(part)
		if err != nil {
			return nil, fmt.Errorf("normalize tracked file %q: %w", part, err)
		}

		trackedFiles = append(trackedFiles, normalized)
	}

	sort.Strings(trackedFiles)

	return trackedFiles, nil
}

func parseNULTerminated(data []byte) []string {
	parts := bytes.Split(data, []byte{0})
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		values = append(values, string(part))
	}

	return values
}
