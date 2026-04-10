package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

type templatePublishLocalJSON struct {
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Commit  string `json:"commit,omitempty"`
}

type templatePublishRemoteJSON struct {
	Attempted bool   `json:"attempted"`
	Success   bool   `json:"success"`
	Remote    string `json:"remote,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type templatePublishResultJSON struct {
	OrbitID         string                    `json:"orbit_id"`
	PublishRef      string                    `json:"publish_ref"`
	Branch          string                    `json:"branch"`
	SourceBranch    string                    `json:"source_branch"`
	DefaultTemplate bool                      `json:"default_template"`
	LocalPublish    templatePublishLocalJSON  `json:"local_publish"`
	RemotePush      templatePublishRemoteJSON `json:"remote_push"`
}

// NewTemplatePublishCommand creates the orbit template publish command.
func NewTemplatePublishCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish one orbit from the current source or template revision",
		Long: "Publish one orbit from the current source revision or validate and publish the current template revision.\n" +
			"This command is author-facing, requires a clean tracked worktree, and does not push by default.",
		Example: "" +
			"  orbit template publish\n" +
			"  orbit template publish --default\n" +
			"  orbit template publish --push --remote origin\n" +
			"  orbit template publish --json\n",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo, err := repoFromCommand(cmd)
			if err != nil {
				return err
			}

			orbitID, err := cmd.Flags().GetString("orbit")
			if err != nil {
				return fmt.Errorf("read --orbit flag: %w", err)
			}
			defaultTemplate, err := cmd.Flags().GetBool("default")
			if err != nil {
				return fmt.Errorf("read --default flag: %w", err)
			}
			pushEnabled, err := cmd.Flags().GetBool("push")
			if err != nil {
				return fmt.Errorf("read --push flag: %w", err)
			}
			remoteName, err := cmd.Flags().GetString("remote")
			if err != nil {
				return fmt.Errorf("read --remote flag: %w", err)
			}
			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}
			if remoteName != "" && !pushEnabled {
				return fmt.Errorf("--remote can only be used with --push")
			}
			defaultTemplateSet := cmd.Flags().Changed("default")

			result, err := orbittemplate.PublishTemplate(cmd.Context(), orbittemplate.TemplatePublishInput{
				RepoRoot:           repo.Root,
				OrbitID:            orbitID,
				DefaultTemplate:    defaultTemplate,
				DefaultTemplateSet: defaultTemplateSet,
				Push:               pushEnabled,
				Remote:             remoteName,
			})
			if err != nil {
				var publishErr *orbittemplate.PublishError
				if errors.As(err, &publishErr) {
					if emitErr := emitTemplatePublishResult(cmd, publishErr.Result, jsonOutput); emitErr != nil {
						return emitErr
					}
					return fmt.Errorf("publish template branch: %w", publishErr.Err)
				}
				return fmt.Errorf("publish template branch: %w", err)
			}

			return emitTemplatePublishResult(cmd, result, jsonOutput)
		},
	}

	cmd.Flags().String("orbit", "", "Explicit orbit id to verify against the single source orbit")
	cmd.Flags().Bool("default", false, "Mark the published template branch as the default template")
	cmd.Flags().Bool("push", false, "Push the published template branch after local publish succeeds")
	cmd.Flags().String("remote", "", "Remote name to use with --push; defaults to origin")
	addJSONFlag(cmd)

	return cmd
}

func templatePublishResultPayload(result orbittemplate.TemplatePublishResult) templatePublishResultJSON {
	payload := templatePublishResultJSON{
		OrbitID:         result.Preview.OrbitID,
		PublishRef:      "refs/heads/" + result.Preview.PublishBranch,
		Branch:          result.Preview.PublishBranch,
		SourceBranch:    result.Preview.SourceBranch,
		DefaultTemplate: result.Preview.DefaultTemplate,
		LocalPublish: templatePublishLocalJSON{
			Success: true,
			Changed: result.Changed,
		},
		RemotePush: templatePublishRemoteJSON{
			Attempted: result.RemotePush.Attempted,
			Success:   result.RemotePush.Success,
			Remote:    result.RemotePush.Remote,
			Reason:    result.RemotePush.Reason,
		},
	}
	if result.Changed {
		payload.LocalPublish.Commit = result.Commit
	}

	return payload
}

func emitTemplatePublishResult(cmd *cobra.Command, result orbittemplate.TemplatePublishResult, jsonOutput bool) error {
	if jsonOutput {
		return emitJSON(cmd.OutOrStdout(), templatePublishResultPayload(result))
	}

	if _, err := fmt.Fprintf(
		cmd.OutOrStdout(),
		"orbit_id: %s\npublish_ref: refs/heads/%s\nsource_branch: %s\ndefault_template: %t\nlocal_publish.success: true\nlocal_publish.changed: %t\n",
		result.Preview.OrbitID,
		result.Preview.PublishBranch,
		result.Preview.SourceBranch,
		result.Preview.DefaultTemplate,
		result.Changed,
	); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if result.Changed {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "local_publish.commit: %s\n", result.Commit); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if _, err := fmt.Fprintf(
		cmd.OutOrStdout(),
		"remote_push.attempted: %t\nremote_push.success: %t\n",
		result.RemotePush.Attempted,
		result.RemotePush.Success,
	); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if result.RemotePush.Remote != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "remote_push.remote: %s\n", result.RemotePush.Remote); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if result.RemotePush.Reason != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "remote_push.reason: %s\n", result.RemotePush.Reason); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}

	return nil
}
