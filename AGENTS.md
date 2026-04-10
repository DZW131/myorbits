# Orbit

Agent entry file for this repository.

Read this first, then follow the referenced project docs. Keep decisions narrow, explicit, and aligned with the current v0.4 unified control plane plus Orbit surface model.

## Source of Truth

Read in this order:

1. `docs/orbit_v0_4_prd.md`
2. `docs/orbit_v0_4_technical_spec.md`
3. `docs/orbit_v0_4_development_plan.md`
4. `docs/testing-strategy.md`
5. `CONTRIBUTING.md`
6. `AGENTS.md`

Read these additional docs when the change touches those areas:

1. `docs/orbit_template_authoring_guide.md` for direct template authoring / source authoring / publish flows
2. `docs/orbit_brief_lane_v0_4_technical_spec.md` for shared `orbit brief materialize` / `orbit brief backfill` semantics across `runtime / source / orbit_template`
3. `docs/orbit_member_runtime_technical_spec.md` and `docs/orbit_member_runtime_development_plan.md` for orchestration, brief, and `.git/orbit/state/*` ledger behavior
4. `docs/context/orbit_v0_4_design_archive.md` for the relationship between the new v0.4 baseline and archived design docs
5. `docs/harness_centric_runtime_prd.md`, `docs/harness_centric_runtime_technical_spec.md`, and `docs/harness_centric_runtime_development_plan.md` only as v0.3 historical background
6. `docs/context/mvp-product-requirements.md`, `docs/context/mvp-technical-architecture.md`, `docs/context/orbit_storage_boundary.md`, `docs/context/orbit_phase2_prd.md`, `docs/context/orbit_phase2_technical_spec.md`, and `docs/context/orbit_phase2_development_plan.md` for inherited Git-native kernel constraints
7. `docs/context/orbit_agents_md_development.md` only when validating the historical V0.2 compatibility lane
8. `docs/orbit_positioning_and_personas.md` for product positioning, personas, and external-control boundary guidance
9. `docs/worker_guide.md`, `docs/orbit_author_guide.md`, and `docs/harness_author_guide.md` for user-facing flows and quickstart routing

Read these sections before implementing:

- `docs/context/mvp-product-requirements.md`
  - `2.2 核心原则`
  - `4. MVP Scope`
  - `8. Core Architecture & Patterns`
  - `9. Tools / Features`
- `docs/context/mvp-technical-architecture.md`
  - `2. Technology Stack`
  - `3. Technical Architecture`
  - `4. Code File Structure`
  - `5. Security & Configuration`
- `docs/testing-strategy.md`
  - `2. MVP Test Pyramid`
  - `3. Minimum Coverage Matrix`
  - `4. Test Harness Rules`

If a change conflicts with those docs, update the docs first or stop and ask.
For all new work, the v0.4 docs above win. Historical v0.3 docs remain useful background, but they no longer outrank the v0.4 baseline. MVP / Phase 2 docs still govern inherited sparse-checkout, pathspec, Git-native state, and single-repo kernel constraints.

For inherited MVP / Phase 2 behavior:

- MVP docs still govern inherited core boundaries such as single-repo, single-workspace, sparse-checkout projection, pathspec-scoped writes, and `.git/orbit/state/` as repo-local runtime state only.
- `docs/context/orbit_storage_boundary.md` is the storage contract for `.orbit/`, `.git/orbit/state/`, Git DAG, sparse projection, and `refs/orbits/*`.
- `docs/context/orbit_phase2_prd.md` defines product behavior; `docs/context/orbit_phase2_technical_spec.md` defines implementation contracts; `docs/context/orbit_phase2_development_plan.md` defines sequencing only.
- Post-v0.4, root `AGENTS.md` is an orchestration artifact, not authored truth; `docs/orbit_member_runtime_technical_spec.md` and `docs/orbit_member_runtime_development_plan.md` remain specialized supplements rather than top-level source of truth.
- When defining conflicts or metadata, use only schema-backed concepts already frozen by the docs, such as `orbit-id`, manifest fields, install records, and concrete target paths. Do not reintroduce undefined legacy concepts.

## Docs Layout

- Frequently used docs live directly under `docs/`.
- Phase background and occasional-reference docs live under `docs/context/`.
- Current layout:
  - `docs/orbit_v0_4_prd.md`
  - `docs/orbit_v0_4_technical_spec.md`
  - `docs/orbit_v0_4_development_plan.md`
  - `docs/orbit_brief_lane_v0_4_technical_spec.md`
  - `docs/testing-strategy.md`
  - `docs/technical-debt.md`
  - `docs/quickstart.md`
  - `docs/orbit_positioning_and_personas.md`
  - `docs/worker_guide.md`
  - `docs/orbit_author_guide.md`
  - `docs/harness_author_guide.md`
  - `docs/orbit_template_authoring_guide.md`
  - `docs/orbit_member_runtime_technical_spec.md`
  - `docs/orbit_member_runtime_development_plan.md`
  - `docs/context/orbit_v0_4_design_archive.md`
  - `docs/context/orbit_storage_boundary.md`
  - `docs/context/orbit_phase2_prd.md`
  - `docs/context/orbit_phase2_technical_spec.md`
  - `docs/context/orbit_phase2_development_plan.md`
  - `docs/context/orbit_state_and_workflow_unification.md`
  - `docs/context/orbit_content_and_state_optimization.md`
  - `docs/context/orbit_member_filesystem_behavior.md`
  - `docs/context/harness_single_control_plane_proposal.md`
  - `docs/context/harness_single_control_plane_technical_spec.md`
  - `docs/context/harness_single_control_plane_development_plan.md`
  - `docs/context/orbit_agents_md_development.md`
  - `docs/context/orbit_agents_md_implementation_plan.md`
  - `docs/issues/README.md`
  - `docs/issues/open/*.md`
  - `docs/issues/closed/*.md`
  - `docs/context/mvp-product-requirements.md`
  - `docs/context/mvp-technical-architecture.md`
  - `docs/context/mvp-development-plan.md`
  - `docs/context/two-scope-refactor.md`
- New PRDs, technical docs, and development plans should start in `docs/`; move them into `docs/context/` once they are no longer high-frequency references.

## Local Issue Tracking

- The repository keeps a local file-based issue tracker under `docs/issues/`.
- One Markdown file equals one issue.
- Open issues live in `docs/issues/open/`; closed issues move to `docs/issues/closed/`.
- Prefer filenames like `NNNN-slug.md` for stable local references.

## Product Boundary

Orbit is a Git-native CLI for file-scoped workspace views inside a single Git repository.

Current v0.4 baseline:

- single repository
- single project
- single workspace
- single branch background
- multiple orbits allowed
- revision identity in `.harness/manifest.yaml`
- versioned orbit definitions in `.harness/orbits/*.yaml`
- runtime provenance in `.harness/vars.yaml`, `.harness/installs/*.yaml`, and `.harness/bundles/*.yaml`
- repo-local runtime state in `.git/orbit/state/`
- one current orbit at a time
- sparse-checkout based view projection
- pathspec based orbit-scoped operations
- outside changes handled in warning mode
- root `AGENTS.md` as orchestration artifact, not authored truth

Do not introduce:

- worktrees
- web services or HTTP APIs
- background daemons
- auth / multitenancy / SaaS concerns
- databases as canonical state
- block-level or semantic orbit logic
- automatic push requirements for `refs/orbits/*`
- long-lived dual-read / dual-write compatibility lanes

Historical v0.3 / harness-centric runtime docs remain useful background, but v0.4 supersedes their control-plane layout. The v0.4 direction explicitly retires these legacy control files from steady-state use:

- `.orbit/source.yaml`
- `.orbit/template.yaml`
- `.harness/runtime.yaml`
- `.harness/template.yaml`

## Architecture

Use the planned command-centric structure:

- `cmd/orbit/cli/commands`
  - commands, flags, stdout/stderr
- `cmd/orbit/cli/orbit`
  - config loading, validation, scope resolution
- `cmd/orbit/cli/view`
  - enter/leave, current orbit, status classification
- `cmd/orbit/cli/scoped`
  - files, diff, log, commit, restore
- `cmd/orbit/cli/git`
  - Git access
- `cmd/orbit/cli/state`
  - file-backed runtime state
- `cmd/orbit/cli/ids`
  - id and path validation helpers

Rules:

- Keep command files thin.
- Do not put sparse-checkout mutation logic in command code.
- Do not write `.git/orbit/state/*` files from command code.

## Git and Path Rules

- Use system `git` for repo-root, git-dir, ls-files, status, sparse-checkout, diff, log, add, commit, restore, update-ref.
- Prefer explicit `exec.CommandContext` argument lists.
- Do not use `sh -c` for normal Git operations.
- Use repo root, not current working directory, for repo-relative paths unless the current directory is explicitly needed.
- Validate identifiers before using them in paths or ref names:
  - `orbit-id`
- Normalize repo-relative paths before matching or writing state.
- Prefer `-z` / NUL-delimited Git I/O for path lists.

## State Rules

Treat these as separate stores:

- Git commit DAG: canonical history
- `.harness/manifest.yaml`: revision identity
- `.harness/orbits/*.yaml`: authored OrbitSpec
- `.harness/vars.yaml`, `.harness/installs/*.yaml`, `.harness/bundles/*.yaml`: runtime provenance
- `.git/orbit/state/`: repo-local runtime state and caches
- root `AGENTS.md`: orchestration artifact
- `refs/orbits/*`: optional auxiliary refs only

Do not blur them:

- `.harness/*` is versioned truth, not cache
- `.git/orbit/state/` is runtime state, not history
- `resolved_scope/*.txt` is cache, not canonical definition
- sparse-checkout view is projection, not source of truth
- root `AGENTS.md` is materialized output, not authored truth
- `refs/orbits/*` cannot become required for correctness

## Command Intent

Keep these behaviors stable:

- `harness create` / `harness init`
  - initialize a `kind=runtime` revision
- `harness install`
  - install an orbit template or harness template into runtime
- `orbit add`
  - create a hosted orbit definition skeleton
- `orbit validate`
  - validate hosted OrbitSpec and scope resolution
  - fail closed on structural problems
- `orbit enter`
  - project one orbit into the current workspace
  - block when dirty tracked paths would be hidden
- `orbit leave`
  - restore full tracked workspace view
- `orbit status`
  - classify in-scope and out-of-scope changes
- `orbit diff` / `orbit log`
  - operate on current orbit `orbit_write` surface
- `orbit commit`
  - commit only current orbit `orbit_write` surface
  - must not silently include scope-outside changes
- `orbit restore`
  - restore only current orbit `orbit_write` surface to a revision
  - create a normal Git commit
- `orbit brief materialize`
  - materialize an editable root `AGENTS.md` entry from orchestration truth
- `orbit brief backfill`
  - write the current orbit brief back to structured orchestration truth
- `orbit template save` / `orbit template publish`
  - save or publish only the `export` surface of an orbit
- `harness template save`
  - export a runtime as a harness template payload

## Output and Logging

- User-facing output goes to stdout/stderr and should be stable.
- Machine-readable commands should support `--json`.
- Do not log full file bodies unless explicitly required.
- Prefer IDs, paths, counts, and warning summaries over raw content.

## Testing

- Testing rules live in `docs/testing-strategy.md`.
- Add unit tests for orbit config, scope resolution, Git adapters, status classification, scoped commit, restore, and file-backed state.
- Tests that touch Git state must use isolated temp repositories.
- Use `t.Parallel()` by default.
- Do not use `t.Parallel()` for tests that modify process-global state such as cwd or env.

## Validation

Before commit:

```bash
mise run fmt
mise run lint
mise run test:ci
```

Once the Go code exists, these checks are required.
