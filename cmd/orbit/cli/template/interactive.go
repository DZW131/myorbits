package orbittemplate

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
)

// BindingPrompter fills missing required bindings during interactive apply flows.
type BindingPrompter interface {
	PromptBindings(ctx context.Context, unresolved []bindings.UnresolvedBinding) (map[string]bindings.VariableBinding, error)
}

// LineBindingPrompter prompts for one binding per line via stdin/stderr style streams.
type LineBindingPrompter struct {
	Reader io.Reader
	Writer io.Writer
}

// PromptBindings prompts for each unresolved binding and returns the filled values.
func (prompter LineBindingPrompter) PromptBindings(ctx context.Context, unresolved []bindings.UnresolvedBinding) (map[string]bindings.VariableBinding, error) {
	if prompter.Reader == nil {
		return nil, fmt.Errorf("interactive input reader must be configured")
	}
	if prompter.Writer == nil {
		return nil, fmt.Errorf("interactive prompt writer must be configured")
	}

	reader := bufio.NewReader(prompter.Reader)
	values := make(map[string]bindings.VariableBinding, len(unresolved))
	for _, missing := range unresolved {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("prompt context canceled: %w", err)
		}

		if _, err := fmt.Fprint(prompter.Writer, promptForBinding(missing)); err != nil {
			return nil, fmt.Errorf("write prompt for %s: %w", missing.Name, err)
		}

		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read interactive value for %s: %w", missing.Name, err)
		}

		value := strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("interactive value for %s must not be empty", missing.Name)
		}

		values[missing.Name] = bindings.VariableBinding{
			Value:       value,
			Description: missing.Description,
		}
	}

	return values, nil
}

func promptForBinding(binding bindings.UnresolvedBinding) string {
	if binding.Description != "" {
		return fmt.Sprintf("%s (%s): ", binding.Name, binding.Description)
	}

	return fmt.Sprintf("%s: ", binding.Name)
}
