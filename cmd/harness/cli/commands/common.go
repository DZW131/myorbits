package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type contextKey string

const workingDirContextKey contextKey = "working_dir"

// WithWorkingDir injects the working directory used by command tests.
func WithWorkingDir(ctx context.Context, workingDir string) context.Context {
	return context.WithValue(ctx, workingDirContextKey, workingDir)
}

func addJSONFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output machine-readable JSON")
}

func addPathFlag(cmd *cobra.Command) {
	cmd.Flags().String("path", "", "Resolve the harness root starting from this path")
}

func wantJSON(cmd *cobra.Command) (bool, error) {
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return false, fmt.Errorf("read json flag: %w", err)
	}

	return jsonOutput, nil
}

func emitJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode json output: %w", err)
	}

	return nil
}

func workingDirFromCommand(cmd *cobra.Command) (string, error) {
	if cmd.Context() != nil {
		if workingDir, ok := cmd.Context().Value(workingDirContextKey).(string); ok && strings.TrimSpace(workingDir) != "" {
			return workingDir, nil
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return workingDir, nil
}

func pathFromCommand(cmd *cobra.Command) (string, error) {
	workingDir, err := workingDirFromCommand(cmd)
	if err != nil {
		return "", err
	}

	pathValue, err := cmd.Flags().GetString("path")
	if err != nil {
		return "", fmt.Errorf("read path flag: %w", err)
	}
	if strings.TrimSpace(pathValue) == "" {
		return filepath.Clean(workingDir), nil
	}
	if filepath.IsAbs(pathValue) {
		return filepath.Clean(pathValue), nil
	}

	return filepath.Clean(filepath.Join(workingDir, pathValue)), nil
}

func absolutePathFromArg(cmd *cobra.Command, value string) (string, error) {
	workingDir, err := workingDirFromCommand(cmd)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}

	return filepath.Clean(filepath.Join(workingDir, value)), nil
}
