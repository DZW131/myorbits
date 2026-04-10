package harness

import (
	"context"
	"errors"
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"

	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

// CheckFindingKind is one stable harness-check diagnostic kind.
type CheckFindingKind string

const (
	CheckFindingManifestSchemaInvalid CheckFindingKind = "manifest_schema_invalid"
	CheckFindingMissingDefinition     CheckFindingKind = "missing_definition"
	CheckFindingInstallMemberMismatch CheckFindingKind = "install_member_mismatch"
	CheckFindingInstallRecordInvalid  CheckFindingKind = "install_record_invalid"
	CheckFindingInstallPathMismatch   CheckFindingKind = "install_path_mismatch"
	CheckFindingBundleMemberMismatch  CheckFindingKind = "bundle_member_mismatch"
	CheckFindingBundlePathMismatch    CheckFindingKind = "bundle_path_mismatch"
)

// CheckFinding captures one stable harness-check diagnostic.
type CheckFinding struct {
	Kind    CheckFindingKind `json:"kind"`
	OrbitID string           `json:"orbit_id,omitempty"`
	Path    string           `json:"path,omitempty"`
	Message string           `json:"message"`
}

// CheckResult captures the current harness-check result.
type CheckResult struct {
	HarnessID    string         `json:"harness_id,omitempty"`
	OK           bool           `json:"ok"`
	FindingCount int            `json:"finding_count"`
	Findings     []CheckFinding `json:"findings"`
}

// CheckRuntime analyzes the current harness runtime and reports stable diagnostics.
func CheckRuntime(ctx context.Context, repoRoot string) (CheckResult, error) {
	manifestData, readErr := os.ReadFile(ManifestPath(repoRoot))
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return CheckResult{}, fmt.Errorf("read %s: %w", ManifestPath(repoRoot), readErr)
		}
		return CheckResult{}, fmt.Errorf("read %s: %w", ManifestPath(repoRoot), readErr)
	}

	runtimeFile, schemaFinding := parseManifestFileForCheck(manifestData)
	if schemaFinding != nil {
		findings := []CheckFinding{*schemaFinding}
		return CheckResult{
			OK:           false,
			FindingCount: len(findings),
			Findings:     findings,
		}, nil
	}

	findings, err := collectMembershipFindings(ctx, repoRoot, runtimeFile)
	if err != nil {
		return CheckResult{}, err
	}
	sortCheckFindings(findings)

	return CheckResult{
		HarnessID:    runtimeFile.Harness.ID,
		OK:           len(findings) == 0,
		FindingCount: len(findings),
		Findings:     findings,
	}, nil
}

func parseManifestFileForCheck(data []byte) (RuntimeFile, *CheckFinding) {
	manifestFile, err := ParseManifestFileData(data)
	if err == nil {
		runtimeFile, convertErr := RuntimeFileFromManifestFile(manifestFile)
		if convertErr == nil {
			return runtimeFile, nil
		}
		err = convertErr
	}

	return RuntimeFile{}, &CheckFinding{
		Kind:    CheckFindingManifestSchemaInvalid,
		Path:    ManifestRepoPath(),
		Message: err.Error(),
	}
}

func collectMembershipFindings(ctx context.Context, repoRoot string, runtimeFile RuntimeFile) ([]CheckFinding, error) {
	findings := make([]CheckFinding, 0)
	validInstallRecords, installFindings, err := scanInstallRecords(repoRoot)
	if err != nil {
		return nil, err
	}
	findings = append(findings, installFindings...)
	validBundleRecords, bundleFindings, err := scanBundleRecords(repoRoot)
	if err != nil {
		return nil, err
	}
	findings = append(findings, bundleFindings...)

	installMembers := make(map[string]struct{})
	bundleMembers := make(map[string]struct{})

	for _, member := range runtimeFile.Members {
		definitionPath, err := orbitpkg.HostedDefinitionRelativePath(member.OrbitID)
		if err != nil {
			return nil, fmt.Errorf("build definition path for %q: %w", member.OrbitID, err)
		}

		if _, err := orbitpkg.LoadHostedOrbitSpec(ctx, repoRoot, member.OrbitID); err != nil {
			findings = append(findings, CheckFinding{
				Kind:    CheckFindingMissingDefinition,
				OrbitID: member.OrbitID,
				Path:    definitionPath,
				Message: fmt.Sprintf("member %q points to a missing or invalid orbit definition", member.OrbitID),
			})
		}

		switch member.Source {
		case MemberSourceInstallOrbit:
			installMembers[member.OrbitID] = struct{}{}
			if _, ok := validInstallRecords[member.OrbitID]; !ok {
				installPath, pathErr := InstallRecordRepoPath(member.OrbitID)
				if pathErr != nil {
					return nil, fmt.Errorf("build install record path for %q: %w", member.OrbitID, pathErr)
				}
				findings = append(findings, CheckFinding{
					Kind:    CheckFindingInstallMemberMismatch,
					OrbitID: member.OrbitID,
					Path:    installPath,
					Message: fmt.Sprintf("install-backed member %q has missing install record", member.OrbitID),
				})
				continue
			}

			driftFindings, err := orbittemplate.AnalyzeInstalledTemplateDrift(ctx, repoRoot, member.OrbitID)
			if err != nil {
				return nil, fmt.Errorf("analyze installed template drift for %q: %w", member.OrbitID, err)
			}
			for _, driftFinding := range driftFindings {
				findings = append(findings, CheckFinding{
					Kind:    CheckFindingKind(driftFinding.Kind),
					OrbitID: member.OrbitID,
					Path:    driftFinding.Path,
					Message: driftMessageForFinding(member.OrbitID, driftFinding),
				})
			}
		case MemberSourceInstallBundle:
			bundleMembers[member.OrbitID] = struct{}{}
			if _, ok := findBundleRecordForMember(validBundleRecords, member.OrbitID); !ok {
				findings = append(findings, CheckFinding{
					Kind:    CheckFindingBundleMemberMismatch,
					OrbitID: member.OrbitID,
					Path:    BundleRecordsDirRepoPath(),
					Message: fmt.Sprintf("bundle-backed member %q has no matching bundle record", member.OrbitID),
				})
			}
		}
	}

	for orbitID, recordPath := range validInstallRecords {
		if _, ok := installMembers[orbitID]; ok {
			continue
		}
		findings = append(findings, CheckFinding{
			Kind:    CheckFindingInstallMemberMismatch,
			OrbitID: orbitID,
			Path:    recordPath,
			Message: fmt.Sprintf("install record %q has no matching install-backed member", orbitID),
		})
	}

	for harnessID, bundlePath := range validBundleRecords {
		if bundleRecordHasAnyMember(bundleMembers, harnessID, validBundleRecords[harnessID]) {
			continue
		}
		findings = append(findings, CheckFinding{
			Kind:    CheckFindingBundleMemberMismatch,
			OrbitID: harnessID,
			Path:    bundlePath.Path,
			Message: fmt.Sprintf("bundle record %q has no matching bundle-backed members", harnessID),
		})
	}

	return findings, nil
}

type validBundleRecord struct {
	Path   string
	Record BundleRecord
}

func scanInstallRecords(repoRoot string) (map[string]string, []CheckFinding, error) {
	entries, err := os.ReadDir(InstallRecordsDirPath(repoRoot))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, []CheckFinding{}, nil
		}
		return nil, nil, fmt.Errorf("read harness install records directory: %w", err)
	}

	validRecords := make(map[string]string)
	findings := make([]CheckFinding, 0)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".yaml" {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		absolutePath := filepath.Join(InstallRecordsDirPath(repoRoot), name)
		repoPath := pathpkg.Join(InstallRecordsDirRepoPath(), name)

		//nolint:gosec // Paths come from the fixed repo-local .harness/installs directory.
		data, err := os.ReadFile(absolutePath)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", absolutePath, err)
		}

		record, err := orbittemplate.ParseInstallRecordData(data)
		if err != nil {
			findings = append(findings, CheckFinding{
				Kind:    CheckFindingInstallRecordInvalid,
				Path:    repoPath,
				Message: fmt.Sprintf("install record is invalid: %v", err),
			})
			continue
		}

		expectedName := record.OrbitID + ".yaml"
		if name != expectedName {
			findings = append(findings, CheckFinding{
				Kind:    CheckFindingInstallPathMismatch,
				OrbitID: record.OrbitID,
				Path:    repoPath,
				Message: fmt.Sprintf("install record path %q does not match orbit_id %q", repoPath, record.OrbitID),
			})
			continue
		}

		validRecords[record.OrbitID] = repoPath
	}

	return validRecords, findings, nil
}

func scanBundleRecords(repoRoot string) (map[string]validBundleRecord, []CheckFinding, error) {
	entries, err := os.ReadDir(BundleRecordsDirPath(repoRoot))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]validBundleRecord{}, []CheckFinding{}, nil
		}
		return nil, nil, fmt.Errorf("read harness bundle records directory: %w", err)
	}

	validRecords := make(map[string]validBundleRecord)
	findings := make([]CheckFinding, 0)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".yaml" {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		absolutePath := filepath.Join(BundleRecordsDirPath(repoRoot), name)
		repoPath := pathpkg.Join(BundleRecordsDirRepoPath(), name)

		//nolint:gosec // Paths come from the fixed repo-local .harness/bundles directory.
		data, err := os.ReadFile(absolutePath)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", absolutePath, err)
		}

		record, err := ParseBundleRecordData(data)
		if err != nil {
			findings = append(findings, CheckFinding{
				Kind:    CheckFindingBundlePathMismatch,
				Path:    repoPath,
				Message: fmt.Sprintf("bundle record is invalid: %v", err),
			})
			continue
		}

		expectedName := record.HarnessID + ".yaml"
		if name != expectedName {
			findings = append(findings, CheckFinding{
				Kind:    CheckFindingBundlePathMismatch,
				OrbitID: record.HarnessID,
				Path:    repoPath,
				Message: fmt.Sprintf("bundle record path %q does not match harness_id %q", repoPath, record.HarnessID),
			})
			continue
		}

		validRecords[record.HarnessID] = validBundleRecord{
			Path:   repoPath,
			Record: record,
		}
	}

	return validRecords, findings, nil
}

func findBundleRecordForMember(records map[string]validBundleRecord, orbitID string) (validBundleRecord, bool) {
	for _, record := range records {
		for _, memberID := range record.Record.MemberIDs {
			if memberID == orbitID {
				return record, true
			}
		}
	}

	return validBundleRecord{}, false
}

func bundleRecordHasAnyMember(bundleMembers map[string]struct{}, harnessID string, record validBundleRecord) bool {
	_ = harnessID
	for _, memberID := range record.Record.MemberIDs {
		if _, ok := bundleMembers[memberID]; ok {
			return true
		}
	}

	return false
}

func sortCheckFindings(findings []CheckFinding) {
	sort.Slice(findings, func(left, right int) bool {
		if findings[left].Path == findings[right].Path {
			if findings[left].OrbitID == findings[right].OrbitID {
				return findings[left].Kind < findings[right].Kind
			}
			return findings[left].OrbitID < findings[right].OrbitID
		}
		return findings[left].Path < findings[right].Path
	})
}

func driftMessageForFinding(orbitID string, finding orbittemplate.InstallDriftFinding) string {
	switch finding.Kind {
	case orbittemplate.DriftKindDefinition:
		return fmt.Sprintf("install-backed member %q definition drift detected", orbitID)
	case orbittemplate.DriftKindRuntimeFile:
		return fmt.Sprintf("install-backed member %q runtime file drift detected", orbitID)
	case orbittemplate.DriftKindProvenanceUnresolvable:
		return fmt.Sprintf("install-backed member %q provenance could not be reconstructed", orbitID)
	default:
		return fmt.Sprintf("install-backed member %q drift detected", orbitID)
	}
}
