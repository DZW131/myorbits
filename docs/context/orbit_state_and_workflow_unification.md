# Orbit State And Workflow Unification

状态：archived design consolidation note

关联文档：
- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/context/orbit_member_filesystem_behavior.md`
- `docs/context/harness_single_control_plane_proposal.md`
- `docs/issues/closed/0108-orbit-brief-backfill-and-agents-orchestration-cutover.md`
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/orbit_member_runtime_technical_spec.md`

---

## 1. 文档目标

本文档把 Orbit / Harness 下一步的使用设计、状态模型、功能边界、状态转化与技术收口方向统一说明清楚。

它要解决的不是“某一个命令怎么实现”，而是先冻结以下问题：

1. 系统里到底有哪些状态；
2. 这些状态分别解决什么场景；
3. 哪些操作属于 projection，哪些属于 orbit-local write，哪些属于 export，哪些属于 orchestration；
4. source / template / runtime / projection / ledger 之间如何解耦；
5. 单控制面到底要收口到什么程度；
6. 哪些现有实现是目标态，哪些只是过渡态。

---

## 2. 一句话定义

**Orbit 的未来收口模型应当是：`.harness/manifest.yaml` 负责“当前 revision 是什么”，`.harness/orbits/<orbit-id>.yaml` 负责“orbit 是什么”，`.harness/vars.yaml` / `.harness/installs/*.yaml` / `.harness/bundles/*.yaml` 负责 runtime provenance，`.git/orbit/state/*` 负责本地 projection 与 runtime ledger；projection、orbit_write、export、orchestration 四个行为面必须显式拆开。**

---

## 3. 总体设计原则

### 3.1 一个版本化控制根

steady-state 只保留一个版本化控制根：

```text
.harness/
```

它负责：

- branch / revision identity
- OrbitSpec authored host
- runtime provenance
- template provenance

这意味着 single control plane 不是“只把 `.harness/runtime.yaml` 换成 `.harness/manifest.yaml`”，而是：

- `.orbit/source.yaml`
- `.orbit/template.yaml`
- `.harness/template.yaml`
- `.harness/runtime.yaml`

都应该逐步退场，最终由 `.harness/manifest.yaml` 的 `kind` 和对应字段统一表达。

### 3.2 一个 repo-local runtime ledger

本地执行态、投影态、状态快照继续只进：

```text
.git/orbit/state/
```

它不参与版本化，不承担模板态或 runtime provenance。

### 3.3 Orbit 不接管原生 Git

Orbit 只定义 orbit-local 行为，不重新定义原生 Git：

- 普通 `git add` / `git commit` / `git restore` 继续是原生 Git 语义；
- Orbit 只为 orbit 自己的命令面提供更清晰、更安全的 scope；
- Orbit 可以分类、提示、辅助，但不应把普通 Git 变成 Orbit 特有行为。

### 3.4 四个行为面必须分离

以下四个行为面必须作为正式内核对象存在：

- `projection`
- `orbit_write`
- `export`
- `orchestration`

任何命令都必须明确自己消费哪个 surface，而不是沿用一个模糊的“in scope = everything”模型。

---

## 4. 系统对象模型

## 4.1 Revision Identity

这是“当前这条 revision / branch 是什么”的层。

它回答：

- 当前是不是 runtime
- 当前是不是 source
- 当前是不是 orbit template
- 当前是不是 harness template

目标宿主：

```text
.harness/manifest.yaml
```

它不回答：

- orbit 内部文件成员是什么
- runtime 里具体绑定了什么变量
- 当前本地 enter 了哪个 orbit

## 4.2 OrbitSpec

这是“orbit 是什么”的层。

它回答：

- orbit id / name / description
- orbit 里有哪些文件成员
- 每个成员是什么 role
- role 到四个行为面的映射规则
- brief / orchestration 的结构化输入是什么

目标宿主：

```text
.harness/orbits/<orbit-id>.yaml
```

steady-state 下：

- runtime branch 使用 `.harness/orbits/*.yaml`
- source branch 使用 `.harness/orbits/*.yaml`
- orbit template branch 使用 `.harness/orbits/*.yaml`
- harness template branch 使用 `.harness/orbits/*.yaml`

换句话说，“orbit 的 authored 宿主”不应再跟 branch kind 一起漂移。

## 4.3 Runtime Provenance

这是“当前 runtime 是怎样组合出来的”的层。

它回答：

- runtime 用了哪些共享 vars
- 哪些 orbit 是 install-backed
- 哪些成员来自 harness bundle install
- 安装来源和 bundle 来源是什么

目标宿主：

```text
.harness/vars.yaml
.harness/installs/<orbit-id>.yaml
.harness/bundles/<harness-id>.yaml
```

它不回答：

- 当前进入了哪个 orbit
- 当前 sparse projection 是什么
- 当前正处于哪个执行阶段

## 4.4 Local Projection / Runtime Ledger

这是本地工作区和本地执行观察面的层。

它回答：

- 当前进入了哪个 orbit
- 当前 projection plan 是什么
- 当前 orbit 的 file inventory 是什么
- 当前 orbit-local / global Git 观察面是什么
- 当前本地 phase 是 `entered`、`status` 还是 `left`

目标宿主：

```text
.git/orbit/state/
```

---

## 5. 正式状态模型

这个系统不应再被描述为“只有一个状态机”。

更准确的结构是：

1. revision / branch identity state
2. runtime provenance state
3. projection state
4. local ledger state

这四层彼此相关，但不应互相代替。

## 5.1 Revision / Branch Identity State

目标上正式只有五种形态：

### `plain`

普通分支，没有 Orbit / Harness 版本化控制面。

特征：

- 没有有效 `.harness/manifest.yaml`

用途：

- 普通开发分支
- 还没初始化 Orbit / Harness 的仓库
- 只做非 Orbit/Harness 相关开发的分支

### `source`

专业作者输入态。

用途：

- 开发复杂 orbit 模板
- 保存发布信息、开发辅助文件、作者脚本、测试工具
- 作为 orbit template 发布的上游作者分支

关键点：

- `source` 是作者态，不是 installable template
- `source` 可以带开发辅助文件
- `source` 应该有更强的 authoring ergonomics

### `orbit_template`

单 orbit 的发布产物态。

用途：

- 被 `harness install` 消费
- 作为单 orbit 可复用模板
- 作为作者稳定发布后的消费对象

关键点：

- `orbit_template` 是 installable unit
- 根 `AGENTS.md` 不属于 orbit template payload
- orbit brief / orchestration 输入来自 OrbitSpec，而不是模板根 `AGENTS.md`

### `harness_template`

多 orbit 组合后的发布产物态。

用途：

- 把一整套 runtime 组合打包成 bundle
- 给用户一键安装一个多 orbit 工作工程

关键点：

- `harness_template` 是 installable bundle unit
- 它允许有根 `AGENTS.md` lane
- 它的 provenance 应通过 `.harness/bundles/*.yaml` 记录

### `runtime`

唯一正式 live runtime。

用途：

- 用户实际工作的工程态
- 被投影、安装、检查、导出

关键点：

- runtime 是唯一正式运行态
- projection 仍然是 orbit-centric，而不是 harness-centric
- runtime provenance 与本地 projection/ledger 必须分离

## 5.2 Runtime Provenance State

runtime 成员来源至少分成三类：

- `manual`
- `install_orbit`
- `install_bundle`

它们只负责表达 provenance，不负责表达 projection。

## 5.3 Projection State

projection 只表达：

- 当前工作区看见什么

它不表达：

- 当前 orbit commit 应提交什么
- 当前 export 应导出什么
- 当前 orchestration 应消费什么

projection 是“用户工作视图”，不是“所有命令统一 scope”。

## 5.4 Local Ledger State

本地 ledger 至少应稳定表达：

- 当前 orbit
- 当前 plan hash
- 当前 file inventory
- 当前 orbit-local Git 观察面
- 当前 global Git 观察面
- 当前 phase

---

## 6. 四个行为面

## 6.1 定义

### `projection`

决定：

- 当前 orbit view 中显示什么

### `orbit_write`

决定：

- `orbit diff`
- `orbit log`
- `orbit commit`
- `orbit restore`

这些 orbit-local 命令应该作用什么

### `export`

决定：

- runtime -> orbit template 的导出能带走什么
- runtime -> harness template 的导出能带走什么
- orbit template 发布时哪些 authored 内容是有效 payload

### `orchestration`

决定：

- `AGENTS.md`
- brief
- materialization
- backfill

这些编排相关行为消费什么

## 6.2 关键边界

必须冻结以下边界：

1. `projection` 下可见的内容，可以被普通 Git 提交；
2. 但这不等于都应该被 `orbit commit` 提交；
3. 也不等于都应该被 runtime 反写进 template；
4. runtime 导出 orbit template 时，只能走 `export surface`；
5. `orbit brief backfill` 只负责 orchestration / brief，不负责通用文件反写。

## 6.3 角色到行为面的默认映射

沿用现有 member-runtime 方向：

| role | projection | orbit_write | export | orchestration |
| --- | --- | --- | --- | --- |
| `meta` | yes | yes | yes | structured-source |
| `subject` | yes | no | no | no |
| `rule` | yes | yes | yes | yes |
| `process` | yes | no | no | yes |

这意味着：

- `subject` 默认可见，但不默认属于 orbit-local commit/export；
- `rule` 是 orbit-local write/export 的主对象；
- `process` 主要服务上下文和 orchestration；
- `meta` 是 orbit 自身控制与 brief 的真相源。

---

## 7. 文件系统设计

## 7.1 steady-state 目标

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml
  vars.yaml
  installs/
    <orbit-id>.yaml
  bundles/
    <harness-id>.yaml

.git/orbit/state/
  current_orbit.json
  resolved_scope/
  warnings.json
  last_status.json
  orbits/
    <orbit-id>/
      file_inventory.json
      runtime_state.json
      git_state.json
  orbit.lock
```

## 7.2 过渡文件的定位

以下文件不应是 steady-state 合同：

- `.orbit/source.yaml`
- `.orbit/template.yaml`
- `.harness/template.yaml`
- `.harness/runtime.yaml`

它们可以在迁移期继续存在，但定位必须明确为：

- compatibility marker
- transitional writer host
- migration input

而不能继续作为最终产品模型的第一性对象。

## 7.3 单控制面最重要的含义

single control plane 的真正含义是：

1. branch identity 全收口到 `.harness/manifest.yaml`
2. OrbitSpec authored host 全收口到 `.harness/orbits/*.yaml`
3. `.orbit/*` 不再承担“当前 branch 是什么”的职责
4. template / source / runtime 不再各自拥有独立顶层 marker 文件

---

## 8. `.harness/manifest.yaml` 目标合同

## 8.1 目标 kind

目标上 `manifest.yaml` 的 `kind` 应至少支持：

- `runtime`
- `source`
- `orbit_template`
- `harness_template`

`plain` 的语义是“没有合法 manifest”，而不是 `kind=plain`。

## 8.2 runtime 形态

它负责表达：

- runtime id / name
- created_at / updated_at
- members

## 8.3 source 形态

它负责表达：

- source branch identity
- source orbit id
- 目标发布 ref

它不应再依赖单独 `.orbit/source.yaml`。

## 8.4 orbit_template 形态

它负责表达：

- orbit id
- created_from_branch / commit / created_at
- template provenance

它不应再依赖单独 `.orbit/template.yaml` 作为 branch identity。

## 8.5 harness_template 形态

它负责表达：

- harness id
- members
- created_from_branch / commit / created_at
- includes_root_agents

它不应再依赖单独 `.harness/template.yaml` 作为 branch identity。

---

## 9. 使用场景设计

## 9.1 场景 1：用户安装多个完善 orbit 模板，组合成自己的 harness

这是最核心的用户场景。

路径：

1. 用户初始化 runtime
2. 用户安装一个或多个 orbit template
3. 用户得到一个组合好的 harness runtime
4. 用户在不同 orbit 之间 enter / leave
5. 用户借助 agent 完成工作

此场景下的正式命令线应是：

- `harness init` / `harness create`
- `harness install`
- `harness inspect`
- `harness check`
- `orbit enter`
- `orbit leave`
- `orbit status`
- `orbit diff / log / commit / restore`

### 9.1.1 用户在 runtime 中的修改行为

用户在 runtime 中可能修改：

- 工作文件
- 过程文件
- orbit 规则
- orbit brief / AGENTS block

这里必须明确：

1. 本工具不应阻挡普通 Git 使用；
2. Orbit 工具的职责是澄清 scope，而不是替代 Git；
3. `orbit commit` 只应该作用于 `orbit_write surface`；
4. 普通 `git commit` 继续可以提交用户当前可见、已暂存的内容；
5. 后续可以增强“心智清晰”和“提示质量”，但不应以功能歧义换取所谓自动化。

### 9.1.2 projection 的定位

用户进入某个 orbit 投影区时：

- 只是“工作视图切换”
- 不是“提交模型切换”
- 不是“模板发布模型切换”

projection 只决定“看见什么”，不应该暗含 export 或 publish 语义。

## 9.2 场景 2：开发者直接开发 orbit template

这是轻量 orbit 模板开发场景。

目标：

- 开发者直接在 orbit template branch 上迭代
- 可以临时 materialize 一个根 `AGENTS.md` 入口文件方便编辑
- 可以通过 `orbit brief backfill` 回填到 `meta.agents_template`
- 发布或保存时必须确保 orbit template payload 中不带根 `AGENTS.md`

这个场景的关键结论：

1. orbit template branch 本身仍然不能把根 `AGENTS.md` 当正式 payload；
2. 但 authoring workflow 允许存在临时 root `AGENTS.md` 作为编辑辅助；
3. 发布前应执行：
   - validate
   - brief backfill
   - 根 `AGENTS.md` 排除或清理

### 9.2.1 建议新增的 authoring 对称命令

既然已有：

```text
orbit brief backfill
```

则建议新增一个对称的 materialize 命令，例如：

```text
orbit brief materialize
```

它的职责是：

- 从 `meta.agents_template` / orchestration inputs 生成一个临时 root `AGENTS.md`
- 作为作者编辑入口
- 明确声明该文件不是 orbit template export payload

这样 authoring 体验才闭环：

1. materialize 出来方便编辑
2. 编辑完成后 backfill 回结构化真相源
3. publish/save 时校验并剥离

### 9.2.2 orbit template 分支上的“发布”

这个场景需要一个更明确的命令语义：

- 不是普通 `git commit`
- 而是“模板态提交 + 校验 + brief 回填确认 + payload 合法性检查”

因此建议把 orbit template authoring 的“发布”建模为：

- 对 template branch 本身的 validated publish / save
- 而不是继续混同为 source-only publish

## 9.3 场景 3：开发者使用 source 状态开发复杂 orbit

这是专业作者场景。

`source` 的作用应当明确为：

- orbit template 的专业作者输入态
- 可带作者说明、开发工具、脚本、测试
- 可保存更强的发布配置
- 通过专门 publish 命令发布到稳定 orbit template branch

这个场景的关键点：

1. source 是作者输入态，不是 installable template；
2. source 应允许额外开发辅助内容；
3. publish 是 source -> orbit_template 的显式转换；
4. 这个模型适合复杂 orbit，而不是所有 orbit 都必须经过 source。

## 9.4 场景 4：用户在 runtime 中优化 orbit，并反写回 orbit template

这是非常重要的反馈闭环场景。

用户在 runtime 中可能因为真实工作而优化：

- orbit 规则
- orbit process
- orbit brief
- orbit 局部 authored control

这个场景必须拆成两类反写：

### 9.4.1 brief / orchestration 反写

只走：

```text
orbit brief backfill
```

它只更新：

- `meta.agents_template`

它不负责：

- 通用文件导出
- rule/process 文件反写
- 模板 branch 发布

### 9.4.2 通用 template 反写

这属于 export 行为，应走：

- `orbit template save`
- 以及后续应补齐的 runtime publish 流程

这里必须明确：

- runtime -> template 的通用反写只能带走 `export surface`
- 不能因为用户当前“看得见”某文件，就自动把它带入模板

## 9.5 场景 5：用户把多 orbit runtime 打包成 harness template

这是组合工程复用场景。

路径：

1. 用户在 runtime 中安装多个 orbit
2. 用户调优配置、vars、组合方式
3. 用户执行 `harness template save`
4. 得到一个可复用 bundle

这个能力应建立在前面基础都清晰之后继续增强。

原因：

- 它依赖稳定的 runtime provenance
- 依赖清晰的 export surface
- 依赖根 `AGENTS.md` lane 合同
- 依赖 mixed install / bundle ownership 的稳定诊断

因此它是重要能力，但不应倒逼前面的基础模型继续混乱。

---

## 10. 功能设计与命令模型

## 10.1 Bootstrap / Control Plane

稳定命令：

- `harness create`
- `harness init`
- `harness inspect`
- `harness root`

方向：

- runtime bootstrap 以后以 `.harness/manifest.yaml` 为第一识别入口
- `orbit init` 继续只作为 compatibility bootstrap

## 10.2 Orbit Authoring

稳定命令：

- `orbit add`
- `orbit validate`
- `orbit show`
- `orbit files`

方向：

- authoring host 最终统一到 `.harness/orbits/*.yaml`
- `orbit add` 默认直接写 hosted OrbitSpec skeleton

## 10.3 Projection / Scoped Work

稳定命令：

- `orbit enter`
- `orbit leave`
- `orbit current`
- `orbit status`
- `orbit diff`
- `orbit log`
- `orbit commit`
- `orbit restore`

语义冻结：

- `enter/leave` 控制 projection
- `diff/log/commit/restore` 消费 `orbit_write`
- `status` 同时展示 orbit 视角与外部风险

## 10.4 Template Authoring / Publication

需要区分三条路径：

1. source -> orbit_template 的 publish
2. runtime -> orbit_template 的 export/save/publish
3. orbit_template 分支自身的 authoring + validated publish

现有命令可保留的部分：

- `orbit template save`
- `orbit template publish`

后续应补齐的 authoring 能力：

- `orbit brief materialize`
- template-branch authoring validate/publish
- runtime publish lane

## 10.5 Install / Composition

正式安装入口保持：

```text
harness install
```

它负责：

- 安装 orbit template
- 安装 harness template
- 记录 provenance
- 更新 runtime members

兼容入口：

- `orbit template apply`

但它只应继续作为 compatibility wrapper。

## 10.6 Brief / Orchestration

稳定命令：

- `orbit brief backfill`

后续建议补齐：

- `orbit brief materialize`

边界冻结：

- brief 命令只处理 orchestration / brief 真相源
- 不承担通用 export/writeback 语义

---

## 11. 状态转化模型

## 11.1 Revision / Branch 转化

```text
plain
  -> harness init/create
runtime

plain or runtime authoring branch
  -> source bootstrap
source

source
  -> orbit template publish
orbit_template

runtime
  -> orbit template save / runtime publish
orbit_template

runtime
  -> harness template save
harness_template

orbit_template
  -> harness install
runtime

harness_template
  -> harness install
runtime
```

## 11.2 Runtime 成员转化

```text
no member
  -> harness add
manual member

no member
  -> harness install orbit_template
install_orbit member

no member set
  -> harness install harness_template
install_bundle members
```

## 11.3 Projection 转化

```text
full workspace
  -> orbit enter <id>
projected workspace

projected workspace
  -> orbit leave
full workspace
```

## 11.4 Brief 转化

```text
meta.agents_template / orchestration inputs
  -> brief materialize
temporary runtime or authoring AGENTS.md

temporary edited AGENTS.md
  -> orbit brief backfill
meta.agents_template
```

---

## 12. 技术设计收口方向

## 12.1 第一优先：统一 branch identity

先做：

1. 把 `source` 正式并入 `.harness/manifest.yaml`
2. 把 orbit_template / harness_template / runtime 的 branch identity 全收口到 manifest
3. branch classifier 只以 manifest 为真相源

这一步完成后：

- `.orbit/source.yaml`
- `.orbit/template.yaml`
- `.harness/template.yaml`
- `.harness/runtime.yaml`

都应该降级为过渡文件，而不是 branch identity 的正式合同。

## 12.2 第二优先：统一 OrbitSpec authored host

先做：

1. loader 支持 source/template/runtime 全部从 `.harness/orbits/*.yaml` 读取
2. writer 默认只写 `.harness/orbits/*.yaml`
3. `.orbit/orbits/*.yaml` 只保留迁移兼容角色

这是 single control plane 真正成立的前提。

## 12.3 第三优先：把四个 surface 变成命令层正式合同

先做：

1. 所有 projection 命令只消费 `projection`
2. 所有 orbit-local 写命令只消费 `orbit_write`
3. 所有 template 导出命令只消费 `export`
4. 所有 brief / AGENTS materialization 命令只消费 `orchestration`

这样用户心智才会稳定：

- 看见什么
- 允许 orbit-local 提交什么
- 允许导出什么
- 允许生成/回填 brief 什么

分别是四件事，不再混为一谈。

## 12.4 第四优先：补齐 authoring 对称命令

必须补齐：

- materialize
- backfill

否则 authoring workflow 永远只解决了“回填”，没有解决“方便编辑”的正向入口。

## 12.5 第五优先：明确 runtime publish lane

需要正式支持：

- runtime 中优化 orbit 后导出 / 发布 orbit template

这条路径必须显式建模为 export，不应再借用 source-only 语义硬套。

---

## 13. 当前实现与目标态的关系

## 13.1 已经接近目标态的部分

当前代码已经有较强基础：

- `.harness/manifest.yaml` 已经在 branch classify 中成为主识别锚点
- runtime 已经有 `.harness/orbits/*.yaml`
- role-aware `ProjectionPlan` 已存在
- `PathClassification` 已显式区分 `projection / orbit_write / export / orchestration`
- `.git/orbit/state/orbits/<id>/...` ledger 已存在
- `orbit brief backfill` 已存在
- `harness install` 已经支持 orbit template 和 harness template
- bundle provenance 已经有 `.harness/bundles/*.yaml`

## 13.2 仍然是过渡态的部分

当前还没有彻底收口：

- `source` 还没有正式并入 manifest-based classifier 主路径
- source/template authored OrbitSpec 还没有全部统一到 `.harness/orbits/*.yaml`
- `.orbit/source.yaml` / `.orbit/template.yaml` / `.harness/template.yaml` / `.harness/runtime.yaml` 仍在不同实现路径中存在
- authoring-friendly brief materialize 入口还没有正式闭环
- runtime publish lane 还没有正式冻结成命令合同

---

## 14. 推荐实现顺序

### Phase A：冻结统一模型文档

先把本文件里的对象、状态、surface 和场景冻结下来。

### Phase B：manifest-only branch identity

完成：

- `kind=source`
- classifier / inspect / list 收口
- 过渡 marker 降级

### Phase C：OrbitSpec host 收口

完成：

- `.harness/orbits/*.yaml` 作为 steady-state authored host
- source/template/runtime 同 host

### Phase D：surface-aware command hardening

完成：

- orbit-local write/export/orchestration 的命令边界钉死

### Phase E：authoring workflow 闭环

完成：

- brief materialize
- template-branch validated publish
- runtime publish lane

### Phase F：harness template hard化

完成：

- bundle save/install/check
- mixed install 继续增强

---

## 15. 最终结论

下一步要做的不是继续增加一个个离散命令，而是先把整套世界观正式冻结成下面这句：

**branch identity 看 manifest，orbit authored truth 看 hosted OrbitSpec，runtime provenance 看 vars/installs/bundles，本地执行与投影看 `.git/orbit/state`；projection、orbit_write、export、orchestration 四个 surface 明确分治。**

只要这句没有被实现层彻底收口，source/template/runtime/projection/brief 的语义就还会继续互相污染。
