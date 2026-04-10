# ISSUE-0113 v0.4 Brief Materialize And Backfill

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-08

## Summary

正式实现并收口 brief lane：`orbit brief materialize` 负责把结构化 brief 真相源物化为当前 orbit 的容器 block，`orbit brief backfill` 负责把当前 orbit block 回填为 `meta.agents_template`，两者都只处理 orchestration brief，不处理 export 或通用文件反写。

本 issue 作为 brief lane 的 umbrella issue，负责冻结总目标、验收口径与子 issue 依赖关系；具体实现按拆分 issue 分步落地。

## Scope

- 新增 `orbit brief materialize`
- 硬化 `orbit brief backfill`
- 固化 root `AGENTS.md` 作为 orchestration artifact 的行为
- 支持 `runtime / source / orbit_template` 三类 revision kind
- 基于 `.harness/vars.yaml` 完成 reverse replacement
- 明确 drifted block、invalid container、missing truth 等状态

## Done When

- `materialize` / `backfill` 能在 `runtime / source / orbit_template` 下稳定往返
- 回填只写回 `.harness/orbits/<orbit-id>.yaml` 的 `meta.agents_template`
- 命令不会把整份根 `AGENTS.md` 误当作 authored truth
- brief lane 的状态机和错误路径有完整测试覆盖

## Notes

- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 8、9 节
- 对应 `docs/orbit_brief_lane_v0_4_technical_spec.md`
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 4
- 推荐按以下顺序分别落地：
  - ISSUE-0119：revision gating 与 orbit targeting
  - ISSUE-0120：`orbit brief materialize` 与 block-preserving container patch
  - ISSUE-0121：`orbit brief backfill` 的 revision matrix 与 local-only write contract
  - ISSUE-0122：brief lane drift diagnostics

## Resolution

- `orbit brief materialize` 与 `orbit brief backfill` 已作为正式命令落地，并覆盖 `runtime / source / orbit_template` 三态。
- brief lane 已从 publish/export/writeback 中拆开；root `AGENTS.md` 按 orchestration artifact 处理，不再作为 authored truth。
- `materialize` / `backfill` / `--check` 的主要往返、drift、invalid container、missing truth 与 revision-matrix 集成测试已到位。
- umbrella 目标已由 ISSUE-0119 / 0120 / 0121 / 0122 的实现共同完成，后续如有补强应开新 issue，而不是继续保持本 umbrella 为 open。
