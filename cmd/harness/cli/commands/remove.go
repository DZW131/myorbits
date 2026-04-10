package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	harnesspkg "github.com/zack-nova/orbit/cmd/orbit/cli/harness"
)

type removeOutput struct {
	HarnessRoot  string `json:"harness_root"`
	OrbitID      string `json:"orbit_id"`
	ManifestPath string `json:"manifest_path"`
	MemberCount  int    `json:"member_count"`
}

// NewRemoveCommand creates the harness remove command.
func NewRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <orbit-id>",
		Short: "Remove one member from the harness runtime without deleting related files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath, err := pathFromCommand(cmd)
			if err != nil {
				return err
			}

			resolved, err := harnesspkg.ResolveRoot(cmd.Context(), targetPath)
			if err != nil {
				return fmt.Errorf("resolve harness root: %w", err)
			}

			result, err := harnesspkg.RemoveMember(resolved.Repo.Root, args[0], time.Now().UTC())
			if err != nil {
				return fmt.Errorf("remove harness member: %w", err)
			}

			output := removeOutput{
				HarnessRoot:  resolved.Repo.Root,
				OrbitID:      args[0],
				ManifestPath: result.ManifestPath,
				MemberCount:  len(result.Runtime.Members),
			}

			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}
			if jsonOutput {
				return emitJSON(cmd.OutOrStdout(), output)
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "removed orbit %s from harness %s\n", args[0], resolved.Repo.Root); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}

			return nil
		},
	}
	addPathFlag(cmd)
	addJSONFlag(cmd)

	return cmd
}
