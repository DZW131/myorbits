# ISSUE-0033 AGENTS Template Apply Merge

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-23
- Updated: 2026-03-23

## Summary

让 `template apply` 通过 AGENTS 专用 writer 执行 `replace-block + create-if-absent`，而不是把 `AGENTS.md` 当普通文件整份覆盖。

## Scope

- source loader / apply preview 识别模板态共享 `AGENTS.md` payload。
- 渲染 payload 后，在 apply 时包裹运行态 marker。
- 目标 repo 不存在 `AGENTS.md`：
  - 创建新文件
- 目标 repo 已存在同 orbit block：
  - 原地替换 block 内容
  - 发出 warning
- 目标 repo 已存在 `AGENTS.md` 但不存在同 orbit block：
  - 在 EOF 追加新 block
  - 保证前置空行和结尾换行
- runtime `AGENTS.md` 异常 marker 场景全部 fail-closed。

## Done When

- local apply 与 remote apply 都能正确处理 shared `AGENTS.md`。
- 同 orbit block 替换、无 block 追加、文件不存在创建这三条路径都有集成测试保护。
- `AGENTS.md` 不会被普通 overwrite 路径误处理。
- malformed runtime `AGENTS.md` 会阻断 apply。

## Notes

- 当前 apply/conflict 逻辑是文件级，本 issue 需要为 `AGENTS.md` 建专用 merge lane。
- 这是 AGENTS 扩展里最容易误伤主线路径的点，必须保持变更集中。
- 参考文档：
  - `docs/context/orbit_agents_md_development.md`
- Completed:
  - `template apply` 现在会把 shared `AGENTS.md` 作为专用 payload 渲染，并在运行态执行 replace-block / create-if-absent。
  - 同 orbit block 已存在时会 warning，并在 apply 时原地替换。
  - malformed runtime `AGENTS.md` 会在 preview / apply 两条路径 fail-closed。
