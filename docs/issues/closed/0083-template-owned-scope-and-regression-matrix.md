# ISSUE-0083 Template Owned Scope And Regression Matrix

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

把 orbit / harness template 相关流程统一切到 owned scope，并补齐新 scope 语义下的回归矩阵与文档。

## Scope

- 更新 `orbit template save` / `orbit template publish`：
  - 只消费 `OwnedPaths`
  - `projection_visible` 文件不进入 orbit template
- 更新 `harness template save` member candidate：
  - 只消费 `OwnedPaths`
- 明确 `AGENTS.md` 特殊 lane：
  - 仅在属于 owned scope 时进入 template 处理
  - 不能因 `projection_visible` 命中而自动进入 template
- 更新测试与文档：
  - `docs/testing-strategy.md`
  - `README.md`
  - 必要的 context 文档与 upgrade note

## Done When

- `projection_visible` 文件不会进入 orbit template save/publish。
- `projection_visible` 文件不会进入 harness template member candidate。
- 新术语 `shared_scope` / `projection_visible` 已反映到主要测试矩阵与对外说明。

## Notes

- 这一项应在 `0081` 与 `0082` 落定后再做，避免 template 测试基线反复漂移。
- 已完成：
  - `orbit template save/publish` 只消费 owned scope
  - `harness template` member candidate 明确切到 owned scope
  - `AGENTS.md` 仅在属于 owned scope 时进入 template special lane
  - README 与 testing strategy 已同步新术语和行为矩阵
