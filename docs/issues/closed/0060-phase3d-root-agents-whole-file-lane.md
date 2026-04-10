# ISSUE-0060 Phase 3D Root Agents Whole File Lane

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

为 `harness template save` 实现根目录 `AGENTS.md` 的 whole-file lane，把当前 runtime root 的 `AGENTS.md` 按普通模板文件处理后写入 harness template branch。

## Scope

- runtime root 存在 `AGENTS.md` 时：
  - 整文件读取
  - 用 `.harness/vars.yaml` 做 replacement
  - 原样保留 marker / 注释 / 分隔结构
  - 写入 harness template branch 根目录 `AGENTS.md`
  - 在 `.harness/template.yaml` 中写 `includes_root_agents=true`
- runtime root 不存在 `AGENTS.md` 时：
  - 不生成根 `AGENTS.md`
  - `includes_root_agents=false`

## Done When

- root `AGENTS.md` replacement 测试通过。
- marker/comment preservation 测试通过。
- lane 行为与普通 candidate merge 解耦，不误复用 v0.2 shared-file lane。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 11.3 / 任务 3。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 8.7。
- Completed:
  - `BuildRootAgentsTemplateFile` 已落地，根目录 `AGENTS.md` 按普通模板文件做 whole-file replacement。
  - marker、注释和分隔结构会原样保留，不再走 v0.2 orbit block 抽取逻辑。
  - 文件缺失时稳定返回 `includes_root_agents=false`，相关单测已覆盖。
