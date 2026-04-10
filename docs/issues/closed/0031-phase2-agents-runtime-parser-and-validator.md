# ISSUE-0031 AGENTS Runtime Parser And Validator

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-23
- Updated: 2026-03-23

## Summary

为运行态 `AGENTS.md` 建立专用 parser / validator 原语，统一 save、apply、validate 三条链路对 marker block 的理解。

## Scope

- 解析运行态 `AGENTS.md` 为有序片段序列：
  - unmarked span
  - marker block
- 支持 V0.2 冻结 marker 语法：
  - `<!-- orbit:begin orbit_id="..." -->`
  - `<!-- orbit:end orbit_id="..." -->`
- 校验并 fail-closed：
  - begin / end 不配对
  - `orbit_id` 不一致
  - nested block
  - duplicate same-orbit block
  - 无法解析属性
- 提供 helper，把模板态 payload 包装为运行态 marker block。

## Done When

- parser 能稳定返回片段序列和 orbit block 元信息。
- 所有异常 marker 场景都以结构化错误 fail-closed。
- 生成 runtime marker block 的 helper 与 parser 使用同一套合同。
- 有完整单元测试覆盖正常与异常场景。

## Notes

- 这一层应保持纯逻辑，不依赖 command / git / state。
- 本 issue 不接命令，也不直接改 `template save` / `template apply` 行为。
- 参考文档：
  - `docs/context/orbit_agents_md_development.md`
- Completed:
  - 新增了 runtime `AGENTS.md` parser / validator 原语。
  - 冻结并测试了 begin/end marker 语法、duplicate block、nested block、mismatch、malformed marker 等 fail-closed 行为。
  - 新增了 runtime block 包装 helper，供后续 apply 路径复用。
