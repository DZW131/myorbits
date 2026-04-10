package branchinfo

import (
	"context"
	"fmt"
	pathpkg "path"
	"sort"
	"strings"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	harnesspkg "github.com/zack-nova/orbit/cmd/orbit/cli/harness"
	"github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

const (
	installRecordsDir      = ".harness/installs"
	runtimeDefinitionsDir  = ".harness/orbits"
	templateDefinitionsDir = ".orbit/orbits"
)

type definitionLoader struct {
	definitionsDir string
	parser         func([]byte, string) (orbit.OrbitSpec, error)
}

// Inspection captures the stable branch-inspection details surfaced by CLI commands.
type Inspection struct {
	Classification     Classification `json:"classification"`
	SourceBranch       string         `json:"source_branch,omitempty"`
	PublishOrbitID     string         `json:"publish_orbit_id,omitempty"`
	HarnessID          string         `json:"harness_id,omitempty"`
	MemberCount        int            `json:"member_count"`
	MemberIDs          []string       `json:"member_ids"`
	DefinitionCount    int            `json:"definition_count"`
	DefinitionIDs      []string       `json:"definition_ids"`
	InstallCount       int            `json:"install_count"`
	InstallIDs         []string       `json:"install_ids"`
	IncludesRootAgents *bool          `json:"includes_root_agents,omitempty"`
}

// InspectRevision classifies one revision and loads the documented branch summary fields.
func InspectRevision(ctx context.Context, repoRoot string, rev string) (Inspection, error) {
	classification, err := ClassifyRevision(ctx, repoRoot, rev)
	if err != nil {
		return Inspection{}, err
	}

	installRecords, err := loadValidInstallRecords(ctx, repoRoot, rev)
	if err != nil {
		return Inspection{}, err
	}

	definitions, err := loadValidDefinitions(ctx, repoRoot, rev, classification)
	if err != nil {
		return Inspection{}, err
	}

	inspection := Inspection{
		Classification: classification,
		MemberIDs:      []string{},
		DefinitionIDs:  definitionIDs(definitions),
		InstallIDs:     installRecordIDs(installRecords),
	}
	inspection.DefinitionCount = len(inspection.DefinitionIDs)
	inspection.InstallCount = len(inspection.InstallIDs)

	switch classification.Kind {
	case KindSource:
		manifestFile, err := loadManifestFileAtRev(ctx, repoRoot, rev)
		if err != nil {
			return Inspection{}, err
		}

		inspection.SourceBranch = manifestFile.Source.SourceBranch
		inspection.PublishOrbitID = manifestFile.Source.OrbitID
	case KindTemplate:
		if err := populateTemplateInspection(ctx, repoRoot, rev, classification, &inspection); err != nil {
			return Inspection{}, err
		}
	case KindRuntime:
		manifestFile, err := loadManifestFileAtRev(ctx, repoRoot, rev)
		if err != nil {
			return Inspection{}, err
		}

		inspection.HarnessID = manifestFile.Runtime.ID
		inspection.MemberIDs = manifestMemberIDs(manifestFile.Members)
		inspection.MemberCount = len(inspection.MemberIDs)
	}

	return inspection, nil
}

func populateTemplateInspection(ctx context.Context, repoRoot string, rev string, classification Classification, inspection *Inspection) error {
	switch classification.TemplateKind {
	case TemplateKindOrbit, TemplateKindHarness:
		manifest, err := loadManifestFileAtRev(ctx, repoRoot, rev)
		if err != nil {
			return err
		}

		if classification.TemplateKind == TemplateKindHarness {
			inspection.HarnessID = manifest.Template.HarnessID
		}
		inspection.MemberIDs = manifestMemberIDs(manifest.Members)
		inspection.MemberCount = len(inspection.MemberIDs)
		if classification.TemplateKind == TemplateKindHarness {
			includesRootAgents := manifest.IncludesRootAgents
			inspection.IncludesRootAgents = &includesRootAgents
		}

		return nil
	case "":
		return nil
	default:
		return fmt.Errorf("unsupported template kind %q", classification.TemplateKind)
	}
}

func loadManifestFileAtRev(ctx context.Context, repoRoot string, rev string) (harnesspkg.ManifestFile, error) {
	data, err := gitpkg.ReadFileAtRev(ctx, repoRoot, rev, manifestRelativePath)
	if err != nil {
		return harnesspkg.ManifestFile{}, fmt.Errorf("read %s at %s: %w", manifestRelativePath, rev, err)
	}

	file, err := harnesspkg.ParseManifestFileData(data)
	if err != nil {
		return harnesspkg.ManifestFile{}, fmt.Errorf("parse %s at %s: %w", manifestRelativePath, rev, err)
	}

	return file, nil
}

func loadValidDefinitions(
	ctx context.Context,
	repoRoot string,
	rev string,
	classification Classification,
) ([]orbit.Definition, error) {
	loaders := definitionLoadersForClassification(classification)
	for _, loader := range loaders {
		paths, err := gitpkg.ListFilesAtRev(ctx, repoRoot, rev, loader.definitionsDir)
		if err != nil {
			return nil, fmt.Errorf("list %s at %s: %w", loader.definitionsDir, rev, err)
		}
		if len(paths) == 0 {
			continue
		}

		definitions := make([]orbit.Definition, 0, len(paths))
		for _, definitionPath := range paths {
			data, err := gitpkg.ReadFileAtRev(ctx, repoRoot, rev, definitionPath)
			if err != nil {
				return nil, fmt.Errorf("read %s at %s: %w", definitionPath, rev, err)
			}

			spec, err := loader.parser(data, definitionPath)
			if err != nil {
				continue
			}
			definitions = append(definitions, orbit.Definition{ID: spec.ID})
		}

		return definitions, nil
	}

	return nil, nil
}

func definitionLoadersForClassification(classification Classification) []definitionLoader {
	if classification.Kind == KindTemplate || classification.Kind == KindSource {
		return []definitionLoader{
			{
				definitionsDir: runtimeDefinitionsDir,
				parser:         orbit.ParseHostedOrbitSpecData,
			},
			{
				definitionsDir: templateDefinitionsDir,
				parser:         orbit.ParseOrbitSpecData,
			},
		}
	}

	return []definitionLoader{
		{
			definitionsDir: runtimeDefinitionsDir,
			parser:         orbit.ParseHostedOrbitSpecData,
		},
	}
}

func loadValidInstallRecords(ctx context.Context, repoRoot string, rev string) ([]orbittemplate.InstallRecord, error) {
	paths, err := gitpkg.ListFilesAtRev(ctx, repoRoot, rev, installRecordsDir)
	if err != nil {
		return nil, fmt.Errorf("list %s at %s: %w", installRecordsDir, rev, err)
	}

	records := make([]orbittemplate.InstallRecord, 0, len(paths))
	for _, recordPath := range paths {
		data, err := gitpkg.ReadFileAtRev(ctx, repoRoot, rev, recordPath)
		if err != nil {
			return nil, fmt.Errorf("read %s at %s: %w", recordPath, rev, err)
		}

		record, err := orbittemplate.ParseInstallRecordData(data)
		if err != nil {
			continue
		}

		expectedOrbitID := strings.TrimSuffix(pathpkg.Base(recordPath), pathpkg.Ext(recordPath))
		if record.OrbitID != expectedOrbitID {
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

func definitionIDs(definitions []orbit.Definition) []string {
	idsList := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		idsList = append(idsList, definition.ID)
	}
	sort.Strings(idsList)

	return idsList
}

func installRecordIDs(records []orbittemplate.InstallRecord) []string {
	idsList := make([]string, 0, len(records))
	for _, record := range records {
		idsList = append(idsList, record.OrbitID)
	}
	sort.Strings(idsList)

	return idsList
}

func manifestMemberIDs(members []harnesspkg.ManifestMember) []string {
	idsList := make([]string, 0, len(members))
	for _, member := range members {
		idsList = append(idsList, member.OrbitID)
	}
	sort.Strings(idsList)

	return idsList
}
