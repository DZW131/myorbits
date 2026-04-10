# ISSUE-0045 Phase 3A-0 Path Host Helpers And Harness Shared Primitives

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-25
- Updated: 2026-03-26

## Summary

抽离 `.orbit/vars.yaml` 与 `.orbit/installs/*` 的硬编码路径，并建立 `.harness/vars.yaml`、`.harness/installs/*` 的统一 host helper，为后续 bindings、install records 和 runtime 写入提供单一入口。

## Scope

- 把现有代码中的 `.orbit/vars.yaml` 固定宿主路径责任从调用点中抽离。
- 把现有代码中的 `.orbit/installs/*` 固定宿主路径责任从 `template` 侧收口到 `harness` 侧。
- 在 `cmd/orbit/cli/harness` 中建立 `.harness/vars.yaml` / `.harness/installs/*` 的 path helper、load / write 入口或等价 shared primitive。
- 保持 `bindings` 包只负责 schema / merge / skeleton，不拥有固定宿主路径。
- 为后续命令避免直接拼接 `.harness/*` 路径。

## Done When

- 业务代码不再直接依赖硬编码 `.orbit/vars.yaml` / `.orbit/installs/*` 宿主路径。
- `.harness/vars.yaml` 与 `.harness/installs/*` 的路径解析与读写入口集中在共享 helper 中。
- 新增单测覆盖 host helper 的合法路径、repo root 解析依赖和 fail-closed 行为。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 6.3 / 任务 1。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 5.1、5.3、5.4、7.2、7.3。
- Completed:
  - 在 `cmd/orbit/cli/harness/paths.go` 中建立 `.harness/runtime.yaml`、`.harness/vars.yaml`、`.harness/installs/*`、`.harness/template.yaml` 的统一 host helper。
  - 在 `cmd/orbit/cli/harness/vars.go` 与 `cmd/orbit/cli/harness/install.go` 中建立 vars / install record 的 host wrapper。
  - `bindings` 与 `template` 包保留 Phase 2 兼容 `AtPath` 读写原语，业务调用点已可通过 harness helper 避免直接拼接 runtime 宿主路径。
