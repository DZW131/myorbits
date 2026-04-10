# ISSUE-0090 Mixed Disjoint Conflict Policy

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

实现 mixed install 第一阶段的 fail-closed disjoint conflict policy，让 orbit template 与 harness template 可以安全并存，但不允许 member / path / variable / AGENTS lane 冲突。

## Scope

- 定义并实现 mixed install 冲突集：
  - member identity conflict
  - ordinary path conflict
  - variable declaration conflict
  - `AGENTS.md` lane conflict
- 普通路径规则：
  - 相同路径相同内容允许
  - 相同路径不同内容 fail-closed
- 变量规则：
  - 继续沿用“兼容即合并，不兼容即失败”
- `AGENTS.md` 不按普通路径处理，接入专门 lane 预检查接口
- 更新 dry-run / json / text 输出中的 conflict payload
- 补 integration tests：
  - orbit + orbit disjoint
  - orbit + harness disjoint
  - harness + harness disjoint
  - 各类冲突 fail-closed

## Done When

- mixed install 第一阶段只允许 disjoint install，并且冲突诊断可观察。
- 不需要 `--overwrite-existing` 也能安全并存安装多个 disjoint install unit。
- 对现有 orbit-only install 路径不引入回归。

## Notes

- 依赖 `0088` 和 `0089`。
- `AGENTS.md` 的真正 block-lane 写入由 `0091` 负责，这里先冻结 conflict policy 与预检查接口。
- 已完成 harness template dry-run preview 的 mixed disjoint conflict analyzer，覆盖 member / path / variable / AGENTS lane 预检查，并锁定 text/json 输出契约。
