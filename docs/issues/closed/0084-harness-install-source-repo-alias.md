# ISSUE-0084 Harness Install Source Repo Alias

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

让 `harness install <repo-url>` 在面对 orbit template source repo 时支持 source alias 解析，降低消费者必须手动理解 source branch / published template branch 的心智负担。

## Scope

- 协调本批次的执行子项：
  - 远端 source branch alias 选择原语
  - `harness install` 的 source-alias 接线
  - text/json 输出与 acceptance smoke
- 保持 install 的正式消费对象仍是 orbit template branch。
- 不自动 publish，不进入 harness template install。

## Done When

- source repo URL 可作为 `harness install` 的输入别名使用。
- install 的真实解析对象仍是已发布的 orbit template branch。
- `--ref <source-branch>` 也能触发同样 alias。
- 文档、测试与输出合同已冻结。

## Notes

- 对应 `docs/harness_install_source_repo_resolution_technical_spec.md`。
- 依赖 `docs/orbit_template_publish_technical_spec.md` 中的 source branch / `publish.orbit_id` 合同。
- 不处理 harness template install 或 mixed install。
