# ISSUE-0002 Implement Revision Branch Classifier

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

实现一个基于文件合同而不是 branch 名称的 revision classifier，用来把任意 revision 判定为 `template`、`runtime` 或 `plain`，并输出原因解释，供后续 `branch status`、`branch inspect`、`branch list` 复用。

## Scope

- 提供类似 `ClassifyRevision(repo, rev)` 的底层接口。
- 按 `.orbit/template.yaml`、`.orbit/config.yaml`、`.orbit/orbits/*.yaml` 的存在性和合法性进行分类。
- 支持当前已 checkout 分支和未 checkout revision 的一致判定。
- 产出分类结果结构体及解释字段。
- 为主要分类路径和失败路径补单测。

## Done When

- `template` 判定仅依赖合法 `.orbit/template.yaml`。
- `runtime` 判定要求不是 template，且存在合法 `.orbit/config.yaml` 和至少一个合法 orbit definition。
- `plain` 作为稳定兜底结果，而不是直接报错。
- branch name 改变不会影响分类结果。
- 单测覆盖当前 branch 与任意 revision 的一致性。

## Notes

- 对应 `docs/context/orbit_phase2_development_plan.md` 的 Phase 2A-0 / 任务 2。
- 主要依据：`docs/context/orbit_phase2_technical_spec.md` 的 14.1 和 `docs/context/orbit_phase2_development_plan.md` 的 5.3 任务 2。
- 2026-03-21: 已实现 `branchinfo.ClassifyRevision(repo, rev)`，按合法 `.orbit/template.yaml`、`.orbit/config.yaml` 和 `.orbit/orbits/*.yaml` 对任意 revision 分类为 `template`、`runtime` 或 `plain`，并补齐 temp repo 单元测试。
