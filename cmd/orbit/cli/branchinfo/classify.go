package branchinfo

import (
	"context"
	"fmt"
	"strings"

	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	harnesspkg "github.com/zack-nova/orbit/cmd/orbit/cli/harness"
)

const manifestRelativePath = ".harness/manifest.yaml"

// Kind is the stable classifier output for a revision.
type Kind string

const (
	KindTemplate Kind = "template"
	KindSource   Kind = "source"
	KindRuntime  Kind = "runtime"
	KindPlain    Kind = "plain"
)

// Classification captures the branch kind and a short human-readable reason.
type Classification struct {
	Kind         Kind   `json:"kind"`
	TemplateKind string `json:"template_kind,omitempty"`
	Reason       string `json:"reason"`
}

const (
	TemplateKindOrbit   = "orbit"
	TemplateKindHarness = "harness"
)

// ClassifyRevision classifies a revision using the single manifest contract, not branch naming.
func ClassifyRevision(ctx context.Context, repoRoot string, rev string) (Classification, error) {
	trimmedRevision := strings.TrimSpace(rev)
	if trimmedRevision == "" {
		return Classification{}, fmt.Errorf("revision must not be empty")
	}

	exists, err := gitpkg.RevisionExists(ctx, repoRoot, trimmedRevision)
	if err != nil {
		return Classification{}, fmt.Errorf("check revision %q: %w", trimmedRevision, err)
	}
	if !exists {
		return Classification{}, fmt.Errorf("revision %q not found", trimmedRevision)
	}

	manifest, reason, err := loadManifestClassificationInput(ctx, repoRoot, trimmedRevision)
	if err != nil {
		return Classification{}, err
	}
	if manifest == nil {
		return Classification{
			Kind:   KindPlain,
			Reason: reason,
		}, nil
	}

	switch manifest.Kind {
	case harnesspkg.ManifestKindSource:
		return Classification{
			Kind:   KindSource,
			Reason: "valid .harness/manifest.yaml present with kind=source",
		}, nil
	case harnesspkg.ManifestKindRuntime:
		return Classification{
			Kind:   KindRuntime,
			Reason: "valid .harness/manifest.yaml present with kind=runtime",
		}, nil
	case harnesspkg.ManifestKindOrbitTemplate:
		return Classification{
			Kind:         KindTemplate,
			TemplateKind: TemplateKindOrbit,
			Reason:       "valid .harness/manifest.yaml present with kind=orbit_template",
		}, nil
	case harnesspkg.ManifestKindHarnessTemplate:
		return Classification{
			Kind:         KindTemplate,
			TemplateKind: TemplateKindHarness,
			Reason:       "valid .harness/manifest.yaml present with kind=harness_template",
		}, nil
	default:
		return Classification{
			Kind:   KindPlain,
			Reason: fmt.Sprintf("invalid %s: unsupported kind %q", manifestRelativePath, manifest.Kind),
		}, nil
	}
}

func loadManifestClassificationInput(
	ctx context.Context,
	repoRoot string,
	rev string,
) (*harnesspkg.ManifestFile, string, error) {
	exists, err := gitpkg.PathExistsAtRev(ctx, repoRoot, rev, manifestRelativePath)
	if err != nil {
		return nil, "", fmt.Errorf("check %s at %s: %w", manifestRelativePath, rev, err)
	}
	if !exists {
		return nil, "no valid .harness/manifest.yaml found", nil
	}

	data, err := gitpkg.ReadFileAtRev(ctx, repoRoot, rev, manifestRelativePath)
	if err != nil {
		return nil, "", fmt.Errorf("read %s at %s: %w", manifestRelativePath, rev, err)
	}

	manifest, err := harnesspkg.ParseManifestFileData(data)
	if err != nil {
		return nil, fmt.Sprintf("invalid %s: %v", manifestRelativePath, err), nil
	}

	return &manifest, "", nil
}
