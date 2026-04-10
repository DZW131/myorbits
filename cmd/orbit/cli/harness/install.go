package harness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zack-nova/orbit/cmd/orbit/cli/ids"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

// LoadInstallRecord reads, decodes, and validates one install record from .harness/installs/.
func LoadInstallRecord(repoRoot string, orbitID string) (orbittemplate.InstallRecord, error) {
	filename, err := InstallRecordPath(repoRoot, orbitID)
	if err != nil {
		return orbittemplate.InstallRecord{}, fmt.Errorf("build install record path: %w", err)
	}

	record, err := orbittemplate.LoadInstallRecordFile(filename)
	if err != nil {
		return orbittemplate.InstallRecord{}, fmt.Errorf("load harness install record: %w", err)
	}
	if record.OrbitID != orbitID {
		return orbittemplate.InstallRecord{}, fmt.Errorf("validate %s: orbit_id must match install path", filename)
	}

	return record, nil
}

// WriteInstallRecord validates and writes one install record into .harness/installs/.
func WriteInstallRecord(repoRoot string, record orbittemplate.InstallRecord) (string, error) {
	filename, err := InstallRecordPath(repoRoot, record.OrbitID)
	if err != nil {
		return "", fmt.Errorf("build install record path: %w", err)
	}

	written, err := orbittemplate.WriteInstallRecordFile(filename, record)
	if err != nil {
		return "", fmt.Errorf("write harness install record: %w", err)
	}

	return written, nil
}

// ListInstallRecordIDs lists valid install-record ids from .harness/installs/ with stable ordering.
func ListInstallRecordIDs(repoRoot string) ([]string, error) {
	entries, err := os.ReadDir(InstallRecordsDirPath(repoRoot))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read harness install records directory: %w", err)
	}

	idsList := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		orbitID := strings.TrimSuffix(entry.Name(), ".yaml")
		if err := ids.ValidateOrbitID(orbitID); err != nil {
			return nil, fmt.Errorf("validate install record filename %q: %w", entry.Name(), err)
		}

		idsList = append(idsList, orbitID)
	}

	sort.Strings(idsList)

	return idsList, nil
}
