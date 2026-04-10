package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zack-nova/orbit/cmd/orbit/cli/branchinfo"
	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
)

type branchStatusOutput struct {
	Branch         string                    `json:"branch"`
	Classification branchinfo.Classification `json:"classification"`
}

// NewBranchStatusCommand creates the orbit branch status command.
func NewBranchStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Classify the current branch as template, runtime, or plain",
		Long: "Classify the current branch using Orbit's file-contract rules rather than branch naming.\n" +
			"The result explains whether the current revision looks like a template branch, a runtime branch, or plain Git history.",
		Example: "" +
			"  orbit branch status\n" +
			"  orbit branch status --json\n",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, err := repoFromCommand(cmd)
			if err != nil {
				return err
			}

			currentBranch, err := gitpkg.CurrentBranch(cmd.Context(), repo.Root)
			if err != nil {
				return fmt.Errorf("resolve current branch: %w", err)
			}

			classification, err := branchinfo.ClassifyRevision(cmd.Context(), repo.Root, currentBranch)
			if err != nil {
				return fmt.Errorf("classify current branch %q: %w", currentBranch, err)
			}

			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}
			if jsonOutput {
				return emitJSON(cmd.OutOrStdout(), branchStatusOutput{
					Branch:         currentBranch,
					Classification: classification,
				})
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "kind: %s\nreason: %s\n", classification.Kind, classification.Reason); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}

			return nil
		},
	}

	addJSONFlag(cmd)

	return cmd
}
