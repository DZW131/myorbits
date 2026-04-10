# ISSUE-0123 v0.4 Harness Author Multi-Orbit Flow Hardening

- Status: open
- Priority: high
- Owner:
- Created: 2026-04-09
- Updated: 2026-04-09

## Summary

收敛并硬化 harness 作者的主流程，使“安装多个 orbit -> 统一变量绑定 -> 编排根 `AGENTS.md` -> 验证 worker 入口”成为当前 v0.4 下最清晰、最轻量、最 agent-friendly 的标准路径。

## Scope

- 明确 harness 作者主流程的产品合同与命令顺序
- 把 `install lane`、`bindings lane`、`orchestration lane` 的边界写清楚
- 识别当前流程中的高认知负担点
- 为后续薄封装命令预留清晰的职责边界
- 作为 ISSUE-0124、ISSUE-0125、ISSUE-0126 的总揽 issue

## Done When

- harness 作者可以不理解底层实现细节，也能完成多 orbit runtime 组合
- 共享 `.harness/vars.yaml` 被明确为 runtime 级变量宿主
- 根 `AGENTS.md` 被明确为宏观编排入口与 orbit brief block 容器
- 后续命令级优化有清晰拆分，不再混成一个“大而全”的 harness authoring 功能

## Notes

- 相关文档：
  - `docs/harness_author_guide.md`
  - `docs/quickstart.md`
  - `docs/orbit_brief_lane_v0_4_technical_spec.md`
- 推荐按下面 3 张子 issue 拆分实现：
  - ISSUE-0124 `harness bindings plan`
  - ISSUE-0125 `harness install batch`
  - ISSUE-0126 `harness agents compose`
