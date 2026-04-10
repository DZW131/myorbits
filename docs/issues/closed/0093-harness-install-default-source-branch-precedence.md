# ISSUE-0093 Harness Install Default Source Branch Precedence

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-02
- Updated: 2026-04-02

## Summary

调整 `harness install <repo-url>` 的远端选择优先级：当远端默认分支是合法 source branch 时，优先按 source alias 逻辑解析并安装对应的 published orbit template branch，而不是先把远端唯一 installable orbit template branch 当成直接候选。

## Scope

- 更新 source repo resolution 合同：
  - 无 `--ref` 时，先检查远端默认分支是否为合法 source branch
  - 若是，则优先走 source alias
  - 只有默认分支不是合法 source branch 时，才回落到现有 orbit template candidate 自动选择逻辑
- 保持以下边界不变：
  - 不自动 publish
  - source branch 仍不是 installable template branch
  - 显式 `--ref` 逻辑保持现有 source alias 合同
  - ambiguity 仍然 fail-closed
- 更新 text/json / progress 回归测试：
  - 无 `--ref` 的 source repo URL 应得到 `resolution_kind=source_alias`
  - 无 `--ref` 的 source repo URL progress 应出现 source alias 专用阶段

## Done When

- `harness install <repo-url>` 在默认分支为 source branch 时优先走 source alias
- 现有 source repo URL acceptance smoke 断言更新为 `source_alias`
- 现有 source repo progress 测试不再需要显式 `--ref main`
- 相关文档与实现保持一致

## Notes

- 依赖 `docs/harness_install_source_repo_resolution_technical_spec.md`
- 这是 source repo 直装体验优化，不涉及 mixed install
