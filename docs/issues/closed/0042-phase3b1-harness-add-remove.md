# ISSUE-0042 Phase 3B-1 Harness Add And Remove

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

实现 `harness add` 与 `harness remove`，建立 runtime members 的基础人工管理路径。

## Scope

- `harness add`
  - 校验 `.orbit/orbits/<orbit-id>.yaml` 存在且合法
  - 若 member 已存在则 fail-closed
  - 以 `source=manual` 写入 `.harness/runtime.yaml`
- `harness remove`
  - 若 member 不存在则 fail-closed
  - 只更新 `.harness/runtime.yaml`
  - 不自动删除 definition / install record / runtime files

## Done When

- `harness add` / `harness remove` CLI 可用。
- `members` 更新符合 technical spec 中的 `source=manual` / remove 语义。
- 对应单测与 temp repo 集成测试补齐。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 8.3 / 任务 2。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.3、8.4。
- Completed:
  - `harness add` 已按 `source=manual` 写入 `.harness/runtime.yaml`，并要求目标 orbit definition 已存在且合法。
  - `harness remove` 已按合同只更新 members，不自动删除 definition / install record / runtime files。
  - CLI、domain 单测与 temp-repo 集成测试已覆盖 add/remove 的主路径。
