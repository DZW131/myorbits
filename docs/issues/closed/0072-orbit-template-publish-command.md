# ISSUE-0072 Orbit Template Publish Command

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

新增作者侧命令 `orbit template publish`，把模板源仓库中的 source branch 发布成最新 orbit template branch，降低作者维护 source branch / release branch 的心智负担，同时保持 `harness install` 仍只消费已发布模板态。

## Scope

- 新增 `orbit template publish`
- 支持默认解析：
  - 单 orbit repo 自动解析 orbit id
  - 默认发布 ref 为 `orbit-template/<orbit-id>`
- 支持扩展参数：
  - `--orbit`
  - `--push`
  - `--remote`
  - `--default`
- 保持 `harness install` 不隐式触发 publish

## Done When

- 已有一份冻结的产品行为文档
- 已有一份冻结的实现合同文档
- 命令边界与默认解析规则明确
- 本地 issue tracker 中已有实现入口 issue

## Notes

- 对应 `docs/orbit_template_publish_prd.md`
- 对应 `docs/orbit_template_publish_technical_spec.md`
- 这是 post-v0.3 enhancement，不回滚现有双分支 source/release 模型
