# ISSUE-0096 Orbit Add And Show Member Schema Compatibility

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-04
- Updated: 2026-04-04

## Summary

让 `orbit add` 和 `orbit show` 对成员化 schema 有稳定出口，避免底层已经支持新对象模型，但创建和展示入口仍然停留在旧 path-list 心智。

## Scope

- 为 `orbit add` 明确兼容策略：
  - 保持旧 skeleton 为默认
  - 或提供新 schema skeleton 开关
- 让 `orbit show` 能稳定展示：
  - orbit 元信息
  - member 列表
  - role -> scope 结果
- 补回归测试，覆盖旧 schema 和新 schema 的 add/show 行为。
- 不在本 issue 内切换 `status/diff/log/commit/restore` 到最终 role-aware 行为。

## Done When

- `orbit add` 对下一阶段 authoring model 有清晰、稳定的入口。
- `orbit show` 能展示成员化结构，且不破坏旧输出的核心可读性。
- add/show 的旧仓库回归测试通过。

## Notes

- 对应 `docs/orbit_member_runtime_development_plan.md` 的 Phase 2 收尾项。
- 依赖 ISSUE-0095 完成新旧 schema 的稳定解析。
- Completed:
  - `orbit add` 默认仍写 legacy skeleton，同时新增 `--member-schema` 入口创建成员化 OrbitSpec。
  - `orbit show` 在 legacy schema 下保留原有输出，在 member schema 下展示结构化 `members` 与 `role_scopes` 结果。
  - 新增 add/show 的 member-schema 集成测试，并保留 existing legacy add/show 回归。
  - `go test ./cmd/orbit/cli/...` 与 `mise run lint` 通过。
