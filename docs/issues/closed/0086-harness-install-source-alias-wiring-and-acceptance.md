# ISSUE-0086 Harness Install Source Alias Wiring And Acceptance

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

把 source alias 解析结果接入 `harness install` 的 preview/result 输出和端到端安装链路，确保 source repo URL 对用户来说就是稳定可用的安装输入。

## Scope

- 在 `harness install` 接线上接收 source alias resolution 结果。
- 冻结 text/json 输出：
  - `requested_ref`
  - `resolved_ref`
  - `resolution_kind=source_alias|template_branch`
- 补 acceptance smoke：
  - source repo -> install
  - source repo + `--ref main` -> install
  - source repo 但 published branch 缺失 -> fail-closed
- 明确错误提示：
  - 缺少 `publish.orbit_id`
  - published branch 不存在
  - published branch 非合法 orbit template branch

## Done When

- `harness install` 在 source alias 成功时能稳定完成 preview 和 real install。
- text/json 输出能区分用户输入 ref 与真实安装 ref。
- 端到端测试覆盖 source repo 直装。
- 不改变 overwrite-existing、bindings、drift 等既有 install 合同。

## Notes

- 依赖 `ISSUE-0085`。
- 只接 orbit template 的 source alias，不进入 harness template install。
- 这项完成后，README / quickstart 是否公开推荐 source repo URL 直装，可作为后续文档决策单独处理。
