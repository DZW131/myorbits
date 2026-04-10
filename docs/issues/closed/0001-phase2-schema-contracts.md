# ISSUE-0001 Freeze Phase 2 Schema Contracts

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

实现并冻结 Phase 2 的三个版本化文件合同：`.orbit/vars.yaml`、`.orbit/template.yaml`、`.orbit/installs/<orbit-id>.yaml`，作为后续 `template save`、`template apply` 和 branch info 的共同基础。

## Scope

- 定义三个 schema 的 Go model 和 YAML codec。
- 为必要字段、`schema_version`、枚举值和结构约束实现 validation。
- 为 optional `description` 字段建立稳定序列化/反序列化行为。
- 产出清晰、稳定、可测试的错误信息。
- 为三类 schema 增加单元测试。

## Done When

- 三类文件都能稳定完成读写和校验。
- 缺少必要字段、字段类型错误或非法枚举时会 fail-closed。
- install record 只记录模板来源，不记录 projection 或 runtime state。
- 单测锁定合法样例、非法样例和向后兼容约束。

## Notes

- 对应 `docs/context/orbit_phase2_development_plan.md` 的 Phase 2A-0 / 任务 1。
- 主要依据：`docs/context/orbit_phase2_technical_spec.md` 的 5.1、6.1、6.2、6.3。
- 2026-03-21: 已实现 `.orbit/vars.yaml`、`.orbit/template.yaml`、`.orbit/installs/<orbit-id>.yaml` 的 schema model、YAML codec、validation 与单元测试。
