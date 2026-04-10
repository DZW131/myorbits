# ISSUE-0030 AGENTS Manifest And Source Routing

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-23
- Updated: 2026-03-23

## Summary

为 `AGENTS.md` 扩展增加最小 manifest 合同，并把模板态根目录 `AGENTS.md` 从普通 owned-file 路径中分流出来，进入专用 shared-file lane。

## Scope

- 扩展 `.orbit/template.yaml` schema，支持最小 `shared_files` 合同。
- 仅允许 V0.2 冻结的 AGENTS entry：
  - `path: AGENTS.md`
  - `kind: agents_fragment`
  - `merge_mode: replace-block`
  - `include_unmarked_content: true|false`
- 更新 template manifest parse / validate。
- 更新 local / remote template source loader，使声明后的根目录 `AGENTS.md` 不再作为普通模板文件加载。
- 更新 content builder，使 `template save` 不会把 shared `AGENTS.md` 同时走普通 owned-file 路径。

## Done When

- 合法 `shared_files` manifest 能稳定 parse / validate。
- 非法 `path`、`kind`、`merge_mode`、重复 entry 等场景 fail-closed。
- 当模板声明 shared `AGENTS.md` 时，source loader 能把它识别为专用 payload，而不是普通 `CandidateFile`。
- 没有声明 `shared_files` 的现有模板行为不变。
- 有单元测试和最小 source-loading 回归测试保护上述合同。

## Notes

- 这是后续 parser、save、apply、validate 的总开关。
- 本 issue 不负责 runtime marker 解析，也不负责 merge / replace-block 行为。
- 参考文档：
  - `docs/context/orbit_agents_md_development.md`
  - `docs/context/orbit_phase2_technical_spec.md`
- Completed:
  - 扩展了 `shared_files` manifest 合同，并冻结了 AGENTS entry 的最小验证规则。
  - local / remote source loader 现在会把声明后的根目录 `AGENTS.md` 分流为专用 payload，不再当普通模板文件读取。
  - 若模板 branch 包含根目录 `AGENTS.md` 但未声明 `shared_files`，会 fail-closed。
