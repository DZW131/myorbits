# ISSUE-0121 v0.4 Brief Backfill Revision Matrix And Local Write

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-08

## Summary

硬化现有 `orbit brief backfill`，让它在 `runtime / source / orbit_template` 三态下共享同一语义，并冻结 local-only write contract：它只从当前 orbit block 回填 `meta.agents_template`，不触碰 export、publish、template source 或其他 provenance 文件。

## Scope

- 为现有 backfill 接入 revision matrix 合同
- 保持只提取当前 orbit block，不吸收其他 block 与 unmarked prose
- 保持 reverse replacement 只消费 `.harness/vars.yaml`
- 冻结 local-only write：只改 `.harness/orbits/<orbit-id>.yaml` 的 `meta.agents_template`
- 明确不自动触发 `save/publish`
- 补齐允许态、拒绝态与 fail-closed integration tests

## Done When

- `backfill` 在 `runtime / source / orbit_template` 下行为一致
- `plain / harness_template` 被稳定拒绝
- 回填只写本地 hosted OrbitSpec，不改根 `AGENTS.md`、install provenance 或模板分支
- reverse replacement 歧义、非法容器、缺失 block 等错误路径有完整测试

## Notes

- 对应 `docs/orbit_brief_lane_v0_4_technical_spec.md` 的第 3、6、7、8 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 4
- 依赖 ISSUE-0119

## Resolution

- `orbit brief backfill` 已在 `runtime / source / orbit_template` 三态下共享统一语义，并稳定拒绝 `plain / harness_template`。
- backfill 已冻结为 local-only write：只回填 hosted OrbitSpec 的 `meta.agents_template`，不自动触发 save/publish/export。
- 当前 orbit block 提取、reverse replacement、非法容器/缺失 block 等 fail-closed 场景已有集成测试覆盖。
