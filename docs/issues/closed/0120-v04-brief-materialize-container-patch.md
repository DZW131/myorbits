# ISSUE-0120 v0.4 Brief Materialize Container Patch

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-08

## Summary

正式实现 `orbit brief materialize`，把结构化 brief 真相源物化为 repo 根 `AGENTS.md` 中当前 orbit block，并采用 block-preserving container patch 语义，避免把整份容器文件误当作当前 orbit 所有。

## Scope

- 新增 `orbit brief materialize`
- 从 `meta.agents_template` 与 orchestration inputs 生成当前 orbit block
- 创建或更新当前 orbit block
- 保留其他 orbit block 与 unmarked prose
- 为 drifted block overwrite 定义显式 `--force` 或等价确认合同
- 补齐 text / json 输出契约与 integration tests

## Done When

- `orbit brief materialize` 可在允许的 revision kind 下创建或更新当前 orbit block
- 命令不会重写其他 orbit block 或 unmarked prose
- drifted block 不会被静默覆盖
- 有覆盖 block create/update/preserve 的往返测试

## Notes

- 对应 `docs/orbit_brief_lane_v0_4_technical_spec.md` 的第 3、4、7、8 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 4
- 依赖 ISSUE-0111、ISSUE-0119

## Resolution

- `orbit brief materialize` 已正式实现，并采用 block-preserving root `AGENTS.md` container patch 语义。
- 命令已覆盖 create/update/preserve 路径，并在 drifted block 下默认 fail-closed，仅在 `--force` 下覆盖。
- `runtime / source / orbit_template` 三态 materialize 行为与 integration tests 已到位。
