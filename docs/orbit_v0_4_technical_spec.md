# Orbit v0.4 Unified State Model Technical Spec

版本：v0.4
状态：proposal，可进入实现拆分
对应 PRD：`docs/orbit_v0_4_prd.md`
对应开发计划：`docs/orbit_v0_4_development_plan.md`
关联文档：
- `docs/orbit_brief_lane_v0_4_technical_spec.md`
- `docs/orbit_template_authoring_guide.md`
- `docs/orbit_member_runtime_technical_spec.md`
- `docs/testing-strategy.md`
- `docs/context/orbit_v0_4_design_archive.md`
- `docs/context/orbit_state_and_workflow_unification.md`

---

## 1. 文档目标

本文档把 v0.4 的统一状态模型收口为可实现的技术合同，重点回答：

1. 版本化与本地状态各自的 canonical host 是什么；
2. `.harness/manifest.yaml` 需要承载哪些 revision identity；
3. OrbitSpec 在各 revision kind 中如何落盘；
4. `projection / orbit_write / export / orchestration` 应如何被统一建模；
5. 命令应消费哪一个 surface；
6. 兼容与迁移策略应如何处理。

---

## 2. 总体技术原则

整个 v0.4 实现遵守以下原则：

1. 只有一个版本化控制平面：`.harness/`
2. 只有一个 repo-local ledger 平面：`.git/orbit/state/`
3. 根 `AGENTS.md` 永远不是 Orbit authored truth
4. 普通 Git 工作流不被 Orbit 命令面重新定义
5. 新实现优先服务 clean-break 目标态，不保留长期 compatibility lane
6. 若需要迁移，只允许显式一跳迁移，不允许长期 dual-read / dual-write

---

## 3. Canonical Storage Contract

### 3.1 版本化存储

| 路径 | canonical contract |
| --- | --- |
| `.harness/manifest.yaml` | revision identity |
| `.harness/orbits/<orbit-id>.yaml` | authored OrbitSpec |
| `.harness/vars.yaml` | runtime variable bindings |
| `.harness/installs/<orbit-id>.yaml` | orbit install provenance |
| `.harness/bundles/<harness-id>.yaml` | bundle / harness template provenance |

### 3.2 repo-local 存储

| 路径 | canonical contract |
| --- | --- |
| `.git/orbit/state/current_orbit.json` | 当前 projection target |
| `.git/orbit/state/orbits/<orbit-id>/file_inventory.json` | orbit file inventory ledger |
| `.git/orbit/state/orbits/<orbit-id>/runtime_state.json` | orbit-local runtime ledger |
| `.git/orbit/state/orbits/<orbit-id>/git_state.json` | orbit-local git ledger |

### 3.3 artifact

| 路径 | contract |
| --- | --- |
| 根 `AGENTS.md` | orchestration materialized container |

artifact 不是 authored truth。

---

## 4. Manifest Contract

### 4.1 通用要求

`.harness/manifest.yaml` 必须满足：

1. `kind` 必填；
2. `kind` 只允许：
   - `runtime`
   - `source`
   - `orbit_template`
   - `harness_template`
3. 能独立完成 revision classification；
4. 不依赖其他 marker 文件才能判断 revision kind；
5. 每种 `kind` 只允许自己的字段集合；
6. mixed-kind fields 必须 fail-closed。

### 4.2 `runtime`

`kind=runtime` 至少需要表达：

- harness identity
- runtime members
- member source / provenance summary

建议结构：

```yaml
version: 1
kind: runtime
runtime:
  harness_id: main
  name: Main Harness
  members:
    - orbit_id: docs
      source: install_orbit
    - orbit_id: planning
      source: manual
```

### 4.3 `source`

`kind=source` 至少需要表达：

- 该 source 服务哪个 `orbit_id`
- 默认 publish target 是什么
- source-only authoring metadata

建议结构：

```yaml
version: 1
kind: source
source:
  orbit_id: docs
  publish:
    target_branch: orbit-template/docs
    remote: origin
```

### 4.4 `orbit_template`

`kind=orbit_template` 至少需要表达：

- installable template 对应哪个 `orbit_id`
- 该 revision 是 installable orbit template，而不是 source

建议结构：

```yaml
version: 1
kind: orbit_template
orbit_template:
  orbit_id: docs
  name: Documentation Orbit
```

### 4.5 `harness_template`

`kind=harness_template` 至少需要表达：

- 模板对应哪个 harness identity
- bundle summary

建议结构：

```yaml
version: 1
kind: harness_template
harness_template:
  harness_id: writing-stack
  name: Writing Stack
```

### 4.6 `plain`

`plain` 不是 manifest kind。

它只是 classifier 在“没有有效 manifest”时返回的外部观察结果。

---

## 5. OrbitSpec Host Contract

### 5.1 正式 authored host

Orbit authored truth 统一落盘到：

```text
.harness/orbits/<orbit-id>.yaml
```

它必须在以下 revision kind 中都保持稳定：

- `runtime`
- `source`
- `orbit_template`
- `harness_template`

### 5.2 角色模型

Orbit 固定角色维持为：

- `meta`
- `subject`
- `rule`
- `process`

`meta` 不应作为普通 `members[]` 再重复声明。

它由 OrbitSpec 宿主文件天然承担。

### 5.3 默认 surface 映射

| role | projection | orbit_write | export | orchestration |
| --- | --- | --- | --- | --- |
| `meta` | yes | yes | yes | structured-source |
| `subject` | yes | no | no | no |
| `rule` | yes | yes | yes | yes |
| `process` | yes | no | no | yes |

### 5.4 override 规则

第一版只允许在 member 上显式覆盖：

- `scopes.write`
- `scopes.export`
- `scopes.orchestration`

不允许：

- 改写 `meta` 的 companion 语义
- 让普通 Git commit 进入 Orbit scope contract

### 5.5 `meta.agents_template`

`meta.agents_template` 是 orbit brief 的结构化 authored truth。

它应满足：

1. 内容短、稳定、可变量化；
2. 可 materialize 成 runtime 根 `AGENTS.md` 中当前 orbit block；
3. 可通过 `orbit brief backfill` 反向更新；
4. 不等于整份 `AGENTS.md` 文件。

---

## 6. Runtime Provenance Contract

### 6.1 `vars.yaml`

`.harness/vars.yaml` 负责：

- runtime 级变量绑定；
- export / publish / install 中使用的变量解析；
- brief reverse replacement 的值域来源。

### 6.2 install record

`.harness/installs/<orbit-id>.yaml` 负责：

- 当前 orbit 从哪个 template source 安装；
- 当前安装的分支/来源；
- 当前 install unit 的稳定 provenance。

### 6.3 bundle record

`.harness/bundles/<harness-id>.yaml` 负责：

- 当前 bundle / harness template 的来源；
- 由 bundle 带入的 member 列表与关联信息。

### 6.4 runtime member source

运行态成员来源至少区分：

- `manual`
- `install_orbit`
- `install_bundle`

member source 不属于 projection，也不属于 local ledger。

---

## 7. Local Ledger Contract

### 7.1 ledger 的职责

`.git/orbit/state/*` 只负责 repo-local 状态：

- 当前是否进入某 orbit；
- 当前 orbit 的局部文件 inventory；
- 当前 orbit 的本地运行态；
- 当前 orbit 与全局 Git 状态的分类摘要。

### 7.2 ledger 的边界

ledger 不应承担：

- authored truth
- publish history
- install provenance
- revision identity

### 7.3 stale state

若本地 ledger 与当前 revision 不匹配，应进入显式 stale / invalid 状态。

CLI 应：

- 给出诊断；
- 阻止基于错误 ledger 的 Orbit scoped 操作；
- 不静默“猜测修复”。

---

## 8. Surface Contract

### 8.1 `projection`

`projection` 只决定当前工作区看见什么。

消费该 surface 的命令包括：

- `orbit enter`
- `orbit leave`
- `orbit files`
- `orbit current`

关键边界：

- projection-visible 文件仍是普通 Git 文件；
- projection 不决定 template export；
- projection 不决定 scoped commit。

### 8.2 `orbit_write`

`orbit_write` 只决定 Orbit scoped 写命令应该作用什么。

消费该 surface 的命令包括：

- `orbit diff`
- `orbit log`
- `orbit commit`
- `orbit restore`

关键边界：

- `subject` 默认不进入 `orbit_write`；
- 普通 Git commit 不属于该 surface；
- `orbit commit` 绝不应静默带上 surface 外文件。

### 8.3 `export`

`export` 只决定 runtime / authoring revision 在保存、反写、发布时应带走什么。

消费该 surface 的命令包括：

- `orbit template save`
- `orbit template publish`
- `harness template save`

关键边界：

- export 不自动吸收所有 projection-visible 文件；
- export 不自动吸收运行态根 `AGENTS.md`；
- export 是有意设计的 payload，不是工作区快照。

### 8.4 `orchestration`

`orchestration` 只决定 brief / `AGENTS.md` materialization / backfill 的输入集合。

消费该 surface 的命令包括：

- `orbit brief materialize`
- `orbit brief backfill`
- runtime root `AGENTS.md` builder

关键边界：

- orchestration 不负责通用 authored file 回写；
- orchestration 与 export 不是同一套 payload；
- `AGENTS.md` 是 materialized artifact。

`orchestration` lane 在 v0.4 中推荐按 `runtime / source / orbit_template` 共享 lane 收口；详细合同见 `docs/orbit_brief_lane_v0_4_technical_spec.md`。

---

## 9. Command Contract

### 9.1 `harness install`

`harness install` 是唯一正式安装入口。

它必须：

1. 支持 orbit template 与 harness template source；
2. 写入 runtime provenance；
3. 正确设置 runtime member source；
4. 不依赖 `orbit template apply` 的长期兼容外壳。

### 9.2 `orbit brief materialize`

建议新增正式命令：

```bash
orbit brief materialize
```

合同：

1. 按当前 orbit 的 orchestration truth materialize 当前可编辑 block；
2. 默认写根 `AGENTS.md`；
3. 若覆盖现有编辑内容，应显式提示或要求 `--force`；
4. 该命令不修改 authored truth。

补充：

- v0.4 推荐支持 `runtime / source / orbit_template` 三态共享语义；
- 详细 revision gating、drift state 与 container patch 合同见 `docs/orbit_brief_lane_v0_4_technical_spec.md`。

### 9.3 `orbit brief backfill`

合同：

1. 从 materialized runtime / authoring 容器中提取当前 orbit block；
2. 只处理当前 orbit block，不吸收整份文件的其他内容；
3. 对当前 `vars.yaml` 做 reverse replacement；
4. 成功后只写回 `meta.agents_template`；
5. 不直接改写根 `AGENTS.md`。

补充：

- v0.4 推荐支持 `runtime / source / orbit_template` 三态共享语义；
- runtime 下的 backfill 仍然只写本地 structured truth，不自动进入 export/writeback/publish lane；
- 详细合同见 `docs/orbit_brief_lane_v0_4_technical_spec.md`。

### 9.3.1 Brief Lane Delivery Order

v0.4 中 brief lane 的推荐实现顺序固定为：

1. revision gating 与 target resolution
2. `materialize` 的 block-preserving container patch
3. `backfill` 的 revision matrix 与 local-only write contract
4. brief drift diagnostics
5. 再把 runtime / authoring 场景接到 save / publish 主链路上

不要反过来先做 runtime 自动回写模板、harness-level `AGENTS.md` lane 合并，或把 brief lane 扩张成通用同步器。

### 9.4 `orbit template publish`

合同：

1. 在 `source` 中，构建并发布 installable orbit template；
2. 在 `orbit_template` 中，直接校验并发布当前 template；
3. 若存在 materialized 根 `AGENTS.md` 与结构化 brief 差异，应给出显式 backfill 流程；
4. installable payload 中不允许携带根 `AGENTS.md`。

### 9.5 `orbit template save`

合同：

1. 输入必须是 `runtime`；
2. 输出必须只取 `export` surface；
3. 不自动吸收 `subject`；
4. 不自动把 runtime 根 `AGENTS.md` 作为 template file 导出。
5. 若当前 runtime member 的 `source=install_orbit`，且 `.harness/installs/<orbit-id>.yaml` 存在，则允许省略 `--to`，并默认回写到该 install record 的 `template.source_ref`；否则必须显式提供 `--to`。
6. orbit template consumer 必须能够从 `.harness/manifest.yaml` 解析 branch-level template metadata；`.orbit/template.yaml` 不再是安装 / apply / publish 校验的唯一必需输入，`orbit template save` 产出的 template branch 也允许不再携带该文件。

### 9.6 `harness template save`

合同：

1. 输入必须是 `runtime`；
2. 输出为 `kind=harness_template` revision；
3. bundle provenance 必须同时落盘；
4. 只导出正式 harness template payload，而不是 runtime 杂项工作目录快照。

---

## 10. Compatibility And Migration Policy

### 10.1 退场文件

以下文件不再作为 v0.4 steady-state 的 canonical control plane：

- `.orbit/source.yaml`
- `.orbit/template.yaml`
- `.harness/runtime.yaml`
- `.harness/template.yaml`

### 10.2 兼容策略

v0.4 不保留长期兼容策略。

允许的只有：

- 文档化的一次性 migration command；
- 测试或过渡分支中短期存在的迁移代码；
- 明确标记为 deprecated、待删除的兼容入口。

不允许：

- 长期 dual-read / dual-write
- 继续把旧 marker 文件作为正式 branch classifier 前置
- 把 hidden compatibility wrapper 当成正式产品命令

### 10.3 fixtures 与测试口径

新的测试夹具、示例和 acceptance path 应优先使用：

- manifest-based revision kinds
- `.harness/orbits/*.yaml`
- surface-based command behavior

---

## 11. 测试要求

至少补齐以下测试矩阵：

1. manifest per-kind schema / validator / codec
2. revision classifier 对 `runtime/source/orbit_template/harness_template/plain` 的识别
3. hosted OrbitSpec loader / writer / validator
4. role -> surface 解析矩阵
5. projection / orbit_write / export / orchestration 的命令消费测试
6. `orbit brief materialize` / `orbit brief backfill` 往返测试
7. `orbit template publish` 在 `source` 与 `orbit_template` 两态的行为测试
8. runtime -> orbit template / harness template export acceptance smoke

---

## 12. 完成标准

以下条件满足后，可认为 v0.4 技术合同已可进入实现：

1. 所有正式状态面都能映射到唯一 canonical host；
2. 所有正式命令都能映射到唯一 surface；
3. direct template authoring 与 source authoring 的发布路径被清楚定义；
4. runtime writeback / export / publish 的边界不再依赖隐式约定；
5. 兼容策略、迁移策略、测试策略都已有明确书面合同。
