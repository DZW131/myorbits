package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zack-nova/orbit/cmd/orbit/cli/commands"
)

// NewRootCommand builds the top-level orbit command tree.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "orbit",
		Short: "Git-native CLI for file-scoped workspace views",
		Long: "Git-native CLI for orbit definition, projection, scoped operations, and single-orbit template authoring.\n" +
			"Use orbit when you are working with orbit definitions, current projection state, or orbit template branches.",
		Example: "" +
			"  orbit add docs\n" +
			"  orbit enter docs\n" +
			"  orbit template save docs --to orbit-template/docs\n" +
			"  orbit bindings init orbit-template/docs --out .harness/vars.yaml\n",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		commands.NewInitCommand(),
		commands.NewAddCommand(),
		commands.NewValidateCommand(),
		commands.NewListCommand(),
		commands.NewShowCommand(),
		commands.NewFilesCommand(),
		commands.NewCurrentCommand(),
		commands.NewEnterCommand(),
		commands.NewLeaveCommand(),
		commands.NewStatusCommand(),
		commands.NewDiffCommand(),
		commands.NewLogCommand(),
		commands.NewCommitCommand(),
		commands.NewRestoreCommand(),
		commands.NewBranchCommand(),
		commands.NewBriefCommand(),
		commands.NewBindingsCommand(),
		commands.NewTemplateCommand(),
	)

	return rootCmd
}

// Execute runs the root command with the provided context.
func Execute(ctx context.Context) error {
	if err := NewRootCommand().ExecuteContext(ctx); err != nil {
		return fmt.Errorf("execute root command: %w", err)
	}

	return nil
}
