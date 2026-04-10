# ISSUE-0047 Phase 3A-0 Harness Binary Entrypoints And Minimal Build Surface

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-25
- Updated: 2026-03-26

## Summary

新增 `harness` 顶层二进制与最小 CLI root，并把它纳入最小 build / test 入口，确保后续 `harness` 命令不是仅存在于代码目录中，而是可被真实构建、发现和验证。

## Scope

- 新增 `cmd/harness/main.go` 与 `cmd/harness/cli/root.go`。
- 建立稳定的 `harness` root command 与共享入口面。
- 保持现有 `orbit` 二进制入口不变。
- 让 `harness` 与 `orbit` 共享已有底层包和新增的 `cmd/orbit/cli/harness` 领域包。
- 把 `harness` 纳入最小 build / test 入口或等价 CI 检查。

## Done When

- `harness` 二进制可被构建并成功启动 root command。
- `orbit` 现有入口不受回归影响。
- 最小 build / test 流程已覆盖 `harness` 入口存在性。
- 为后续在 `harness` 下挂正式命令留出稳定入口。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 6.3 / 任务 3。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 7.1、7.2。
- Completed:
  - `cmd/harness/main.go` 与 `cmd/harness/cli/root.go` 已落地，`harness` 成为正式独立二进制入口。
  - root/help 测试与 CLI 集成测试已覆盖入口存在性与基本命令面。
  - `orbit` 入口保持不变，后续 `harness` 命令继续复用同一入口面扩展。
