# ISSUE-0077 Orbit Template Source Branch Contract Alignment

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

把 `.orbit/source.yaml` 与 `orbit template publish` 的实现合同对齐到新的 single-orbit source branch 模型：`base_branch` 改为 `source_branch`，新增可选 `publish.orbit_id`，并把 source branch 收紧为单 orbit、无 `.harness/*` 的专业作者分支。

## Scope

- `.orbit/source.yaml` schema 从 `base_branch` 迁移到 `source_branch`
- 新增可选 `publish.orbit_id`
- `orbit template publish` 强制 source branch 恰好一个 orbit definition
- source branch 出现 `.harness/*` 或 `.orbit/template.yaml` 时 fail-closed
- `publish.orbit_id`、唯一 orbit definition、显式 `--orbit` 三者一致性校验
- 更新 `branch status|list|inspect`、`template publish` 的 text/json 字段与 reason 名称

## Done When

- 代码与文档都只使用 `source_branch`
- `.orbit/source.yaml` 支持可选 `publish.orbit_id`
- source branch 的单 orbit / 无 `.harness/*` 合同被测试锁定
- `source_branch_not_up_to_date` 等新输出契约稳定

## Notes

- `orbit template save` 的 runtime-like 提取路径保持不变
- 这是 publish/source branch 模型的核心收口 issue
