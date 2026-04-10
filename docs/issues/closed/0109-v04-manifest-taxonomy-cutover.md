# ISSUE-0109 v0.4 Manifest Taxonomy Cutover

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-07
- Updated: 2026-04-07

## Summary

把 `.harness/manifest.yaml` 收口为 v0.4 唯一 revision identity 控制面，完整覆盖 `runtime / source / orbit_template / harness_template` 四种正式 revision kind，并删除主链路对旧 marker 文件的依赖。

## Scope

- 扩展 manifest schema、validator、codec 与 per-kind field contract
- 统一 branch classifier、inspect、root resolution、bootstrap writer
- 明确 `plain` 只作为“无有效 manifest”时的 classifier 结果
- 删除主链路对 `.orbit/source.yaml`、`.orbit/template.yaml`、`.harness/runtime.yaml`、`.harness/template.yaml` 的依赖
- 若需要迁移，定义显式一跳迁移路径，不保留长期 dual-read / dual-write

## Done When

- `branch inspect` 可只靠 `.harness/manifest.yaml` 稳定识别四类正式 revision kind
- `harness init/create`、template bootstrap、source bootstrap 全部写 manifest 而不是旧 marker 文件
- 主链路命令不再要求旧 marker 文件存在
- manifest 的 schema / validation / integration tests 覆盖四类 revision kind

## Notes

- 对应 `docs/orbit_v0_4_technical_spec.md` 的第 4、10 节
- 对应 `docs/orbit_v0_4_development_plan.md` 的 Phase 1
- 这是后续 hosted OrbitSpec 与命令 cutover 的共同前置项

## Progress

- `kind=source` / `runtime` / `orbit_template` / `harness_template` 的 manifest schema 与 branch classify/inspect 主链路已落地。
- `harness init` / `harness create` 现在只初始化 `.harness/manifest.yaml` 与 `.harness/orbits/`，不再额外生成 `.harness/runtime.yaml`。
- `ResolveRoot` 与 `harness check` 已切到 manifest-first 语义：无效 legacy `.harness/runtime.yaml` 不再污染 runtime 主入口，schema 诊断也开始指向 `.harness/manifest.yaml`。
- runtime member mutation、orbit install、harness-template install 与 `harness add/remove --json` 现在都开始把 `.harness/manifest.yaml` 作为正式写入/返回路径，不再把 `.harness/runtime.yaml` 暴露成主链路结果。
- repo-root 级别的 runtime compatibility view 已切成 manifest-backed：`LoadRuntimeFile` / `WriteRuntimeFile` 现在都只围绕 `.harness/manifest.yaml` 工作，坏的 legacy `.harness/runtime.yaml` 不再影响 `harness add`、`harness template save` 等主链路命令。
- shared harness path helpers 里已经不再暴露 `.harness/runtime.yaml`；legacy runtime path 只剩 explicit codec / targeted tests 会直接按字面路径访问。

## Resolution

- `.harness/manifest.yaml` 已经成为 `runtime / source / orbit_template / harness_template` 四类正式 revision kind 的 canonical identity document，主链 branch classify / inspect / bootstrap / root resolution 都已 manifest-first。
- `harness init/create`、runtime member mutation、orbit install、harness-template install、`harness add/remove --json` 的正式写入与用户可见路径，都已收口到 `.harness/manifest.yaml`，不再把 `.harness/runtime.yaml` 暴露成主链结果。
- repo-root 级 runtime compatibility view 已 manifest-backed，坏的 legacy `.harness/runtime.yaml` 不再污染 `harness check`、`ResolveRoot`、`harness add`、`harness template save` 等主入口。
- shared harness path helpers 已不再暴露 `.harness/runtime.yaml`；剩余 legacy runtime 引用已收敛到 explicit compatibility codec 与 targeted tests，不再阻塞 v0.4 manifest taxonomy 主线完成。
- `.orbit/template.yaml` / `.harness/template.yaml` 的剩余 payload bridge 已转入后续 host/payload cutover 与技术债跟踪，不再属于 revision identity taxonomy 的未完成项。

因此本 issue 按 v0.4 Phase 1 的主链完成标准关闭；残余兼容尾项转入 `docs/technical-debt.md` 跟踪，而不再继续阻塞 manifest taxonomy cutover。
