# ISSUE-0116 v0.4 Runtime Export And Writeback

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-09

## Summary

把 runtime 中对 orbit 的优化、保存与发布正式收口为 export/writeback lane，确保 runtime -> template 的反写只消费 `export` surface，而不是误用 projection-visible 文件或整个 `AGENTS.md` 容器。

## Scope

- 明确 `orbit template save` 只消费 `export` surface
- 定义 runtime -> orbit template writeback / publish 主路径
- 同步更新 runtime provenance 与 member source 信息
- 拦截把 `subject`、容器态 `AGENTS.md` 或其他非 export 内容误导出的情况

## Done When

- runtime 中的 orbit 改动可通过正式 export/writeback 进入 template
- `orbit template save` 的 payload 与 `export` surface 严格一致
- runtime writeback 有独立 integration tests 与 acceptance smoke

## Notes

- 对应 `docs/orbit_v0_4_prd.md` 的第 6、7 节
- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 8、9 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 6
- 依赖 ISSUE-0111、ISSUE-0113 与 ISSUE-0122

## Progress

- `orbit template save` 已稳定只消费 `export` surface，不再把 `subject`、runtime 根 `AGENTS.md` 或 projection-only 内容误带入 template payload。
- runtime 安装来源现在已开始进入正式 writeback 主路径：当 `.harness/installs/<orbit-id>.yaml` 存在时，`orbit template save` 允许省略 `--to`，并默认回写到 install record 的 `template.source_ref`。
- runtime 的 versioned truth 现在已能保留完整 member source taxonomy：`.harness/manifest.yaml kind=runtime` 不再把 `install_orbit` / `install_bundle` 压扁成单个 `install`，bundle install 与 orbit install 的 provenance 分层已进入正式 manifest contract。
- `orbit template save` 现在会在命令入口 fail-closed 校验当前 revision 必须是 `runtime`，不再允许 plain/source/template repo 直接走 runtime writeback lane。
- `orbit template save` 现在只会为 `source=install_orbit` 的 runtime member 复用 `.harness/installs/<orbit-id>.yaml` 推断默认 writeback target；`manual` / `install_bundle` member 即使残留 install record，也必须显式提供 `--to`，避免 stale provenance 劫持 writeback lane。
- quickstart acceptance 现已覆盖 clean runtime 与 migrated runtime 两条 runtime writeback 基本路径：既验证 installed runtime 的默认 writeback，也验证残留兼容 `.orbit/config.yaml` 的 migrated runtime 不会把 legacy config 卷回 template payload。
- 对应 CLI help 已同步补齐该入口，runtime 下“安装后优化再回写”的主路径不再只靠手动记忆目标 branch。

## Resolution

- `orbit template save` 现在已经稳定地只消费 `export` surface，并在命令入口 fail-closed 要求当前 revision 必须是 `runtime`，不会再把 `subject`、projection-only 文件或 runtime 根 `AGENTS.md` 误导出到 template payload。
- runtime writeback 主路径已经完成：`source=install_orbit` 的 runtime member 可以省略 `--to` 并复用 `.harness/installs/<orbit-id>.yaml` 里的 `template.source_ref`，而 `manual` / `install_bundle` member 必须显式指定目标，避免 stale provenance 劫持 writeback lane。
- runtime versioned truth 已经保留完整 member source taxonomy：`.harness/manifest.yaml kind=runtime`、`.harness/installs/*.yaml` 与 `.harness/bundles/*.yaml` 现在都能稳定区分 `manual / install_orbit / install_bundle`。
- clean runtime 与 migrated runtime 的 runtime writeback acceptance smoke 已完成，且验证了 legacy `.orbit/config.yaml` 不会被卷回 template payload。

因此本 issue 按 v0.4 Phase 6 的主链完成标准关闭。剩余关于 writeback 后 install provenance 仍保留 install-time 语义、不会自动刷新成“最后一次回写提交”的一致性硬化，转入 `docs/technical-debt.md` 跟踪，而不再继续阻塞 runtime export / writeback lane。
