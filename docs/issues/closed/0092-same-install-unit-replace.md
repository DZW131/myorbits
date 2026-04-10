# ISSUE-0092 Same Install Unit Replace

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

实现 mixed install 第二阶段：只允许同 install unit 覆盖同 install unit，不开放跨 install unit 的危险冲突覆盖。

## Scope

- orbit install unit：
  - 继续沿用当前 `orbit_id` + `--overwrite-existing` 语义
- harness install unit：
  - 以 `harness_id` 为 replace identity
  - 只允许 bundle 覆盖同 bundle
- 不支持：
  - bundle 内单 member replace
  - bundle 覆盖散装 orbit install
  - orbit 覆盖 bundle 内单 member
- 结合 `0089` 的 provenance 与 `0091` 的 `AGENTS.md` lane：
  - bundle owned files cleanup
  - bundle block replace
  - replace 后 `harness check` / drift 基线
- 补 overwrite-focused integration tests 与 acceptance smoke。

## Done When

- `--overwrite-existing` 在 mixed install 场景下只允许 same install unit replace。
- 不同 install unit 之间的 overwrite 仍然 fail-closed。
- bundle replace 后 cleanup / provenance / `AGENTS.md` lane 行为稳定。

## Notes

- 依赖 `0089`、`0090`、`0091`。
- 这项明确不实现危险覆盖模式。
