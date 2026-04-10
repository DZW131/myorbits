package bindings

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	"github.com/zack-nova/orbit/cmd/orbit/cli/internal/contractutil"
)

const (
	varsRelativePath   = ".orbit/vars.yaml"
	varsSchemaVersion  = 1
	variablesFieldName = "variables"
)

// VarsFile is the schema-backed versioned bindings document.
type VarsFile struct {
	SchemaVersion int                        `yaml:"schema_version"`
	Variables     map[string]VariableBinding `yaml:"variables"`
}

// VariableBinding stores a concrete string value and optional description.
type VariableBinding struct {
	Value       string `yaml:"value"`
	Description string `yaml:"description,omitempty"`
}

type rawVarsFile struct {
	SchemaVersion *int                          `yaml:"schema_version"`
	Variables     map[string]rawVariableBinding `yaml:"variables"`
}

type rawVariableBinding struct {
	Value       *string `yaml:"value"`
	Description *string `yaml:"description"`
}

// VarsPath returns the absolute path to .orbit/vars.yaml.
func VarsPath(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(varsRelativePath))
}

// LoadVarsFile reads, decodes, and validates the bindings file at the fixed Phase 2 host path.
func LoadVarsFile(repoRoot string) (VarsFile, error) {
	return LoadVarsFileAtPath(VarsPath(repoRoot))
}

// LoadVarsFileAtPath reads, decodes, and validates one bindings document from an absolute path.
func LoadVarsFileAtPath(filename string) (VarsFile, error) {
	//nolint:gosec // The path is repo-local and built from the fixed bindings contract path.
	data, err := os.ReadFile(filename)
	if err != nil {
		return VarsFile{}, fmt.Errorf("read %s: %w", filename, err)
	}

	file, err := ParseVarsData(data)
	if err != nil {
		return VarsFile{}, fmt.Errorf("validate %s: %w", filename, err)
	}

	return file, nil
}

// LoadVarsFileWorktreeOrHEAD reads the bindings file from the fixed Phase 2 host path.
func LoadVarsFileWorktreeOrHEAD(ctx context.Context, repoRoot string) (VarsFile, error) {
	return LoadVarsFileWorktreeOrHEADAtRepoPath(ctx, repoRoot, varsRelativePath)
}

// LoadVarsFileWorktreeOrHEADAtRepoPath reads a bindings file from the worktree when visible
// and falls back to HEAD when sparse-checkout currently hides it.
func LoadVarsFileWorktreeOrHEADAtRepoPath(ctx context.Context, repoRoot string, repoPath string) (VarsFile, error) {
	filename := filepath.Join(repoRoot, filepath.FromSlash(repoPath))
	data, err := gitpkg.ReadFileWorktreeOrHEAD(ctx, repoRoot, repoPath)
	if err != nil {
		return VarsFile{}, fmt.Errorf("read %s: %w", filename, err)
	}

	file, err := ParseVarsData(data)
	if err != nil {
		return VarsFile{}, fmt.Errorf("validate %s: %w", filename, err)
	}

	return file, nil
}

// ParseVarsData decodes and validates .orbit/vars.yaml bytes.
func ParseVarsData(data []byte) (VarsFile, error) {
	var raw rawVarsFile
	if err := contractutil.DecodeKnownFields(data, &raw); err != nil {
		return VarsFile{}, fmt.Errorf("decode vars file: %w", err)
	}

	file, err := raw.toVarsFile()
	if err != nil {
		return VarsFile{}, err
	}

	return file, nil
}

// ValidateVarsFile validates the bindings schema contract.
func ValidateVarsFile(file VarsFile) error {
	if file.SchemaVersion != varsSchemaVersion {
		return fmt.Errorf("schema_version must be %d", varsSchemaVersion)
	}
	if file.Variables == nil {
		return fmt.Errorf("%s must be present", variablesFieldName)
	}

	for _, name := range contractutil.SortedKeys(file.Variables) {
		if err := contractutil.ValidateVariableName(name); err != nil {
			return fmt.Errorf("variables.%s: %w", name, err)
		}
	}

	return nil
}

// WriteVarsFile validates and writes the bindings file at the fixed Phase 2 host path.
func WriteVarsFile(repoRoot string, file VarsFile) (string, error) {
	return WriteVarsFileAtPath(VarsPath(repoRoot), file)
}

// WriteVarsFileAtPath validates and writes one bindings document with stable field ordering.
func WriteVarsFileAtPath(filename string, file VarsFile) (string, error) {
	data, err := MarshalVarsFile(file)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", filename, err)
	}

	if err := contractutil.AtomicWriteFile(filename, data); err != nil {
		return "", fmt.Errorf("write %s: %w", filename, err)
	}

	return filename, nil
}

// MarshalVarsFile validates and encodes a bindings document with stable field ordering.
func MarshalVarsFile(file VarsFile) ([]byte, error) {
	if err := ValidateVarsFile(file); err != nil {
		return nil, fmt.Errorf("validate vars file: %w", err)
	}

	data, err := contractutil.EncodeYAMLDocument(varsFileNode(file))
	if err != nil {
		return nil, fmt.Errorf("encode vars file: %w", err)
	}

	return data, nil
}

func (raw rawVarsFile) toVarsFile() (VarsFile, error) {
	if raw.SchemaVersion == nil {
		return VarsFile{}, fmt.Errorf("schema_version must be present")
	}
	if raw.Variables == nil {
		return VarsFile{}, fmt.Errorf("%s must be present", variablesFieldName)
	}

	file := VarsFile{
		SchemaVersion: *raw.SchemaVersion,
		Variables:     make(map[string]VariableBinding, len(raw.Variables)),
	}

	for name, rawBinding := range raw.Variables {
		if rawBinding.Value == nil {
			return VarsFile{}, fmt.Errorf("variables.%s.value must be present", name)
		}

		binding := VariableBinding{
			Value: *rawBinding.Value,
		}
		if rawBinding.Description != nil {
			binding.Description = *rawBinding.Description
		}

		file.Variables[name] = binding
	}

	if err := ValidateVarsFile(file); err != nil {
		return VarsFile{}, err
	}

	return file, nil
}

func varsFileNode(file VarsFile) *yaml.Node {
	root := contractutil.MappingNode()
	contractutil.AppendMapping(root, "schema_version", contractutil.IntNode(file.SchemaVersion))

	variables := contractutil.MappingNode()
	for _, name := range contractutil.SortedKeys(file.Variables) {
		binding := file.Variables[name]
		bindingNode := contractutil.MappingNode()
		contractutil.AppendMapping(bindingNode, "value", contractutil.StringNode(binding.Value))
		if binding.Description != "" {
			contractutil.AppendMapping(bindingNode, "description", contractutil.StringNode(binding.Description))
		}
		contractutil.AppendMapping(variables, name, bindingNode)
	}

	contractutil.AppendMapping(root, variablesFieldName, variables)

	return root
}
