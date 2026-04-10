# Issues Orbit Authoring Entry

This root `AGENTS.md` exists only in the source-branch example.

Use it as a temporary authoring entry when editing the issues-orbit brief. After
editing, the author should backfill the orbit block into
`.harness/orbits/issues.yaml` under `meta.agents_template` and publish without
keeping this file in the final orbit-template payload.

## Current Orbit: `issues`

- Keep the issue loop observable.
- Keep `Status` and `Outcome` explicit.
- Do not close an issue without evidence.
- Use `tools/check-issues.sh` only for a quick `ready` / `not_ready` signal.
- Update the closure note when the issue reaches `closed`.
