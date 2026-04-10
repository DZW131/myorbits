package orbit

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/zack-nova/orbit/cmd/orbit/cli/ids"
)

// ValidateGlobalConfig validates the documented MVP behavior configuration.
func ValidateGlobalConfig(config GlobalConfig) error {
	if config.Version <= 0 {
		return errors.New("config version must be positive")
	}

	switch config.Behavior.OutsideChangesMode {
	case OutsideChangesModeWarn:
	default:
		return fmt.Errorf("outside_changes_mode must be %q", OutsideChangesModeWarn)
	}

	switch config.Behavior.SparseCheckoutMode {
	case SparseCheckoutModeNoCone:
	default:
		return fmt.Errorf("sparse_checkout_mode must be %q", SparseCheckoutModeNoCone)
	}

	for index, pattern := range config.SharedScope {
		normalizedPattern, err := normalizePattern(pattern)
		if err != nil {
			return fmt.Errorf("shared_scope[%d]: %w", index, err)
		}
		if err := validateScopeOverlayPattern(normalizedPattern); err != nil {
			return fmt.Errorf("shared_scope[%d]: %w", index, err)
		}
	}

	for index, pattern := range config.ProjectionVisible {
		normalizedPattern, err := normalizePattern(pattern)
		if err != nil {
			return fmt.Errorf("projection_visible[%d]: %w", index, err)
		}
		if err := validateScopeOverlayPattern(normalizedPattern); err != nil {
			return fmt.Errorf("projection_visible[%d]: %w", index, err)
		}
	}

	return nil
}

// ValidateOrbitSpec validates either the legacy path-list schema or the new
// member-model schema used by the compatibility parser.
func ValidateOrbitSpec(spec OrbitSpec) error {
	return validateOrbitSpecWithPathBuilder(spec, true, DefinitionRelativePath)
}

// ValidateHostedOrbitSpec validates a hosted OrbitSpec whose steady-state control path lives under .harness/orbits/.
func ValidateHostedOrbitSpec(spec OrbitSpec) error {
	return validateOrbitSpecWithPathBuilder(spec, true, HostedDefinitionRelativePath)
}

func validateOrbitSpecWithPathBuilder(
	spec OrbitSpec,
	validateSourcePath bool,
	pathBuilder func(string) (string, error),
) error {
	if err := ids.ValidateOrbitID(spec.ID); err != nil {
		return fmt.Errorf("validate orbit id: %w", err)
	}
	if validateSourcePath && spec.SourcePath != "" {
		expectedName := spec.ID + ".yaml"
		if filepath.Base(spec.SourcePath) != expectedName {
			return fmt.Errorf("definition filename must match orbit id: expected %q", expectedName)
		}
	}

	if !spec.HasMemberSchema() {
		return validateDefinition(spec.LegacyDefinition(), validateSourcePath)
	}

	if len(spec.Include) > 0 || len(spec.Exclude) > 0 {
		return errors.New("legacy include/exclude cannot be combined with member schema")
	}
	if spec.Meta == nil {
		return errors.New("meta must be present when using member schema")
	}

	expectedMetaFile, err := pathBuilder(spec.ID)
	if err != nil {
		return fmt.Errorf("build expected meta.file: %w", err)
	}
	if spec.Meta.File != expectedMetaFile {
		return fmt.Errorf("meta.file must match %q", expectedMetaFile)
	}

	seenKeys := make(map[string]struct{}, len(spec.Members))
	for index, member := range spec.Members {
		if strings.TrimSpace(member.Key) == "" {
			return fmt.Errorf("members[%d].key must be present", index)
		}
		if _, ok := seenKeys[member.Key]; ok {
			return fmt.Errorf("members[%d].key must be unique", index)
		}
		seenKeys[member.Key] = struct{}{}

		if !member.Role.IsValid() {
			return fmt.Errorf("members[%d].role: invalid orbit member role %q", index, member.Role)
		}
		if len(member.Paths.Include) == 0 {
			return fmt.Errorf("members[%d].paths.include must not be empty", index)
		}

		for patternIndex, pattern := range member.Paths.Include {
			if _, err := normalizePattern(pattern); err != nil {
				return fmt.Errorf("members[%d].paths.include[%d]: %w", index, patternIndex, err)
			}
		}
		for patternIndex, pattern := range member.Paths.Exclude {
			if _, err := normalizePattern(pattern); err != nil {
				return fmt.Errorf("members[%d].paths.exclude[%d]: %w", index, patternIndex, err)
			}
		}
	}

	if spec.Rules != nil {
		if err := validateOrbitSpecRoleList(spec.Rules.Scope.ProjectionRoles, "rules.scope.projection_roles"); err != nil {
			return err
		}
		if err := validateOrbitSpecRoleList(spec.Rules.Scope.WriteRoles, "rules.scope.write_roles"); err != nil {
			return err
		}
		if err := validateOrbitSpecRoleList(spec.Rules.Scope.ExportRoles, "rules.scope.export_roles"); err != nil {
			return err
		}
		if err := validateOrbitSpecRoleList(spec.Rules.Scope.OrchestrationRoles, "rules.scope.orchestration_roles"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDefinition validates an individual orbit definition.
func ValidateDefinition(definition Definition) error {
	return validateDefinition(definition, true)
}

func validateDefinition(definition Definition, validateSourcePath bool) error {
	if err := ids.ValidateOrbitID(definition.ID); err != nil {
		return fmt.Errorf("validate orbit id: %w", err)
	}
	if validateSourcePath && definition.SourcePath != "" {
		expectedName := definition.ID + ".yaml"
		if filepath.Base(definition.SourcePath) != expectedName {
			return fmt.Errorf("definition filename must match orbit id: expected %q", expectedName)
		}
	}
	if len(definition.Include) == 0 {
		return errors.New("include must not be empty")
	}

	for index, pattern := range definition.Include {
		if _, err := normalizePattern(pattern); err != nil {
			return fmt.Errorf("include[%d]: %w", index, err)
		}
	}
	for index, pattern := range definition.Exclude {
		if _, err := normalizePattern(pattern); err != nil {
			return fmt.Errorf("exclude[%d]: %w", index, err)
		}
	}

	return nil
}

func validateOrbitSpecRoleList(roles []OrbitMemberRole, field string) error {
	for index, role := range roles {
		if !role.IsValid() {
			return fmt.Errorf("%s[%d]: invalid orbit member role %q", field, index, role)
		}
	}

	return nil
}

// ValidateRepositoryConfig validates the full repository config set.
func ValidateRepositoryConfig(config GlobalConfig, definitions []Definition) error {
	if err := ValidateGlobalConfig(config); err != nil {
		return fmt.Errorf("validate global config: %w", err)
	}

	seen := make(map[string]string, len(definitions))

	for index, definition := range definitions {
		source := definition.SourcePath
		if source == "" {
			source = fmt.Sprintf("orbit[%d]", index)
		}

		if definition.ID != "" {
			if previousSource, exists := seen[definition.ID]; exists {
				return fmt.Errorf("duplicate orbit id %q in %s and %s", definition.ID, previousSource, source)
			}

			seen[definition.ID] = source
		}

		if err := ValidateDefinition(definition); err != nil {
			label := definition.ID
			if label == "" {
				label = fmt.Sprintf("orbit[%d]", index)
			}
			return fmt.Errorf("%s: %w", label, err)
		}
	}

	return nil
}

func validateScopeOverlayPattern(normalizedPattern string) error {
	matchesConfig, err := matchNormalizedAny([]string{normalizedPattern}, configRelativePath)
	if err != nil {
		return fmt.Errorf("match config path: %w", err)
	}
	if matchesConfig {
		return fmt.Errorf("pattern %q must not match Orbit control-plane paths", normalizedPattern)
	}

	sampleDefinitionPath := path.Join(orbitsRelativeDir, "sample.yaml")

	matchesDefinition, err := matchNormalizedAny([]string{normalizedPattern}, sampleDefinitionPath)
	if err != nil {
		return fmt.Errorf("match orbit definition path: %w", err)
	}
	if matchesDefinition {
		return fmt.Errorf("pattern %q must not match Orbit control-plane paths", normalizedPattern)
	}

	if strings.HasPrefix(normalizedPattern, orbitDirName+"/") {
		return fmt.Errorf("pattern %q must not match Orbit control-plane paths", normalizedPattern)
	}

	if strings.HasPrefix(normalizedPattern, ".harness/") {
		return fmt.Errorf("pattern %q must not match Orbit control-plane paths or runtime-only paths", normalizedPattern)
	}

	if strings.HasPrefix(normalizedPattern, ".git/orbit/state/") {
		return fmt.Errorf("pattern %q must not match Orbit control-plane paths or runtime-only paths", normalizedPattern)
	}

	return nil
}
