# ISSUE-0087 Install Variable Conflict Compatibility Policy

- Status: open
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

把 install / apply 路径里的同名变量冲突从“整个 `.harness/vars.yaml` 文件级冲突”收成“先看变量声明兼容性，再决定是否继续”的稳定合同，减少常见变量名重叠对工具使用的阻塞。

## Scope

- 在 install / apply 路径中显式区分：
  - 变量声明冲突
  - runtime 现有变量值复用
- 声明兼容时：
  - 自动继续安装
  - 默认复用 runtime 中已有值
  - 更新后的声明按“兼容即合并”规则收口
- 声明不兼容时：
  - 在写入 install 结果前 fail-closed
  - 输出至少包含变量名与冲突来源
- 与 mixed install 的变量规则对齐：
  - 描述相同可合并
  - 一边描述为空时可补齐
  - `required` 按 OR 合并
  - 两边描述都非空且不同则失败
- 明确 `--var-conflicts=prefer-runtime` 等逃生阀不在本 issue 范围内。

## Done When

- install / apply 前会做 per-variable 声明兼容检查，而不是只依赖 `.harness/vars.yaml` 文件级 diff。
- 以下场景有稳定测试：
  - runtime 中已有同名变量且声明兼容 -> 自动继续并复用已有值
  - runtime 中已有同名变量且声明不兼容 -> fail-closed
  - 新变量正常写入
  - diagnostics 至少包含变量名与来源
- 不引入变量 namespacing、自动 rename、交互式 rename。

## Notes

- 依赖 `docs/harness_mixed_install_technical_spec.md` 第 9 节。
- 这是 mixed install 之前就值得先做的 install kernel 收口项。
- `--var-conflicts=...` 若未来要做，应另开 issue，不和本批次混做。
