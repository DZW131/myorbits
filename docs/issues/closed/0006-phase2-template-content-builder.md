# ISSUE-0006 Implement Template Content Builder

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

实现 template content builder，从 runtime repo 和目标 orbit 的 user scope 构建模板候选文件树，并正确注入 companion orbit definition，同时严格排除不应进入模板 branch 的控制面和运行态文件。

## Scope

- 输入 runtime repo root、orbit id、resolved user scope、orbit definition 和 bindings。
- 只读取 orbit user scope 下的用户文件，并加入 `.orbit/orbits/<orbit-id>.yaml`。
- 明确排除 `.orbit/config.yaml`、`.orbit/vars.yaml`、`.orbit/installs/*` 和 `.git/orbit/state/*`。
- 读取规则与 technical spec 保持一致：可见路径读 worktree，被 sparse 隐藏的 clean tracked path 回退到 index 或 `HEAD` snapshot。
- 输出模板候选文件树、替换摘要和 ambiguity 集合。
- 为 scope 边界、隐藏路径回退和二进制文件复制增加测试。

## Done When

- scope 外文件不会进入模板候选树。
- `.orbit/config.yaml`、`.orbit/vars.yaml` 和 install records 不会进入模板。
- 当前 orbit definition path 会作为 companion file 进入模板。
- 不依赖 `.git/orbit/state/resolved_scope/*` cache 也能稳定构建结果。
- 单测覆盖可见文件、被隐藏 clean tracked 文件和二进制文件复制。

## Notes

- 对应 `docs/context/orbit_phase2_development_plan.md` 的 Phase 2A-0 / 任务 5。
- 主要依据：`docs/context/orbit_phase2_technical_spec.md` 的 4.6、8.1、11.2、11.3。
- 2026-03-21: 已实现 template content builder，支持基于 user scope 构建候选模板树、注入 companion orbit definition、过滤 `.orbit/config.yaml` / `.orbit/vars.yaml` / install records、对 hidden clean tracked path 回退到 `HEAD`、复制二进制文件并汇总 replacement/ambiguity 结果，相关单测已补齐。
