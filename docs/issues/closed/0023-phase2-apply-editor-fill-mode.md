# ISSUE-0023 Phase 2 Apply Editor Fill Mode

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement `orbit template apply --editor` so missing required bindings can be filled through an editor-backed skeleton flow instead of only line-based interactive prompts.

## Scope

- Extend `orbit template apply` bindings resolution with `--editor`.
- For unresolved required variables:
  - generate a temporary bindings skeleton
  - open the configured editor
  - read the edited YAML back
  - merge the result after `--bindings` and repo `.orbit/vars.yaml`
- Keep the existing precedence contract:
  - `--bindings` > repo vars > editor
- Only ask the editor flow for still-missing variables.
- Return clear errors for:
  - invalid YAML
  - unsupported schema
  - still-missing required values after editing

## Done When

- `orbit template apply <source> --editor` succeeds when the only blocker is missing required bindings.
- Existing explicit values are not silently overwritten by editor-provided values.
- The edit flow does not mutate the runtime repo before the real apply write path runs.
- Local and remote apply paths both gain coverage for the editor-backed fill mode.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-2 / task 2.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 10.3, 12.2, and 17.4.
- This issue is closely related to technical-debt entry `0010`; editor command parsing hardening is tracked separately in ISSUE-0027.
- 2026-03-21: 已实现 `orbit template apply --editor`，仅对仍缺失的必填变量生成临时 bindings skeleton，调用 `EDITOR` 后读取编辑结果并按既有优先级合并；本地和远程 apply 路径都已补齐 coverage，非法 YAML 和“保留空值未填写”都会稳定失败。
