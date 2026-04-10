package harness

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

// BindingsPlanSource describes one template source that contributed to the merged bindings plan.
type BindingsPlanSource struct {
	Kind    string
	Repo    string
	Ref     string
	Commit  string
	OrbitID string
}

// BindingsPlanResult captures the merged shared bindings skeleton for multiple template sources.
type BindingsPlanResult struct {
	Sources         []BindingsPlanSource
	Bindings        bindings.VarsFile
	MissingRequired []string
	ReusedValues    []string
}

// BindingsPlanVariableConflictError reports one conflicting variable declaration across multiple template sources.
type BindingsPlanVariableConflictError struct {
	Name    string
	Sources []string
}

func (err *BindingsPlanVariableConflictError) Error() string {
	return fmt.Sprintf("variable conflict for %q (sources: %s)", err.Name, strings.Join(err.Sources, ", "))
}

type mergedBindingsPlanVariable struct {
	Description string
	Required    bool
}

// BuildBindingsPlan merges multiple bindings-init previews into one shared runtime bindings skeleton.
func BuildBindingsPlan(
	previews []orbittemplate.BindingsInitPreview,
	repoVars bindings.VarsFile,
) (BindingsPlanResult, error) {
	result := BindingsPlanResult{
		Sources: make([]BindingsPlanSource, 0, len(previews)),
		Bindings: bindings.VarsFile{
			SchemaVersion: 1,
			Variables:     map[string]bindings.VariableBinding{},
		},
	}

	if repoVars.SchemaVersion != 0 {
		result.Bindings.SchemaVersion = repoVars.SchemaVersion
	}

	mergedVariables := make(map[string]mergedBindingsPlanVariable)
	contributors := make(map[string][]string)

	for _, preview := range previews {
		result.Sources = append(result.Sources, BindingsPlanSource{
			Kind:    preview.Source.SourceKind,
			Repo:    preview.Source.SourceRepo,
			Ref:     preview.Source.SourceRef,
			Commit:  preview.Source.TemplateCommit,
			OrbitID: preview.Manifest.Template.OrbitID,
		})

		for _, name := range sortedVariableNames(preview.Manifest.Variables) {
			next := preview.Manifest.Variables[name]
			current, ok := mergedVariables[name]
			if !ok {
				mergedVariables[name] = mergedBindingsPlanVariable{
					Description: next.Description,
					Required:    next.Required,
				}
				contributors[name] = appendContributor(contributors[name], preview.Source.SourceRef)
				continue
			}

			merged, err := mergeBindingsPlanVariable(name, current, next)
			if err != nil {
				return BindingsPlanResult{}, &BindingsPlanVariableConflictError{
					Name:    name,
					Sources: appendContributor(contributors[name], preview.Source.SourceRef),
				}
			}
			mergedVariables[name] = merged
			contributors[name] = appendContributor(contributors[name], preview.Source.SourceRef)
		}
	}

	for _, name := range sortedMergedBindingsPlanNames(mergedVariables) {
		spec := mergedVariables[name]
		binding := bindings.VariableBinding{
			Value:       "",
			Description: spec.Description,
		}

		if existing, ok := repoVars.Variables[name]; ok && strings.TrimSpace(existing.Value) != "" {
			binding.Value = existing.Value
			result.ReusedValues = append(result.ReusedValues, name)
		}
		if binding.Value == "" && spec.Required {
			result.MissingRequired = append(result.MissingRequired, name)
		}

		result.Bindings.Variables[name] = binding
	}

	return result, nil
}

func mergeBindingsPlanVariable(
	name string,
	current mergedBindingsPlanVariable,
	next orbittemplate.VariableSpec,
) (mergedBindingsPlanVariable, error) {
	switch {
	case current.Description == next.Description:
	case current.Description == "":
		current.Description = next.Description
	case next.Description == "":
	default:
		return mergedBindingsPlanVariable{}, fmt.Errorf("variable conflict for %q", name)
	}

	current.Required = current.Required || next.Required

	return current, nil
}

func sortedVariableNames(values map[string]orbittemplate.VariableSpec) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedMergedBindingsPlanNames(values map[string]mergedBindingsPlanVariable) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
