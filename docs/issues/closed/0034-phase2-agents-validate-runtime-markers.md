# ISSUE-0034 AGENTS Validate Runtime Markers

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-23
- Updated: 2026-03-23

## Summary

把运行态 `AGENTS.md` marker 合法性检查接入 `orbit validate`，让结构错误能够在专门的验证命令里 fail-closed 暴露出来。

## Scope

- 复用 AGENTS runtime parser / validator。
- 为 `orbit validate` 增加 `AGENTS.md` 结构校验。
- 只在仓库明确使用 shared `AGENTS.md` 能力时触发该校验。
- 保持 `validate` 的 fail-closed 语义，不做自动修复。
- 覆盖 human-readable 输出和必要的错误摘要。

## Done When

- 合法 runtime `AGENTS.md` 不影响现有 validate 成功路径。
- malformed / nested / duplicate marker 会让 `orbit validate` 失败。
- 普通未启用 AGENTS shared lane 的仓库不会被误报。
- 有集成测试覆盖 validate 的接线和错误传播。

## Notes

- 触发条件应尽量保守，避免把普通 repo 的自定义 `AGENTS.md` 误当 Orbit runtime block 文件。
- 若 validate 输出结构需要补充 `--json`，应沿用现有命令契约，不另造一套格式。
- Completed:
  - `orbit validate` 现在会在检测到 runtime Orbit marker 前缀时，复用 AGENTS parser 对 root `AGENTS.md` 做专项校验。
  - malformed marker 会 fail-closed；普通不含 Orbit marker 的 `AGENTS.md` 不会被误报。
