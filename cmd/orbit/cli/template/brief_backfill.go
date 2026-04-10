package orbittemplate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
)

const (
	runtimeAgentsRepoPath = "AGENTS.md"
	briefBackfillVarsPath = ".harness/vars.yaml"
)

// BriefBackfillInput captures the repo and orbit targeted by one explicit brief backfill.
type BriefBackfillInput struct {
	RepoRoot string
	OrbitID  string
}

// BriefBackfillResult reports the hosted definition updated by one successful backfill.
type BriefBackfillResult struct {
	OrbitID        string
	DefinitionPath string
	Replacements   []ReplacementSummary
}

type reverseReplacementOccurrence struct {
	Variable string
	Literal  string
	Start    int
	End      int
}

// BackfillOrbitBrief extracts one current revision root AGENTS block, reverse-variableizes it,
// and writes the result back into meta.agents_template.
func BackfillOrbitBrief(ctx context.Context, input BriefBackfillInput) (BriefBackfillResult, error) {
	if strings.TrimSpace(input.RepoRoot) == "" {
		return BriefBackfillResult{}, errors.New("repo root must not be empty")
	}
	if err := ensureBriefBackfillRevisionAllowed(input.RepoRoot); err != nil {
		return BriefBackfillResult{}, err
	}

	runtimeData, err := os.ReadFile(filepath.Join(input.RepoRoot, filepath.FromSlash(runtimeAgentsRepoPath)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return BriefBackfillResult{}, fmt.Errorf("root AGENTS.md is missing")
		}
		return BriefBackfillResult{}, fmt.Errorf("read root AGENTS.md: %w", err)
	}

	document, err := ParseRuntimeAgentsDocument(runtimeData)
	if err != nil {
		return BriefBackfillResult{}, fmt.Errorf("parse root AGENTS.md: %w", err)
	}

	payload, err := extractRuntimeAgentsBlock(document, input.OrbitID)
	if err != nil {
		return BriefBackfillResult{}, err
	}

	runtimeBindings, err := loadOptionalRuntimeVars(input.RepoRoot)
	if err != nil {
		return BriefBackfillResult{}, err
	}

	replaced, err := ReverseVariableizeBrief(payload, runtimeBindings)
	if err != nil {
		return BriefBackfillResult{}, err
	}

	spec, err := orbitpkg.LoadHostedOrbitSpec(ctx, input.RepoRoot, input.OrbitID)
	if err != nil {
		return BriefBackfillResult{}, fmt.Errorf("load hosted orbit spec: %w", err)
	}
	if !spec.HasMemberSchema() || spec.Meta == nil {
		return BriefBackfillResult{}, fmt.Errorf("hosted orbit %q must use member schema before brief backfill", input.OrbitID)
	}

	spec.Meta.AgentsTemplate = string(replaced.Content)

	definitionPath, err := orbitpkg.WriteHostedOrbitSpec(input.RepoRoot, spec)
	if err != nil {
		return BriefBackfillResult{}, fmt.Errorf("write hosted orbit spec: %w", err)
	}

	return BriefBackfillResult{
		OrbitID:        input.OrbitID,
		DefinitionPath: definitionPath,
		Replacements:   replaced.Replacements,
	}, nil
}

// ReverseVariableizeBrief converts runtime literals back into variable placeholders.
func ReverseVariableizeBrief(content []byte, variables map[string]bindings.VariableBinding) (ReplacementResult, error) {
	result := ReplacementResult{
		Content: append([]byte(nil), content...),
	}

	entries, ambiguities, err := buildReplacementPlan(variables)
	if err != nil {
		return ReplacementResult{}, fmt.Errorf("build reverse replacement plan: %w", err)
	}
	if len(ambiguities) > 0 {
		return ReplacementResult{}, fmt.Errorf("reverse replacement is ambiguous: %s", formatReplacementAmbiguities(ambiguities))
	}

	if left, right, ok := findOverlappingReplacementEntries(string(content), entries); ok {
		return ReplacementResult{}, fmt.Errorf(
			"reverse replacement is ambiguous: overlapping runtime values %q (%s) and %q (%s)",
			left.Literal,
			left.Variable,
			right.Literal,
			right.Variable,
		)
	}

	text := string(content)
	summaries := make([]ReplacementSummary, 0, len(entries))
	for _, entry := range entries {
		count := strings.Count(text, entry.Literal)
		if count == 0 {
			continue
		}

		text = strings.ReplaceAll(text, entry.Literal, "$"+entry.Variable)
		summaries = append(summaries, ReplacementSummary{
			Variable: entry.Variable,
			Literal:  entry.Literal,
			Count:    count,
		})
	}

	sort.Slice(summaries, func(left, right int) bool {
		if summaries[left].Count == summaries[right].Count {
			return summaries[left].Variable < summaries[right].Variable
		}
		return summaries[left].Count > summaries[right].Count
	})

	result.Content = []byte(text)
	result.Replacements = summaries

	return result, nil
}

func ensureBriefBackfillRevisionAllowed(repoRoot string) error {
	revisionKind, err := resolveBriefRevisionKind(repoRoot)
	if err != nil {
		return fmt.Errorf("load current revision manifest: %w", err)
	}

	return validateBriefRevisionKindAllowed(revisionKind, "backfill")
}

func ensureBriefMaterializeRevisionAllowed(repoRoot string) error {
	revisionKind, err := resolveBriefRevisionKind(repoRoot)
	if err != nil {
		return fmt.Errorf("load current revision manifest: %w", err)
	}

	return validateBriefRevisionKindAllowed(revisionKind, "materialize")
}

func resolveBriefRevisionKind(repoRoot string) (string, error) {
	revisionKind, err := loadCurrentRevisionManifestKind(repoRoot)
	if errors.Is(err, os.ErrNotExist) {
		return "plain", nil
	}
	if err != nil {
		return "", err
	}

	return revisionKind, nil
}

func validateBriefRevisionKindAllowed(revisionKind string, operation string) error {
	switch revisionKind {
	case "", "plain":
		return fmt.Errorf(`brief %s supports only runtime, source, or orbit_template revisions; current revision kind is "plain"`, operation)
	case "runtime", "source", "orbit_template":
		return nil
	default:
		return fmt.Errorf("brief %s supports only runtime, source, or orbit_template revisions; current revision kind is %q", operation, revisionKind)
	}
}

func extractRuntimeAgentsBlock(document AgentsRuntimeDocument, orbitID string) ([]byte, error) {
	for _, segment := range document.Segments {
		if segment.Kind != AgentsRuntimeSegmentBlock {
			continue
		}
		if segment.OrbitID != orbitID {
			continue
		}

		return append([]byte(nil), segment.Content...), nil
	}

	return nil, fmt.Errorf("root AGENTS.md does not contain orbit block %q", orbitID)
}

func loadOptionalRuntimeVars(repoRoot string) (map[string]bindings.VariableBinding, error) {
	filename := filepath.Join(repoRoot, filepath.FromSlash(briefBackfillVarsPath))
	file, err := bindings.LoadVarsFileAtPath(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]bindings.VariableBinding{}, nil
		}
		return nil, fmt.Errorf("load .harness/vars.yaml: %w", err)
	}

	return file.Variables, nil
}

func formatReplacementAmbiguities(ambiguities []ReplacementAmbiguity) string {
	parts := make([]string, 0, len(ambiguities))
	for _, ambiguity := range ambiguities {
		parts = append(parts, fmt.Sprintf("%q => %s", ambiguity.Literal, strings.Join(ambiguity.Variables, ", ")))
	}
	sort.Strings(parts)

	return strings.Join(parts, "; ")
}

func findOverlappingReplacementEntries(content string, entries []replacementEntry) (reverseReplacementOccurrence, reverseReplacementOccurrence, bool) {
	occurrences := make([]reverseReplacementOccurrence, 0)
	for _, entry := range entries {
		start := 0
		for {
			offset := strings.Index(content[start:], entry.Literal)
			if offset < 0 {
				break
			}

			matchStart := start + offset
			occurrences = append(occurrences, reverseReplacementOccurrence{
				Variable: entry.Variable,
				Literal:  entry.Literal,
				Start:    matchStart,
				End:      matchStart + len(entry.Literal),
			})
			start = matchStart + 1
		}
	}

	sort.Slice(occurrences, func(left, right int) bool {
		if occurrences[left].Start == occurrences[right].Start {
			if occurrences[left].End == occurrences[right].End {
				return occurrences[left].Variable < occurrences[right].Variable
			}
			return occurrences[left].End < occurrences[right].End
		}
		return occurrences[left].Start < occurrences[right].Start
	})

	for left := 0; left < len(occurrences); left++ {
		for right := left + 1; right < len(occurrences); right++ {
			if occurrences[right].Start >= occurrences[left].End {
				break
			}
			if occurrences[left].Literal == occurrences[right].Literal {
				continue
			}

			return occurrences[left], occurrences[right], true
		}
	}

	return reverseReplacementOccurrence{}, reverseReplacementOccurrence{}, false
}
