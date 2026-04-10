# Orbit Brief Lane v0.4 Technical Spec

版本：v0.4
状态：proposal，可进入实现拆分
定位：`orbit brief materialize` / `orbit brief backfill` 专项技术方案
关联文档：
- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_v0_4_development_plan.md`
- `docs/orbit_template_authoring_guide.md`
- `docs/orbit_member_runtime_technical_spec.md`

---

## 1. 文档目标

本文档专门冻结 v0.4 中 brief lane 的统一语义，回答以下问题：

1. `orbit brief materialize` / `orbit brief backfill` 是否只属于模板作者场景；
2. 它们在 `runtime / source / orbit_template` 三种 revision kind 下是否应共享一套语义；
3. 这两个命令到底处理什么，不处理什么；
4. 面对多个状态层时，怎样定义一条依然清晰的 brief lane；
5. v0.4 应支持到什么程度，哪些泛化应明确延后。

---

## 2. 一句话定义

**v0.4 中的 brief lane 应被定义为一条跨 `runtime / source / orbit_template` 共享的 `orchestration` lane：`materialize` 负责把结构化 brief 真相源物化为当前 orbit 的容器 block，`backfill` 负责把当前 orbit 的容器 block 回收为结构化 brief 真相源；它们只处理 `brief / meta.agents_template`，不承担 export、通用文件反写或 publish。**

---

## 3. 已冻结结论

### 3.1 它不是“作者专属命令”

虽然 `materialize -> edit -> backfill -> publish` 最常出现在 `source` 和 `orbit_template` 作者工作流里，但命令本身不应按“作者态”定义，而应按“brief lane 的方向”定义。

因此：

- `orbit brief materialize`
  - `structured truth -> container block`
- `orbit brief backfill`
  - `container block -> structured truth`

### 3.2 v0.4 只做有限泛化

v0.4 推荐支持的 revision kind：

- `runtime`
- `source`
- `orbit_template`

v0.4 明确不支持的 revision kind：

- `plain`
- `harness_template`

原因：

1. `plain` 没有稳定控制面与 Orbit authored truth；
2. `harness_template` 的 `AGENTS.md` lane 仍是 harness-level whole-file / bundle lane，不能和 orbit brief lane 混写；
3. 如果把这两类也一起收进来，会让命令语义从“orbit brief 同步”滑向“通用 AGENTS 同步器”。

### 3.3 两个命令的职责必须固定

`materialize` 只负责：

- 从 `meta.agents_template` 和 orchestration inputs 生成当前 orbit block；
- 把 block 写进 repo 根 `AGENTS.md` 容器；
- 保留其余 block 和容器外文本。

`backfill` 只负责：

- 从 repo 根 `AGENTS.md` 中提取当前 orbit block；
- 反向变量化；
- 写回 `.harness/orbits/<orbit-id>.yaml` 的 `meta.agents_template`。

### 3.4 两个命令都不做 export / publish / writeback

必须明确分层：

- brief lane 处理 `orchestration`
- template save / publish 处理 `export`
- runtime -> template 反写属于 export/writeback lane

因此：

- `backfill` 不会把改动自动发布到 template branch
- `materialize` 不会把容器文件自动标记为 export payload
- runtime 中的 brief 更新，若要进入 template，必须后续显式走 `save/publish`

### 3.5 面对多状态时，语义依然清晰

只要把 brief lane 视为一条独立于 revision kind 的 surface lane，而不是“某种分支专属工具”，语义就仍然清晰：

- revision kind 决定“下一步能做什么”
- brief lane 只决定“当前 orbit brief 如何在 truth 与 container block 间来回”

---

## 4. Brief Lane 的对象模型

### 4.1 结构化真相源

brief 的结构化真相源固定为：

```text
.harness/orbits/<orbit-id>.yaml
  -> meta.agents_template
```

这是所有支持的 revision kind 下都一致的 contract。

### 4.2 容器态

brief 的容器态固定为：

```text
AGENTS.md
  -> current orbit block
```

容器是 repo 根文件，不是 orbit 自有文件。

brief lane 永远只拥有：

- 当前 orbit block

而不拥有：

- 整份 `AGENTS.md`
- 其他 orbit block
- unmarked prose

### 4.3 local ledger 与 provenance 不参与 brief lane

brief lane 不直接读写：

- `.git/orbit/state/*`
- `.harness/installs/*.yaml`
- `.harness/bundles/*.yaml`

它唯一允许消费的 runtime provenance 是：

- `.harness/vars.yaml`

且用途仅限 reverse replacement。

---

## 5. 支持的状态模型

为了避免把 revision state、projection state、brief state 混成一团，v0.4 推荐把 brief lane 单独抽出一层状态。

### 5.1 Brief Lane State

对单个 orbit 而言，brief lane 只需要下面 5 种状态：

1. `structured_only`
   - `meta.agents_template` 存在
   - 根 `AGENTS.md` 中当前 orbit block 不存在
2. `materialized_in_sync`
   - 当前 orbit block 存在
   - block 与结构化 truth 等价
3. `materialized_drifted`
   - 当前 orbit block 存在
   - block 与结构化 truth 不等价
4. `invalid_container`
   - 根 `AGENTS.md` 存在
   - 但 marker 非法或无法稳定解析
5. `missing_truth`
   - OrbitSpec 缺失、非法，或不支持 `meta.agents_template`

这 5 种状态与 `runtime/source/orbit_template` 是正交的。

### 5.2 为什么这套状态足够

它回答的不是“当前 repo 是什么分支状态”，而是：

- 当前 orbit brief 是否已经物化；
- 物化内容是否与真相源一致；
- 当前是否允许安全执行 `materialize/backfill`。

这样多状态下的语义仍然清楚，不需要把命令再做成“每个 revision kind 一套特殊语义”。

---

## 6. Revision Matrix

| revision kind | materialize | backfill | 下一步常见动作 |
| --- | --- | --- | --- |
| `runtime` | allow | allow | `orbit template save` / runtime writeback / publish |
| `source` | allow | allow | `orbit template publish` |
| `orbit_template` | allow | allow | `orbit template publish` |
| `plain` | deny | deny | 先进入正式 control plane |
| `harness_template` | deny | deny | 走独立 harness-level lane |

关键理解：

- 命令语义不变；
- 变的是执行后可进入的下一条产品流程。

---

## 7. Command Contract

### 7.1 `orbit brief materialize`

合同：

1. 支持 `runtime/source/orbit_template`
2. 默认目标 orbit 为 current orbit，也支持 `--orbit <id>`
3. 读取 `meta.agents_template` 与其他 orchestration inputs
4. 只生成或更新当前 orbit block
5. 保留其他 orbit block 与 unmarked prose
6. 默认写 repo 根 `AGENTS.md`
7. 若将覆盖 drifted block，必须显式 `--force` 或交互确认
8. 不改写结构化 truth

建议输出至少包含：

- target revision kind
- target orbit id
- container path
- block status
- whether overwrite happened

### 7.2 `orbit brief backfill`

合同：

1. 支持 `runtime/source/orbit_template`
2. 默认目标 orbit 为 current orbit，也支持 `--orbit <id>`
3. 读取 repo 根 `AGENTS.md`
4. 只提取当前 orbit block
5. 不吸收其他 orbit block 或 unmarked prose
6. 使用 `.harness/vars.yaml` 做 reverse replacement
7. 成功后只写回 `meta.agents_template`
8. 不直接改写根 `AGENTS.md`
9. 不直接触发 `save/publish`

建议输出至少包含：

- target revision kind
- target orbit id
- definition path
- updated field
- replacement summary

### 7.3 `--check` / drift diagnostics

v0.4 推荐为 brief lane 增加轻量 check 能力，但不要求一定作为独立命令先落地。

最小能力可以是以下之一：

1. `orbit brief materialize --check`
2. `orbit brief backfill --check`
3. `orbit status` / `harness check` 输出 brief drift diagnostics

目标不是自动修复，而是让用户明确知道自己现在处于：

- in sync
- container drift
- invalid container

---

## 8. 运行态泛化后必须补的功能

如果要把 brief lane 正式扩到运行态，至少需要补齐下面这些能力：

### 8.1 revision gating

命令需要先判断当前 revision kind，只允许：

- `runtime`
- `source`
- `orbit_template`

### 8.2 orbit targeting

必须统一支持：

- current orbit
- `--orbit <id>`

### 8.3 block-preserving container patch

materialize 不能把根 `AGENTS.md` 简化成“整文件重写”。

它必须能够：

- 创建当前 orbit block
- 覆盖当前 orbit block
- 保留其他 orbit block
- 保留 unmarked prose

### 8.4 local-only backfill

runtime 中执行 `backfill` 时，只能更新当前 runtime repo 里的 hosted OrbitSpec。

它不能偷偷：

- 回写远端 template branch
- 更新 install source
- 更新 bundle source

### 8.5 handoff to export / publish

runtime 下的 brief lane 完成后，需要有清晰的下一步：

- `backfill` 之后，若要把更新带回模板，必须再显式执行 `orbit template save` 或后续 runtime publish lane

### 8.6 fail-closed behavior

以下情况必须失败：

- 当前 orbit 未知
- 根 `AGENTS.md` 缺失
- marker 非法
- 当前 orbit block 缺失
- reverse replacement 歧义
- hosted OrbitSpec 缺失或不支持 member schema

---

## 9. v0.4 明确不做的泛化

为了保持语义清晰，v0.4 明确不做下面这些扩展：

1. harness template whole-file `AGENTS.md` 反写
2. 通用 authored file 回写
3. rule / process / subject 文件同步
4. 跨分支自动回写上游 template
5. `backfill` 成功后自动 `save/publish`
6. 批量多 orbit backfill/materialize
7. 把 brief lane 做成“AGENTS 通用同步器”

---

## 10. 为什么这轮就做是合适的

我建议这轮就做“有限泛化”，原因是：

1. 这不会破坏四个 surface 的分层，反而会把 `orchestration` lane 固化清楚；
2. 当前作者态和运行态已经都存在对 brief 同步的真实需求；
3. 如果继续把它只写成作者命令，后面 runtime writeback lane 会重新发明一套相似但更混乱的语义；
4. 只要严格禁止它承担 export/writeback/publish，就不会把命令边界搞糊。

换句话说：

- 现在做“共享 brief lane”，是收口；
- 现在做“万能同步器”，才是扩散。

---

## 11. 推荐的 v0.4 实施顺序

建议按下面顺序实现：

1. 先为 `materialize/backfill` 增加 revision gating
2. 再把 `materialize` 定义为 block-preserving container patch
3. 再补 brief drift diagnostics
4. 再把 runtime 场景接到 `save/publish` 主链路上

不要反过来先做：

- runtime 自动回写模板
- harness-level AGENTS lane 合并
- 通用文件同步

---

## 12. 完成标准

满足以下条件后，可认为 v0.4 的 brief lane 泛化已经收口：

1. `runtime/source/orbit_template` 三态下都能安全执行 `materialize/backfill`
2. brief lane 的输出与副作用在三态下保持同一语义
3. 运行态下的 brief 更新不会被误认为 template publish
4. 主文档、作者文档、实现排期都已经明确这条 lane 只属于 `orchestration`
5. harness-level `AGENTS.md` lane 与通用 export/writeback lane 仍保持独立
