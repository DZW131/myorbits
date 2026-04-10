# Orbit Template Authoring Guide

版本：v0.4
状态：authoring baseline guide
适用范围：最新 unified state model 设计口径

关联文档：
- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_v0_4_development_plan.md`
- `docs/orbit_brief_lane_v0_4_technical_spec.md`
- `docs/context/orbit_state_and_workflow_unification.md`
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/context/orbit_member_filesystem_behavior.md`
- `docs/context/harness_single_control_plane_proposal.md`
- `docs/issues/closed/0108-orbit-brief-backfill-and-agents-orchestration-cutover.md`

---

## 1. 文档范围

这份文档只讲 **orbit template 的作者工作流**，不讲 runtime 安装与用户使用。

并且只聚焦两种作者场景：

1. **直接开发 orbit template branch**
2. **通过 source branch 开发，再发布到 orbit template branch**

本文直接按最新规则来写，不解释旧兼容路径，只采用最新的作者模型和控制面边界。

若本文与 `docs/orbit_v0_4_prd.md`、`docs/orbit_v0_4_technical_spec.md`、`docs/orbit_brief_lane_v0_4_technical_spec.md` 冲突，以专项技术文档与 v0.4 主文档为准。本文只覆盖作者视角，不覆盖 runtime 下的 brief lane 全量语义。

如果当前 CLI 个别细节还没有完全实现，以本文的目标态口径为准。

---

## 2. 先把 Orbit 讲清楚

## 2.1 Orbit 不是“一个文件夹模板”

一个 orbit 不是整仓库模板，也不是完整 harness。

它更准确的定位是：

- harness 里一个**具备基本留痕和局部闭环的最小子模块**
- 它对某一类任务负责
- 它有清晰边界
- 它最低要能留下记录，并让成功 / 失败 / 未完成可被看见
- 它可以带很小的自检，但不应自带沉重质量系统

换句话说：

- harness 负责把 agent 的任务、状态、工具调用、证据、验收、记忆、约束外显成可追踪工件；
- orbit 不负责整个 harness 的全部可观测性；
- orbit 只对某一类任务的某些维度负责，只实现自己的局部记录、状态和结果信号。

## 2.2 Orbit 的最小闭环是什么

一个写得好的 orbit，至少应回答下面 6 个问题：

1. 这类任务的**观测面**是什么。
2. 这类任务的**最小留痕工件**是什么。
3. 成功 / 失败 / 未完成怎样被看见。
4. agent 可以操作哪些**工作对象**。
5. 如果需要，一个**极轻的自检 / 探针**是什么。
6. 哪些内容应该被复用、安装、导出。

如果一个 orbit 只能回答“有几份文档和一个 AGENTS.md”，但回答不了上面 6 个问题，它通常还不是一个稳定的 orbit。

这里尤其要避免一个误区：

- 不要默认把 orbit 写成“重质量检测系统”
- 因为 orbit 本身就可能是整个 harness 的检测 / 质量保证分支之一
- 如果每个 orbit 都再内置一套沉重检测，循环会不断叠加，带来明显的心智负担和资源损耗

所以对 orbit 来说，最低合同不是“必须很会检查别人”，而是：

- 自己能留下最小记录
- 自己的当前状态和结果可被看见
- 若有自检，它必须足够轻，最好只是一个 cheap probe / ready-not-ready 信号

## 2.2A Orbit framework 与 orbit template 的边界

在作者视角里，最容易混淆的是：

- `Orbit framework` 到底负责什么
- `orbit template` 到底应该自己携带什么

建议用下面这张表记住边界：

| 层 | 应负责 | 不应负责 |
| --- | --- | --- |
| `Orbit framework` | 宿主、surface、brief lane、template/save/publish/writeback lane、统一退出词汇 | 具体任务定义、probe 实现、budget 检测实现、agent 内部策略 |
| `orbit template` | 某类工作的目标、边界、规则、done probe、成功/失败/异常退出语义、记录要求 | Git / manifest 底层机制、通用 runtime 预算检测、全局编排实现 |
| `harness / agent runtime` | 预算来源、外部中断、调度、重试与异常检测 | 模板 authored truth 本身 |

换句话说：

- framework 提供“通用控制能力”
- template 提供“某类工作的 authored contract”
- runtime 决定“哪些控制语义能被真实执行”

## 2.2B 一个成熟 orbit template 最少应具备什么

建议至少具备下面这些 authored contract：

1. `objective`
2. `scope boundary`
3. `rules`
4. `done probe`
5. `failure condition`
6. `abnormal exit hint`
7. `record target`
8. `record minimum`

如果一个模板只有文件集合、缺少这些合同，它通常还只是“内容包”，不是一个成熟的 orbit template。

## 2.2C 退出模型要写清楚

一个 orbit template 最好至少显式支持下面几类退出：

- `success`
  - done probe 满足
- `failure`
  - 明确不满足该 orbit 目标
- `abnormal_exit`
  - 还不能判定 success / failure，但继续消耗已经不划算
- `external_stop`
  - 被父级 harness、人工或全局预算中断

作者最少要做的是把这些退出的语义写清楚，而不是把检测实现都塞进模板。

## 2.2D 预算限制更适合写成 hint，而不是模板内核能力

像下面这些限制：

- retry 次数
- 时长
- agent turns
- tokens
- cost

更适合作为：

- template 的 authored hint
- harness / agent runtime 的检测能力
- 外层编排的退出输入

而不是 Orbit framework 或 template 本身必须完整实现的通用机制。

这点非常重要，因为：

- 不同 runtime 可观测能力不同
- 不同 agent 框架预算来源不同
- 如果把这些实现强塞进 Orbit 内核，系统会迅速变重

## 2.3 Orbit 与 Harness 的关系

一个 harness 可以由多个 orbit 组成，例如：

- requirements orbit
- planning orbit
- coding / review orbit
- issues orbit
- release orbit

而每个 orbit 只负责自己那一小块任务闭环。

最终目标不是“orbit 越大越好”，而是：

- orbit 边界越清楚越好
- 可观测、可组合、可复用、可替换、可比较越好

这样 orbit 才能成为 harness 的稳定子模块，也才有可能进一步成为可执行、可比较、可消融的研究对象。

---

## 3. 作者模型的最新规则

## 3.1 单控制面

作者模型只认下面这几个正式宿主：

```text
.harness/manifest.yaml
.harness/orbits/<orbit-id>.yaml
```

其中：

- `.harness/manifest.yaml`
  - 表达当前 revision / branch 是什么
- `.harness/orbits/<orbit-id>.yaml`
  - 表达 orbit 是什么

## 3.2 source / template 的身份都进 manifest

目标态下：

- `source`
- `orbit_template`
- `harness_template`
- `runtime`

都应由 `.harness/manifest.yaml` 的 `kind` 表达。

本文不再把旧 marker 文件当正式作者入口。

## 3.3 Orbit authored truth 只看 OrbitSpec

orbit 的 authored 真相源应当是：

```text
.harness/orbits/<orbit-id>.yaml
```

它负责表达：

- orbit id / name / description
- members
- role -> scope 规则
- `meta.agents_template`

而不是靠根 `AGENTS.md` 充当真相源。

## 3.4 四个 surface 必须分开

orbit 作者必须清楚 4 个 surface：

- `projection`
  - 看见什么
- `orbit_write`
  - `orbit diff/log/commit/restore` 作用什么
- `export`
  - runtime / authoring 内容能带走什么
- `orchestration`
  - brief / `AGENTS.md` materialization / backfill 吃什么

关键边界：

1. `projection` 可见，不等于 `orbit_write`
2. `orbit_write` 不等于 `export`
3. `export` 不等于 `orchestration`
4. `orbit brief backfill` 只处理 orchestration / brief，不处理通用文件反写

## 3.5 已发布 orbit template 不应带根 `AGENTS.md`

这是最新规则里最重要的一条之一。

orbit template branch：

- 可以在作者工作流里临时 materialize 一个根 `AGENTS.md` 方便编辑
- 但已发布的 orbit template payload **不应包含根 `AGENTS.md`**

orbit 级 brief 真相源应写回：

- `meta.agents_template`

发布前：

- 可以 materialize
- 可以编辑
- 可以 backfill

发布后：

- 根 `AGENTS.md` 不属于 orbit template payload

---

## 4. 两种作者场景怎么选

## 4.1 直接开发 orbit template branch

适合：

- orbit 边界已经比较稳定
- 作者不需要很多 source-only 开发辅助文件
- 想直接维护 installable template 本身
- 想让模板分支本身成为主要协作面

## 4.2 通过 source branch 开发和发布

适合：

- orbit 比较复杂
- 需要保留作者脚本、发布说明、开发辅助文件、实验文件
- 需要更强的发布流程控制
- 希望把“作者输入态”和“已发布消费态”严格分开

## 4.3 简单选择法

如果你犹豫用哪种，就用下面这条：

- **小而稳定的 orbit，用 direct template branch**
- **复杂、长期演化、带作者工具链的 orbit，用 source branch**

---

## 5. 写 Orbit 之前，先做这张最小闭环清单

无论用哪种作者场景，先把 orbit 本身设计清楚。

建议先写出下面这张清单：

### 5.0 先写 authored contract，再写文件布局

写 orbit template 时，最容易犯的错误是：

- 先堆一组文件
- 再试图解释这些文件为什么构成一个 orbit

更稳的顺序应该反过来：

1. 先定义这类工作是什么
2. 先定义 worker 进入后要看到什么
3. 先定义什么算成功、失败、异常退出
4. 先定义结果记录到哪里、最少记录什么
5. 再决定哪些文件承载这些合同
6. 最后再决定它们落在哪个 branch / surface

也就是说：

- 先写 `contract`
- 再写 `files`

而不是：

- 先有文件堆
- 再强行把它包装成 orbit

### 5.1 任务边界

- 这个 orbit 负责哪一类任务？
- 它明确不负责什么？

### 5.2 观测面

- agent 完成这类任务时，需要观察哪些文件？
- 哪些文件是状态输入？
- 哪些文件是证据输入？

### 5.3 留痕合同

- 哪个文件或字段记录这类任务的当前状态？
- 哪个文件或字段记录结果，例如 `success` / `failed` / `pending`？
- 哪些证据或说明会被留下来，供后续 orbit 或 harness 消费？

### 5.4 结果信号

- 什么信号能让人或 agent 一眼看出当前是成功、失败、未完成还是待继续？
- 这个结果信号是结构化字段、文件位置、还是一段标准化文字？

建议把异常退出也一起纳入结果信号设计：

- `success`
- `failure`
- `abnormal_exit`
- `external_stop`

### 5.5 轻量自检（可选）

- 如果需要一个快速 ready / not_ready 判断，最便宜的探针是什么？
- 它是否足够轻，不会把 orbit 变成一个重 QA 子系统？

### 5.5A 预算与异常退出提示

如果你预计这类工作很容易因为预算、重试或场景不匹配而退出，建议在模板里显式写出：

- 建议重试上限
- 建议时长上限
- 建议资源预算上限
- 超限后推荐动作

但请记住：

- 这些是 authored hint
- 不是模板自己必须实现的检测器
- 真实检测能力由 harness / agent runtime 决定

### 5.6 export surface

- 哪些内容应该进入已发布 template？
- 哪些内容只是 source branch 的作者工具，不应该发布？

### 5.7 最小模板撰写卡

如果你想快速判断一个模板是否已经“像一个 orbit”，可以先写出下面这张卡：

```text
orbit:
  objective:
  not_in_scope:
  worker_entry:
  visible_scope:
  write_scope:
  rules:
  done_probe:
  failure_condition:
  abnormal_exit_hint:
  record_target:
  record_minimum:
  export_surface:
```

这张卡里每一项都应该能回答一个稳定问题：

- `objective`
  - 这类工作要完成什么
- `not_in_scope`
  - 这类工作明确不负责什么
- `worker_entry`
  - worker 一进入 orbit 必须先知道什么
- `visible_scope`
  - worker 需要看见什么
- `write_scope`
  - orbit 命令允许改什么
- `rules`
  - 必须遵守什么
- `done_probe`
  - 什么算成功
- `failure_condition`
  - 什么情况应明确判为失败
- `abnormal_exit_hint`
  - 什么情况应停止继续消耗
- `record_target`
  - 结果写到哪里
- `record_minimum`
  - 最少写什么
- `export_surface`
  - 哪些内容能随模板发布

如果这张卡还写不出来，通常说明模板仍然在“文件包”阶段，还没有到“工作模块”阶段。

---

## 6. 推荐的 OrbitSpec 思维方式

orbit 作者不应该先想“一堆 include/exclude path”。

更好的方式是先想成员和角色：

- `meta`
  - orbit 自身元信息、brief 真相源
- `subject`
  - 任务工作对象
- `rule`
  - 规则、状态机、轻量自检探针
- `process`
  - 推进说明、流程、协作约束

推荐默认映射：

| role | projection | orbit_write | export | orchestration |
| --- | --- | --- | --- | --- |
| `meta` | yes | yes | yes | structured-source |
| `subject` | yes | no | no | no |
| `rule` | yes | yes | yes | yes |
| `process` | yes | no | no | yes |

如果你的 orbit 需要某个 `process` 文件进入已发布 template，请显式 patch 它的 `export`，不要把“看得见”误当成“应该发布”。

### 6.1 OrbitSpec 不是模板的全部，但应当承载模板的核心合同

不要把 OrbitSpec 理解成“路径配置文件”。

更合理的理解是：

- OrbitSpec 不是模板的全部内容
- 但它应承载模板最核心的 authored contract

至少应在 OrbitSpec 或与之直接关联的 authored 内容里稳定表达：

- 这个 orbit 是什么
- 这个 orbit 为谁服务
- 这个 orbit 的入口是什么
- 这个 orbit 成功 / 失败 / 异常退出的语义是什么
- 这个 orbit 的结果记录目标是什么

而更长的过程材料、详细规则、示例、脚本，则可以继续留在 rule / process / source-only 文件中。

### 6.2 模板写的是外部合同，不是内部方法

一个模板最容易写重的地方，是把“agent 内部应该怎么想”也写进去。

更稳的做法是只写外部合同：

- 要遵守什么规则
- 要满足什么 probe
- 要留下什么结果
- 超限后应该退出到哪里

不要强制写死：

- 必须怎样拆分任务
- 必须怎样总结 session
- 必须怎样进行内部反思
- 必须怎样计量 token / turn / cost

模板越像“内部方法说明书”，它就越脆弱，也越难长期复用。

---

## 7. 场景 A：直接开发 orbit template branch

## 7.1 这个场景的 branch 合同

直接开发 template branch 时，这个分支本身就是 installable orbit template。

最小结构：

```text
.harness/
  manifest.yaml         # kind=orbit_template
  orbits/
    <orbit-id>.yaml

<export surface files...>
```

已发布 branch 不应包含：

- 根 `AGENTS.md`
- runtime provenance
- repo-local state

## 7.2 这个场景的作者工作流

### 第一步：在 template branch 上写 OrbitSpec

核心真相源是：

```text
.harness/orbits/<orbit-id>.yaml
```

你要在这里写清楚：

- orbit 的任务边界
- members
- rule / process / subject 的分层
- `meta.agents_template`
- 成功 / 失败 / 异常退出语义
- 结果记录目标

### 第二步：维护 export surface 文件

template branch 上真正会被安装带走的，是 export surface 文件。

所以你在这个分支上维护的文件，应该默认满足：

- 可以复用
- 可以安装
- 不依赖某台机器或某个 runtime 痕迹

### 第三步：临时 materialize 根 `AGENTS.md` 方便编辑

如果作者需要更舒服地编辑 brief，可以临时生成一个根 `AGENTS.md` 入口文件。

但要记住：

- 它只是作者辅助文件
- 不是 orbit template payload
- 不是最终真相源

### 第四步：执行 `orbit brief backfill`

编辑完临时 `AGENTS.md` 后，应把当前 orbit block 回填进：

- `meta.agents_template`

它只负责 brief / orchestration 真相源回收，不负责其它普通文件反写。

### 第五步：发布前做边界清晰的 publish

template branch 的正式发布，不应只是普通 `git commit`。

它应当是：

- 检查 export payload
- 检查根 `AGENTS.md` 不会进入 published payload
- 在无冲突时允许先 backfill brief
- 若 orbit 提供 cheap probe，可选地跑一次极轻自检
- 然后再发布 / 更新 template branch

换句话说，template-branch authoring 下的 publish 更像：

- **带 payload 检查、带 brief 收口、必要时带轻量自检的正式提交**

## 7.3 这个场景最常见的坑

### 把作者辅助文件当成 published payload

例如：

- 临时草稿
- authoring checklist
- 本地实验文件
- 临时 root `AGENTS.md`

这些都不应该随 orbit template 发布。

### 把 projection 误当 export

某个文件在 orbit view 里可见，不代表它应该进 template。

决定它是否发布的，是 `export surface`，不是 `projection`。

---

## 8. 场景 B：通过 source branch 开发，再发布到 orbit template branch

## 8.1 这个场景的 branch 合同

source branch 是作者输入态，不是 installable template。

最小结构：

```text
.harness/
  manifest.yaml         # kind=source
  orbits/
    <orbit-id>.yaml

<source-only authoring files...>
<orbit export/orchestration source files...>
```

它和 template branch 的最大区别是：

- source branch 可以带作者工具、发布说明、实验文件、脚本
- orbit template branch 只保留已发布 payload

## 8.2 source branch 允许有什么

source branch 可以有两类内容：

### orbit 真相源

例如：

- `.harness/orbits/<orbit-id>.yaml`
- rule 文件
- process 文件
- 需要进入 export surface 的模板文件

### source-only 作者辅助内容

例如：

- 发布 checklist
- 测试脚本
- authoring 笔记
- 本地 authoring probe
- 生成脚本

这类内容只服务作者，不应自动进入 orbit template payload。

## 8.3 这个场景的作者工作流

### 第一步：把 source branch 当成专业作者输入态

先在 source branch 上把 orbit 的真相源写完整：

- OrbitSpec
- rule / process / subject 结构
- brief 真相源
- 作者辅助脚本
- authored contract

这里建议显式检查一次：

- 这个模板到底是“某类工作的合同”
- 还是“若干作者文件的集合”

如果更像后者，先回到第 5 节把 authored contract 写清楚，再继续做 source authoring。

### 第二步：像维护模块一样维护 orbit 的闭环

source branch 的价值不是“多一个分支名”，而是：

- 让 orbit 的观测面、留痕工件、结果信号、轻量自检更完整
- 让作者可以保留开发工具链

### 第三步：必要时 materialize / backfill brief

source branch 也可以：

- 临时 materialize 一个根 `AGENTS.md`
- 编辑
- backfill 回 `meta.agents_template`

brief 的生命周期与 template branch 相同：

- 真相源在 OrbitSpec
- 根 `AGENTS.md` 只是作者辅助入口

### 第四步：执行 publish

source branch 的 publish 负责把 source 输入态发布成稳定 orbit template branch。

这里 publish 的职责不是“普通提交”，而是：

- 根据 source branch 的真相源生成 published orbit_template
- 做 payload 合法性检查
- 保证 source-only 文件不泄漏到 template branch
- 在需要时先完成 brief backfill / brief 一致性检查

## 8.4 什么时候必须升级成 source branch

如果你的 orbit 出现下面几类情况，就不建议只在 template branch 上裸写：

- 有作者脚本和发布工具链
- 有较多 source-only authoring 工具或实验文件
- 需要保留 source-only 开发资料
- 需要比较稳定的发布流程
- orbit 长期演化，且作者协作比较频繁

这时 source branch 往往更清晰。

---

## 9. 一个好 orbit 应该怎样体现“留痕和可观测性”

## 9.1 Orbit 的观测性不是“大而全”

orbit 的观测性不是把 harness 的所有状态都包进来。

更好的理解是：

- orbit 对自己负责的那类任务建立一个局部可观测闭环

例如一个 `issues orbit`：

- 它不负责整个 coding harness 的所有证据
- 但它应该能回答 issue 工作流本身有没有闭环

## 9.2 建议每个 orbit 至少具备这 6 类工件

### 工作对象

真正被处理的对象。

例如：

- issue 模板
- issue 状态文件
- 需求条目

### 过程说明

解释任务如何推进。

例如：

- triage 流程
- close 条件
- escalation 规则

### 留痕记录

任务推进时，最少要留下什么。

例如：

- 当前状态
- 当前结果
- 证据链接
- 关闭说明

### 结果信号

让人或 agent 能快速知道“现在怎么样”。

例如：

- `pending`
- `success`
- `failed`
- `ready` / `not_ready`

### 轻量规则 / 自检探针

定义哪些状态合法、哪些动作不合法。

例如：

- README 里的状态机规则
- 低成本的 `check-issues.sh`
- checklist / schema 约束

### brief

给 agent 的最小执行入口。

例如：

- `meta.agents_template`

这 6 类工件一起，才更像一个完整 orbit。

## 9.3 记录不是越多越好，而是越稳越好

模板作者最容易误判的一点是：以为“记录越多越完整”。

更好的标准是：

- 记录是否稳定
- 记录是否可复用
- 记录是否真的会被后续 orbit / harness 消费

因此推荐优先记录：

- 当前结果
- 关键证据
- 异常退出原因
- 下一步建议动作

而不是默认把整段内部过程都变成模板合同。

---

## 10. 参考示例：新的 `issues orbit`

本文改用仓库内的新参考示例。

参考路径：

- source branch 示例：
  - [docs/examples/issues_orbit_source_branch](/Users/zack/Code/Vocation/orbit/docs/examples/issues_orbit_source_branch)
- orbit template branch 示例：
  - [docs/examples/issues_orbit_template_branch](/Users/zack/Code/Vocation/orbit/docs/examples/issues_orbit_template_branch)

## 10.1 这个例子为什么适合讲闭环

`issues orbit` 很适合做闭环示例，因为它天然有：

- 明确工作对象
  - issue 模板、issue 状态文件
- 明确留痕
  - status、outcome、evidence、closure note
- 明确规则
  - issue 状态流转规则
  - 关闭前必须补证据并写清结果
- 明确过程
  - triage / update / close 流程
- 明确结果信号
  - `pending` / `success` / `failed`
  - cheap probe 给出 `ready` / `not_ready`

## 10.2 这个例子怎么体现 source 与 template 的区别

source branch 示例里会多出：

- `authoring/publish-checklist.md`
- 根 `AGENTS.md`

它们代表：

- 作者辅助文件
- brief 编辑入口

而 template branch 示例里只保留：

- OrbitSpec
- 规则文件
- 过程文件
- 模板文件

不会保留根 `AGENTS.md`。

## 10.3 这个例子怎么体现四个 surface

以示例里的 `issues orbit` 来说：

- `projection`
  - 让作者或用户看见 issue 模板、规则、过程说明
- `orbit_write`
  - 只让 orbit-local 命令作用于 `meta` 和 `rule`
- `export`
  - 决定哪些文件会随 template 发布
- `orchestration`
  - 决定哪些内容进入 brief / `AGENTS.md`

这正是 Orbit 应该具备的清晰边界。

---

## 11. 作者工作流建议命令口径

本文采用下面这组**目标态作者命令心智**：

- `orbit brief materialize`
  - 物化临时根 `AGENTS.md`，方便作者编辑
  - 它表达的是一个明确的作者辅助动作；若当前 CLI 尚未提供同名命令，也应按这个动作语义理解
- `orbit brief backfill`
  - 把当前 orbit block 回填进 `meta.agents_template`
- `orbit template publish`
  - 作者正式发布入口

其中：

- 在 **template branch** 下，`publish` 更像“带边界检查、带 brief 收口的正式提交”
- 在 **source branch** 下，`publish` 更像“把 source 输入态发布成 orbit template branch”

无论哪种场景：

- `publish` 都不应把根 `AGENTS.md` 当成 orbit template payload 发布

---

## 12. 不要这样写 orbit

### 12.1 一个 orbit 承担整个 harness

如果一个 orbit 同时包：

- requirements
- planning
- coding
- review
- release
- issues

那它通常已经不是 orbit，而是一整套 harness。

### 12.2 没有留痕和可观测性

如果 orbit 只有文件，却看不出当前状态、结果信号和失败原因，它很难称得上“稳定的原子子模块”。

### 12.3 自带过重的质量检测

如果 orbit 自己内置一套重量级 lint / test / review pipeline，会让资源损耗和心智负担快速膨胀。

尤其当这个 orbit 本来就是 harness 的检测 / 质量保证分支时，更容易形成“检查器再检查检查器”的套娃。

orbit 级自检应尽量保持 cheap、局部、可组合。

### 12.4 根 `AGENTS.md` 成了唯一真相源

这是需要避免的。

正确口径应该是：

- 真相源在 OrbitSpec
- 根 `AGENTS.md` 是 materialized authoring / runtime entry

### 12.5 source-only 文件泄漏到 template

source branch 可以有作者辅助内容，但 publish 时必须剥离。

### 12.6 用 brief backfill 做通用文件反写

`orbit brief backfill` 只负责 brief / orchestration 真相源回收。

它不应该变成：

- 任意文件同步器
- 通用 export 工具
- 规则文件回写器

### 12.7 模板只是一组文件，没有 authored contract

如果一个模板只有：

- include / exclude
- 几个 Markdown
- 一个 brief

但没有：

- objective
- scope boundary
- done probe
- failure / abnormal exit 语义
- record target

那么它通常还不是一个成熟的 orbit template。

它更像：

- 一个内容包
- 一个作者草稿
- 一个尚未收口成工作模块的文件集合

### 12.8 把预算检测实现强塞进模板

模板可以写：

- budget hint
- retry hint
- timeout hint
- 超限后的推荐动作

但模板不应该被迫自己实现：

- token 计量
- turn 计量
- time 计量
- cost 计量

这些更适合由 harness / agent runtime 提供。

---

## 13. 一句话总结

写 orbit template，不是在写“一组文件”，而是在写 **一类常见工作的 authored contract**：它既要定义边界、入口、退出和记录，也要能被 Orbit framework 稳定物化、复用和编排。

直接开发 template branch，适合小而稳定的 orbit；通过 source branch 开发再发布，适合复杂、长期演化、带作者工具链的 orbit。

无论采用哪种场景，都应坚持这几条：

1. branch identity 看 `.harness/manifest.yaml`
2. orbit authored truth 看 `.harness/orbits/<orbit-id>.yaml`
3. published orbit template 不带根 `AGENTS.md`
4. export 只看 `export surface`
5. brief backfill 只负责 orchestration / brief
6. orbit 的最低合同是“留痕 + 可观测结果”，自检若存在必须足够轻
