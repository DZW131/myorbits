# ISSUE-0124 v0.4 Harness Bindings Plan For Multi-Orbit Installs

- Status: open
- Priority: high
- Owner:
- Created: 2026-04-09
- Updated: 2026-04-09

## Summary

为 harness 作者提供一个面向多个 orbit template 的统一 bindings planning 入口，避免反复手工运行 `orbit bindings init`、手工合并 skeleton，并降低共享 `.harness/vars.yaml` 的维护负担。

## Scope

- 设计 `harness bindings plan <template...>` 或等价命令
- 一次解析多个 orbit template 的变量声明
- 输出共享 `.harness/vars.yaml` skeleton 或等价 preview
- 报告变量冲突、描述不一致、缺失值和潜在复用值
- 保持当前 `.harness/vars.yaml` 作为 runtime 级正式宿主

## Done When

- harness 作者可以一次看到多个模板所需变量的合并结果
- 共享值与冲突值有明确诊断输出
- 结果可直接写成 `.harness/vars.yaml`，供后续 install 复用
- 该能力不绕过现有 bindings merge contract，而是建立在其之上

## Notes

- 当前可复用的现有 primitive：
  - `orbit bindings init`
  - `.harness/vars.yaml`
  - bindings merge precedence
- 这张 issue 只负责“统一变量规划”，不负责实际安装
- 依赖 ISSUE-0123 的主流程定义
