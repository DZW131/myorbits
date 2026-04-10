package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	statepkg "github.com/zack-nova/orbit/cmd/orbit/cli/state"
	viewpkg "github.com/zack-nova/orbit/cmd/orbit/cli/view"
)

// NewEnterCommand creates the orbit enter command.
func NewEnterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enter <orbit-id>",
		Short: "Project an orbit into the current workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}

			repo, err := repoFromCommand(cmd)
			if err != nil {
				return err
			}

			config, err := loadValidatedRepositoryConfig(cmd.Context(), repo.Root)
			if err != nil {
				return err
			}

			definition, err := definitionByID(config, args[0])
			if err != nil {
				return err
			}

			store, err := statepkg.NewFSStore(repo.GitDir)
			if err != nil {
				return fmt.Errorf("create state store: %w", err)
			}

			result, err := viewpkg.Enter(cmd.Context(), repo, store, config, definition)
			if err != nil {
				if !jsonOutput {
					if warningErr := printWarnings(cmd, result.WarningMessages); warningErr != nil {
						return warningErr
					}
				}
				return fmt.Errorf("enter orbit: %w", err)
			}

			if jsonOutput {
				return emitJSON(cmd.OutOrStdout(), result)
			}

			if err := printWarnings(cmd, result.WarningMessages); err != nil {
				return err
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "entered orbit %s (%d file(s))\n", result.Orbit, result.ScopeCount); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}

			return nil
		},
	}
	addJSONFlag(cmd)

	return cmd
}

func printWarnings(cmd *cobra.Command, warnings []string) error {
	for _, warningMessage := range warnings {
		if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warningMessage); err != nil {
			return fmt.Errorf("write warning output: %w", err)
		}
	}

	return nil
}
