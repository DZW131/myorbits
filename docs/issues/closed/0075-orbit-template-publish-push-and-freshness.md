# ISSUE-0075 Orbit Template Publish Push And Freshness

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

为 `orbit template publish` 增加受限的 `--push` 路径，包含 remote 解析、主分支 freshness 检查、普通 push 合同和本地/远端分离输出。

## Scope

- 支持 `--push`
- 支持 `--remote`
- 默认 remote 为 `origin`
- `--remote` 只能和 `--push` 一起使用
- base branch equal / ahead 时允许 push
- base branch behind / diverged 时阻止 push
- push 失败不回滚本地 publish 结果

## Done When

- `--push` 能向默认或显式 remote 做普通 push
- behind / diverged 时阻止 push 并返回非零
- 输出明确区分 `local_publish` 与 `remote_push`
- 相关单元测试与集成测试通过

## Notes

- 第一版不支持 force push
- 第一版不做自动切 branch 或远端自动重建
