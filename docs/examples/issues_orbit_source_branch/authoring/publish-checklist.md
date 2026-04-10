# Publish Checklist

Use this checklist before publishing the `issues` orbit from source to template.

1. Confirm `.harness/manifest.yaml` declares `kind: source`.
2. Review `.harness/orbits/issues.yaml` as the authored truth.
3. If the root `AGENTS.md` changed, backfill the orbit block into `meta.agents_template`.
4. Verify that source-only files stay source-only:
   - `AGENTS.md`
   - `authoring/`
5. Check the rules and process docs for trace and observability consistency.
6. If needed, run the cheap probe in `tools/check-issues.sh` for a quick ready / not_ready signal.
7. Publish only the export surface.
