package commands

import "github.com/spf13/cobra"

// NewTemplateCommand creates the harness template command tree.
func NewTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Harness template export commands",
	}

	cmd.AddCommand(NewTemplateSaveCommand())

	return cmd
}
