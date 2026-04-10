# ISSUE-0049 Phase 3B-2 Overwrite And Reinstall Contract

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-26
- Updated: 2026-03-26

## Summary

实现 `harness install --overwrite-existing` 的正式覆盖更新合同，并把 repeated install / remove 后 reinstall 的行为收紧到一致、可测试的 fail-closed 语义。

## Scope

- `harness install`
  - 同一 `orbit-id` 默认继续失败
  - 只有显式 `--overwrite-existing` 才进入覆盖更新路径
  - 首阶段不支持 `--as`
- install-backed member 已存在时：
  - 未传 `--overwrite-existing` 时 fail-closed
  - 传入 `--overwrite-existing` 时进入 overwrite 路径
- remove 后 install record 仍存在时：
  - 同一 `orbit-id` 的 reinstall 不视为全新安装
  - 继续按 overwrite 路径处理
- manual member 仍占用同一 `orbit-id` 时继续 fail-closed，不自动替换

## Done When

- `harness install --overwrite-existing` CLI 可用。
- remove 后同 ID reinstall 进入 overwrite 路径并有集成测试覆盖。
- manual member / install-backed member / install record 三类占用关系的错误语义稳定。

## Notes

- 对应 `docs/harness_centric_runtime_development_plan.md` 的 9.3 / 任务 1。
- 相关技术合同见 `docs/harness_centric_runtime_technical_spec.md` 的 6.7、8.2、8.4。
- Completed:
  - `harness install --overwrite-existing` 已开放，并保留默认重复安装 fail-closed。
  - remove 后只要 install record 仍在，同一 `orbit-id` 的 reinstall 会继续进入 overwrite 路径。
  - install-backed member 在 overwrite 成功后复用或重建为单一 `source=install` member，不会重复追加。
