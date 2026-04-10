package orbittemplate

import (
	"fmt"
	"path/filepath"
	"strings"

	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	"gopkg.in/yaml.v3"
)

func stableOrbitSpecData(spec orbitpkg.OrbitSpec) ([]byte, error) {
	stable := spec
	stable.SourcePath = ""

	data, err := yaml.Marshal(stable)
	if err != nil {
		return nil, fmt.Errorf("marshal orbit spec: %w", err)
	}

	return append(data, '\n'), nil
}

func rewriteRuntimeCompanionDataForTemplate(orbitID string, runtimeCompanionPath string, data []byte) ([]byte, error) {
	hostedPath, err := orbitpkg.HostedDefinitionRelativePath(orbitID)
	if err != nil {
		return nil, fmt.Errorf("build hosted companion path: %w", err)
	}
	if runtimeCompanionPath != hostedPath {
		return data, nil
	}

	spec, err := orbitpkg.ParseHostedOrbitSpecData(data, filepath.Join("/repo", filepath.FromSlash(hostedPath)))
	if err != nil {
		return nil, fmt.Errorf("parse hosted runtime companion: %w", err)
	}
	return stableOrbitSpecData(spec)
}

func plannedRuntimeCompanionWrite(source LocalTemplateSource) ([]byte, error) {
	repoPath, err := orbitpkg.HostedDefinitionRelativePath(source.Manifest.Template.OrbitID)
	if err != nil {
		return nil, fmt.Errorf("build runtime companion path: %w", err)
	}

	if !source.Spec.HasMemberSchema() {
		data, err := stableDefinitionData(source.Definition)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	runtimeSpec := source.Spec
	runtimeSpec.SourcePath = ""
	if runtimeSpec.Meta != nil {
		runtimeSpec.Meta.File = repoPath
	}

	data, err := stableOrbitSpecData(runtimeSpec)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func orbitAgentsBody(spec orbitpkg.OrbitSpec) []byte {
	if spec.Meta != nil && strings.TrimSpace(spec.Meta.AgentsTemplate) != "" {
		return ensureTrailingNewline([]byte(spec.Meta.AgentsTemplate))
	}

	includeDescription := spec.Meta != nil && spec.Meta.IncludeDescriptionInOrchestration
	if spec.Rules != nil && spec.Rules.Orchestration.IncludeOrbitDescription {
		includeDescription = true
	}
	if includeDescription && strings.TrimSpace(spec.Description) != "" {
		return ensureTrailingNewline([]byte(spec.Description))
	}

	return nil
}

func companionAgentsCandidate(files []CandidateFile, orbitID string) (CandidateFile, bool, error) {
	hostedPath, legacyPath, err := templateCompanionPaths(orbitID)
	if err != nil {
		return CandidateFile{}, false, fmt.Errorf("build companion path: %w", err)
	}

	var companion *CandidateFile
	for _, file := range files {
		switch file.Path {
		case hostedPath:
			fileCopy := file
			companion = &fileCopy
		case legacyPath:
			if companion == nil {
				fileCopy := file
				companion = &fileCopy
			}
		}
	}
	if companion == nil {
		return CandidateFile{}, false, nil
	}

	var spec orbitpkg.OrbitSpec
	switch companion.Path {
	case hostedPath:
		spec, err = orbitpkg.ParseHostedOrbitSpecData(companion.Content, companion.Path)
	case legacyPath:
		spec, err = orbitpkg.ParseOrbitSpecData(companion.Content, companion.Path)
	default:
		return CandidateFile{}, false, fmt.Errorf("parse companion definition %s: unexpected path", companion.Path)
	}
	if err != nil {
		return CandidateFile{}, false, fmt.Errorf("parse companion definition %s: %w", companion.Path, err)
	}

	body := orbitAgentsBody(spec)
	if len(body) == 0 {
		return CandidateFile{}, false, nil
	}

	return CandidateFile{
		Path:    sharedFilePathAgents,
		Content: body,
		Mode:    companion.Mode,
	}, true, nil
}

func ensureTrailingNewline(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return append([]byte(nil), data...)
	}

	return append(append([]byte(nil), data...), '\n')
}
