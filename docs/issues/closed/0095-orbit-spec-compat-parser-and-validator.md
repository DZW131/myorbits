# ISSUE-0095 Orbit Spec Compatibility Parser And Validator

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-04
- Updated: 2026-04-04

## Summary

扩展 `.orbit/orbits/*.yaml` 的解析与校验链路，让 Orbit 能同时接受旧模型 `id + description + include + exclude` 和新成员化模型 `meta + members + rules`，并保持 fail-closed 的结构校验。

## Scope

- 扩展 OrbitSpec 解析器，接受新字段：
  - `meta`
  - `members`
  - `rules`
- 建立兼容投影：
  - 旧 schema 可自动进入新内部模型
  - 新 schema 可直接进入新内部模型
- 增加至少以下校验：
  - role 合法性
  - member key 唯一性
  - `meta.file` 与 orbit id 一致性
  - role-scope patch 合法性
- 保持未知结构、非法组合继续 fail-closed。

## Done When

- 旧 schema 和新 schema 都能稳定解析。
- 非法 role、重复 key、非法 meta/file 配置会稳定报错。
- 旧仓库不需要立刻迁移也能继续工作。
- 解析与校验矩阵测试已覆盖主要合法/非法场景。

## Notes

- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 2。
- 依赖 ISSUE-0094 提供的新内部对象模型。
- Completed:
  - `ParseOrbitSpecData` 现已同时接受 legacy schema 与 member schema，并对 mixed schema、非法 role、重复 key、错误 `meta.file`、非法 scope patch 失败关闭。
  - `ParseDefinitionData` 现通过 OrbitSpec 兼容投影回当前 `Definition`，让现有命令面仍可消费。
  - control loader 已切到兼容解析链路，并保留 repo 级 duplicate-id 校验顺序。
  - 新增 parser / validator / loader 相关测试矩阵，`go test ./cmd/orbit/cli/orbit`、`go test ./cmd/orbit/cli/...`、`mise run lint` 通过。
