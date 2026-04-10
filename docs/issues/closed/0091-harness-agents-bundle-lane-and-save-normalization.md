# ISSUE-0091 Harness AGENTS Bundle Lane And Save Normalization

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

为 harness template install 增加 bundle-level `AGENTS.md` block lane，并让 `harness template save` 在导出时规范化 / 剥离 runtime block markers，避免 mixed install 后导出回路把 runtime markers 卷回模板。

## Scope

- 设计 harness bundle block identity：
  - block identity = `harness_id`
- 设计并实现 runtime `AGENTS.md` 中 bundle block 的写入 / 替换 / 删除原语。
- mixed install 中：
  - orbit install 继续用 `orbit_id` block
  - harness install 改用 `harness_id` block
- `AGENTS.md` lane 冲突按 block identity 判定：
  - 新 identity append
  - same install unit replace 时 replace
  - 不同 install unit 试图改写已有 block fail-closed
- 更新 `harness template save`：
  - 读取 runtime `AGENTS.md`
  - 规范化并剥离 runtime block markers
  - 模板根 `AGENTS.md` 保留 payload，不保留 runtime markers
- 补 tests：
  - orbit block + bundle block 并存
  - bundle replace
  - mixed runtime 导出时 marker strip

## Done When

- harness template install 不再把根 `AGENTS.md` 当普通文件覆盖。
- mixed runtime 的 `AGENTS.md` 可容纳 orbit block 与 bundle block。
- `harness template save` 导出结果不含 runtime block markers。

## Notes

- 依赖 `0090`。
- 实现时需要谨慎复用现有 orbit AGENTS parser / marker 原语，避免两套 marker 语义发散。
- 已完成：
  - runtime AGENTS bundle block append / replace / remove primitives
  - `harness template save` 的 runtime marker strip / payload normalization
- 已完成：
  - harness template install 写路径接入 bundle lane
- `same install unit replace` 的 bundle block replace/remove 接线继续留在 `0092`。
