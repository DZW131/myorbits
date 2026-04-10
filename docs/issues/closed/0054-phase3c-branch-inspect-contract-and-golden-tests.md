# ISSUE-0054 Phase 3C Branch Inspect Contract And Golden Tests

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

把 `orbit branch inspect` 升级到 harness-aware 合同，并用 text/json golden tests 冻结输出，确保后续分支诊断结果稳定。

## Scope

- `orbit branch inspect` 输出至少包含：
  - `kind`
  - `template_kind`
  - `harness_id`
  - `member_count`
  - `member_ids`
  - `definition_count`
  - `definition_ids`
  - `install_count`
  - `install_ids`
  - `includes_root_agents`
- inspect 只基于 branch 上可证明的信息：
  - 不读 `.git/orbit/state/*`
- 覆盖：
  - zero-member runtime
  - orbit template
  - harness template
  - runtime with installs
- 用 text/json golden tests 冻结输出

## Done When

- `orbit branch inspect` text / json 契约与 technical spec 对齐。
- golden tests 覆盖 zero-member runtime、template subtype、install-backed runtime 摘要。
- `orbit branch inspect` 不再沿用单 orbit runtime 的旧摘要模型。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 10.3 / 任务 2。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 7.3、8.5。
- Completed:
  - `orbit branch inspect` 已切到 harness-aware counts/ids 合同，不再输出旧的单-orbit runtime 摘要模型。
  - text/json golden tests 已覆盖 orbit template、harness template、install-backed runtime、zero-member runtime。
  - inspect 只读取 branch 上的 `.harness/*` / `.orbit/*` 版本化合同，不依赖 `.git/orbit/state/*`。
