# ISSUE-0078 Orbit Template Init-Source Command

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

新增作者入口 `orbit template init-source`，用来初始化 single-orbit source branch 的最小环境，避免作者手工编写 `.orbit/source.yaml`。

## Scope

- 新增 `orbit template init-source`
- 读取当前 branch 名并写入 `.orbit/source.yaml`
- 默认把 `source_branch` 设为当前 branch
- 在恰好一个 orbit definition 时写入 `publish.orbit_id`
- 在 detached HEAD、多个 orbit definitions、或存在 `.harness/*` / `.orbit/template.yaml` 时 fail-closed
- 增加 text/json 输出与帮助示例

## Done When

- 作者可以在合法单 orbit branch 上直接运行 `orbit template init-source`
- 生成的 `.orbit/source.yaml` 符合新 schema
- 非法上下文会明确拒绝而不是写半成品配置
- 相关单元测试与集成测试通过

## Notes

- 这不是 publish 的远端行为扩展，而是 source branch bootstrap 工具
- 不负责迁移多 orbit repo；遇到多 orbit 直接失败
