package orbittemplate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/zack-nova/orbit/cmd/orbit/cli/ids"
	"github.com/zack-nova/orbit/cmd/orbit/cli/internal/contractutil"
)

const (
	installRecordDirName         = ".orbit/installs"
	installRecordSchemaVersion   = 1
	InstallSourceKindLocalBranch = "local_branch"
	InstallSourceKindRemoteGit   = "remote_git"
)

// InstallRecord stores the schema-backed template source for one orbit installation.
type InstallRecord struct {
	SchemaVersion int       `yaml:"schema_version"`
	OrbitID       string    `yaml:"orbit_id"`
	Template      Source    `yaml:"template"`
	AppliedAt     time.Time `yaml:"applied_at"`
}

// Source captures where the installed template came from.
type Source struct {
	SourceKind     string `yaml:"source_kind"`
	SourceRepo     string `yaml:"source_repo"`
	SourceRef      string `yaml:"source_ref"`
	TemplateCommit string `yaml:"template_commit"`
}

type rawInstallRecord struct {
	SchemaVersion *int               `yaml:"schema_version"`
	OrbitID       *string            `yaml:"orbit_id"`
	Template      *rawTemplateSource `yaml:"template"`
	AppliedAt     *time.Time         `yaml:"applied_at"`
}

type rawTemplateSource struct {
	SourceKind     *string `yaml:"source_kind"`
	SourceRepo     *string `yaml:"source_repo"`
	SourceRef      *string `yaml:"source_ref"`
	TemplateCommit *string `yaml:"template_commit"`
}

// InstallRecordPath returns the absolute path to one install record.
func InstallRecordPath(repoRoot string, orbitID string) (string, error) {
	repoPath, err := InstallRecordRepoPath(orbitID)
	if err != nil {
		return "", err
	}

	return filepath.Join(repoRoot, filepath.FromSlash(repoPath)), nil
}

// InstallRecordRepoPath returns the repository-relative path to one install record.
func InstallRecordRepoPath(orbitID string) (string, error) {
	if err := ids.ValidateOrbitID(orbitID); err != nil {
		return "", fmt.Errorf("validate orbit id: %w", err)
	}

	return installRecordDirName + "/" + orbitID + ".yaml", nil
}

// LoadInstallRecord reads, decodes, and validates one install record from the fixed Phase 2 host path.
func LoadInstallRecord(repoRoot string, orbitID string) (InstallRecord, error) {
	filename, err := InstallRecordPath(repoRoot, orbitID)
	if err != nil {
		return InstallRecord{}, fmt.Errorf("build install record path: %w", err)
	}

	record, err := LoadInstallRecordFile(filename)
	if err != nil {
		return InstallRecord{}, err
	}
	if record.OrbitID != orbitID {
		return InstallRecord{}, fmt.Errorf("validate %s: orbit_id must match install path", filename)
	}

	return record, nil
}

// LoadInstallRecordFile reads, decodes, and validates one install record from an absolute path.
func LoadInstallRecordFile(filename string) (InstallRecord, error) {
	//nolint:gosec // The path is repo-local and built from a validated orbit id.
	data, err := os.ReadFile(filename)
	if err != nil {
		return InstallRecord{}, fmt.Errorf("read %s: %w", filename, err)
	}

	record, err := ParseInstallRecordData(data)
	if err != nil {
		return InstallRecord{}, fmt.Errorf("parse %s: %w", filename, err)
	}

	return record, nil
}

// ParseInstallRecordData decodes and validates install-record bytes.
func ParseInstallRecordData(data []byte) (InstallRecord, error) {
	var raw rawInstallRecord
	if err := contractutil.DecodeKnownFields(data, &raw); err != nil {
		return InstallRecord{}, fmt.Errorf("decode install record: %w", err)
	}

	record, err := raw.toInstallRecord()
	if err != nil {
		return InstallRecord{}, fmt.Errorf("validate install record: %w", err)
	}

	return record, nil
}

// ValidateInstallRecord validates the install-record schema contract.
func ValidateInstallRecord(record InstallRecord) error {
	if record.SchemaVersion != installRecordSchemaVersion {
		return fmt.Errorf("schema_version must be %d", installRecordSchemaVersion)
	}
	if err := ids.ValidateOrbitID(record.OrbitID); err != nil {
		return fmt.Errorf("orbit_id: %w", err)
	}

	switch record.Template.SourceKind {
	case InstallSourceKindLocalBranch:
	case InstallSourceKindRemoteGit:
		if record.Template.SourceRepo == "" {
			return fmt.Errorf("template.source_repo must not be empty for %q", InstallSourceKindRemoteGit)
		}
	default:
		return fmt.Errorf("template.source_kind must be one of %q or %q", InstallSourceKindLocalBranch, InstallSourceKindRemoteGit)
	}

	if record.Template.SourceRef == "" {
		return fmt.Errorf("template.source_ref must not be empty")
	}
	if record.Template.TemplateCommit == "" {
		return fmt.Errorf("template.template_commit must not be empty")
	}
	if record.AppliedAt.IsZero() {
		return fmt.Errorf("applied_at must be set")
	}

	return nil
}

// WriteInstallRecord validates and writes one install record with stable ordering to the fixed Phase 2 host path.
func WriteInstallRecord(repoRoot string, record InstallRecord) (string, error) {
	filename, err := InstallRecordPath(repoRoot, record.OrbitID)
	if err != nil {
		return "", fmt.Errorf("build install record path: %w", err)
	}

	return WriteInstallRecordFile(filename, record)
}

// WriteInstallRecordFile validates and writes one install record with stable ordering.
func WriteInstallRecordFile(filename string, record InstallRecord) (string, error) {
	if err := ValidateInstallRecord(record); err != nil {
		return "", fmt.Errorf("validate install record: %w", err)
	}

	data, err := contractutil.EncodeYAMLDocument(installRecordNode(record))
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", filename, err)
	}

	if err := contractutil.AtomicWriteFile(filename, data); err != nil {
		return "", fmt.Errorf("write %s: %w", filename, err)
	}

	return filename, nil
}

func (raw rawInstallRecord) toInstallRecord() (InstallRecord, error) {
	switch {
	case raw.SchemaVersion == nil:
		return InstallRecord{}, fmt.Errorf("schema_version must be present")
	case raw.OrbitID == nil:
		return InstallRecord{}, fmt.Errorf("orbit_id must be present")
	case raw.Template == nil:
		return InstallRecord{}, fmt.Errorf("template must be present")
	case raw.AppliedAt == nil:
		return InstallRecord{}, fmt.Errorf("applied_at must be present")
	}

	templateSource, err := raw.Template.toTemplateSource()
	if err != nil {
		return InstallRecord{}, err
	}

	record := InstallRecord{
		SchemaVersion: *raw.SchemaVersion,
		OrbitID:       *raw.OrbitID,
		Template:      templateSource,
		AppliedAt:     raw.AppliedAt.UTC(),
	}

	if err := ValidateInstallRecord(record); err != nil {
		return InstallRecord{}, err
	}

	return record, nil
}

func (raw rawTemplateSource) toTemplateSource() (Source, error) {
	switch {
	case raw.SourceKind == nil:
		return Source{}, fmt.Errorf("template.source_kind must be present")
	case raw.SourceRepo == nil:
		return Source{}, fmt.Errorf("template.source_repo must be present")
	case raw.SourceRef == nil:
		return Source{}, fmt.Errorf("template.source_ref must be present")
	case raw.TemplateCommit == nil:
		return Source{}, fmt.Errorf("template.template_commit must be present")
	default:
		return Source{
			SourceKind:     *raw.SourceKind,
			SourceRepo:     *raw.SourceRepo,
			SourceRef:      *raw.SourceRef,
			TemplateCommit: *raw.TemplateCommit,
		}, nil
	}
}

func installRecordNode(record InstallRecord) *yaml.Node {
	root := contractutil.MappingNode()
	contractutil.AppendMapping(root, "schema_version", contractutil.IntNode(record.SchemaVersion))
	contractutil.AppendMapping(root, "orbit_id", contractutil.StringNode(record.OrbitID))

	templateNode := contractutil.MappingNode()
	contractutil.AppendMapping(templateNode, "source_kind", contractutil.StringNode(record.Template.SourceKind))
	contractutil.AppendMapping(templateNode, "source_repo", contractutil.StringNode(record.Template.SourceRepo))
	contractutil.AppendMapping(templateNode, "source_ref", contractutil.StringNode(record.Template.SourceRef))
	contractutil.AppendMapping(templateNode, "template_commit", contractutil.StringNode(record.Template.TemplateCommit))
	contractutil.AppendMapping(root, "template", templateNode)

	contractutil.AppendMapping(root, "applied_at", contractutil.TimestampNode(record.AppliedAt))

	return root
}
