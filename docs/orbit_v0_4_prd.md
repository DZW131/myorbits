# Orbit v0.4 Unified State Model PRD（产品需求文档）

版本：v0.4
状态：clean-break 基线，可进入实现拆分
关联文档：
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_v0_4_development_plan.md`
- `docs/orbit_brief_lane_v0_4_technical_spec.md`
- `docs/orbit_template_authoring_guide.md`
- `docs/context/orbit_v0_4_design_archive.md`
- `docs/context/orbit_state_and_workflow_unification.md`
- `docs/issues/closed/0108-orbit-brief-backfill-and-agents-orchestration-cutover.md`

---

## 1. 文档目标

本文档把 Orbit 下一阶段统一收口为一套正式的 v0.4 产品模型。

目标不是继续在 v0.3 的局部优化上叠加概念，而是把下面几件事一次讲清楚：

1. 当前系统里到底有哪些状态层；
2. 哪些文件是版本化真相源，哪些只是本地运行态，哪些只是派生产物；
3. `projection / orbit_write / export / orchestration` 四个行为面分别解决什么问题；
4. 用户态、模板作者态、source 开发态之间如何转换；
5. 下一阶段正式命令面与工作流应该如何组织；
6. 哪些旧控制文件与兼容路径应被正式退场。

---

## 2. 一句话定义

**v0.4 的目标模型是：`.harness/manifest.yaml` 负责“当前 revision 是什么”，`.harness/orbits/<orbit-id>.yaml` 负责“orbit 是什么”，`.harness/vars.yaml` / `.harness/installs/*.yaml` / `.harness/bundles/*.yaml` 负责 runtime provenance，`.git/orbit/state/*` 负责本地 projection 与 ledger，根 `AGENTS.md` 只是 orchestration 派生产物；projection、orbit_write、export、orchestration 四个行为面必须彻底拆开。**

---

## 3. 设计驱动力

当前代码和文档已经走到一个关键点：功能并不算少，但状态模型还没有完全收口，导致后续继续开发时会同时遇到四类问题。

### 3.1 状态宿主分散

当前 revision identity、orbit authored truth、runtime provenance、projection ledger、`AGENTS.md` 入口仍分散在多个历史文件与行为路径中，用户和开发者都容易混淆：

- 哪个文件是正式版本化真相源；
- 哪个文件只是兼容入口；
- 哪个文件只是 runtime 容器；
- 哪个文件只是 repo-local 的本地状态。

### 3.2 几种行为面被混在一起

当前很多命令虽然已经开始区分 scope，但用户心智里仍容易把下面几件事误认为一件事：

- “看见什么”
- “orbit 命令应该提交什么”
- “runtime 发布时能带走什么”
- “agent orchestration 会吃什么”

v0.4 必须把它们作为四个不同 surface 冻结下来。

### 3.3 作者工作流不够对称

目前已经存在三类真实作者/使用路径：

1. 用户在 runtime 中 install 多个 orbit 来完成任务；
2. 开发者直接维护 orbit template branch；
3. 开发者维护更专业的 source branch，然后发布到 template。

但这三类路径在“如何编辑 brief / AGENTS”、“如何 publish”、“哪些内容可导出”上还没有完全对齐。

### 3.4 兼容路径会拖慢后续收口

用户已经明确倾向于不保留长期兼容功能。

因此 v0.4 的立场应当是：

- 允许一次性迁移工具；
- 不保留长期 dual-read / dual-write；
- 不让旧 marker 文件继续占据正式 source of truth 地位；
- 文档与代码以最终模型的优美性、完整性和清晰性为先。

---

## 4. 已冻结结论

### 4.1 Single Control Plane 不只是替掉 `runtime.yaml`

v0.4 的 single control plane 不是“只把 `.harness/runtime.yaml` 吃进 manifest”，而是把 revision identity 整体收口到：

```text
.harness/manifest.yaml
```

它应统一表达：

- `runtime`
- `source`
- `orbit_template`
- `harness_template`

被替代并逐步退场的文件包括：

- `.orbit/source.yaml`
- `.orbit/template.yaml`
- `.harness/runtime.yaml`
- `.harness/template.yaml`

### 4.2 Orbit authored truth 统一进入 `.harness/orbits/*.yaml`

后续正式 authored OrbitSpec 统一收口到：

```text
.harness/orbits/<orbit-id>.yaml
```

它负责回答“orbit 是什么”，包括：

- orbit identity
- orbit 描述
- members
- role -> scope 规则
- orchestration 输入，如 `meta.agents_template`

### 4.3 Runtime provenance 单独存放

runtime provenance 不属于 revision identity，也不属于 projection ledger。

它应稳定放在：

- `.harness/vars.yaml`
- `.harness/installs/<orbit-id>.yaml`
- `.harness/bundles/<harness-id>.yaml`

### 4.4 本地 projection / ledger 继续只属于 `.git/orbit/state/*`

repo-local 状态继续只放在：

```text
.git/orbit/state/*
```

它负责：

- 当前进入哪个 orbit
- 当前 orbit 的 file inventory
- 本地 runtime state / git state
- 其他只应存在于本地、不能进入 Git DAG 的缓存与 ledger

### 4.5 `AGENTS.md` 是 artifact，不是 authored truth

v0.4 正式冻结以下边界：

- 根 `AGENTS.md` 是 orchestration artifact；
- runtime 根 `AGENTS.md` 是 harness 入口容器；
- orbit brief 的 authored truth 在 OrbitSpec 中，而不在根 `AGENTS.md`；
- `orbit brief backfill` 只负责 orchestration/brief 回填；
- brief backfill 不负责通用文件反写；
- 发布后的 orbit template 不应携带根 `AGENTS.md`。

### 4.6 四个 surface 必须拆开

| surface | 回答的问题 |
| --- | --- |
| `projection` | 当前看见什么 |
| `orbit_write` | `orbit diff/log/commit/restore` 作用什么 |
| `export` | runtime / authoring 内容反写、保存、发布时能带走什么 |
| `orchestration` | brief / `AGENTS.md` materialization / backfill 吃什么 |

必须明确：

1. `projection` 可见，不等于 `orbit_write`；
2. `orbit_write` 不等于 `export`；
3. `export` 不等于 `orchestration`；
4. 普通 Git 提交与 Orbit scoped commit 不是一回事。

### 4.7 普通 Git 继续保持普通 Git

Orbit 不应阻挡也不应重新定义用户的普通 Git 行为。

也就是说：

- 用户进入 projection 后，仍可对可见文件执行普通 `git add/commit`；
- Orbit 只定义 Orbit 自己的 scoped commands；
- `orbit commit` 必须严格受 `orbit_write` 约束；
- 普通 `git commit` 不是 Orbit scope 系统的一部分。

### 4.8 v0.4 采用 clean-break 方向

v0.4 的正式开发方向应当是：

- 新文档直接定义目标态；
- 新实现优先服务目标态；
- 若需要迁移，只允许显式、一次性的 migration path；
- 不保留长期 compatibility lane 作为正式主路径。

---

## 5. 核心对象与状态模型

### 5.1 存储矩阵

| 存储位置 | 正式回答的问题 | 是否进 Git | 说明 |
| --- | --- | --- | --- |
| `.harness/manifest.yaml` | 当前 revision 是什么 | yes | revision identity control plane |
| `.harness/orbits/<orbit-id>.yaml` | 当前 orbit 是什么 | yes | authored OrbitSpec |
| `.harness/vars.yaml` | runtime 变量绑定是什么 | yes | runtime provenance |
| `.harness/installs/*.yaml` | orbit 从哪里安装而来 | yes | runtime provenance |
| `.harness/bundles/*.yaml` | bundle / harness template provenance 是什么 | yes | runtime provenance |
| `.git/orbit/state/*` | 当前本地 projection / ledger 是什么 | no | repo-local runtime state |
| 根 `AGENTS.md` | 当前 orchestration artifact 长什么样 | usually yes | materialized container，不是 authored truth |

### 5.2 Revision kinds

v0.4 只认下面几种 revision identity：

1. `plain`
   - 没有有效 `.harness/manifest.yaml`
   - 不是 Orbit 正式 revision kind
2. `runtime`
   - 正式用户运行态工程
   - 可 install、enter、commit、save、publish
3. `source`
   - 专业 orbit 作者开发态
   - 可包含 author-only 文件、发布配置、开发辅助工具
   - 不直接作为 installable template 消费
4. `orbit_template`
   - installable 的单 orbit 模板态
   - 可直接被 `harness install`
   - 也应支持直接维护与直接 publish
5. `harness_template`
   - installable 的多 orbit 模板态
   - 可直接被 `harness install`

### 5.3 Runtime member source

runtime 中的 orbit 成员来源态独立于 revision kind。

它至少应区分：

- `manual`
- `install_orbit`
- `install_bundle`

这样 runtime 才能清楚表达：

- 这是手工加进去的 orbit；
- 这是从单 orbit template 安装来的；
- 这是从 harness template / bundle 安装来的。

### 5.4 Projection state

projection 不是 branch kind，也不是 template/runtime kind。

它只是 repo-local 的当前工作视图状态：

- full workspace
- entered orbit
- stale / mismatched local ledger

它的变化不应改变 revision identity。

---

## 6. 主要使用场景

### 6.1 用户组装 runtime

用户按任务类型 install 若干已经完善的 orbit template，组合成自己的 harness runtime，再借助 agent 完成任务。

这条路径是系统最核心的消费路径。

### 6.2 用户在 runtime 中按阶段切换 orbit

用户可能根据任务推进进入不同 orbit 的 projection 中工作。

在这个过程中：

- 工作文件、过程文件、规则文件都可能被修改；
- 普通 Git 行为不应被 Orbit 阻挡；
- Orbit 相关 scoped 命令应更清晰地表达自己的作用边界；
- 复杂但不应歧义的功能可以逐步增强，但基础心智模型必须先冻结清楚。

### 6.3 开发者直接维护 orbit template

开发者可以直接在 `orbit_template` revision 中维护 installable 模板本身。

该场景需要：

- 可临时 materialize 根 `AGENTS.md` 作为编辑入口；
- 可通过 `orbit brief backfill` 回填结构化 brief；
- 发布时验证 installable payload 不应包含根 `AGENTS.md`；
- 支持在 template 态直接 publish，而不是强迫所有人先走 source。

### 6.4 开发者通过 source branch 进行专业开发

复杂 orbit 的作者需要更专业的开发态。

`source` revision 应服务以下需求：

- 保留作者脚本、实验文件、发布信息、开发辅助工具；
- 与 installable template 分离；
- 提供清晰的 source -> orbit_template publish 路径。

### 6.5 用户在 runtime 中优化 orbit，再反写回 template

实际使用中，用户可能在 runtime 中对 orbit 规则或结构进行改良。

v0.4 应正式支持：

- 在 runtime 中修改 orbit authored truth；
- 通过 `export` surface 显式反写/保存/发布回 orbit template；
- 但不把 projection 或普通 Git 提交误当成 template publish。

### 6.6 用户把组合好的 runtime 打包成 harness template

当一组 orbit 已经被组合出稳定用法后，用户应能把整个 runtime 打包为 harness template，以便后续快速复用。

这条路径建立在前面几个基础能力稳定之后，再继续增强。

---

## 7. 正式命令与流程模型

### 7.1 命令分层

| 意图 | 正式命令 | 说明 |
| --- | --- | --- |
| 创建 runtime | `harness create` / `harness init` | 建立 `kind=runtime` revision |
| 安装模板 | `harness install` | 唯一正式 install 入口，可消费 orbit/harness template |
| 进入/退出视图 | `orbit enter` / `orbit leave` | 只处理 projection |
| 查看 orbit scoped 状态 | `orbit status` / `orbit diff` / `orbit log` | `status` 应显式展示 surface 归属 |
| orbit scoped 写入 | `orbit commit` / `orbit restore` | 只作用 `orbit_write` surface |
| 编辑 brief 入口 | `orbit brief materialize` | 生成临时根 `AGENTS.md` 供作者编辑 |
| brief 回填 | `orbit brief backfill` | 只回填 orchestration brief，不做通用文件反写 |
| runtime 导出 orbit template | `orbit template save` | 只导出 `export` surface |
| template / source 发布 | `orbit template publish` | source 与 orbit_template 都应支持 publish |
| runtime 打包 harness template | `harness template save` | 导出 runtime 组合产物 |

### 7.2 `orbit brief materialize`

v0.4 建议把“临时生成根 `AGENTS.md` 供编辑”正式化为一个独立命令：

```bash
orbit brief materialize
```

它的职责是：

- 把当前 orchestration truth materialize 为开发者可编辑的根 `AGENTS.md`；
- 方便 direct template authoring 和 source authoring；
- 作为作者入口，而不是正式真相源。

v0.4 进一步建议把这条命令泛化为 `runtime / source / orbit_template` 共享的 brief lane；详细合同见 `docs/orbit_brief_lane_v0_4_technical_spec.md`。

### 7.3 `orbit brief backfill`

`orbit brief backfill` 的职责固定为：

- 从当前 runtime / authoring 容器中提取当前 orbit brief；
- 反向变量化；
- 写回 `meta.agents_template`；
- 不修改通用 authored 文件；
- 不把整份根 `AGENTS.md` 当作 Orbit authored truth。

v0.4 的推荐收口是“有限泛化”：让 `backfill` 在 `runtime / source / orbit_template` 三态下共享同一语义，但仍然只属于 `orchestration` lane，而不扩张成 export/writeback/publish 命令。详见 `docs/orbit_brief_lane_v0_4_technical_spec.md`。

### 7.4 `orbit template publish`

`orbit template publish` 应成为统一的 orbit 发布命令，但按当前 revision kind 表现不同：

1. 在 `source` 中：
   - 从 source 构建 installable orbit template；
   - 校验；
   - 写入目标 template branch；
   - 发布。
2. 在 `orbit_template` 中：
   - 直接对当前 template branch 做校验；
   - 必要时引导或执行 brief backfill；
   - 确保最终 installable payload 不含根 `AGENTS.md`；
   - 发布。

### 7.5 `harness install`

`harness install` 是唯一正式安装入口。

v0.4 不再把 `orbit template apply` 作为主路径，也不应继续保留它的长期兼容地位。

### 7.6 runtime -> template 的写回边界

runtime 发布 orbit template 或 harness template 时：

- 只能从 `export` surface 带走内容；
- 不能自动把所有 projection-visible 文件都视为 template payload；
- 不能自动把运行态根 `AGENTS.md` 整体反写回 authored truth。

---

## 8. 范围与非目标

### 8.1 v0.4 范围内

v0.4 需要正式收口：

- single control plane
- OrbitSpec host 统一
- 四个 surface 拆分
- direct template authoring + source authoring 的对称工作流
- runtime -> orbit_template export / publish
- harness template save / install 的正式语义
- `AGENTS.md` brief materialize / backfill 的正式模型

### 8.2 v0.4 非目标

以下内容不属于本轮主目标：

- 多 repo / 多 workspace
- 多 harness 并存
- 后台服务、守护进程、数据库型 canonical state
- 自动隐式迁移
- 长期 dual-read / dual-write compatibility lane
- 自动把普通 Git 工作流“Orbit 化”

---

## 9. 成功标准

满足以下条件即可认为 v0.4 文档基线完成，并可进入开发实施：

1. 新主文档能独立回答“状态模型、场景模型、命令模型、边界模型是什么”；
2. 开发者不再需要同时查多份互相重叠的 proposal 才能理解正式方向；
3. 新文档能够明确区分 authored truth、runtime provenance、local ledger、artifact；
4. 新文档能够支持 issue 拆分、实现排期和测试矩阵设计；
5. 旧 proposal 已归档为背景材料，而不再与正式 source of truth 并列。
