# ISSUE-0043 Phase 3B-1 Shared Apply Service And Harness Install

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

把现有 `orbit template apply` 内核收敛成共享 apply service，并提供 `harness install` 的正式 basic path CLI 封装。

## Scope

- 抽共享 apply service，统一：
  - runtime file write
  - orbit definition write
  - install record write
  - vars write
  - member update
- 新增 `harness install`
  - 支持 local branch / remote git source
  - 同一 `orbit-id` 首阶段默认失败
  - 非 `--dry-run` 时落盘到 runtime repo
- 现有 `orbit template apply` 暂保留为 hidden/internal wrapper

## Done When

- `harness install` basic path 可用。
- shared apply write path 只认 `.harness/*` runtime metadata。
- `orbit template apply` wrapper 回归测试通过。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 8.3 / 任务 3、任务 4。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.2。
- Completed:
  - `harness install` 已支持 local branch 与 remote git source 两条 basic path。
  - install 结果会统一写入 runtime files、`.orbit/orbits/*.yaml`、`.harness/installs/*.yaml`、`.harness/runtime.yaml`，并在需要时更新 `.harness/vars.yaml`。
  - 同一 `orbit-id` 的重复安装首阶段默认 fail-closed；`orbit template apply` 继续作为 hidden/internal wrapper 保留且已有回归测试。
