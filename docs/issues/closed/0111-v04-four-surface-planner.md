# ISSUE-0111 v0.4 Four-Surface Planner

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-08

## Summary

把当前以 projection 为主的 path planning 收口成统一的 four-surface planner，明确产出 `projection / orbit_write / export / orchestration` 四个消费面。

## Scope

- 固化 `meta / subject / rule / process` 的默认 role -> surface 映射
- 升级 `projection_plan.go`，同时产出四类 surface path set
- 保留必要的 member scope override，但不允许打破 `meta` companion 语义
- 为后续 status、ledger、commit、publish、brief 命令提供统一 planner 输出

## Done When

- planner 可稳定生成四个 surface 的 path 集合
- role -> surface matrix 有完整单测
- 新 planner 可被下游命令消费，而不依赖单个模糊的 “in scope” 布尔值

## Notes

- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 5、8 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 3
- 依赖 ISSUE-0110

## Resolution

- `projection_plan.go` 已稳定产出 `projection / orbit_write / export / orchestration` 四个 surface，并有 role -> surface matrix tests 覆盖。
- `path_classification.go`、status snapshot、ledger snapshot、`enter/files/diff/log/commit/restore/template save/publish/brief materialize/backfill` 等正式命令链路已经切到明确 surface 消费。
- `harness template save` 的 member candidate 现在按 `export` surface 取文件，不再回退到 legacy `OwnedPaths` 语义；`orbit validate` 也已改为直接消费 planner，member-schema 的 process projection 不再被漏算。
- `ScopeSet` 兼容结构仍作为内部桥接保留，但已不再驱动 v0.4 正式命令行为；后续若要删除兼容桥，应另开清理 issue，而不是继续保持本 issue 为 open。
