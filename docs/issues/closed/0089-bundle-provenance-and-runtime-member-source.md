# ISSUE-0089 Bundle Provenance And Runtime Member Source

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

为 harness template install 引入 bundle-level provenance 与更细粒度的 runtime member source，避免 mixed install 时 ownership / drift / overwrite 无法归因。

## Scope

- 新增 bundle install record schema 与存储位置。
- 设计并实现 runtime member source 扩展，至少区分：
  - `manual`
  - `install_orbit`
  - `install_bundle`
- bundle install record 至少记录：
  - `harness_id`
  - source kind / repo / ref / commit
  - member ids
  - applied_at
  - includes_root_agents
- 若实现成本合理，顺手记录 bundle owned paths，为后续 replace 做准备。
- 更新 `harness inspect` / `harness check` 所需的最小读取逻辑。
- 补 schema unit tests 和 repo-local IO tests。

## Done When

- harness template install 有独立 provenance 存储，不复用 orbit install record。
- runtime members 可区分 orbit install 与 bundle install。
- 现有 orbit install 路径保持兼容。

## Notes

- 依赖 `0088`。
- 这是 mixed install 与 same install unit replace 的核心控制面基础。
- 已实现 bundle install record、`install_orbit/install_bundle` runtime member source、以及 `harness inspect/check` 的最小 bundle 读取逻辑。
