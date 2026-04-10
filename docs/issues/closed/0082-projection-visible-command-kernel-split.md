# ISSUE-0082 Projection Visible Command Kernel Split

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

让 `projection_visible` 只影响 orbit projection 与 status，不进入 scoped operations。

## Scope

- 更新 `orbit enter` / `orbit status` / `orbit files`：
  - projection 继续使用 `ProjectionPaths`
  - hidden-dirty gate 与 sparse-checkout 继续基于 `ProjectionPaths`
- 更新 scoped kernel：
  - `diff/log/commit/restore` 改用 `ScopedOperationPaths`
  - projection cache 继续缓存 `ProjectionPaths`
- 保持 command 层薄，不在命令层重新拼接 scope。
- 视需要补充 `orbit files --json` 的 scope 可观测性，但不强制在首轮修改 text 输出。

## Done When

- `projection_visible` 文件会进入 orbit 投影视图与 status in-scope 分类。
- 同一批文件不会进入 `diff/log/commit/restore` 的 scoped pathspec。
- 现有 projection cache 与 hidden-dirty gate 行为不被破坏。

## Notes

- 这是本批次改动量最大的内核子项。
- 需要 CLI integration tests 锁住 enter/status/files/commit/restore 的新边界。
- 已完成：
  - `diff/log/commit/restore` 切到 `ScopedOperationPaths`
  - projection cache 继续缓存 `ProjectionPaths`
  - 新增 `projection_visible` 集成测试，覆盖 enter/status/files/diff/commit/log/restore
