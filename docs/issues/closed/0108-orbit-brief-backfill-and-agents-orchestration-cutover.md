# ISSUE-0108 Orbit Brief Backfill And AGENTS Orchestration Cutover

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-06
- Updated: 2026-04-07

## Summary

把 orbit 级 `AGENTS.md` 从模板 payload 文件模型切到结构化 brief + orchestration 派生模型：运行态根 `AGENTS.md` 只作为整个 harness 的入口容器，orbit 自身进入其中的内容改由 OrbitSpec 中的 `meta.agents_template` 持有，并通过显式命令 `orbit brief backfill` 把运行态当前 orbit block 回填为真相源。

## Scope

- 在 OrbitSpec / hosted OrbitSpec schema 中冻结 `meta.agents_template`：
  - 保存 variableized 的 orbit brief
  - 作为进入运行态 `AGENTS.md` 当前 orbit block 的首选真相源
- 新增显式命令：
  - `orbit brief backfill`
  - 从运行态根 `AGENTS.md` 中只提取当前 orbit block
  - 反向变量化后写回 OrbitSpec
- 切换 orbit template lane：
  - orbit template branch 不再写根 `AGENTS.md`
  - orbit template save / publish / apply / install / replay 改为消费 `meta.agents_template` 或 orchestration output
  - 不提供对 legacy orbit template 根 `AGENTS.md` payload 格式的 backward compatibility
- 保持 harness template 的根 `AGENTS.md` lane 独立存在，不与 orbit template 的 cutover 混写。
- 更新相关 validate / drift / replay / source-loader 测试与文档。

## Done When

- 运行态根 `AGENTS.md` 被稳定定义为容器，而不是 orbit 真相源。
- OrbitSpec 中已有 `meta.agents_template` 合同，并被相关命令消费。
- `orbit brief backfill` 能 fail-closed 地把当前 orbit block 回填成结构化真相源。
- orbit template branch 不再包含根 `AGENTS.md`。
- 现有 harness template 根 `AGENTS.md` lane 不回归。

## Notes

- 设计依据：
  - `docs/orbit_member_runtime_technical_spec.md`
  - `docs/orbit_member_runtime_development_plan.md`
- 当前 single-control-plane steady-state runtime OrbitSpec host 是 `.harness/orbits/<orbit-id>.yaml`。
- 这项不要求兼容旧的 orbit template 根 `AGENTS.md` payload 格式。
