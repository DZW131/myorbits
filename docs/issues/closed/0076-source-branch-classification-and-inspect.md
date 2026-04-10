# ISSUE-0076 Source Branch Classification And Inspect

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

把 source branch 纳入 branch taxonomy，使 `orbit branch status|list|inspect` 能按 `.orbit/source.yaml` 正确表达 source branch，而不是把它误归类为 runtime 或 plain。

## Scope

- 在 branch classifier 中新增 source 类型
- source branch 识别基于 `.orbit/source.yaml`
- source/template 冲突时 fail-closed
- 更新 inspect / status / list 的 text/json 合同

## Done When

- `orbit branch status|list|inspect` 能稳定输出 source branch
- source branch 不再被误表达为普通 runtime
- text/json 契约有测试锁定

## Notes

- 不阻塞 `orbit template publish` 的第一版落地
- 但会明显改善作者侧可观测性
