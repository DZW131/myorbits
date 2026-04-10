package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zack-nova/orbit/cmd/harness/cli/commands"
)

// NewRootCommand builds the top-level harness command tree.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "harness",
		Short: "Git-native CLI for harness runtime engineering",
		Long: "Git-native CLI for harness runtime bootstrap, install flows, member management, diagnostics, and harness template export.\n" +
			"Use harness when you are operating on the runtime as a whole rather than on one projected orbit view.",
		Example: "" +
			"  harness create demo-repo\n" +
			"  harness bindings plan orbit-template/docs orbit-template/cmd\n" +
			"  harness init\n" +
			"  harness install orbit-template/docs --bindings .harness/vars.yaml\n" +
			"  harness check\n" +
			"  harness template save --to harness-template/workspace\n",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(
		commands.NewAddCommand(),
		commands.NewBindingsCommand(),
		commands.NewCheckCommand(),
		commands.NewCreateCommand(),
		commands.NewInitCommand(),
		commands.NewInspectCommand(),
		commands.NewInstallCommand(),
		commands.NewRemoveCommand(),
		commands.NewRootPathCommand(),
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
