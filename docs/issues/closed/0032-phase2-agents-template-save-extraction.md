# ISSUE-0032 AGENTS Template Save Extraction

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-23
- Updated: 2026-03-23

## Summary

让 `template save` 能按 AGENTS 专用规则从运行态 `AGENTS.md` 提取模板态 payload，并把结果写入模板 branch 根目录 `AGENTS.md`。

## Scope

- 在 `template save` 中接入 runtime AGENTS parser。
- 按冻结语义抽取 payload：
  - 保留所有 unmarked span
  - 保留当前 orbit block 的内部内容
  - 删除其他 orbit block
  - 保持原始顺序
- 若缺少当前 orbit block：
  - 发出 warning
  - 继续，仅保留 unmarked span
- 若提取结果为空：
  - 不生成 `shared_files` entry
  - 不写模板态根目录 `AGENTS.md`
- `--edit-template` 仍以模板态 payload 为编辑对象，不回写 runtime `AGENTS.md`。

## Done When

- `template save` 能为 shared `AGENTS.md` 生成模板态 payload。
- payload 不会被当普通 owned file 重复写入。
- warning 行为稳定，且不静默猜测无 marker 内容的归属。
- 本地 save 路径和 `--edit-template` 路径都有测试保护。

## Notes

- 用户已接受 `include_unmarked_content: true` 的默认风险；该行为应被测试固定，而不是留给后续“顺手优化”。
- 参考文档：
  - `docs/context/orbit_agents_md_development.md`
  - `docs/technical-debt.md`
- Completed:
  - `template save` 现在会按 AGENTS 专用规则提取 payload，并在非空时生成 shared `AGENTS.md` 与 `shared_files` manifest entry。
  - 缺少当前 orbit marker 会 warning 后继续，仅保留 unmarked 内容。
  - `--edit-template` 现在可直接编辑模板态根目录 `AGENTS.md` payload，且不会回写 runtime `AGENTS.md`。
