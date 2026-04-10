package harness

import (
	"fmt"
	"os"
	"sort"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/zack-nova/orbit/cmd/orbit/cli/ids"
	"github.com/zack-nova/orbit/cmd/orbit/cli/internal/contractutil"
)

const (
	manifestSchemaVersion = 1

	ManifestKindSource          = "source"
	ManifestKindRuntime         = "runtime"
	ManifestKindOrbitTemplate   = "orbit_template"
	ManifestKindHarnessTemplate = "harness_template"

	ManifestMemberSourceManual        = "manual"
	ManifestMemberSourceInstallOrbit  = "install_orbit"
	ManifestMemberSourceInstallBundle = "install_bundle"
)

// ManifestFile is the schema-backed single-control-plane manifest stored in .harness/manifest.yaml.
type ManifestFile struct {
	SchemaVersion      int                             `yaml:"schema_version"`
	Kind               string                          `yaml:"kind"`
	Source             *ManifestSourceMetadata         `yaml:"source,omitempty"`
	Runtime            *ManifestRuntimeMetadata        `yaml:"runtime,omitempty"`
	Template           *ManifestTemplateMetadata       `yaml:"template,omitempty"`
	Members            []ManifestMember                `yaml:"members,omitempty"`
	Variables          map[string]ManifestVariableSpec `yaml:"variables,omitempty"`
	IncludesRootAgents bool                            `yaml:"includes_root_agents,omitempty"`
}

// ManifestSourceMetadata stores source-branch authoring identity.
type ManifestSourceMetadata struct {
	OrbitID      string `yaml:"orbit_id"`
	SourceBranch string `yaml:"source_branch"`
}

// ManifestRuntimeMetadata stores runtime identity and timestamps.
type ManifestRuntimeMetadata struct {
	ID        string    `yaml:"id"`
	Name      string    `yaml:"name,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

// ManifestTemplateMetadata stores template provenance for either orbit or harness templates.
type ManifestTemplateMetadata struct {
	OrbitID           string    `yaml:"orbit_id,omitempty"`
	HarnessID         string    `yaml:"harness_id,omitempty"`
	DefaultTemplate   bool      `yaml:"default_template,omitempty"`
	CreatedFromBranch string    `yaml:"created_from_branch"`
	CreatedFromCommit string    `yaml:"created_from_commit"`
	CreatedAt         time.Time `yaml:"created_at"`
}

// ManifestMember stores one declared member in either runtime or harness-template manifests.
type ManifestMember struct {
	OrbitID string    `yaml:"orbit_id"`
	Source  string    `yaml:"source,omitempty"`
	AddedAt time.Time `yaml:"added_at,omitempty"`
}

// ManifestVariableSpec captures template variable metadata embedded into the single-control-plane branch manifest.
type ManifestVariableSpec struct {
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required"`
}

type rawManifestFile struct {
	SchemaVersion      *int                               `yaml:"schema_version"`
	Kind               *string                            `yaml:"kind"`
	Source             *rawManifestSource                 `yaml:"source"`
	Runtime            *rawManifestRuntime                `yaml:"runtime"`
	Template           *rawManifestTemplate               `yaml:"template"`
	Members            []rawManifestMember                `yaml:"members"`
	Variables          map[string]rawManifestVariableSpec `yaml:"variables"`
	IncludesRootAgents *bool                              `yaml:"includes_root_agents"`
}

type rawManifestSource struct {
	OrbitID      *string `yaml:"orbit_id"`
	SourceBranch *string `yaml:"source_branch"`
}

type rawManifestRuntime struct {
	ID        *string    `yaml:"id"`
	Name      *string    `yaml:"name"`
	CreatedAt *time.Time `yaml:"created_at"`
	UpdatedAt *time.Time `yaml:"updated_at"`
}

type rawManifestTemplate struct {
	OrbitID           *string    `yaml:"orbit_id"`
	HarnessID         *string    `yaml:"harness_id"`
	DefaultTemplate   *bool      `yaml:"default_template"`
	CreatedFromBranch *string    `yaml:"created_from_branch"`
	CreatedFromCommit *string    `yaml:"created_from_commit"`
	CreatedAt         *time.Time `yaml:"created_at"`
}

type rawManifestMember struct {
	OrbitID *string    `yaml:"orbit_id"`
	Source  *string    `yaml:"source"`
	AddedAt *time.Time `yaml:"added_at"`
}

type rawManifestVariableSpec struct {
	Description *string `yaml:"description"`
	Required    *bool   `yaml:"required"`
}

// LoadManifestFile reads, decodes, and validates .harness/manifest.yaml.
func LoadManifestFile(repoRoot string) (ManifestFile, error) {
	return LoadManifestFileAtPath(ManifestPath(repoRoot))
}

// LoadManifestFileAtPath reads, decodes, and validates one single-control-plane manifest.
func LoadManifestFileAtPath(filename string) (ManifestFile, error) {
	//nolint:gosec // The path is repo-local and built from the fixed manifest contract path.
	data, err := os.ReadFile(filename)
	if err != nil {
		return ManifestFile{}, fmt.Errorf("read %s: %w", filename, err)
	}

	file, err := ParseManifestFileData(data)
	if err != nil {
		return ManifestFile{}, fmt.Errorf("validate %s: %w", filename, err)
	}

	return file, nil
}

// ParseManifestFileData decodes and validates .harness/manifest.yaml bytes.
func ParseManifestFileData(data []byte) (ManifestFile, error) {
	var raw rawManifestFile
	if err := contractutil.DecodeKnownFields(data, &raw); err != nil {
		return ManifestFile{}, fmt.Errorf("decode manifest file: %w", err)
	}

	file, err := raw.toManifestFile()
	if err != nil {
		return ManifestFile{}, err
	}

	return file, nil
}

// ValidateManifestFile validates the top-level manifest schema contract.
func ValidateManifestFile(file ManifestFile) error {
	if file.SchemaVersion != manifestSchemaVersion {
		return fmt.Errorf("schema_version must be %d", manifestSchemaVersion)
	}

	switch file.Kind {
	case ManifestKindSource:
		return ValidateSourceManifestFile(file)
	case ManifestKindRuntime:
		return ValidateRuntimeManifestFile(file)
	case ManifestKindOrbitTemplate:
		return ValidateOrbitTemplateManifestFile(file)
	case ManifestKindHarnessTemplate:
		return ValidateHarnessTemplateManifestFile(file)
	default:
		return fmt.Errorf(
			"kind must be one of %q, %q, %q, or %q",
			ManifestKindSource,
			ManifestKindRuntime,
			ManifestKindOrbitTemplate,
			ManifestKindHarnessTemplate,
		)
	}
}

// ValidateSourceManifestFile validates the source branch form of .harness/manifest.yaml.
func ValidateSourceManifestFile(file ManifestFile) error {
	if file.SchemaVersion != manifestSchemaVersion {
		return fmt.Errorf("schema_version must be %d", manifestSchemaVersion)
	}
	if file.Kind != ManifestKindSource {
		return fmt.Errorf("kind must be %q", ManifestKindSource)
	}
	if file.Source == nil {
		return fmt.Errorf("source must be present")
	}
	if file.Runtime != nil {
		return fmt.Errorf("runtime must not be present")
	}
	if file.Template != nil {
		return fmt.Errorf("template must not be present")
	}
	if file.Members != nil {
		return fmt.Errorf("members must not be present")
	}
	if file.IncludesRootAgents {
		return fmt.Errorf("includes_root_agents must not be present")
	}
	if err := ids.ValidateOrbitID(file.Source.OrbitID); err != nil {
		return fmt.Errorf("source.orbit_id: %w", err)
	}
	if file.Source.SourceBranch == "" {
		return fmt.Errorf("source.source_branch must not be empty")
	}

	return nil
}

// ValidateRuntimeManifestFile validates the runtime branch form of .harness/manifest.yaml.
func ValidateRuntimeManifestFile(file ManifestFile) error {
	if file.SchemaVersion != manifestSchemaVersion {
		return fmt.Errorf("schema_version must be %d", manifestSchemaVersion)
	}
	if file.Kind != ManifestKindRuntime {
		return fmt.Errorf("kind must be %q", ManifestKindRuntime)
	}
	if file.Runtime == nil {
		return fmt.Errorf("runtime must be present")
	}
	if file.Template != nil {
		return fmt.Errorf("template must not be present")
	}
	if file.Members == nil {
		return fmt.Errorf("members must be present")
	}
	if file.IncludesRootAgents {
		return fmt.Errorf("includes_root_agents must not be present")
	}
	if err := ids.ValidateOrbitID(file.Runtime.ID); err != nil {
		return fmt.Errorf("runtime.id: %w", err)
	}
	if file.Runtime.CreatedAt.IsZero() {
		return fmt.Errorf("runtime.created_at must be set")
	}
	if file.Runtime.UpdatedAt.IsZero() {
		return fmt.Errorf("runtime.updated_at must be set")
	}

	seenOrbitIDs := make(map[string]struct{}, len(file.Members))
	for index, member := range file.Members {
		if err := ids.ValidateOrbitID(member.OrbitID); err != nil {
			return fmt.Errorf("members[%d].orbit_id: %w", index, err)
		}
		if _, ok := seenOrbitIDs[member.OrbitID]; ok {
			return fmt.Errorf("members[%d].orbit_id must be unique", index)
		}
		seenOrbitIDs[member.OrbitID] = struct{}{}

		switch member.Source {
		case ManifestMemberSourceManual, ManifestMemberSourceInstallOrbit, ManifestMemberSourceInstallBundle:
		default:
			return fmt.Errorf(
				"members[%d].source must be one of %q, %q, or %q",
				index,
				ManifestMemberSourceManual,
				ManifestMemberSourceInstallOrbit,
				ManifestMemberSourceInstallBundle,
			)
		}
		if member.AddedAt.IsZero() {
			return fmt.Errorf("members[%d].added_at must be set", index)
		}
	}

	return nil
}

// ValidateOrbitTemplateManifestFile validates the orbit-template branch form of .harness/manifest.yaml.
func ValidateOrbitTemplateManifestFile(file ManifestFile) error {
	if file.SchemaVersion != manifestSchemaVersion {
		return fmt.Errorf("schema_version must be %d", manifestSchemaVersion)
	}
	if file.Kind != ManifestKindOrbitTemplate {
		return fmt.Errorf("kind must be %q", ManifestKindOrbitTemplate)
	}
	if file.Runtime != nil {
		return fmt.Errorf("runtime must not be present")
	}
	if file.Template == nil {
		return fmt.Errorf("template must be present")
	}
	if file.Members != nil {
		return fmt.Errorf("members must not be present")
	}
	if file.IncludesRootAgents {
		return fmt.Errorf("includes_root_agents must not be present")
	}
	if file.Template.OrbitID == "" {
		return fmt.Errorf("template.orbit_id must be present")
	}
	if err := ids.ValidateOrbitID(file.Template.OrbitID); err != nil {
		return fmt.Errorf("template.orbit_id: %w", err)
	}
	if file.Template.HarnessID != "" {
		return fmt.Errorf("template.harness_id must not be present")
	}
	if file.Template.CreatedFromBranch == "" {
		return fmt.Errorf("template.created_from_branch must not be empty")
	}
	if file.Template.CreatedFromCommit == "" {
		return fmt.Errorf("template.created_from_commit must not be empty")
	}
	if file.Template.CreatedAt.IsZero() {
		return fmt.Errorf("template.created_at must be set")
	}
	if file.Variables != nil {
		for _, name := range contractutil.SortedKeys(file.Variables) {
			if err := contractutil.ValidateVariableName(name); err != nil {
				return fmt.Errorf("variables.%s: %w", name, err)
			}
		}
	}

	return nil
}

// ValidateHarnessTemplateManifestFile validates the harness-template branch form of .harness/manifest.yaml.
func ValidateHarnessTemplateManifestFile(file ManifestFile) error {
	if file.SchemaVersion != manifestSchemaVersion {
		return fmt.Errorf("schema_version must be %d", manifestSchemaVersion)
	}
	if file.Kind != ManifestKindHarnessTemplate {
		return fmt.Errorf("kind must be %q", ManifestKindHarnessTemplate)
	}
	if file.Runtime != nil {
		return fmt.Errorf("runtime must not be present")
	}
	if file.Template == nil {
		return fmt.Errorf("template must be present")
	}
	if file.Members == nil {
		return fmt.Errorf("members must be present")
	}
	if file.Variables != nil {
		return fmt.Errorf("variables must not be present")
	}
	if len(file.Members) == 0 {
		return fmt.Errorf("members must contain at least one member")
	}
	if file.Template.HarnessID == "" {
		return fmt.Errorf("template.harness_id must be present")
	}
	if err := ids.ValidateOrbitID(file.Template.HarnessID); err != nil {
		return fmt.Errorf("template.harness_id: %w", err)
	}
	if file.Template.OrbitID != "" {
		return fmt.Errorf("template.orbit_id must not be present")
	}
	if file.Template.CreatedFromBranch == "" {
		return fmt.Errorf("template.created_from_branch must not be empty")
	}
	if file.Template.CreatedFromCommit == "" {
		return fmt.Errorf("template.created_from_commit must not be empty")
	}
	if file.Template.CreatedAt.IsZero() {
		return fmt.Errorf("template.created_at must be set")
	}

	seenOrbitIDs := make(map[string]struct{}, len(file.Members))
	for index, member := range file.Members {
		if err := ids.ValidateOrbitID(member.OrbitID); err != nil {
			return fmt.Errorf("members[%d].orbit_id: %w", index, err)
		}
		if _, ok := seenOrbitIDs[member.OrbitID]; ok {
			return fmt.Errorf("members[%d].orbit_id must be unique", index)
		}
		seenOrbitIDs[member.OrbitID] = struct{}{}

		if member.Source != "" {
			return fmt.Errorf("members[%d].source must not be present", index)
		}
		if !member.AddedAt.IsZero() {
			return fmt.Errorf("members[%d].added_at must not be present", index)
		}
	}

	return nil
}

// MarshalManifestFile validates and encodes one single-control-plane manifest with stable ordering.
func MarshalManifestFile(file ManifestFile) ([]byte, error) {
	if err := ValidateManifestFile(file); err != nil {
		return nil, fmt.Errorf("validate manifest file: %w", err)
	}

	data, err := contractutil.EncodeYAMLDocument(manifestFileNode(file))
	if err != nil {
		return nil, fmt.Errorf("encode manifest file: %w", err)
	}

	return data, nil
}

// WriteManifestFile validates and writes .harness/manifest.yaml with stable ordering.
func WriteManifestFile(repoRoot string, file ManifestFile) (string, error) {
	return WriteManifestFileAtPath(ManifestPath(repoRoot), file)
}

// WriteManifestFileAtPath validates and writes one single-control-plane manifest.
func WriteManifestFileAtPath(filename string, file ManifestFile) (string, error) {
	data, err := MarshalManifestFile(file)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", filename, err)
	}

	if err := contractutil.AtomicWriteFile(filename, data); err != nil {
		return "", fmt.Errorf("write %s: %w", filename, err)
	}

	return filename, nil
}

func (raw rawManifestFile) toManifestFile() (ManifestFile, error) {
	switch {
	case raw.SchemaVersion == nil:
		return ManifestFile{}, fmt.Errorf("schema_version must be present")
	case raw.Kind == nil:
		return ManifestFile{}, fmt.Errorf("kind must be present")
	}

	file := ManifestFile{
		SchemaVersion: *raw.SchemaVersion,
		Kind:          *raw.Kind,
	}

	switch file.Kind {
	case ManifestKindSource:
		if raw.Source == nil {
			return ManifestFile{}, fmt.Errorf("source must be present")
		}
		if raw.Runtime != nil {
			return ManifestFile{}, fmt.Errorf("runtime must not be present")
		}
		if raw.Template != nil {
			return ManifestFile{}, fmt.Errorf("template must not be present")
		}
		if raw.Members != nil {
			return ManifestFile{}, fmt.Errorf("members must not be present")
		}
		if raw.Variables != nil {
			return ManifestFile{}, fmt.Errorf("variables must not be present")
		}
		if raw.IncludesRootAgents != nil {
			return ManifestFile{}, fmt.Errorf("includes_root_agents must not be present")
		}

		sourceMetadata, err := raw.Source.toManifestSourceMetadata()
		if err != nil {
			return ManifestFile{}, err
		}
		file.Source = &sourceMetadata

	case ManifestKindRuntime:
		if raw.Runtime == nil {
			return ManifestFile{}, fmt.Errorf("runtime must be present")
		}
		if raw.Template != nil {
			return ManifestFile{}, fmt.Errorf("template must not be present")
		}
		if raw.Members == nil {
			return ManifestFile{}, fmt.Errorf("members must be present")
		}
		if raw.Variables != nil {
			return ManifestFile{}, fmt.Errorf("variables must not be present")
		}
		if raw.IncludesRootAgents != nil {
			return ManifestFile{}, fmt.Errorf("includes_root_agents must not be present")
		}

		runtimeMetadata, err := raw.Runtime.toManifestRuntimeMetadata()
		if err != nil {
			return ManifestFile{}, err
		}
		file.Runtime = &runtimeMetadata
		file.Members = make([]ManifestMember, 0, len(raw.Members))
		for index, rawMember := range raw.Members {
			member, err := rawMember.toRuntimeManifestMember(index)
			if err != nil {
				return ManifestFile{}, err
			}
			file.Members = append(file.Members, member)
		}

	case ManifestKindOrbitTemplate:
		if raw.Runtime != nil {
			return ManifestFile{}, fmt.Errorf("runtime must not be present")
		}
		if raw.Template == nil {
			return ManifestFile{}, fmt.Errorf("template must be present")
		}
		if raw.Members != nil {
			return ManifestFile{}, fmt.Errorf("members must not be present")
		}
		if raw.IncludesRootAgents != nil {
			return ManifestFile{}, fmt.Errorf("includes_root_agents must not be present")
		}

		templateMetadata, err := raw.Template.toOrbitTemplateMetadata()
		if err != nil {
			return ManifestFile{}, err
		}
		file.Template = &templateMetadata
		if raw.Variables != nil {
			file.Variables = make(map[string]ManifestVariableSpec, len(raw.Variables))
			for name, rawSpec := range raw.Variables {
				if rawSpec.Required == nil {
					return ManifestFile{}, fmt.Errorf("variables.%s.required must be present", name)
				}

				spec := ManifestVariableSpec{
					Required: *rawSpec.Required,
				}
				if rawSpec.Description != nil {
					spec.Description = *rawSpec.Description
				}
				file.Variables[name] = spec
			}
		}

	case ManifestKindHarnessTemplate:
		if raw.Runtime != nil {
			return ManifestFile{}, fmt.Errorf("runtime must not be present")
		}
		if raw.Template == nil {
			return ManifestFile{}, fmt.Errorf("template must be present")
		}
		if raw.Members == nil {
			return ManifestFile{}, fmt.Errorf("members must be present")
		}
		if raw.Variables != nil {
			return ManifestFile{}, fmt.Errorf("variables must not be present")
		}

		templateMetadata, err := raw.Template.toHarnessTemplateMetadata()
		if err != nil {
			return ManifestFile{}, err
		}
		file.Template = &templateMetadata
		file.Members = make([]ManifestMember, 0, len(raw.Members))
		for index, rawMember := range raw.Members {
			member, err := rawMember.toTemplateManifestMember(index)
			if err != nil {
				return ManifestFile{}, err
			}
			file.Members = append(file.Members, member)
		}
		if raw.IncludesRootAgents != nil {
			file.IncludesRootAgents = *raw.IncludesRootAgents
		}

	default:
		return ManifestFile{}, fmt.Errorf(
			"kind must be one of %q, %q, %q, or %q",
			ManifestKindSource,
			ManifestKindRuntime,
			ManifestKindOrbitTemplate,
			ManifestKindHarnessTemplate,
		)
	}

	if err := ValidateManifestFile(file); err != nil {
		return ManifestFile{}, err
	}

	return file, nil
}

func (raw rawManifestSource) toManifestSourceMetadata() (ManifestSourceMetadata, error) {
	switch {
	case raw.OrbitID == nil:
		return ManifestSourceMetadata{}, fmt.Errorf("source.orbit_id must be present")
	case raw.SourceBranch == nil:
		return ManifestSourceMetadata{}, fmt.Errorf("source.source_branch must be present")
	}

	return ManifestSourceMetadata{
		OrbitID:      *raw.OrbitID,
		SourceBranch: *raw.SourceBranch,
	}, nil
}

func (raw rawManifestRuntime) toManifestRuntimeMetadata() (ManifestRuntimeMetadata, error) {
	switch {
	case raw.ID == nil:
		return ManifestRuntimeMetadata{}, fmt.Errorf("runtime.id must be present")
	case raw.CreatedAt == nil:
		return ManifestRuntimeMetadata{}, fmt.Errorf("runtime.created_at must be present")
	case raw.UpdatedAt == nil:
		return ManifestRuntimeMetadata{}, fmt.Errorf("runtime.updated_at must be present")
	}

	metadata := ManifestRuntimeMetadata{
		ID:        *raw.ID,
		CreatedAt: raw.CreatedAt.UTC(),
		UpdatedAt: raw.UpdatedAt.UTC(),
	}
	if raw.Name != nil {
		metadata.Name = *raw.Name
	}

	return metadata, nil
}

func (raw rawManifestTemplate) toOrbitTemplateMetadata() (ManifestTemplateMetadata, error) {
	switch {
	case raw.OrbitID == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.orbit_id must be present")
	case raw.HarnessID != nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.harness_id must not be present")
	case raw.CreatedFromBranch == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.created_from_branch must be present")
	case raw.CreatedFromCommit == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.created_from_commit must be present")
	case raw.CreatedAt == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.created_at must be present")
	}

	return ManifestTemplateMetadata{
		OrbitID:           *raw.OrbitID,
		DefaultTemplate:   raw.DefaultTemplate != nil && *raw.DefaultTemplate,
		CreatedFromBranch: *raw.CreatedFromBranch,
		CreatedFromCommit: *raw.CreatedFromCommit,
		CreatedAt:         raw.CreatedAt.UTC(),
	}, nil
}

func (raw rawManifestTemplate) toHarnessTemplateMetadata() (ManifestTemplateMetadata, error) {
	switch {
	case raw.HarnessID == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.harness_id must be present")
	case raw.OrbitID != nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.orbit_id must not be present")
	case raw.CreatedFromBranch == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.created_from_branch must be present")
	case raw.CreatedFromCommit == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.created_from_commit must be present")
	case raw.CreatedAt == nil:
		return ManifestTemplateMetadata{}, fmt.Errorf("template.created_at must be present")
	}

	return ManifestTemplateMetadata{
		HarnessID:         *raw.HarnessID,
		DefaultTemplate:   raw.DefaultTemplate != nil && *raw.DefaultTemplate,
		CreatedFromBranch: *raw.CreatedFromBranch,
		CreatedFromCommit: *raw.CreatedFromCommit,
		CreatedAt:         raw.CreatedAt.UTC(),
	}, nil
}

func (raw rawManifestMember) toRuntimeManifestMember(index int) (ManifestMember, error) {
	switch {
	case raw.OrbitID == nil:
		return ManifestMember{}, fmt.Errorf("members[%d].orbit_id must be present", index)
	case raw.Source == nil:
		return ManifestMember{}, fmt.Errorf("members[%d].source must be present", index)
	case raw.AddedAt == nil:
		return ManifestMember{}, fmt.Errorf("members[%d].added_at must be present", index)
	}

	return ManifestMember{
		OrbitID: *raw.OrbitID,
		Source:  *raw.Source,
		AddedAt: raw.AddedAt.UTC(),
	}, nil
}

func (raw rawManifestMember) toTemplateManifestMember(index int) (ManifestMember, error) {
	switch {
	case raw.OrbitID == nil:
		return ManifestMember{}, fmt.Errorf("members[%d].orbit_id must be present", index)
	case raw.Source != nil:
		return ManifestMember{}, fmt.Errorf("members[%d].source must not be present", index)
	case raw.AddedAt != nil:
		return ManifestMember{}, fmt.Errorf("members[%d].added_at must not be present", index)
	}

	return ManifestMember{
		OrbitID: *raw.OrbitID,
	}, nil
}

func manifestFileNode(file ManifestFile) *yaml.Node {
	root := contractutil.MappingNode()
	contractutil.AppendMapping(root, "schema_version", contractutil.IntNode(file.SchemaVersion))
	contractutil.AppendMapping(root, "kind", contractutil.StringNode(file.Kind))

	switch file.Kind {
	case ManifestKindSource:
		contractutil.AppendMapping(root, "source", sourceNode(file.Source))

	case ManifestKindRuntime:
		runtimeNode := contractutil.MappingNode()
		contractutil.AppendMapping(runtimeNode, "id", contractutil.StringNode(file.Runtime.ID))
		if file.Runtime.Name != "" {
			contractutil.AppendMapping(runtimeNode, "name", contractutil.StringNode(file.Runtime.Name))
		}
		contractutil.AppendMapping(runtimeNode, "created_at", contractutil.TimestampNode(file.Runtime.CreatedAt))
		contractutil.AppendMapping(runtimeNode, "updated_at", contractutil.TimestampNode(file.Runtime.UpdatedAt))
		contractutil.AppendMapping(root, "runtime", runtimeNode)
		contractutil.AppendMapping(root, "members", runtimeMembersNode(file.Members))

	case ManifestKindOrbitTemplate:
		contractutil.AppendMapping(root, "template", orbitTemplateNode(file.Template))
		if file.Variables != nil {
			contractutil.AppendMapping(root, "variables", manifestVariablesNode(file.Variables))
		}

	case ManifestKindHarnessTemplate:
		contractutil.AppendMapping(root, "template", harnessTemplateNode(file.Template))
		contractutil.AppendMapping(root, "members", templateMembersNode(file.Members))
		contractutil.AppendMapping(root, "includes_root_agents", contractutil.BoolNode(file.IncludesRootAgents))
	}

	return root
}

func sourceNode(source *ManifestSourceMetadata) *yaml.Node {
	node := contractutil.MappingNode()
	contractutil.AppendMapping(node, "orbit_id", contractutil.StringNode(source.OrbitID))
	contractutil.AppendMapping(node, "source_branch", contractutil.StringNode(source.SourceBranch))
	return node
}

func orbitTemplateNode(template *ManifestTemplateMetadata) *yaml.Node {
	node := contractutil.MappingNode()
	contractutil.AppendMapping(node, "orbit_id", contractutil.StringNode(template.OrbitID))
	contractutil.AppendMapping(node, "default_template", contractutil.BoolNode(template.DefaultTemplate))
	contractutil.AppendMapping(node, "created_from_branch", contractutil.StringNode(template.CreatedFromBranch))
	contractutil.AppendMapping(node, "created_from_commit", contractutil.StringNode(template.CreatedFromCommit))
	contractutil.AppendMapping(node, "created_at", contractutil.TimestampNode(template.CreatedAt))
	return node
}

func harnessTemplateNode(template *ManifestTemplateMetadata) *yaml.Node {
	node := contractutil.MappingNode()
	contractutil.AppendMapping(node, "harness_id", contractutil.StringNode(template.HarnessID))
	contractutil.AppendMapping(node, "default_template", contractutil.BoolNode(template.DefaultTemplate))
	contractutil.AppendMapping(node, "created_from_branch", contractutil.StringNode(template.CreatedFromBranch))
	contractutil.AppendMapping(node, "created_from_commit", contractutil.StringNode(template.CreatedFromCommit))
	contractutil.AppendMapping(node, "created_at", contractutil.TimestampNode(template.CreatedAt))
	return node
}

func runtimeMembersNode(members []ManifestMember) *yaml.Node {
	node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, member := range sortedManifestMembers(members) {
		memberNode := contractutil.MappingNode()
		contractutil.AppendMapping(memberNode, "orbit_id", contractutil.StringNode(member.OrbitID))
		contractutil.AppendMapping(memberNode, "source", contractutil.StringNode(member.Source))
		contractutil.AppendMapping(memberNode, "added_at", contractutil.TimestampNode(member.AddedAt))
		node.Content = append(node.Content, memberNode)
	}

	return node
}

func templateMembersNode(members []ManifestMember) *yaml.Node {
	node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, member := range sortedManifestMembers(members) {
		memberNode := contractutil.MappingNode()
		contractutil.AppendMapping(memberNode, "orbit_id", contractutil.StringNode(member.OrbitID))
		node.Content = append(node.Content, memberNode)
	}

	return node
}

func manifestVariablesNode(variables map[string]ManifestVariableSpec) *yaml.Node {
	node := contractutil.MappingNode()
	for _, name := range contractutil.SortedKeys(variables) {
		spec := variables[name]
		specNode := contractutil.MappingNode()
		if spec.Description != "" {
			contractutil.AppendMapping(specNode, "description", contractutil.StringNode(spec.Description))
		}
		contractutil.AppendMapping(specNode, "required", contractutil.BoolNode(spec.Required))
		contractutil.AppendMapping(node, name, specNode)
	}

	return node
}

func sortedManifestMembers(members []ManifestMember) []ManifestMember {
	sorted := append([]ManifestMember(nil), members...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OrbitID < sorted[j].OrbitID
	})
	return sorted
}
