# ISSUE-0073 Orbit Template Publish Source Marker

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

为 `orbit template publish` 引入 `.orbit/source.yaml` schema、读写与校验原语，明确 source branch 的显式 marker 合同，并把它和 installable template branch 区分开。

## Scope

- 新增 `.orbit/source.yaml` 的 schema、codec、validation
- 提供 source marker 读取与合法性检查原语
- 明确 source marker 与 `.orbit/template.yaml` 的冲突失败合同
- 确保 template content builder 不把 `.orbit/source.yaml` 带入模板输出

## Done When

- 代码中有稳定的 `.orbit/source.yaml` schema-backed 读写/解析入口
- source marker 缺失、无效、与 `.orbit/template.yaml` 冲突时都能 fail-closed
- orbit template 输出不包含 `.orbit/source.yaml`
- 相关单元测试覆盖上述合同

## Notes

- 对应 `docs/orbit_template_publish_technical_spec.md`
- 这是 `orbit template publish` 的前置基础，不直接引入 push
- 已完成：
  - `.orbit/source.yaml` schema、codec、validation
  - publish 前置 source marker 校验
  - `.orbit/source.yaml` 已从 orbit template 输出中排除
