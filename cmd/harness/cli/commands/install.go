package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/zack-nova/orbit/cmd/orbit/cli/bindings"
	gitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/git"
	harnesspkg "github.com/zack-nova/orbit/cmd/orbit/cli/harness"
	orbitpkg "github.com/zack-nova/orbit/cmd/orbit/cli/orbit"
	orbittemplate "github.com/zack-nova/orbit/cmd/orbit/cli/template"
)

type installSourceJSON struct {
	Kind           string `json:"kind"`
	Repo           string `json:"repo,omitempty"`
	Ref            string `json:"ref"`
	RequestedRef   string `json:"requested_ref,omitempty"`
	ResolvedRef    string `json:"resolved_ref,omitempty"`
	ResolutionKind string `json:"resolution_kind,omitempty"`
	Commit         string `json:"commit"`
}

type installBindingJSON struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

type installPreviewJSON struct {
	DryRun            bool                          `json:"dry_run"`
	HarnessRoot       string                        `json:"harness_root"`
	TemplateKind      string                        `json:"template_kind,omitempty"`
	OverwriteExisting bool                          `json:"overwrite_existing"`
	Source            installSourceJSON             `json:"source"`
	OrbitID           string                        `json:"orbit_id"`
	HarnessID         string                        `json:"harness_id,omitempty"`
	MemberIDs         []string                      `json:"member_ids,omitempty"`
	Bindings          []installBindingJSON          `json:"bindings"`
	Files             []string                      `json:"files"`
	Warnings          []string                      `json:"warnings,omitempty"`
	Conflicts         []orbittemplate.ApplyConflict `json:"conflicts"`
}

type installResultJSON struct {
	DryRun       bool              `json:"dry_run"`
	HarnessRoot  string            `json:"harness_root"`
	Source       installSourceJSON `json:"source"`
	OrbitID      string            `json:"orbit_id"`
	WrittenPaths []string          `json:"written_paths"`
	Warnings     []string          `json:"warnings,omitempty"`
	MemberCount  int               `json:"member_count"`
}

type harnessTemplateInstallResultJSON struct {
	DryRun       bool              `json:"dry_run"`
	HarnessRoot  string            `json:"harness_root"`
	TemplateKind string            `json:"template_kind"`
	Source       installSourceJSON `json:"source"`
	HarnessID    string            `json:"harness_id"`
	MemberIDs    []string          `json:"member_ids"`
	WrittenPaths []string          `json:"written_paths"`
	Warnings     []string          `json:"warnings,omitempty"`
	MemberCount  int               `json:"member_count"`
	BundleCount  int               `json:"bundle_count"`
}

type installTargetState struct {
	Member            *harnesspkg.RuntimeMember
	HasDefinition     bool
	HasInstallRecord  bool
	ExistingRecord    orbittemplate.InstallRecord
	RequiresOverwrite bool
}

const (
	installTemplateKindOrbit   = "orbit_template"
	installTemplateKindHarness = "harness_template"
)

// NewInstallCommand creates the harness install command.
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <template-branch|git-url>",
		Short: "Install one local or remote orbit template into the current harness runtime",
		Long: "Install one local or remote orbit template into the current harness runtime.\n" +
			"By default this command writes the runtime immediately; use --dry-run to preview without mutating the repository.",
		Example: "" +
			"  harness install orbit-template/docs --bindings .harness/vars.yaml\n" +
			"  harness install https://example.com/acme/templates.git --ref orbit-template/docs --bindings .harness/vars.yaml\n" +
			"  harness install orbit-template/docs --overwrite-existing --bindings .harness/vars.yaml --json\n",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath, err := pathFromCommand(cmd)
			if err != nil {
				return err
			}

			resolved, err := harnesspkg.ResolveRoot(cmd.Context(), targetPath)
			if err != nil {
				return fmt.Errorf("resolve harness root: %w", err)
			}

			bindingsPath, err := cmd.Flags().GetString("bindings")
			if err != nil {
				return fmt.Errorf("read --bindings flag: %w", err)
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return fmt.Errorf("read --dry-run flag: %w", err)
			}
			overwriteExisting, err := cmd.Flags().GetBool("overwrite-existing")
			if err != nil {
				return fmt.Errorf("read --overwrite-existing flag: %w", err)
			}
			interactive, err := cmd.Flags().GetBool("interactive")
			if err != nil {
				return fmt.Errorf("read --interactive flag: %w", err)
			}
			editorMode, err := cmd.Flags().GetBool("editor")
			if err != nil {
				return fmt.Errorf("read --editor flag: %w", err)
			}
			requestedRef, err := cmd.Flags().GetString("ref")
			if err != nil {
				return fmt.Errorf("read --ref flag: %w", err)
			}
			jsonOutput, err := wantJSON(cmd)
			if err != nil {
				return err
			}
			progressMode, err := cmd.Flags().GetString("progress")
			if err != nil {
				return fmt.Errorf("read --progress flag: %w", err)
			}
			progress, err := newInstallProgressEmitter(cmd.ErrOrStderr(), progressMode)
			if err != nil {
				return err
			}

			sourceArg := args[0]
			prompter := buildInstallPrompter(cmd, interactive)
			var editor orbittemplate.Editor
			if editorMode {
				editor, err = orbittemplate.NewEnvironmentEditor()
				if err != nil {
					return fmt.Errorf("configure bindings editor: %w", err)
				}
			}

			if err := progress.Stage("resolving install source"); err != nil {
				return err
			}
			localSource, err := installSourceUsesLocalRevision(cmd.Context(), resolved.Repo.Root, sourceArg)
			if err != nil {
				return err
			}

			now := time.Now().UTC()
			if localSource {
				if strings.TrimSpace(requestedRef) != "" {
					return fmt.Errorf("--ref is only supported when installing from a remote Git URL")
				}

				harnessSource, err := harnesspkg.ResolveLocalTemplateInstallSource(cmd.Context(), resolved.Repo.Root, sourceArg)
				if err == nil {
					installSource := orbittemplate.Source{
						SourceKind:     orbittemplate.InstallSourceKindLocalBranch,
						SourceRepo:     "",
						SourceRef:      harnessSource.Ref,
						TemplateCommit: harnessSource.Commit,
					}
					if err := progress.Stage("resolving bindings"); err != nil {
						return err
					}
					if dryRun {
						preview, err := harnesspkg.BuildTemplateInstallPreview(cmd.Context(), harnesspkg.TemplateInstallPreviewInput{
							RepoRoot:          resolved.Repo.Root,
							Source:            harnessSource,
							InstallSource:     installSource,
							BindingsFilePath:  bindingsPath,
							OverwriteExisting: overwriteExisting,
							Interactive:       interactive,
							Prompter:          prompter,
							EditorMode:        editorMode,
							Editor:            editor,
							Now:               now,
						})
						if err != nil {
							return fmt.Errorf("build harness template install preview: %w", err)
						}
						if err := progress.Stage("checking conflicts"); err != nil {
							return err
						}
						return emitHarnessTemplateInstallPreview(
							cmd,
							resolved.Repo.Root,
							installSourceJSON{
								Kind:   orbittemplate.InstallSourceKindLocalBranch,
								Ref:    harnessSource.Ref,
								Commit: harnessSource.Commit,
							},
							preview,
							jsonOutput,
						)
					}
					preview, err := harnesspkg.BuildTemplateInstallPreview(cmd.Context(), harnesspkg.TemplateInstallPreviewInput{
						RepoRoot:                resolved.Repo.Root,
						Source:                  harnessSource,
						InstallSource:           installSource,
						BindingsFilePath:        bindingsPath,
						OverwriteExisting:       overwriteExisting,
						Interactive:             interactive,
						Prompter:                prompter,
						EditorMode:              editorMode,
						Editor:                  editor,
						RequireResolvedBindings: true,
						Now:                     now,
					})
					if err != nil {
						return fmt.Errorf("build harness template install preview: %w", err)
					}
					if err := progress.Stage("checking conflicts"); err != nil {
						return err
					}
					if err := progress.Stage("writing files"); err != nil {
						return err
					}
					result, err := harnesspkg.ApplyTemplateInstallPreview(cmd.Context(), resolved.Repo.Root, preview, overwriteExisting)
					if err != nil {
						return fmt.Errorf("install harness template: %w", err)
					}
					if err := progress.Stage("updating runtime metadata"); err != nil {
						return err
					}
					if err := progress.Stage("install complete"); err != nil {
						return err
					}
					return emitHarnessTemplateInstallResult(cmd, resolved.Repo.Root, result, jsonOutput)
				}
				var notHarnessTemplateErr *harnesspkg.LocalTemplateInstallSourceNotFoundError
				if !errors.As(err, &notHarnessTemplateErr) {
					return fmt.Errorf("resolve local harness template source: %w", err)
				}

				previewInput := orbittemplate.TemplateApplyPreviewInput{
					RepoRoot:         resolved.Repo.Root,
					SourceRef:        sourceArg,
					BindingsFilePath: bindingsPath,
					Interactive:      interactive,
					Prompter:         prompter,
					EditorMode:       editorMode,
					Editor:           editor,
					Now:              now,
				}
				if err := progress.Stage("resolving bindings"); err != nil {
					return err
				}
				preview, err := orbittemplate.BuildTemplateApplyPreview(cmd.Context(), previewInput)
				if err != nil {
					return fmt.Errorf("build harness install preview: %w", err)
				}
				if err := progress.Stage("checking conflicts"); err != nil {
					return err
				}
				targetState, err := inspectInstallTargetState(resolved.Repo.Root, resolved.Runtime, preview.Source.Manifest.Template.OrbitID)
				if err != nil {
					return err
				}
				if err := validateInstallTargetState(targetState, preview.Source.Manifest.Template.OrbitID, overwriteExisting); err != nil {
					return err
				}
				var cleanupPlan orbittemplate.InstallOwnedCleanupPlan
				if targetState.RequiresOverwrite {
					cleanupPlan, err = orbittemplate.BuildInstallOwnedCleanupPlan(cmd.Context(), resolved.Repo.Root, targetState.ExistingRecord, preview)
					if err != nil {
						return fmt.Errorf("reconstruct existing install ownership: %w", err)
					}
				}
				if dryRun {
					if err := progress.Stage("install complete"); err != nil {
						return err
					}
					return emitInstallPreview(cmd, resolved.Repo.Root, preview, jsonOutput, overwriteExisting)
				}

				previewInput.OverwriteExisting = overwriteExisting
				if err := progress.Stage("writing files"); err != nil {
					return err
				}
				result, err := orbittemplate.ApplyLocalTemplate(cmd.Context(), orbittemplate.TemplateApplyInput{Preview: previewInput})
				if err != nil {
					return fmt.Errorf("install local template: %w", err)
				}
				if targetState.RequiresOverwrite {
					removedPaths, err := orbittemplate.ApplyInstallOwnedCleanup(resolved.Repo.Root, result.Preview.Source.Manifest.Template.OrbitID, cleanupPlan)
					if err != nil {
						return fmt.Errorf("remove stale install-owned paths: %w", err)
					}
					result.WrittenPaths = append(result.WrittenPaths, removedPaths...)
				}
				if err := progress.Stage("updating runtime metadata"); err != nil {
					return err
				}
				memberResult, err := upsertInstallMemberForState(cmd.Context(), resolved.Repo.Root, result.Preview.Source.Manifest.Template.OrbitID, now, targetState)
				if err != nil {
					return fmt.Errorf("record install-backed member: %w", err)
				}
				if err := progress.Stage("install complete"); err != nil {
					return err
				}
				return emitInstallResult(cmd, resolved.Repo.Root, result, memberResult.ManifestPath, memberResult.Runtime, jsonOutput)
			}

			previewInput := orbittemplate.RemoteTemplateApplyPreviewInput{
				RepoRoot:         resolved.Repo.Root,
				RemoteURL:        sourceArg,
				RequestedRef:     requestedRef,
				BindingsFilePath: bindingsPath,
				Interactive:      interactive,
				Prompter:         prompter,
				EditorMode:       editorMode,
				Editor:           editor,
				Now:              now,
			}
			if strings.TrimSpace(requestedRef) == "" {
				if err := progress.Stage("resolving remote template candidates"); err != nil {
					return err
				}
			}
			if err := progress.Stage("fetching selected template"); err != nil {
				return err
			}
			if strings.TrimSpace(requestedRef) != "" {
				harnessCandidate, harnessSource, harnessErr := harnesspkg.ResolveRemoteTemplateInstallSource(
					cmd.Context(),
					resolved.Repo.Root,
					sourceArg,
					requestedRef,
				)
				if harnessErr == nil {
					installSource := orbittemplate.Source{
						SourceKind:     orbittemplate.InstallSourceKindRemoteGit,
						SourceRepo:     harnessCandidate.RepoURL,
						SourceRef:      harnessCandidate.Branch,
						TemplateCommit: harnessSource.Commit,
					}
					if err := progress.Stage("resolving bindings"); err != nil {
						return err
					}
					if dryRun {
						preview, err := harnesspkg.BuildTemplateInstallPreview(cmd.Context(), harnesspkg.TemplateInstallPreviewInput{
							RepoRoot:          resolved.Repo.Root,
							Source:            harnessSource,
							InstallSource:     installSource,
							BindingsFilePath:  bindingsPath,
							OverwriteExisting: overwriteExisting,
							Interactive:       interactive,
							Prompter:          prompter,
							EditorMode:        editorMode,
							Editor:            editor,
							Now:               now,
						})
						if err != nil {
							return fmt.Errorf("build harness template install preview: %w", err)
						}
						if err := progress.Stage("checking conflicts"); err != nil {
							return err
						}
						if err := progress.Stage("install complete"); err != nil {
							return err
						}
						return emitHarnessTemplateInstallPreview(
							cmd,
							resolved.Repo.Root,
							installSourceJSON{
								Kind:   orbittemplate.InstallSourceKindRemoteGit,
								Repo:   harnessCandidate.RepoURL,
								Ref:    harnessCandidate.Branch,
								Commit: harnessSource.Commit,
							},
							preview,
							jsonOutput,
						)
					}
					preview, err := harnesspkg.BuildTemplateInstallPreview(cmd.Context(), harnesspkg.TemplateInstallPreviewInput{
						RepoRoot:                resolved.Repo.Root,
						Source:                  harnessSource,
						InstallSource:           installSource,
						BindingsFilePath:        bindingsPath,
						OverwriteExisting:       overwriteExisting,
						Interactive:             interactive,
						Prompter:                prompter,
						EditorMode:              editorMode,
						Editor:                  editor,
						RequireResolvedBindings: true,
						Now:                     now,
					})
					if err != nil {
						return fmt.Errorf("build harness template install preview: %w", err)
					}
					if err := progress.Stage("checking conflicts"); err != nil {
						return err
					}
					if err := progress.Stage("writing files"); err != nil {
						return err
					}
					result, err := harnesspkg.ApplyTemplateInstallPreview(cmd.Context(), resolved.Repo.Root, preview, overwriteExisting)
					if err != nil {
						return fmt.Errorf("install harness template: %w", err)
					}
					if err := progress.Stage("updating runtime metadata"); err != nil {
						return err
					}
					if err := progress.Stage("install complete"); err != nil {
						return err
					}
					return emitHarnessTemplateInstallResult(cmd, resolved.Repo.Root, result, jsonOutput)
				}
				var notHarnessTemplateErr *harnesspkg.RemoteTemplateInstallNotFoundError
				if !errors.As(harnessErr, &notHarnessTemplateErr) {
					return fmt.Errorf("resolve remote harness template source: %w", harnessErr)
				}
			}

			preview, err := orbittemplate.BuildRemoteTemplateApplyPreview(cmd.Context(), previewInput)
			if err != nil {
				var notOrbitTemplateErr *orbittemplate.RemoteTemplateNotFoundError
				if strings.TrimSpace(requestedRef) == "" && errors.As(err, &notOrbitTemplateErr) {
					harnessCandidate, harnessSource, harnessErr := harnesspkg.ResolveRemoteTemplateInstallSource(
						cmd.Context(),
						resolved.Repo.Root,
						sourceArg,
						"",
					)
					if harnessErr == nil {
						installSource := orbittemplate.Source{
							SourceKind:     orbittemplate.InstallSourceKindRemoteGit,
							SourceRepo:     harnessCandidate.RepoURL,
							SourceRef:      harnessCandidate.Branch,
							TemplateCommit: harnessSource.Commit,
						}
						if dryRun {
							preview, err := harnesspkg.BuildTemplateInstallPreview(cmd.Context(), harnesspkg.TemplateInstallPreviewInput{
								RepoRoot:          resolved.Repo.Root,
								Source:            harnessSource,
								InstallSource:     installSource,
								BindingsFilePath:  bindingsPath,
								OverwriteExisting: overwriteExisting,
								Interactive:       interactive,
								Prompter:          prompter,
								EditorMode:        editorMode,
								Editor:            editor,
								Now:               now,
							})
							if err != nil {
								return fmt.Errorf("build harness template install preview: %w", err)
							}
							return emitHarnessTemplateInstallPreview(
								cmd,
								resolved.Repo.Root,
								installSourceJSON{
									Kind:   orbittemplate.InstallSourceKindRemoteGit,
									Repo:   harnessCandidate.RepoURL,
									Ref:    harnessCandidate.Branch,
									Commit: harnessSource.Commit,
								},
								preview,
								jsonOutput,
							)
						}
						preview, err := harnesspkg.BuildTemplateInstallPreview(cmd.Context(), harnesspkg.TemplateInstallPreviewInput{
							RepoRoot:                resolved.Repo.Root,
							Source:                  harnessSource,
							InstallSource:           installSource,
							BindingsFilePath:        bindingsPath,
							OverwriteExisting:       overwriteExisting,
							Interactive:             interactive,
							Prompter:                prompter,
							EditorMode:              editorMode,
							Editor:                  editor,
							RequireResolvedBindings: true,
							Now:                     now,
						})
						if err != nil {
							return fmt.Errorf("build harness template install preview: %w", err)
						}
						result, err := harnesspkg.ApplyTemplateInstallPreview(cmd.Context(), resolved.Repo.Root, preview, overwriteExisting)
						if err != nil {
							return fmt.Errorf("install harness template: %w", err)
						}
						return emitHarnessTemplateInstallResult(cmd, resolved.Repo.Root, result, jsonOutput)
					}
				}
				return fmt.Errorf("build harness install preview: %w", err)
			}
			if preview.RemoteResolutionKind == orbittemplate.RemoteTemplateResolutionSourceAlias {
				if err := progress.Stage("source branch detected; resolving published template"); err != nil {
					return err
				}
			}
			if err := progress.Stage("resolving bindings"); err != nil {
				return err
			}
			if err := progress.Stage("checking conflicts"); err != nil {
				return err
			}
			targetState, err := inspectInstallTargetState(resolved.Repo.Root, resolved.Runtime, preview.Source.Manifest.Template.OrbitID)
			if err != nil {
				return err
			}
			if err := validateInstallTargetState(targetState, preview.Source.Manifest.Template.OrbitID, overwriteExisting); err != nil {
				return err
			}
			var cleanupPlan orbittemplate.InstallOwnedCleanupPlan
			if targetState.RequiresOverwrite {
				cleanupPlan, err = orbittemplate.BuildInstallOwnedCleanupPlan(cmd.Context(), resolved.Repo.Root, targetState.ExistingRecord, preview)
				if err != nil {
					return fmt.Errorf("reconstruct existing install ownership: %w", err)
				}
			}
			if dryRun {
				if err := progress.Stage("install complete"); err != nil {
					return err
				}
				return emitInstallPreview(cmd, resolved.Repo.Root, preview, jsonOutput, overwriteExisting)
			}

			previewInput.OverwriteExisting = overwriteExisting
			if err := progress.Stage("writing files"); err != nil {
				return err
			}
			result, err := orbittemplate.ApplyRemoteTemplate(cmd.Context(), orbittemplate.RemoteTemplateApplyInput{Preview: previewInput})
			if err != nil {
				return fmt.Errorf("install remote template: %w", err)
			}
			if targetState.RequiresOverwrite {
				removedPaths, err := orbittemplate.ApplyInstallOwnedCleanup(resolved.Repo.Root, result.Preview.Source.Manifest.Template.OrbitID, cleanupPlan)
				if err != nil {
					return fmt.Errorf("remove stale install-owned paths: %w", err)
				}
				result.WrittenPaths = append(result.WrittenPaths, removedPaths...)
			}
			if err := progress.Stage("updating runtime metadata"); err != nil {
				return err
			}
			memberResult, err := upsertInstallMemberForState(cmd.Context(), resolved.Repo.Root, result.Preview.Source.Manifest.Template.OrbitID, now, targetState)
			if err != nil {
				return fmt.Errorf("record install-backed member: %w", err)
			}
			if err := progress.Stage("install complete"); err != nil {
				return err
			}
			return emitInstallResult(cmd, resolved.Repo.Root, result, memberResult.ManifestPath, memberResult.Runtime, jsonOutput)
		},
	}

	cmd.Flags().String("bindings", "", "Path to an explicit bindings YAML file")
	cmd.Flags().String("ref", "", "Select one remote template branch explicitly when installing from a remote Git URL")
	cmd.Flags().Bool("overwrite-existing", false, "Allow overwriting an existing install-backed orbit and removing stale install-owned files")
	cmd.Flags().Bool("dry-run", false, "Preview harness install without writing files")
	cmd.Flags().String("progress", string(installProgressAuto), "Progress output mode: auto, plain, or quiet")
	cmd.Flags().Bool("interactive", false, "Prompt for missing bindings interactively")
	cmd.Flags().Bool("editor", false, "Open an editor-backed bindings skeleton for missing required values")
	addPathFlag(cmd)
	addJSONFlag(cmd)

	return cmd
}

func installSourceUsesLocalRevision(ctx context.Context, repoRoot string, source string) (bool, error) {
	exists, err := gitpkg.RevisionExists(ctx, repoRoot, source)
	if err != nil {
		return false, fmt.Errorf("check local template source %q: %w", source, err)
	}

	return exists, nil
}

func inspectInstallTargetState(repoRoot string, runtimeFile harnesspkg.RuntimeFile, orbitID string) (installTargetState, error) {
	state := installTargetState{}
	for _, member := range runtimeFile.Members {
		if member.OrbitID != orbitID {
			continue
		}
		memberCopy := member
		state.Member = &memberCopy
		break
	}

	definitionPath, err := orbitpkg.HostedDefinitionPath(repoRoot, orbitID)
	if err != nil {
		return installTargetState{}, fmt.Errorf("build orbit definition path: %w", err)
	}
	if _, err := os.Stat(definitionPath); err == nil {
		state.HasDefinition = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return installTargetState{}, fmt.Errorf("stat %s: %w", definitionPath, err)
	}

	installPath, err := harnesspkg.InstallRecordPath(repoRoot, orbitID)
	if err != nil {
		return installTargetState{}, fmt.Errorf("build install record path: %w", err)
	}
	if _, err := os.Stat(installPath); err == nil {
		record, loadErr := harnesspkg.LoadInstallRecord(repoRoot, orbitID)
		if loadErr != nil {
			return installTargetState{}, fmt.Errorf("load existing install record for orbit %q: %w", orbitID, loadErr)
		}
		state.HasInstallRecord = true
		state.ExistingRecord = record
	} else if !errors.Is(err, os.ErrNotExist) {
		return installTargetState{}, fmt.Errorf("stat %s: %w", installPath, err)
	}

	state.RequiresOverwrite = state.HasInstallRecord

	return state, nil
}

func validateInstallTargetState(state installTargetState, orbitID string, overwriteExisting bool) error {
	if state.Member != nil {
		switch state.Member.Source {
		case harnesspkg.MemberSourceManual:
			return fmt.Errorf("orbit %q is already present as a manual member", orbitID)
		case harnesspkg.MemberSourceInstallOrbit:
			if !overwriteExisting {
				return fmt.Errorf("orbit %q is already installed; repeated install requires --overwrite-existing", orbitID)
			}
		default:
			return fmt.Errorf("orbit %q already exists in harness runtime", orbitID)
		}
	}

	if state.HasInstallRecord {
		if !overwriteExisting {
			return fmt.Errorf("orbit %q is already installed; repeated install requires --overwrite-existing", orbitID)
		}
		return nil
	}

	if state.HasDefinition {
		return fmt.Errorf("orbit definition %q already exists in the runtime repository", orbitID)
	}

	return nil
}

func upsertInstallMemberForState(
	ctx context.Context,
	repoRoot string,
	orbitID string,
	now time.Time,
	state installTargetState,
) (harnesspkg.MutateMembersResult, error) {
	if state.RequiresOverwrite {
		result, err := harnesspkg.UpsertInstallMember(ctx, repoRoot, orbitID, now)
		if err != nil {
			return harnesspkg.MutateMembersResult{}, fmt.Errorf("upsert install-backed member: %w", err)
		}
		return result, nil
	}

	result, err := harnesspkg.AddInstallMember(ctx, repoRoot, orbitID, now)
	if err != nil {
		return harnesspkg.MutateMembersResult{}, fmt.Errorf("add install-backed member: %w", err)
	}

	return result, nil
}

func buildInstallPrompter(cmd *cobra.Command, interactive bool) orbittemplate.BindingPrompter {
	if !interactive {
		return nil
	}

	return orbittemplate.LineBindingPrompter{
		Reader: cmd.InOrStdin(),
		Writer: cmd.ErrOrStderr(),
	}
}

func emitInstallPreview(cmd *cobra.Command, harnessRoot string, preview orbittemplate.TemplateApplyPreview, jsonOutput bool, overwriteExisting bool) error {
	if jsonOutput {
		return emitJSON(cmd.OutOrStdout(), installPreviewJSON{
			DryRun:            true,
			HarnessRoot:       harnessRoot,
			TemplateKind:      installTemplateKindOrbit,
			OverwriteExisting: overwriteExisting,
			Source:            installSourcePayload(preview),
			OrbitID:           preview.Source.Manifest.Template.OrbitID,
			Bindings:          installBindingsPayload(preview.ResolvedBindings),
			Files:             installPreviewPaths(preview),
			Warnings:          append([]string(nil), preview.Warnings...),
			Conflicts:         preview.Conflicts,
		})
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "harness install dry-run from %s\n", preview.Source.Ref); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_ref: %s\n", preview.InstallRecord.Template.SourceRef); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if strings.TrimSpace(preview.RemoteRequestedRef) != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "requested_ref: %s\n", preview.RemoteRequestedRef); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if preview.RemoteResolutionKind != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "resolved_ref: %s\n", preview.InstallRecord.Template.SourceRef); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "resolution_kind: %s\n", preview.RemoteResolutionKind); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_kind: %s\n", preview.InstallRecord.Template.SourceKind); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if strings.TrimSpace(preview.InstallRecord.Template.SourceRepo) != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_repo: %s\n", preview.InstallRecord.Template.SourceRepo); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_commit: %s\n", preview.Source.Commit); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "orbit_id: %s\n", preview.Source.Manifest.Template.OrbitID); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if err := emitInstallWarnings(cmd, preview.Warnings); err != nil {
		return err
	}

	if len(preview.ResolvedBindings) == 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "bindings: none"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	} else {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "bindings:"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
		for _, item := range installBindingsPayload(preview.ResolvedBindings) {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s <- %s\n", item.Name, item.Source); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
		}
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "files:"); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	for _, path := range installPreviewPaths(preview) {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), path); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}

	if len(preview.Conflicts) == 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "conflicts: none"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "conflicts:"); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	for _, conflict := range preview.Conflicts {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", conflict.Path, conflict.Message); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}

	return nil
}

func emitHarnessTemplateInstallPreview(
	cmd *cobra.Command,
	harnessRoot string,
	source installSourceJSON,
	preview harnesspkg.TemplateInstallPreview,
	jsonOutput bool,
) error {
	files := harnessTemplateInstallPreviewPaths(preview)
	memberIDs := preview.Source.MemberIDs()

	if jsonOutput {
		return emitJSON(cmd.OutOrStdout(), installPreviewJSON{
			DryRun:       true,
			HarnessRoot:  harnessRoot,
			TemplateKind: installTemplateKindHarness,
			Source:       source,
			HarnessID:    preview.Source.Manifest.Template.HarnessID,
			MemberIDs:    memberIDs,
			Bindings:     installBindingsPayload(preview.ResolvedBindings),
			Files:        files,
			Warnings:     append([]string(nil), preview.Warnings...),
			Conflicts:    append([]orbittemplate.ApplyConflict(nil), preview.Conflicts...),
		})
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "harness install dry-run from %s\n", source.Ref); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_ref: %s\n", source.Ref); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_kind: %s\n", source.Kind); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if strings.TrimSpace(source.Repo) != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_repo: %s\n", source.Repo); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_commit: %s\n", source.Commit); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "template_kind: %s\n", installTemplateKindHarness); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "harness_id: %s\n", preview.Source.Manifest.Template.HarnessID); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "member_ids:"); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	for _, memberID := range memberIDs {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), memberID); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if len(preview.ResolvedBindings) == 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "bindings: none"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	} else {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "bindings:"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
		for _, item := range installBindingsPayload(preview.ResolvedBindings) {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s <- %s\n", item.Name, item.Source); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
		}
	}
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "files:"); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	for _, path := range files {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), path); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if len(preview.Conflicts) == 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "conflicts: none"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	} else {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "conflicts:"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
		for _, conflict := range preview.Conflicts {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", conflict.Path, conflict.Message); err != nil {
				return fmt.Errorf("write command output: %w", err)
			}
		}
	}
	if err := emitInstallWarnings(cmd, preview.Warnings); err != nil {
		return err
	}

	return nil
}

func harnessTemplateInstallPreviewPaths(preview harnesspkg.TemplateInstallPreview) []string {
	paths := make([]string, 0, len(preview.RenderedDefinitionFiles)+len(preview.RenderedFiles)+4)
	for _, file := range preview.RenderedDefinitionFiles {
		paths = append(paths, file.Path)
	}
	for _, file := range preview.RenderedFiles {
		paths = append(paths, file.Path)
	}
	if preview.RenderedRootAgentsFile != nil {
		paths = append(paths, preview.RenderedRootAgentsFile.Path)
	}
	paths = append(paths,
		fmt.Sprintf(".harness/bundles/%s.yaml", preview.Source.Manifest.Template.HarnessID),
		harnesspkg.ManifestRepoPath(),
	)
	if preview.VarsFile != nil {
		paths = append(paths, ".harness/vars.yaml")
	}

	sort.Strings(paths)

	return paths
}

func emitHarnessTemplateInstallResult(
	cmd *cobra.Command,
	harnessRoot string,
	result harnesspkg.TemplateInstallResult,
	jsonOutput bool,
) error {
	memberIDs := result.Preview.Source.MemberIDs()
	bundleIDs, err := harnesspkg.ListBundleRecordIDs(harnessRoot)
	if err != nil {
		return fmt.Errorf("list harness bundle records: %w", err)
	}
	bundleCount := len(bundleIDs)

	if jsonOutput {
		return emitJSON(cmd.OutOrStdout(), harnessTemplateInstallResultJSON{
			DryRun:       false,
			HarnessRoot:  harnessRoot,
			TemplateKind: installTemplateKindHarness,
			Source: installSourceJSON{
				Kind:   result.Preview.InstallSource.SourceKind,
				Repo:   result.Preview.InstallSource.SourceRepo,
				Ref:    result.Preview.InstallSource.SourceRef,
				Commit: result.Preview.Source.Commit,
			},
			HarnessID:    result.Preview.Source.Manifest.Template.HarnessID,
			MemberIDs:    memberIDs,
			WrittenPaths: append([]string(nil), result.WrittenPaths...),
			Warnings:     append([]string(nil), result.Preview.Warnings...),
			MemberCount:  len(result.Runtime.Members),
			BundleCount:  bundleCount,
		})
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "installed harness template %s into harness %s\n", result.Preview.Source.Manifest.Template.HarnessID, harnessRoot); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_ref: %s\n", result.Preview.InstallSource.SourceRef); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "member_count: %d\n", len(result.Runtime.Members)); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "bundle_count: %d\n", bundleCount); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "files: %d\n", len(result.WrittenPaths)); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if err := emitInstallWarnings(cmd, result.Preview.Warnings); err != nil {
		return err
	}

	return nil
}

func emitInstallResult(cmd *cobra.Command, harnessRoot string, result orbittemplate.TemplateApplyResult, manifestPath string, runtimeFile harnesspkg.RuntimeFile, jsonOutput bool) error {
	writtenPaths := installResultPaths(harnessRoot, result.WrittenPaths, manifestPath)

	if jsonOutput {
		return emitJSON(cmd.OutOrStdout(), installResultJSON{
			DryRun:       false,
			HarnessRoot:  harnessRoot,
			Source:       installSourcePayload(result.Preview),
			OrbitID:      result.Preview.Source.Manifest.Template.OrbitID,
			WrittenPaths: writtenPaths,
			Warnings:     append([]string(nil), result.Preview.Warnings...),
			MemberCount:  len(runtimeFile.Members),
		})
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "installed orbit %s into harness %s\n", result.Preview.Source.Manifest.Template.OrbitID, harnessRoot); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "source_ref: %s\n", result.Preview.InstallRecord.Template.SourceRef); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if strings.TrimSpace(result.Preview.RemoteRequestedRef) != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "requested_ref: %s\n", result.Preview.RemoteRequestedRef); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if result.Preview.RemoteResolutionKind != "" {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "resolved_ref: %s\n", result.Preview.InstallRecord.Template.SourceRef); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "resolution_kind: %s\n", result.Preview.RemoteResolutionKind); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "member_count: %d\n", len(runtimeFile.Members)); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "files: %d\n", len(writtenPaths)); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	if err := emitInstallWarnings(cmd, result.Preview.Warnings); err != nil {
		return err
	}

	return nil
}

func emitInstallWarnings(cmd *cobra.Command, warnings []string) error {
	if len(warnings) == 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "warnings: none"); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "warnings:"); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	for _, warning := range warnings {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), warning); err != nil {
			return fmt.Errorf("write command output: %w", err)
		}
	}

	return nil
}

func installSourcePayload(preview orbittemplate.TemplateApplyPreview) installSourceJSON {
	return installSourceJSON{
		Kind:           preview.InstallRecord.Template.SourceKind,
		Repo:           preview.InstallRecord.Template.SourceRepo,
		Ref:            preview.InstallRecord.Template.SourceRef,
		RequestedRef:   preview.RemoteRequestedRef,
		ResolvedRef:    resolvedInstallRef(preview),
		ResolutionKind: resolvedInstallResolutionKind(preview),
		Commit:         preview.Source.Commit,
	}
}

func resolvedInstallRef(preview orbittemplate.TemplateApplyPreview) string {
	if preview.RemoteResolutionKind == "" {
		return ""
	}

	return preview.InstallRecord.Template.SourceRef
}

func resolvedInstallResolutionKind(preview orbittemplate.TemplateApplyPreview) string {
	if preview.RemoteResolutionKind == "" {
		return ""
	}

	return string(preview.RemoteResolutionKind)
}

func installBindingsPayload(resolvedBindings map[string]bindings.ResolvedBinding) []installBindingJSON {
	names := make([]string, 0, len(resolvedBindings))
	for name := range resolvedBindings {
		names = append(names, name)
	}
	sort.Strings(names)

	items := make([]installBindingJSON, 0, len(names))
	for _, name := range names {
		source := string(resolvedBindings[name].Source)
		if strings.TrimSpace(source) == "" {
			source = "unknown"
		}
		items = append(items, installBindingJSON{
			Name:   name,
			Source: source,
		})
	}

	return items
}

func installPreviewPaths(preview orbittemplate.TemplateApplyPreview) []string {
	paths := make([]string, 0, len(preview.RenderedFiles)+5)
	for _, file := range preview.RenderedFiles {
		paths = append(paths, file.Path)
	}
	if preview.RenderedSharedAgentsFile != nil {
		paths = append(paths, preview.RenderedSharedAgentsFile.Path)
	}
	paths = append(paths,
		fmt.Sprintf(".harness/orbits/%s.yaml", preview.Source.Manifest.Template.OrbitID),
		fmt.Sprintf(".harness/installs/%s.yaml", preview.Source.Manifest.Template.OrbitID),
		harnesspkg.ManifestRepoPath(),
		".harness/vars.yaml",
	)

	sort.Strings(paths)

	return paths
}

func installResultPaths(repoRoot string, writtenPaths []string, manifestPath string) []string {
	merged := append([]string(nil), writtenPaths...)
	if strings.TrimSpace(manifestPath) != "" {
		if relativePath, err := filepath.Rel(repoRoot, manifestPath); err == nil {
			merged = append(merged, filepath.ToSlash(relativePath))
		}
	}

	sort.Strings(merged)

	deduped := merged[:0]
	for _, path := range merged {
		if len(deduped) > 0 && deduped[len(deduped)-1] == path {
			continue
		}
		deduped = append(deduped, path)
	}

	return deduped
}
