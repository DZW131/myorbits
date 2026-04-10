package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

type templateInitSourceResultJSON struct {
	RepoRoot           string `json:"repo_root"`
	SourceManifestPath string `json:"source_manifest_path"`
	SourceBranch       string `json:"source_branch"`
	PublishOrbitID     string `json:"publish_orbit_id"`
	Changed            bool   `json:"changed"`
}

// NewTemplateInitSourceCommand creates the orbit template init-source command.
func NewTemplateInitSourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-source",
		Short: "Initialize the current branch as a single-orbit source branch",
		Long: "Initialize the current branch as a single-orbit source branch for template authoring.\n" +
			"The command writes .harness/manifest.yaml with kind=source, locks source.orbit_id to the single source orbit,\n" +
			"and fails closed for detached HEAD, template branches, or .harness metadata.",
		Example: "" +
			"  orbit template init-source\n" +
			"  orbit template init-source --json\n",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, err := repoFromCommand(cmd)
			if err != nil {
				return err
			}

			result, err := orbittemplate.InitSourceBranch(cmd.Context(), repo.Root)
			if err != nil {
				return fmt.Errorf("initialize source branch: %w", err)
			}

			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}
			if jsonOutput {
				return emitJSON(cmd.OutOrStdout(), templateInitSourceResultJSON{
					RepoRoot:           result.RepoRoot,
					SourceManifestPath: result.SourceManifestPath,
					SourceBranch:       result.SourceBranch,
					PublishOrbitID:     result.PublishOrbitID,
					Changed:            result.Changed,
				})
			}

			if _, err := fmt.Fprintf(
				cmd.OutOrStdout(),
				"initialized template source in %s\nsource_branch: %s\npublish_orbit_id: %s\nchanged: %t\n",
				result.RepoRoot,
				result.SourceBranch,
				result.PublishOrbitID,
				result.Changed,
			); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}

			return nil
		},
	}
	addJSONFlag(cmd)

	return cmd
}
