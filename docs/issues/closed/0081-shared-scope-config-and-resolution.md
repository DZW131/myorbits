# ISSUE-0081 Shared Scope Config And Resolution

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

在 config/schema/resolve 层引入 `shared_scope` 与 `projection_visible`，并把 `always_visible` 兼容读取迁到新的 scope 解析合同。

## Scope

- 更新 `GlobalConfig`：
  - `always_visible` -> `shared_scope`
  - 新增 `projection_visible`
- 兼容读取旧 `always_visible`：
  - 允许旧仓库读取
  - 若与 `shared_scope` 同时出现则 fail-closed
- 更新校验规则：
  - `shared_scope` 和 `projection_visible` 都不能命中 `.orbit/*`、`.harness/*`、`.git/orbit/state/*`
- 重写 `ResolveScopeSet(...)`：
  - `OwnedPaths = (include ∪ shared_scope) - exclude`
  - `ProjectionOnlyPaths = (projection_visible - exclude) - owned`
  - `ScopedOperationPaths = OwnedPaths + CompanionPaths`
  - `ProjectionPaths = OwnedPaths + ProjectionOnlyPaths + CompanionPaths`
- 更新 `PathMatchesOrbit(...)` 的 untracked path 判定到新合同。

## Done When

- 解析层能稳定输出 owned / projection-only / scoped-op / projection 四层 scope。
- `exclude` 能正确覆盖 `shared_scope` 与 `projection_visible`。
- 旧 `always_visible` 仓库在无冲突情况下仍可读取。

## Notes

- 这一项应先用 unit tests 锁死集合公式，再改实现。
- 已完成：
  - `shared_scope` / `projection_visible` schema 与兼容读取
  - `ResolveScopeSet(...)` 四层 scope 输出
  - `PathMatchesOrbit(...)` 新合同
  - 对应 unit tests 与全量 `cmd/orbit/cli/...` 回归通过
