# ISSUE-0105 Harness Single Control Plane Bootstrap And Classification

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-06

## Summary

让 runtime bootstrap 与 branch classification 改为基于 `.harness/manifest.yaml`，把 branch 顶层身份判断从多 marker 文件组合收口为单入口。

## Scope

- `harness init` / `harness create` 改写 `manifest.kind=runtime`。
- `harness root` / `inspect` 改读 manifest。
- branch classifier 改为只看 manifest。
- 明确 `orbit init` 的兼容策略：
  - hidden wrapper
  - 或稳定失败并给迁移提示
- 覆盖：
  - runtime
  - orbit_template
  - harness_template
  - plain

## Done When

- runtime / template 顶层身份只依赖 manifest。
- zero-member runtime 仍被稳定识别。
- init/create/root/inspect 与 branchinfo 集成测试通过。

## Notes

- 对应 `docs/context/harness_single_control_plane_technical_spec.md` 的第 5、7、8.1 节。
- 依赖 ISSUE-0103。
- Completed:
  - `harness init` / `harness create` 已写入 `kind=runtime` 的 `.harness/manifest.yaml`，zero-member runtime 可被稳定识别。
  - branch classifier / inspect 已改为以 `.harness/manifest.yaml` 为 branch 顶层身份入口，覆盖 runtime / orbit_template / harness_template / plain。
  - `harness root` / `inspect` 已改为 manifest-based；`orbit init` 现保留为 deprecated compatibility path，并给出 `harness init` migration hint。
