# ISSUE-0044 Phase 3A-0 Preflight Foundations

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-25
- Updated: 2026-03-26

## Summary

推进 v0.3 开发计划的 `Phase 3A-0`，先完成 host 前置基础设施、共享原语和双二进制最小入口，为后续 harness bootstrap、runtime host 切换和 install 主路径提供稳定地基。

## Scope

- 协调本阶段的 3 个执行子项：
  - path host helper 与 `.harness/*` shared primitives
  - `.harness/runtime.yaml` / `.harness/template.yaml` schema contracts
  - `harness` 顶层二进制与最小 build / test 入口
- 确保本阶段只做 preflight foundations，不提前进入 bootstrap、install 或 branch/check 主逻辑。
- 以 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3A-0` 作为验收基线。

## Done When

- `Phase 3A-0` 的 host helper、schema contracts 与 binary entrypoint 已全部落地。
- 本阶段完成标准与 `docs/harness_centric_runtime_development_plan.md` 的 6.5 保持一致。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 `Phase 3A-0：Preflight Foundations`。
- Completed:
  - 新增 `cmd/orbit/cli/harness/*` 共享领域包，承载 `.harness/*` 路径、schema 与共享原语。
  - 新增 `cmd/harness/main.go` 与 `cmd/harness/cli/root.go`，建立独立 `harness` 二进制入口。
  - 为后续 bootstrap、runtime host 切换与 install 主路径提供统一前置基础设施。
