package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	harnesspkg "github.com/zack-nova/orbit/cmd/orbit/cli/harness"
	statepkg "github.com/zack-nova/orbit/cmd/orbit/cli/state"
)

type inspectOutput struct {
	HarnessRoot       string   `json:"harness_root"`
	HarnessID         string   `json:"harness_id"`
	HarnessName       string   `json:"harness_name,omitempty"`
	MemberCount       int      `json:"member_count"`
	Members           []string `json:"members"`
	VarsCount         int      `json:"vars_count"`
	InstallCount      int      `json:"install_count"`
	BundleCount       int      `json:"bundle_count"`
	CurrentProjection string   `json:"current_projection,omitempty"`
}

// NewInspectCommand creates the harness inspect command.
func NewInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect the current harness runtime",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			targetPath, err := pathFromCommand(cmd)
			if err != nil {
				return err
			}

			resolved, err := harnesspkg.ResolveRoot(cmd.Context(), targetPath)
			if err != nil {
				return fmt.Errorf("resolve harness root: %w", err)
			}

			varsCount, err := countVars(resolved.Repo.Root)
			if err != nil {
				return err
			}
			installIDs, err := harnesspkg.ListInstallRecordIDs(resolved.Repo.Root)
			if err != nil {
				return fmt.Errorf("list harness install records: %w", err)
			}
			bundleIDs, err := harnesspkg.ListBundleRecordIDs(resolved.Repo.Root)
			if err != nil {
				return fmt.Errorf("list harness bundle records: %w", err)
			}

			members := make([]string, 0, len(resolved.Manifest.Members))
			for _, member := range resolved.Manifest.Members {
				members = append(members, member.OrbitID)
			}

			currentProjection := readCurrentProjection(resolved.Repo.GitDir)
			output := inspectOutput{
				HarnessRoot:       resolved.Repo.Root,
				HarnessID:         resolved.Manifest.Runtime.ID,
				HarnessName:       resolved.Manifest.Runtime.Name,
				MemberCount:       len(members),
				Members:           members,
				VarsCount:         varsCount,
				InstallCount:      len(installIDs),
				BundleCount:       len(bundleIDs),
				CurrentProjection: currentProjection,
			}

			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}
			if jsonOutput {
				return emitJSON(cmd.OutOrStdout(), output)
			}

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "harness_root: %s\n", output.HarnessRoot); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "harness_id: %s\n", output.HarnessID); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			if output.HarnessName != "" {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "harness_name: %s\n", output.HarnessName); err != nil {
					return fmt.Errorf("write command output: %w", err)
				}
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "member_count: %d\n", output.MemberCount); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			if len(output.Members) == 0 {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "members: none"); err != nil {
					return fmt.Errorf("write command output: %w", err)
				}
			} else {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "members: %s\n", joinMembers(output.Members)); err != nil {
					return fmt.Errorf("write command output: %w", err)
				}
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "vars_count: %d\n", output.VarsCount); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "install_count: %d\n", output.InstallCount); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "bundle_count: %d\n", output.BundleCount); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
			if output.CurrentProjection == "" {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "current_projection: none"); err != nil {
					return fmt.Errorf("write command output: %w", err)
				}
			} else {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "current_projection: %s\n", output.CurrentProjection); err != nil {
					return fmt.Errorf("write command output: %w", err)
				}
			}

			return nil
		},
	}
	addPathFlag(cmd)
	addJSONFlag(cmd)

	return cmd
}

func countVars(repoRoot string) (int, error) {
	file, err := harnesspkg.LoadVarsFile(repoRoot)
	if err == nil {
		return len(file.Variables), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}

	return 0, fmt.Errorf("load harness vars file: %w", err)
}

func readCurrentProjection(gitDir string) string {
	store, err := statepkg.NewFSStore(gitDir)
	if err != nil {
		return ""
	}

	current, err := store.ReadCurrentOrbit()
	if err != nil {
		return ""
	}

	return current.Orbit
}

func joinMembers(members []string) string {
	return strings.Join(members, ",")
}
