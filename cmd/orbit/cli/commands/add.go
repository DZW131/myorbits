package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
)

type addOutput struct {
	RepoRoot   string `json:"repo_root"`
	Definition any    `json:"definition,omitempty"`
	Orbit      any    `json:"orbit,omitempty"`
	Schema     string `json:"schema,omitempty"`
	File       string `json:"file"`
}

// NewAddCommand creates the orbit add command.
func NewAddCommand() *cobra.Command {
	var memberSchema bool

	cmd := &cobra.Command{
		Use:   "add <orbit-id>",
		Short: "Create an orbit definition skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := repoFromCommand(cmd)
			if err != nil {
				return err
			}

			filename, err := orbitpkg.HostedDefinitionPath(repo.Root, args[0])
			if err != nil {
				return fmt.Errorf("build orbit definition path: %w", err)
			}
			if _, err := os.Stat(filename); err == nil {
				return fmt.Errorf("orbit definition file %q already exists", filename)
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("stat orbit definition: %w", err)
			}

			config, err := loadValidatedAuthoringRepositoryConfig(cmd.Context(), repo.Root)
			if err != nil {
				return err
			}

			orbitID := args[0]
			if _, found := config.OrbitByID(orbitID); found {
				return fmt.Errorf("orbit %q already exists", orbitID)
			}

			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}

			if memberSchema {
				spec, err := orbitpkg.DefaultHostedMemberSchemaSpec(orbitID)
				if err != nil {
					return fmt.Errorf("build orbit skeleton: %w", err)
				}

				filename, err = orbitpkg.WriteHostedOrbitSpec(repo.Root, spec)
				if err != nil {
					return fmt.Errorf("write orbit definition: %w", err)
				}

				if jsonOutput {
					return emitJSON(cmd.OutOrStdout(), addOutput{
						RepoRoot: repo.Root,
						Orbit:    spec,
						Schema:   "members",
						File:     filename,
					})
				}
			} else {
				definition, err := orbitpkg.DefaultDefinition(orbitID)
				if err != nil {
					return fmt.Errorf("build orbit skeleton: %w", err)
				}

				filename, err = orbitpkg.WriteHostedDefinition(repo.Root, definition)
				if err != nil {
					return fmt.Errorf("write orbit definition: %w", err)
				}

				if jsonOutput {
					return emitJSON(cmd.OutOrStdout(), addOutput{
						RepoRoot:   repo.Root,
						Definition: definition,
						File:       filename,
					})
				}
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "created orbit %s at %s\n", orbitID, filename)
			if err != nil {
				return fmt.Errorf("write command output: %w", err)
			}

			return nil
		},
	}
	addJSONFlag(cmd)
	cmd.Flags().BoolVar(&memberSchema, "member-schema", false, "Create a member-schema orbit skeleton")

	return cmd
}
