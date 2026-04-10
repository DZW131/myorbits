# ISSUE-0110 v0.4 Hosted OrbitSpec Cutover

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-09

## Summary

把 `.harness/orbits/<orbit-id>.yaml` 切成所有正式 revision kind 下唯一的 authored OrbitSpec host，并让 Orbit 主链路命令全部围绕 hosted definitions 工作。

## Scope

- 切 Orbit loader、writer、validator 到 `.harness/orbits/*.yaml`
- 统一 `orbit add / show / list / validate` 的 hosted 行为
- 冻结 `meta.file` 与 hosted path 的一致性规则
- 删除新主链路对 `.orbit/orbits/*.yaml` 与 `.orbit/config.yaml` 的依赖
- 如确实需要，补显式 migration tool 或 fixture rewrite script

## Done When

- `runtime / source / orbit_template / harness_template` 四类 revision 都以 `.harness/orbits/*.yaml` 作为 authored truth
- `orbit add / show / list / validate` 在 hosted model 下通过 integration tests
- 新 fixture 与 acceptance path 已不再依赖旧 Orbit host

## Notes

- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 5 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 2
- 依赖 ISSUE-0109

## Progress

- `source / orbit_template / harness_template / branch inspect` 已经全面走 hosted-first 的 `.harness/orbits/*.yaml`，并在 source publish 上 fail-closed 拒绝 legacy orbit host。
- orbit template branch manifest 现在开始承载 template `variables`，local/remote template consumer 在缺失 `.orbit/template.yaml` 时也能继续从 `.harness/manifest.yaml` 解析模板变量合同。
- `orbit template save` 写出的 orbit template branch 已经可以只保留 `.harness/manifest.yaml` 这份 canonical branch manifest；`apply / publish / remote resolve` 主链不再要求 saved branch 一定携带 `.orbit/template.yaml`。
- `TemplateSavePreview` 内部也不再生成 legacy `.orbit/template.yaml` bytes；对应 remote selection / local apply 夹具已经开始切到 branch-manifest-only 口径。
- low-level `git.WriteTemplateBranch` 已不再默认回退到 `.orbit/template.yaml`；所有主链写分支路径都必须显式声明 manifest path。
- `apply / remote enumerate` 在 branch manifest 已有效时，已经收口到 branch-manifest-only；legacy `.orbit/template.yaml` 不再参与这两条消费链的主解析与校验，即使它包含无效内容或遗留 `shared_files` 也不会继续拖垮主链。
- direct `orbit template publish` 也已经开始以 branch manifest 为准；working tree 中遗留的 `.orbit/template.yaml` 不再作为 direct publish 的阻断前置。
- source `orbit template publish` 的 no-op 判断也已经收口到 branch manifest + exported files；已发布 branch 中单独漂移的 legacy `.orbit/template.yaml` 不会再触发一次无意义重写。
- `orbit template init-source` 也已经开始把遗留 `.orbit/template.yaml` 当作显式迁移输入处理；命令会删除这份旧 marker，而不是继续把 source 初始化卡在 legacy control plane 上。
- source `orbit template publish` 本身也已经不再因为 source 分支里残留 `.orbit/template.yaml` 而 fail-closed；legacy marker 不会再阻断 source -> template 的正式发布路径。
- harness-template install source loader 也已经开始忽略遗留 `.orbit/template.yaml`；只要 branch manifest 和 harness template payload 有效，这份旧 orbit-template marker 就不会再阻断 install / enumerate。

## Resolution

- `runtime / source / orbit_template / harness_template` 四类正式 revision 的主链 loader / consumer / publish / install 路径，已经都以 `.harness/orbits/*.yaml` 与 `.harness/manifest.yaml` 为 canonical hosted truth。
- `apply / remote resolve / direct publish / source publish / init-source / harness-template install` 都已不再被遗留 `.orbit/template.yaml` 或 legacy orbit host 阻断。
- 剩余 `.orbit/template.yaml` 引用已经收敛到 compatibility-only parser/codec、显式 legacy marker 测试夹具、以及少量“旧 marker 不应再影响行为”的回归断言。
- 因此本 issue 按 v0.4 主链完成标准关闭；残余兼容尾项转入 `docs/technical-debt.md` 跟踪，而不再继续阻塞 Phase 2 host cutover。
