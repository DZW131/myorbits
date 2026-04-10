# ISSUE-0040 Phase 3B-1 Runtime Host Switch And Install Basic Path

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

推进 v0.3 开发计划的 `Phase 3B-1`，完成 runtime vars / install host 切换，并交付 `harness install` 的 basic path，为后续 overwrite、drift 和 branch/check 升级提供稳定主路径。

## Scope

- 协调本阶段的 3 个执行子项：
  - vars / install host 切换
  - `harness add` / `harness remove`
  - shared apply service 与 `harness install`
- 确保本阶段只交付首次安装成功 / 同 ID 默认失败，不提前进入 overwrite / drift foundations。
- 以 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3B-1` 作为验收基线。

## Done When

- `.harness/vars.yaml` 与 `.harness/installs/*` 成为正式 runtime 宿主。
- `harness add` / `harness remove` 可用。
- `harness install` basic path 可用，且 install record 正式写入 `.harness/installs/*`。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3B-1`。
- 建议实现分支：`feature/v0.3-runtime-host-install-basic`。
- Completed:
  - runtime vars / install record 默认宿主已切到 `.harness/vars.yaml` 与 `.harness/installs/*`。
  - `harness add` / `harness remove` / `harness install` basic path 已落地并接入 temp-repo 集成测试。
  - `mise run lint` 与 `mise run test:ci` 已通过。
