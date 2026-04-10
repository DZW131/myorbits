package orbit

import (
	"fmt"

	"github.com/zack-nova/orbit/cmd/orbit/cli/internal/contractutil"
)

// OrbitSpec is the member-model superset used by the post-v0.3 runtime design.
// It intentionally coexists with the current Definition shape until the
// compatibility parser is switched over in a later phase.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitSpec struct {
	ID          string        `yaml:"id"`
	Name        string        `yaml:"name,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Include     []string      `yaml:"include,omitempty"`
	Exclude     []string      `yaml:"exclude,omitempty"`
	Meta        *OrbitMeta    `yaml:"meta,omitempty"`
	Members     []OrbitMember `yaml:"members,omitempty"`
	Rules       *OrbitRules   `yaml:"rules,omitempty"`
	SourcePath  string        `yaml:"-"`
}

// OrbitMeta captures the singleton metadata member that is stored by the
// definition file itself.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitMeta struct {
	File                              string `yaml:"file"`
	AgentsTemplate                    string `yaml:"agents_template,omitempty"`
	IncludeInProjection               bool   `yaml:"include_in_projection"`
	IncludeInWrite                    bool   `yaml:"include_in_write"`
	IncludeInExport                   bool   `yaml:"include_in_export"`
	IncludeDescriptionInOrchestration bool   `yaml:"include_description_in_orchestration"`
}

// OrbitMemberRole is the fixed role taxonomy for orbit members.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitMemberRole string

const (
	OrbitMemberMeta    OrbitMemberRole = "meta"
	OrbitMemberSubject OrbitMemberRole = "subject"
	OrbitMemberRule    OrbitMemberRole = "rule"
	OrbitMemberProcess OrbitMemberRole = "process"
)

var orbitMemberRoles = []OrbitMemberRole{
	OrbitMemberMeta,
	OrbitMemberSubject,
	OrbitMemberRule,
	OrbitMemberProcess,
}

// AllOrbitMemberRoles returns the fixed role list in stable order.
func AllOrbitMemberRoles() []OrbitMemberRole {
	return append([]OrbitMemberRole(nil), orbitMemberRoles...)
}

// IsValid reports whether the role is one of the fixed schema-backed values.
func (role OrbitMemberRole) IsValid() bool {
	for _, candidate := range orbitMemberRoles {
		if role == candidate {
			return true
		}
	}

	return false
}

// ParseOrbitMemberRole parses one fixed member role.
func ParseOrbitMemberRole(raw string) (OrbitMemberRole, error) {
	role := OrbitMemberRole(raw)
	if !role.IsValid() {
		return "", fmt.Errorf("invalid orbit member role %q", raw)
	}

	return role, nil
}

// OrbitMember stores one authored file member in the member model.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitMember struct {
	Key         string                 `yaml:"key"`
	Name        string                 `yaml:"name,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Role        OrbitMemberRole        `yaml:"role"`
	Paths       OrbitMemberPaths       `yaml:"paths"`
	Lane        string                 `yaml:"lane,omitempty"`
	Scopes      *OrbitMemberScopePatch `yaml:"scopes,omitempty"`
}

// OrbitMemberPaths captures the authored include/exclude path set for one member.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitMemberPaths struct {
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// OrbitMemberScopePatch carries explicit scope overrides for one member.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitMemberScopePatch struct {
	Write         *bool `yaml:"write,omitempty"`
	Export        *bool `yaml:"export,omitempty"`
	Orchestration *bool `yaml:"orchestration,omitempty"`
}

// OrbitRules captures authored behavior rules for the member model.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitRules struct {
	Scope         OrbitScopeRules         `yaml:"scope,omitempty"`
	Orchestration OrbitOrchestrationRules `yaml:"orchestration,omitempty"`
}

// OrbitScopeRules stores role-to-scope behavior selections.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitScopeRules struct {
	ProjectionRoles    []OrbitMemberRole `yaml:"projection_roles,omitempty"`
	WriteRoles         []OrbitMemberRole `yaml:"write_roles,omitempty"`
	ExportRoles        []OrbitMemberRole `yaml:"export_roles,omitempty"`
	OrchestrationRoles []OrbitMemberRole `yaml:"orchestration_roles,omitempty"`
}

// OrbitOrchestrationRules stores orchestration-specific authored behavior.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
type OrbitOrchestrationRules struct {
	IncludeOrbitDescription   bool `yaml:"include_orbit_description"`
	MaterializeAgentsFromMeta bool `yaml:"materialize_agents_from_meta"`
}

// ProjectionPlan is the role-aware internal plan targeted by later runtime phases.
type ProjectionPlan struct {
	OrbitID            string
	ControlPaths       []string
	MetaPaths          []string
	SubjectPaths       []string
	RulePaths          []string
	ProcessPaths       []string
	ProjectionPaths    []string
	OrbitWritePaths    []string
	ExportPaths        []string
	OrchestrationPaths []string
	PlanHash           string
}

// PathsForRole returns the role-owned path group from the plan.
func (plan ProjectionPlan) PathsForRole(role OrbitMemberRole) []string {
	switch role {
	case OrbitMemberMeta:
		return append([]string(nil), plan.MetaPaths...)
	case OrbitMemberSubject:
		return append([]string(nil), plan.SubjectPaths...)
	case OrbitMemberRule:
		return append([]string(nil), plan.RulePaths...)
	case OrbitMemberProcess:
		return append([]string(nil), plan.ProcessPaths...)
	default:
		return nil
	}
}

// ParseOrbitSpecData decodes one member-model orbit spec without switching the
// current Definition-based command surface to the new schema yet.
func ParseOrbitSpecData(data []byte, sourcePath string) (OrbitSpec, error) {
	return parseOrbitSpecData(data, sourcePath, true)
}

// ParseHostedOrbitSpecData decodes one hosted member-model orbit spec from .harness/orbits/.
func ParseHostedOrbitSpecData(data []byte, sourcePath string) (OrbitSpec, error) {
	return parseOrbitSpecDataWithPathBuilder(data, sourcePath, true, HostedDefinitionRelativePath)
}

func parseOrbitSpecData(data []byte, sourcePath string, validateSourcePath bool) (OrbitSpec, error) {
	return parseOrbitSpecDataWithPathBuilder(data, sourcePath, validateSourcePath, DefinitionRelativePath)
}

func parseOrbitSpecDataWithPathBuilder(
	data []byte,
	sourcePath string,
	validateSourcePath bool,
	pathBuilder func(string) (string, error),
) (OrbitSpec, error) {
	var spec OrbitSpec
	if err := contractutil.DecodeKnownFields(data, &spec); err != nil {
		return OrbitSpec{}, fmt.Errorf("decode orbit spec: %w", err)
	}

	spec.SourcePath = sourcePath
	if err := validateOrbitSpecWithPathBuilder(spec, validateSourcePath, pathBuilder); err != nil {
		return OrbitSpec{}, fmt.Errorf("validate orbit spec: %w", err)
	}

	return spec, nil
}

// HasMemberSchema reports whether the spec uses any member-model fields.
func (spec OrbitSpec) HasMemberSchema() bool {
	return spec.Meta != nil || len(spec.Members) > 0 || spec.Rules != nil
}

// OrbitSpecFromDefinition lifts the current legacy definition shape into the
// member-model container without changing semantics.
//
//nolint:revive // Orbit member runtime docs freeze this exported domain name for cross-package consistency.
func OrbitSpecFromDefinition(definition Definition) OrbitSpec {
	return OrbitSpec{
		ID:          definition.ID,
		Description: definition.Description,
		Include:     append([]string(nil), definition.Include...),
		Exclude:     append([]string(nil), definition.Exclude...),
		SourcePath:  definition.SourcePath,
	}
}

// LegacyDefinition projects the member-model container back to the current
// Definition shape used by the existing command surface.
func (spec OrbitSpec) LegacyDefinition() Definition {
	return Definition{
		ID:          spec.ID,
		Description: spec.Description,
		Include:     append([]string(nil), spec.Include...),
		Exclude:     append([]string(nil), spec.Exclude...),
		SourcePath:  spec.SourcePath,
	}
}

func compatibilityDefinitionFromOrbitSpec(spec OrbitSpec) (Definition, error) {
	return compatibilityDefinitionFromOrbitSpecWithValidation(spec, true)
}

func compatibilityDefinitionFromOrbitSpecWithValidation(spec OrbitSpec, validateSourcePath bool) (Definition, error) {
	if !spec.HasMemberSchema() {
		definition := spec.LegacyDefinition()
		if definition.Exclude == nil {
			definition.Exclude = []string{}
		}
		if err := validateDefinition(definition, validateSourcePath); err != nil {
			return Definition{}, fmt.Errorf("validate compatibility orbit definition: %w", err)
		}

		return definition, nil
	}

	include := make([]string, 0, len(spec.Members))
	exclude := make([]string, 0, len(spec.Members))
	seenInclude := make(map[string]struct{})
	seenExclude := make(map[string]struct{})

	for _, member := range spec.Members {
		for _, pattern := range member.Paths.Include {
			if _, ok := seenInclude[pattern]; ok {
				continue
			}
			seenInclude[pattern] = struct{}{}
			include = append(include, pattern)
		}

		for _, pattern := range member.Paths.Exclude {
			if _, ok := seenExclude[pattern]; ok {
				continue
			}
			seenExclude[pattern] = struct{}{}
			exclude = append(exclude, pattern)
		}
	}

	definition := Definition{
		ID:          spec.ID,
		Description: spec.Description,
		Include:     include,
		Exclude:     exclude,
		SourcePath:  spec.SourcePath,
	}

	if err := validateDefinition(definition, validateSourcePath); err != nil {
		return Definition{}, fmt.Errorf("validate compatibility orbit definition: %w", err)
	}

	return definition, nil
}
