# ISSUE-0074 Orbit Template Publish Local Command

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-31
- Updated: 2026-03-31

## Summary

实现 `orbit template publish` 的本地发布路径：source branch 前置校验、orbit id 自动解析、固定分支命名、默认 overwrite、no-op 检测，以及稳定的 text/json 输出。

## Scope

- 新增 `orbit template publish`
- 支持 `--orbit`
- 支持 `--default`
- 固定发布到 `orbit-template/<orbit-id>`
- 断言当前 branch 等于 `.orbit/source.yaml` 中的 `base_branch`
- 断言 worktree clean
- 增加 no-op publish 检测

## Done When

- 单 orbit source branch 可直接 `orbit template publish`
- 多 orbit source branch 未给 `--orbit` 时失败
- 重复 publish 内容未变化时不新增 commit，并输出 `changed=false`
- text/json 输出冻结本地 publish 结果字段
- 相关命令与集成测试通过

## Notes

- 只覆盖本地 publish，不包括远端 push
- 底层继续复用 `orbit template save` / `WriteTemplateBranch`
- 已完成：
  - `orbit template publish`
  - 单 orbit 自动解析 / 多 orbit 强制 `--orbit`
  - 固定分支命名
  - source branch / clean worktree 前置校验
  - 本地 no-op 检测与 text/json 输出
