# ISSUE-0088 Harness Template Install Source And Preview

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

为 `harness install` 增加 harness template install unit 的识别、解析与 preview 基线，让命令可以区分 orbit template 与 harness template，并给后续 mixed install 提供统一输入模型。

## Scope

- 扩展 install source resolution：
  - 本地 revision 可识别 `.harness/template.yaml`
  - 远程 Git URL 可识别 `.harness/template.yaml`
- 定义 install preview 的 template kind：
  - `orbit_template`
  - `harness_template`
- 保持 orbit template install 现有行为不变。
- harness template install 第一版只做 preview / source resolution / manifest 校验，不在本 issue 内完成最终 runtime 写入。
- 补单元测试与集成测试：
  - 本地 harness template branch preview
  - 远程 harness template branch preview
  - source/template/plain 分支误识别防回归

## Done When

- install source 层能稳定区分 orbit template 与 harness template。
- `harness install --dry-run` 能对 harness template branch 给出正确 preview。
- 现有 orbit template install preview 测试无回归。

## Notes

- 依赖当前 `harness template save` 产出的 `.harness/template.yaml` 合同。
- 这是 mixed install 的基础项，后续 issue 都依赖本项。
