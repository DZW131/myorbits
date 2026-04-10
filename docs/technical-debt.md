# Technical Debt

This document tracks intentionally deferred engineering risks that remain after a scoped change is considered complete.

It is not a replacement for open issues. Use it for residual risks that are real, known, and currently accepted so they do not disappear into commit history or chat logs.

## Current Debt

### Harness Single Control Plane Foundations

#### 1. Typed Single-Control-Plane Manifest Validation Loses Some Field-Presence Information (`0103`)

- Source: `docs/issues/closed/0103-harness-single-control-plane-manifest-and-host-paths.md`
- Residual risk:
  The new `.harness/manifest.yaml` parser is fail-closed because it validates against raw YAML with presence-aware fields first, but the exported typed `ManifestFile` model still collapses some branch-only zero values to their Go defaults. In particular, `includes_root_agents` is represented as a plain `bool`, so a programmatically-constructed runtime or orbit-template manifest cannot preserve the distinction between “field absent” and “field explicitly carried as false”.
- Impact:
  Current file decode/write paths remain safe, but future in-memory mutation code that reuses one `ManifestFile` value across kinds could get weaker mixed-field validation than the raw parser provides, which increases the chance of branch-specific contract drift hiding inside typed helper code.
- Follow-up:
  If later command wiring starts constructing or transforming single-control-plane manifests in memory, preserve branch-field presence explicitly, for example with per-kind typed wrappers or pointer-backed optional fields, so typed validation and raw-parse validation stay equally fail-closed.

#### 2. Legacy Runtime Compatibility Codec Still Exists As A Narrow Compatibility Lane (`0109`)

- Source: `docs/issues/closed/0109-v04-manifest-taxonomy-cutover.md`
- Residual risk:
  Mainline runtime bootstrap, root resolution, mutation, install, status, and JSON output paths now read and write `.harness/manifest.yaml`, and the repo-root runtime compatibility view is manifest-backed. The remaining `.harness/runtime.yaml` support is limited to explicit compatibility helpers in `cmd/orbit/cli/harness/runtime.go` plus a small set of targeted legacy tests.
- Impact:
  The v0.4 revision taxonomy is already manifest-first in production paths, but the old runtime codec still leaves a small amount of compatibility-only surface area that could invite accidental reuse if future changes are made without checking the single-control-plane direction first.
- Follow-up:
  If no new migration scenarios appear, either remove the explicit legacy runtime codec entirely or move it behind even more obviously compatibility-only naming and file organization so it cannot drift back into mainline command flows.

#### 3. Legacy Orbit Template Manifest Codec Still Exists As A Compatibility Lane (`0110`)

- Source: `docs/issues/closed/0110-v04-hosted-orbit-spec-cutover.md`
- Residual risk:
  Mainline orbit template consumers and publish/install paths now resolve from `.harness/manifest.yaml` plus hosted `.harness/orbits/*.yaml`, but the legacy `.orbit/template.yaml` parser/writer in `cmd/orbit/cli/template/manifest.go` still exists for compatibility tests and narrow legacy helper coverage.
- Impact:
  The v0.4 control plane is already effectively branch-manifest-first in production paths, but the old codec still leaves some implementation and test surface area that could invite accidental reuse if future changes are made without checking the hosted-control-plane direction first.
- Follow-up:
  If no new compatibility scenarios appear, either remove the legacy template manifest codec entirely or make its compatibility-only status explicit in code organization and naming so it cannot drift back into mainline command flows.

#### 4. Runtime Writeback Keeps Install-Time Provenance Instead Of Refreshing Install Records (`0116`)

- Source: `docs/issues/closed/0116-v04-runtime-export-and-writeback.md`
- Residual risk:
  `orbit template save` now forms a stable runtime writeback lane and can default its target branch from `.harness/installs/<orbit-id>.yaml` for `install_orbit` members, but a successful writeback does not rewrite the install record's `template_commit` or otherwise turn install provenance into "last writeback provenance". The runtime member keeps its original install-time provenance even after exporting newer template content back to the same branch.
- Impact:
  This preserves the intended meaning of install records as "how this runtime member was installed", but it also means drift analysis and reinstall reasoning can continue to anchor on the original install commit until the runtime member is reinstalled. Users who write back improvements and then inspect provenance later may therefore see install-time history rather than the latest exported template commit.
- Follow-up:
  If real workflows need post-writeback provenance to become first-class, introduce an explicit writeback/publication record or another schema-backed "last exported" lane rather than mutating install provenance in place.

### Phase 2A-0 Shared Primitives

#### 5. Template Variable Scanner (`0003`)

- Source: `docs/issues/closed/0003-phase2-template-variable-scanner.md`
- Residual risk:
  Text/binary detection currently uses a minimal heuristic: skip files containing NUL bytes or invalid UTF-8.
- Impact:
  Valid non-UTF-8 text files may be skipped, and some exotic binary payloads that happen to be UTF-8-compatible could still be scanned as text.
- Follow-up:
  Revisit file-type detection if template sources expand beyond UTF-8 text, or if future real-world templates expose false positives / false negatives.

#### 6. Replacement Engine (`0004`)

- Source: `docs/issues/closed/0004-phase2-replacement-engine.md`
- Residual risk:
  The current replacement primitive rejects empty `literal_value` inputs and treats duplicate literals as a global ambiguity across the full bindings set passed into the engine, rather than attempting any file-local or hit-local disambiguation.
- Impact:
  If Phase 2 later needs to preserve empty-string bindings during template save, or allow narrower disambiguation based on actual file matches, the current engine contract will be too strict and may block otherwise valid save flows.
- Follow-up:
  Revisit the replacement contract when `content builder` / `template save` integration lands, and decide whether empty literals and narrower ambiguity scopes should remain fail-closed rules or become explicit product behavior.

### Phase 2B-1 Local Template Apply

#### 7. Interactive Apply Prompt UX (`0012`)

- Source: `docs/issues/closed/0012-phase2-local-template-apply-interactive-bindings.md`
- Residual risk:
  `orbit template apply --interactive` currently uses a minimal line-based stdin prompt. Empty answers fail immediately, there is no retry loop, and the flow is limited to simple single-line values.
- Impact:
  The interactive path is functional, but user input mistakes require re-running the command, and future templates that need richer validation or multi-line values will outgrow the current prompt contract.
- Follow-up:
  Add a retry/validation loop or introduce the planned editor-backed fill path before expanding interactive apply beyond the current MVP-style bindings flow.

#### 8. Repo Vars Conflict Resolution Strategy (`0036`)

- Source: `docs/issues/closed/0036-phase2-vars-optional-and-apply-conflict-hardening.md`
- Residual risk:
  Current install/apply code still resolves same-name variable conflicts too coarsely: runtime reuse, overwrite, or fail are still anchored to whole-file `.harness/vars.yaml` rewrites more than to per-variable declaration compatibility. There is still no schema-backed alias / namespacing model for cases where two templates intentionally need distinct persisted meanings under the same logical variable name.
- Impact:
  This keeps install/apply semantics simple and fail-closed, but repositories that compose templates with colliding variable vocabularies still pay with coarse conflict behavior until install-time compatibility checks move down to the variable declaration level.
- Follow-up:
  The next step is already chosen: move install/apply onto “compatible declaration => auto-continue and reuse runtime value; incompatible declaration => fail before write”, then keep `--var-conflicts=...` and any alias / namespacing model deferred until real workflows justify them.

### Phase 2C-1 Remote Template Apply

#### 9. Remote Template Source Enumeration Fan-Out (`0017`-`0021`)

- Source:
  `docs/issues/closed/0017-phase2-remote-template-source-enumeration.md`,
  `docs/issues/closed/0018-phase2-remote-default-template-resolution.md`,
  `docs/issues/closed/0019-phase2-remote-temp-ref-fetch-and-read.md`,
  `docs/issues/closed/0020-phase2-remote-template-apply-command.md`,
  `docs/issues/closed/0021-phase2-remote-template-apply-integration-tests.md`
- Residual risk:
  Remote template discovery still uses `git ls-remote --heads` plus one shallow temp-ref fetch per candidate branch when the user relies on auto-discovery. Phase 2C-2 now optimizes the explicit `--ref` fast path so remote apply / bindings init can skip `ls-remote` and avoid a duplicate fetch for the selected branch, but large remotes without `--ref` still fan out across candidate heads.
- Impact:
  This still avoids a full clone and keeps correctness simple, but repositories with many remote branches will pay in repeated network round trips and temp-ref churn before the command can resolve a winner unless callers provide `--ref`.
- Follow-up:
  If large auto-discovery remotes become a practical bottleneck, consider a repo-local cache under `.git/orbit/` or another bounded manifest discovery strategy that preserves the current cleanup and fail-closed guarantees.

### AGENTS Extension

#### 10. AGENTS Shared Fragment Save Default (`0032`-`0033`)

- Source:
  `docs/issues/closed/0032-phase2-agents-template-save-extraction.md`,
  `docs/issues/closed/0033-phase2-agents-template-apply-merge.md`
- Residual risk:
  The implemented V0.2 `AGENTS.md` save/apply model includes all unmarked runtime `AGENTS.md` content in the extracted template payload by default, in addition to the current orbit block body when present.
- Impact:
  Repo-level prose may be duplicated across multiple orbit templates, and repeated apply flows may require manual cleanup or editorial discipline from developers because unmarked content is intentionally preserved and then wrapped into the target orbit block on apply.
- Follow-up:
  If this becomes noisy in real usage, revisit whether unmarked content should remain the default save behavior, move behind an explicit bootstrap flow, or gain more precise ownership markers.

#### 11. AGENTS Validate Trigger Heuristic (`0034`)

- Source: `docs/issues/closed/0034-phase2-agents-validate-runtime-markers.md`
- Residual risk:
  `orbit validate` currently runs the dedicated `AGENTS.md` structural check only when the runtime file still contains the Orbit marker prefix `<!-- orbit:`.
- Impact:
  This keeps false positives low for ordinary project `AGENTS.md` files, but if a repository that previously used the AGENTS shared lane has all Orbit markers manually deleted, `validate` will not infer that intent from other runtime metadata alone and may miss that loss.
- Follow-up:
  Revisit the trigger once runtime metadata gains a low-false-positive signal for AGENTS shared usage, or if real-world repos show that “all markers deleted” is a meaningful failure mode worth catching automatically.

### Phase 3B-2 Overwrite / Reinstall / Drift Foundations

#### 12. Remote Install Replay Depends On Recorded Commit Reachability (`0048`-`0051`)

- Source:
  `docs/issues/closed/0048-phase3b2-overwrite-reinstall-drift-foundations.md`,
  `docs/issues/closed/0049-phase3b2-overwrite-and-reinstall-contract.md`,
  `docs/issues/closed/0050-phase3b2-owned-file-reconstruction.md`,
  `docs/issues/closed/0051-phase3b2-drift-replay-primitives.md`
- Residual risk:
  Remote install replay currently resolves overwrite/drift expectations by shallow-fetching the recorded remote branch ref and then trying to replay the recorded `template_commit`. If that commit is no longer reachable from the fetched depth-1 ref, replay falls back to `provenance_unresolvable`.
- Impact:
  Remote template installs remain safe and fail-closed, but overwrite and drift diagnosis for older remote installs may stop working once the remote branch head moves far enough away from the originally installed commit.
- Follow-up:
  If this becomes a practical workflow problem, store or resolve a more replay-stable remote source pin, such as a directly fetchable commit object, immutable tag, or equivalent provenance contract that does not depend on the current branch head.

#### 13. Overwrite Replay Uses Current Runtime Bindings Instead Of Historical Binding Snapshots (`0048`-`0051`)

- Source:
  `docs/issues/closed/0048-phase3b2-overwrite-reinstall-drift-foundations.md`,
  `docs/issues/closed/0049-phase3b2-overwrite-and-reinstall-contract.md`,
  `docs/issues/closed/0050-phase3b2-owned-file-reconstruction.md`,
  `docs/issues/closed/0051-phase3b2-drift-replay-primitives.md`
- Residual risk:
  The new replay primitive rebuilds expected install output from the recorded template source plus the current `.harness/vars.yaml`; it does not have a historical per-install bindings snapshot.
- Impact:
  If runtime bindings drift after an install, overwrite cleanup and drift diagnosis may conservatively fail closed because reconstructed “old owned” content no longer matches the materialized files that were originally rendered with older binding values.
- Follow-up:
  If this begins to block real overwrite workflows, extend the install provenance model with a schema-backed bindings snapshot or another stable replay identity so old owned-file reconstruction does not depend on today’s runtime vars.

### Phase 3C Branch / Check Upgrade

#### 12. Harness Check Stops Drift Analysis Behind Structurally Invalid Install Records (`0052`-`0056`)

- Source:
  `docs/issues/closed/0052-phase3c-branch-check-upgrade.md`,
  `docs/issues/closed/0055-phase3c-harness-check-schema-and-membership.md`,
  `docs/issues/closed/0056-phase3c-harness-check-drift-diagnostics.md`
- Residual risk:
  Drift replay is only attempted after `harness check` accepts an install record as structurally usable. If the record path or payload is already invalid, the command emits membership/path diagnostics and does not attempt any deeper drift reconstruction for that install-backed member.
- Impact:
  This keeps the diagnostic path safe and simple, but one broken install record can hide whether the same member also has definition drift or runtime file drift that would matter during manual repair.
- Follow-up:
  If this becomes a practical debugging problem, add a second-stage “best effort” drift probe for invalid install records that can surface likely drift context without weakening the existing fail-closed provenance contract.

### Phase 3D Harness Template Save

#### 13. Harness Template Save JSON Failure Contract Still Falls Back To Plain Errors For Late Git Write Failures (`0057`-`0061`)

- Source:
  `docs/issues/closed/0057-phase3d-harness-template-save.md`,
  `docs/issues/closed/0058-phase3d-member-candidate-builder.md`,
  `docs/issues/closed/0059-phase3d-candidate-merge-and-conflict-analysis.md`,
  `docs/issues/closed/0060-phase3d-root-agents-whole-file-lane.md`,
  `docs/issues/closed/0061-phase3d-harness-template-save-command-and-integration.md`
- Residual risk:
  `harness template save --json` now emits structured failure payloads for preview-stage conflicts, preview-stage replacement ambiguities, and the stable write-stage `target_branch_exists` case. Lower-level Git write failures that happen later in the branch write path, such as temp-index setup, `write-tree`, `commit-tree`, or `update-ref` failures, still fall back to plain wrapped errors instead of a machine-readable failure payload.
- Impact:
  Human operators still get a clear error, but automation consuming `--json` cannot rely on one complete save-failure schema once the command has entered the actual Git write lane. The contract is therefore strongest for validation/preflight failures and only partially structured for write-time failures.
- Follow-up:
  If save failure handling needs full machine-readable parity, extend the remaining late Git write failures onto the same `stage` / `reason` JSON contract instead of leaving them as plain wrapped errors.

### Phase 3E Release Hardening

### Harness Mixed Install

#### 16. Mixed-Kind Remote Auto-Selection Still Prefers Orbit Template Candidates (`0088`)

- Source:
  `docs/issues/closed/0088-harness-template-install-source-and-preview.md`
- Residual risk:
  `harness install <repo-url>` now recognizes `harness template` branches for dry-run preview, but the no-`--ref` selection path still resolves orbit template candidates first. If a remote exposes both installable orbit template branches and installable harness template branches, the current command will keep the existing orbit-template-first behavior instead of surfacing a mixed-kind ambiguity that forces the caller to choose explicitly.
- Impact:
  Repositories that publish both template kinds remain installable, but callers relying on implicit selection may get an orbit template preview when they expected a harness template preview. The behavior is safe and deterministic, but the precedence is not yet as explicit as the mixed-install design intends.
- Follow-up:
  When mixed install proceeds beyond preview and the selection policy is finalized, revisit remote auto-selection so mixed-kind remotes either require `--ref` or follow one documented precedence rule shared by both orbit and harness template install units.

#### 17. Bundle-Level Check Diagnostics Reuse `orbit_id` For `harness_id` (`0089`)

- Source:
  `docs/issues/closed/0089-bundle-provenance-and-runtime-member-source.md`
- Residual risk:
  `harness check` now emits bundle-level mismatch findings, but the stable `CheckFinding` payload still only has `orbit_id` as its identity slot. Bundle orphan findings currently reuse that field to carry a `harness_id`, which makes the payload semantically imprecise for machine consumers.
- Impact:
  Human-readable output remains understandable, but JSON consumers cannot distinguish “this finding refers to an orbit member” from “this finding refers to a bundle install unit” by schema alone.
- Follow-up:
  When mixed install diagnostics expand beyond the current foundation, split the check payload into explicit identity fields or introduce bundle-specific finding shapes so bundle provenance does not overload `orbit_id`.

#### 18. Mixed-Install Variable Conflict Preview Uses Runtime Vars As The Only Compatibility Proxy (`0090`)

- Source:
  `docs/issues/closed/0090-mixed-disjoint-conflict-policy.md`
- Residual risk:
  The new harness-template mixed-install preview checks variable conflicts only against the current `.harness/vars.yaml` descriptions that happen to be present in the runtime. It does not yet have install-unit-scoped declaration provenance for previously installed orbit or bundle templates.
- Impact:
  The conflict policy remains safe and fail-closed, but manual edits or stale descriptions in `.harness/vars.yaml` can produce conservative false conflicts, and the preview cannot point back to which existing install unit established the conflicting variable contract.
- Follow-up:
  When mixed-install write path and install provenance mature further, add schema-backed variable declaration provenance per install unit so compatibility checks can compare incoming template variables against the owning install contract instead of only the current runtime vars file.

#### 19. Harness Template Install Write Path Is Not Transactional Across Files, Vars, Bundles, And Runtime (`0091`)

- Source:
  `docs/issues/closed/0091-harness-agents-bundle-lane-and-save-normalization.md`
- Residual risk:
  Real harness-template install now writes rendered files, bundle `AGENTS.md`, `.harness/vars.yaml`, `.harness/bundles/<harness-id>.yaml`, and `.harness/manifest.yaml` sequentially without a rollback mechanism. A late failure can therefore leave a partially materialized bundle install in the runtime repository.
- Impact:
  The flow is still fail-closed at the conflict-analysis stage, but unexpected filesystem or write-order failures after preview can leave mixed-install state that requires manual cleanup and may surface as bundle/runtime mismatch findings in `harness check`.
- Follow-up:
  If partial-write recovery becomes a real operational problem, add a bounded install transaction strategy or explicit cleanup journal for harness bundle installs before expanding mixed-install overwrite semantics further.

#### 20. Same-Bundle Replace Cleanup Relies On Recorded `owned_paths`, Not Replayed Old Rendered Content (`0092`)

- Source:
  `docs/issues/closed/0092-same-install-unit-replace.md`
- Residual risk:
  Harness bundle overwrite now supports same-install-unit replace by deleting stale bundle-owned paths from the previous bundle record and replacing the matching `AGENTS.md` block, but bundle provenance still does not record the old resolved bindings/rendered payload needed to replay the prior install exactly. Cleanup therefore trusts the previous bundle record's `owned_paths` set instead of proving that each stale file still matches the old rendered content.
- Impact:
  Same-bundle replace now works for normal install/update flows, but a repository that manually edits a stale bundle-owned path and then runs `--overwrite-existing` can lose that path if it disappears from the new bundle payload. The behavior is deterministic and limited to the same bundle install unit, but it is less conservative than the single-orbit overwrite replay path.
- Follow-up:
  If this becomes a practical safety problem, extend bundle provenance so overwrite can replay the previously installed bundle content or otherwise verify stale-path ownership before deletion, instead of relying only on `owned_paths`.

### Install Progress

#### 21. Harness Install Progress Is Command-Stage Level Only (`0088`-`0091`)

- Source:
  `docs/issues/closed/0088-harness-install-progress.md`,
  `docs/issues/closed/0089-harness-install-progress-mode-and-emitter.md`,
  `docs/issues/closed/0090-harness-install-phase-progress-wiring.md`,
  `docs/issues/closed/0091-harness-install-progress-regression-matrix.md`
- Residual risk:
  `harness install` now emits stable progress lines to `stderr`, but the implementation is intentionally coarse: stages are emitted from the command layer around high-level checkpoints rather than from transport-level Git operations or deeper preview sub-steps. `auto` mode also relies on a simple char-device check against the configured stderr writer instead of a richer terminal capability probe.
- Impact:
  Users now get clear liveness and broad phase context, but progress ordering is still only an approximation of the underlying work, and some wrappers or redirected stderr environments may need explicit `--progress plain` because `auto` can conservatively stay quiet.
- Follow-up:
  If operators need finer-grained network visibility or more robust default auto-detection, add deeper progress hooks inside remote selection / fetch helpers, introduce an optional heartbeat or raw Git progress mode, and replace the current lightweight terminal heuristic with a better-tested terminal capability check.

### V0.4 Brief Lane And Authoring Cutover

#### 22. Source / Template Authoring Still Mixes `.orbit/orbits` And `.harness/orbits` (`0110`, `0113`-`0115`, `0120`)

- Source:
  `docs/issues/open/0110-v04-hosted-orbit-spec-cutover.md`,
  `docs/issues/open/0113-v04-brief-materialize-and-backfill.md`,
  `docs/issues/open/0114-v04-orbit-template-direct-publish.md`,
  `docs/issues/open/0115-v04-source-publish-pipeline.md`,
  `docs/issues/open/0120-v04-brief-materialize-container-patch.md`
- Residual risk:
  The current v0.4 authoring lane still has an unresolved hosted-path split. `orbit template init-source` continues to discover the single source orbit through `.orbit/orbits`, and direct template publish now uses an explicit hosted-first fallback rule: it prefers `.harness/orbits/<orbit-id>.yaml` for brief diagnostics when present, but still falls back to `.orbit/orbits/<orbit-id>.yaml` when the hosted companion is absent. Meanwhile `orbit brief materialize` and `orbit brief backfill` load and rewrite `.harness/orbits/<orbit-id>.yaml`.
- Impact:
  The most dangerous silent split is now reduced because direct template publish no longer ignores hosted authored truth when both paths exist, but authoring flows can still disagree about which OrbitSpec is canonical. A repo that only has the legacy `.orbit/orbits` companion path can still fail `materialize/backfill`, while a repo that mixes source/init-source and hosted-brief flows may still need the explicit fallback rule to keep publish usable during the cutover.
- Follow-up:
  Finish the hosted OrbitSpec cutover so source, orbit_template, publish diagnostics, and brief lane commands all resolve one canonical authoring definition host. The publish-side explicit fallback rule should remain temporary and be removed once source/init-source and template authoring stop depending on `.orbit/orbits`.

#### 23. Brief Lane Forward Rendering And `--check` Diagnostics Currently Depend On Runtime Vars (`0113`, `0120`, `0122`)

- Source:
  `docs/issues/open/0113-v04-brief-materialize-and-backfill.md`,
  `docs/issues/open/0120-v04-brief-materialize-container-patch.md`,
  `docs/issues/open/0122-v04-brief-lane-drift-diagnostics.md`,
  `docs/orbit_brief_lane_v0_4_technical_spec.md`
- Residual risk:
  The shared `materializedOrbitBriefPayload` helper now sits under both `orbit brief materialize` and the new `orbit brief materialize/backfill --check` diagnostics. It loads `.harness/vars.yaml` and performs forward variable rendering before either writing the root `AGENTS.md` block or reporting brief-lane state. That is stricter than the brief-lane design, which currently reserves `.harness/vars.yaml` for reverse replacement during backfill.
- Impact:
  Authoring branches that carry a variableized `meta.agents_template` but do not also carry runtime vars will fail to materialize a temporary root `AGENTS.md`, and they can now also fail to run `brief ... --check` at all instead of receiving a diagnostic state. It also blurs the intended boundary between orchestration truth and runtime provenance, so commands meant to be shared by `runtime / source / orbit_template` now behave differently depending on whether runtime bindings happen to exist.
- Follow-up:
  Narrow the command contract so forward render and diagnostics either preserve placeholders, accept explicit authoring-time bindings, or use another non-runtime input source. The important part is to stop making successful materialization or state inspection silently depend on runtime provenance that the brief-lane spec treats as backfill-only.

### Source Repo Direct Install

#### 24. No-Ref Remote Install Now Pays One Extra Default-Branch Probe On Non-Source Repos (`0093`)

- Source:
  `docs/issues/closed/0093-harness-install-default-source-branch-precedence.md`
- Residual risk:
  `harness install <repo-url>` without `--ref` now checks the remote default branch for a valid source marker before falling back to ordinary orbit-template candidate enumeration. On remotes whose default branch is not a source branch, this adds one extra default-branch resolution plus one source-marker inspection fetch before the command reaches the old candidate path.
- Impact:
  The new precedence is correct and improves source-repo UX, but ordinary non-source template repos now pay a small fixed remote-selection overhead on every no-ref install attempt.
- Follow-up:
  If this added latency becomes noticeable in practice, collapse the default-branch source probe into a cheaper manifest check or cache the default-branch/source-marker result alongside the existing remote candidate discovery path.

### Orbit Member Runtime Compatibility

#### 25. Member-Schema Compatibility Projection Still Flattens `process` Members Into Legacy Owned Scope (`0095`-`0096`)

- Source:
  `docs/issues/closed/0095-orbit-spec-compat-parser-and-validator.md`,
  `docs/issues/closed/0096-orbit-add-show-member-schema-compatibility.md`
- Residual risk:
  The member-schema compatibility bridge currently projects `members[]` back into legacy `Definition.Include/Exclude` lists so the existing command surface can keep working. That flattening preserves authored paths, but it cannot represent `process -> projection_only`, so `process` member patterns are still merged into the same legacy include set consumed by the current owned-scope path.
- Impact:
  Member-schema repos can already use the new parser plus `orbit add --member-schema` and `orbit show`, but any command path that still consumes the legacy `Definition` bridge may over-include `process` files in scoped write / commit / export-style behavior. The flow remains deterministic and tested, but it is broader than the member model intends.
- Follow-up:
  Replace the legacy `Definition` compatibility bridge with a role-aware projection input before widening member-schema usage beyond the current add/show and parser compatibility surface, so `process` members can flow through `ProjectionOnlyPaths` instead of `OwnedPaths`.

### Orbit Member Runtime Phase 3

#### 26. Plain-Text `orbit status` Still Groups By Projection Scope, Not Orbit-Write Scope (`0098`-`0099`)

- Source:
  `docs/issues/closed/0098-orbit-status-role-aware-classification.md`,
  `docs/issues/closed/0099-orbit-scoped-consumers-switch-to-role-aware-paths.md`
- Residual risk:
  The new role-aware snapshot now records `role`, `projection`, `orbit_write`, `export`, and `orchestration` per path, but the human-readable `orbit status` sections still split only on the projection-facing `InScope` flag. In member-schema repos this means `subject` and `process` files can still appear under `in-scope:` even though `orbit commit` and `orbit restore` now operate only on `orbit_write`.
- Impact:
  JSON consumers can distinguish the behavior surfaces correctly, but operators reading the plain-text status view can still overestimate what a scoped write command will touch. That mismatch is especially easy to hit when `status` shows `subject/process` edits as in-scope and the next `commit` warns that those changes remain outside the write surface.
- Follow-up:
  If this causes operator confusion in practice, teach the plain-text status renderer to surface the role-aware scope split explicitly, such as by annotating each path with its write/export flags or by adding a dedicated orbit-write section alongside the projection view.

#### 27. Empty `orbit_write` Plans Still Fall Through To A Failing Git Restore Invocation (`0099`)

- Source:
  `docs/issues/closed/0099-orbit-scoped-consumers-switch-to-role-aware-paths.md`
- Residual risk:
  `orbit restore` now resolves its scope from `ProjectionPlan.OrbitWritePaths`, but it still unconditionally calls the Git restore pathspec helper. If a valid member-schema orbit explicitly disables all write-capable surfaces and therefore resolves to an empty `orbit_write`, Git exits with `fatal: you must specify path(s) to restore` instead of Orbit returning a stable no-op or a first-class validation error.
- Impact:
  The normal default member-schema path is unaffected because `meta` is write-enabled by default, but explicitly write-disabled or highly customized orbits can hit an abrupt low-level Git failure during restore. The failure is safe and fail-closed, yet it leaks transport-level behavior into the user-facing contract.
- Follow-up:
  If zero-write orbits become part of real workflows, short-circuit restore before the Git call with a stable Orbit-level outcome, and decide whether an empty `orbit_write` should be treated as “nothing to restore” or as an invalid runtime configuration.

### Orbit Member Runtime Phase 4

#### 28. Orbit-Local Ledger Readers Do Not Verify Path-To-Payload Identity (`0100`)

- Source:
  `docs/issues/closed/0100-orbit-state-ledger-models-and-fsstore-layout.md`
- Residual risk:
  The new `ReadFileInventorySnapshot`, `ReadRuntimeStateSnapshot`, and `ReadGitStateSnapshot` helpers validate only the requested `orbitID` argument and the JSON syntax on disk. They do not verify that the decoded payload's embedded `orbit` field still matches the orbit implied by the file path being read.
- Impact:
  If a ledger file is manually edited, copied, or renamed under `.git/orbit/state/orbits/<orbit-id>/`, future consumers can silently read a mismatched snapshot as if it belonged to the requested orbit. The current branch does not yet consume these ledgers for behavior-critical decisions, but once they back diagnostics or runtime observation, that mismatch could produce confusing or misleading orbit-local state.
- Follow-up:
  Add shared read-time identity validation so each ledger read fails closed when the payload's embedded `orbit` value does not match the requested orbit or the file location.

#### 29. Ledger Snapshot Writers Still Trust Caller-Supplied Internal Invariants (`0100`)

- Source:
  `docs/issues/closed/0100-orbit-state-ledger-models-and-fsstore-layout.md`
- Residual risk:
  The Phase 4 storage layer currently validates only the top-level orbit id before writing. It does not yet enforce deeper invariants such as normalized file paths and role values in `file_inventory.json`, uniqueness of file entries, or `count == len(paths)` consistency inside `git_state.json`.
- Impact:
  The storage API stays simple and easy to evolve, but a buggy future writer or a hand-edited ledger file can persist internally inconsistent snapshots that still deserialize successfully. That would make the new ledger files a less trustworthy debugging surface and could force later runtime consumers to defend against malformed local state ad hoc.
- Follow-up:
  Introduce schema-backed normalization and invariant checks at write time, and decide whether read paths should also reject inconsistent snapshots or tolerate them only for best-effort diagnostics.

#### 30. Runtime/Git Ledger Refresh Can Fail After The Primary Orbit Mutation Already Succeeded (`0102`)

- Source:
  `docs/issues/closed/0102-orbit-runtime-and-git-state-ledger-updates.md`
- Residual risk:
  The new ledger writes are currently sequenced after the command's primary mutation in several paths. `orbit enter` applies sparse-checkout and writes `current_orbit.json` before `runtime_state.json` / `git_state.json`; `orbit leave` disables sparse-checkout before writing the "left" ledger and clearing current state; `orbit commit` and `orbit restore` both create the Git commit before refreshing the new ledgers; `orbit status` writes `status.json` before the ledger refresh. A late `git status` failure or ledger write failure therefore returns an error after the user-visible side effect has already happened.
- Impact:
  Operators can be told that `enter`, `leave`, `commit`, `restore`, or `status` failed even though the workspace view, current-orbit state, or Git history has already changed. In the worst case this can leave partially refreshed local state, such as a restored full workspace with a stale `current_orbit.json`, or a successful scoped commit whose runtime ledger still looks old.
- Follow-up:
  Decide whether ledger refresh is best-effort observation or part of the primary transaction. If it is observational, downgrade late ledger failures to warnings; if it is required state, add a bounded rollback / compensation strategy so post-mutation failures do not leave mixed runtime state.

#### 31. No-Op `orbit commit` Does Not Refresh The New Ledgers (`0102`)

- Source:
  `docs/issues/closed/0102-orbit-runtime-and-git-state-ledger-updates.md`
- Residual risk:
  `orbit commit` still returns early when the orbit-write scope has no in-scope changes, before it reaches the new runtime/git ledger refresh path. A user can therefore intentionally run `orbit commit` with only out-of-scope changes present, receive the current warning behavior, and still leave `runtime_state.json` / `git_state.json` unchanged from an older command.
- Impact:
  Ledger consumers can observe stale `phase`, stale global change buckets, or both immediately after a real `commit` command invocation. That weakens the new files as an operator-facing debugging surface precisely in the "nothing committed, but why?" workflow where the ledger would otherwise be most useful.
- Follow-up:
  Refresh the runtime/git ledgers on the no-op `commit` path as well, or explicitly freeze the contract to say that no-op commits are not ledger-bearing events and document that behavior for future readers.

### Orbit Brief Orchestration

#### 33. `orbit brief backfill` Canonicalizes Hosted Orbit YAML Instead Of Preserving Non-Schema Structure (`0108`)

- Source:
  `docs/issues/closed/0108-orbit-brief-backfill-and-agents-orchestration-cutover.md`
- Residual risk:
  `orbit brief backfill` now updates `meta.agents_template` by loading the hosted OrbitSpec into the typed member model and rewriting the full `.harness/orbits/<orbit-id>.yaml` file through `yaml.Marshal`. The write is atomic and schema-valid, but it does not preserve comments, anchors, or manual YAML layout outside the schema-backed fields.
- Impact:
  A successful backfill can silently normalize a hand-edited hosted orbit file beyond the intended brief change, which is stricter than ordinary formatting cleanup and weaker than the technical spec's preferred "preserve unrelated structure or fail-closed" contract.
- Follow-up:
  If repositories start depending on comments or other non-round-trippable YAML structure in hosted orbit files, switch backfill to a node-aware patch strategy or explicitly reject backfill when the source file contains structure that the typed rewrite cannot preserve safely.

### V0.4 Source Manifest Cutover

#### 34. Source Authoring Still Uses A Template-Local Manifest Parser Instead Of The Shared Harness Manifest Contract (`0109`-`0115`)

- Source:
  `docs/issues/open/0109-v04-manifest-taxonomy-cutover.md`,
  `docs/issues/open/0115-v04-source-publish-pipeline.md`
- Residual risk:
  The new source branch flow now stores authoring identity in `.harness/manifest.yaml` and the local publish path is hardened to require `source.orbit_id`, but `cmd/orbit/cli/template/source_manifest.go` still carries its own source-manifest codec instead of reusing `cmd/orbit/cli/harness/manifest.go`. That duplication exists because the current package graph would create a cycle if `template` imported `harness`, and the remote source-alias path intentionally keeps a slightly looser parse so it can emit a targeted “missing source.orbit_id” diagnostic.
- Impact:
  The main local authoring flow is now fail-closed, but future source-manifest shape changes can still drift between branch tooling and template authoring / remote alias tooling if both parsers are not updated together. In practice that means a later schema edit could reintroduce subtle disagreement in validation rules, field names, or error surfaces around source branches even though they all read the same file path.
- Follow-up:
  When the v0.4 control-plane cutover advances far enough to change package boundaries safely, move source authoring and remote alias inspection onto one shared source-manifest parser or a small cycle-free shared contract package so `.harness/manifest.yaml kind=source` has exactly one authoritative codec and validator.

#### 35. Brief Lane Diagnostics And Direct Template Publish Still Compare Different Brief Shapes (`0113`, `0114`, `0120`, `0122`)

- Source:
  `docs/issues/open/0113-v04-brief-materialize-and-backfill.md`,
  `docs/issues/open/0114-v04-orbit-template-direct-publish.md`,
  `docs/issues/open/0120-v04-brief-materialize-container-patch.md`,
  `docs/issues/open/0122-v04-brief-lane-drift-diagnostics.md`
- Residual risk:
  The shared brief-lane path now treats `materialize` and `materialize/backfill --check` as forward-rendered operations: `materializedOrbitBriefPayload` loads `.harness/vars.yaml` and compares the current root `AGENTS.md` block against the rendered brief body. Direct template publish does not reuse that rendered comparison. Its `diagnoseDirectTemplateAgentsArtifact` preflight intentionally compares the root `AGENTS.md` block against the unrendered authored body from `orbitAgentsBody`, even after the hosted-first authored-truth cutover.
- Impact:
  A template or source authoring branch that carries both runtime bindings and a fully materialized root `AGENTS.md` block can still see inconsistent results across commands. `orbit brief materialize --check` may classify the block as `materialized_in_sync` because it matches the rendered brief, while `orbit template publish` can still classify the same block as drifted or require removal because it compares against the placeholder-based authored brief shape. That keeps publish fail-closed, but it leaves the user-facing brief status model internally inconsistent across adjacent commands.
- Follow-up:
  Decide on one canonical comparison contract for authoring branches: either publish should compare against the same rendered brief shape as the brief-lane diagnostics, or the brief-lane diagnostics/materialize path should stop depending on runtime-style forward rendering in authoring revisions. Until then, keep the current behavior documented as an accepted temporary split rather than treating the commands as fully equivalent.

#### 36. `orbit template init-source` Source-Host Migration Is Not Transactional (`0109`, `0115`)

- Source:
  `docs/issues/open/0109-v04-manifest-taxonomy-cutover.md`,
  `docs/issues/open/0115-v04-source-publish-pipeline.md`
- Residual risk:
  `orbit template init-source` now migrates a single legacy source definition into `.harness/orbits/<orbit-id>.yaml` before it writes or rewrites the source manifest. The migration path in `cmd/orbit/cli/template/init_source.go` and `cmd/orbit/cli/template/source_branch_definitions.go` therefore spans three separate filesystem mutations: write hosted definition, remove legacy definition, then write `.harness/manifest.yaml`, without a rollback path if a later step fails.
- Impact:
  A late filesystem failure can leave a partially migrated source branch. In the most obvious case, the command can successfully move the definition into `.harness/orbits/` and remove `.orbit/orbits/<orbit-id>.yaml`, then fail while writing `.harness/manifest.yaml`. That leaves the branch in a mixed “definition migrated, source identity missing or stale” state that requires manual repair before the source authoring flow is coherent again.
- Follow-up:
  If partial migration starts showing up in practice, either treat source-host migration plus manifest write as one bounded transaction with compensation, or switch the command to a staged write order that can fail before deleting the legacy definition.

## Update Rule

- Add an entry when a change is intentionally shipped with a known, accepted risk.
- Remove or rewrite an entry when the risk is actually retired, not merely postponed.
- Prefer linking the originating closed issue so the implementation context stays traceable.
