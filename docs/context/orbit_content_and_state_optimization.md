# Orbit Content And State Optimization

版本：v0.1
状态：proposal
阶段：post-v0.3 enhancement
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/orbit_scope_visibility_technical_spec.md`
- `docs/testing-strategy.md`
- `docs/context/orbit_storage_boundary.md`
- `docs/context/mvp-product-requirements.md`
- `docs/context/mvp-technical-architecture.md`

---

## 1. 文档目标

这份文档用于把 Orbit 后续优化方向收敛成一个更清晰的对象模型，解决当前讨论中的三个混淆点：

1. Orbit 的“内容”与“状态”还没有被稳定拆开；
2. 持久协作态与本地运行态还容易被混写；
3. projection、scoped write、template export 目前共享同一个解析内核，但还没有被显式建模为不同的命令消费面。

本文档不直接替代现有 v0.3 规格。
它的定位是：

- 对当前实现做重构收口提案；
- 为后续 projection state、Orbit spec 增强、local execution state 提供统一方向；
- 明确哪些内容继续放 `.orbit/`，哪些内容进入 `.harness/`，哪些内容必须留在 `.git/orbit/state/`。

正式收口后的配套文档为：

- `docs/context/orbit_member_filesystem_behavior.md`
  - 文件系统与行为系统总览，尽量短、只保留冻结结论
- `docs/orbit_member_runtime_technical_spec.md`
  - 成员化模型、role -> scope、运行态 ledger 的技术合同
- `docs/orbit_member_runtime_development_plan.md`
  - 成员化模型的实现阶段、测试面与完成标准

---

## 2. 当前问题

### 2.1 Orbit 内容模型还偏隐式

当前 Orbit 在代码中已经包含：

- 元信息：`id`、`description`
- 文件范围：`include`、`exclude`
- repo 级 overlay：`shared_scope`、`projection_visible`
- 隐式 companion：当前 orbit 自己的 `.orbit/orbits/<orbit-id>.yaml`

但这些内容在概念上还没有被显式拆成：

- Orbit 是什么
- Orbit 处理哪些对象
- Orbit 有哪些规则
- Orbit 自带哪些 companion / process docs

这使得后续要增加 `name`、`subjects`、过程文件、编排说明时，很容易继续堆在一个模糊的 YAML 结构里。

### 2.2 状态模型还容易混淆

当前系统已经有三个状态层，但用户心智上还不够稳定：

- `.orbit/`
  - 版本化 spec / control plane
- `.harness/`
  - 版本化 runtime metadata / install metadata
- `.git/orbit/state/`
  - repo-local runtime state

问题不在于缺层，而在于“执行态”还没有被显式命名。

例如：

- 当前进入哪个 orbit
- 当前 projection 对应哪个解析结果
- 当前是否正在执行某个 orbit-local flow
- 当前处于什么阶段

这些都属于不同强度的本地状态，但现在只有 projection state 被冻结，execution state 还没有合同。

### 2.3 Scope 分层已存在，但没有被提升为正式内核对象

当前实现里的 `ScopeSet` 已经隐含了多层 scope：

- `ControlReadPaths`
- `OwnedPaths`
- `ProjectionOnlyPaths`
- `CompanionPaths`
- `ScopedOperationPaths`
- `ProjectionPaths`

这套设计方向是对的，但它仍然主要表现为“命令内部临时结果”，而不是一个正式的 `ProjectionPlan`。

结果是：

- `enter`
  - 消费 `ProjectionPaths`
- `commit / diff / log / restore`
  - 消费 `ScopedOperationPaths`
- `template save`
  - 消费 `OwnedPaths`

语义是对的，但对象层次还不够清晰。

### 2.4 当前已经有一个真实边界问题

当前 tracked `projection_visible` 文件已经被正确排除在 scoped commit / restore / template save 之外。

但 untracked 路径仍复用了过于宽泛的“属于 orbit 吗”判断，导致：

- untracked `projection_visible` 文件
- 可能被 `orbit commit` 当成 in-scope 新文件带入提交

这说明后续优化不能继续依赖单个布尔 `in_scope`。
必须显式区分：

- visible
- writable
- exportable

---

## 3. 已有稳定前提

后续优化必须继续遵守以下冻结边界：

1. Git 仍是唯一历史真相源；
2. Orbit 仍是文件级 projection，不引入 block-level 或 semantic orbit；
3. 单仓库、单工作区、单 branch 背景不变；
4. `.orbit/` 仍是 spec / control plane，不是 cache；
5. `.harness/` 仍是 runtime versioned metadata；
6. `.git/orbit/state/` 仍是 repo-local runtime state，不是跨机器同步机制；
7. sparse-checkout 仍只是工作区投影，不是真相源；
8. `refs/orbits/*` 仍只是 best-effort auxiliary refs。

---

## 4. 建议对象模型

为避免继续把“内容”和“状态”混在一起，建议把 Orbit 相关对象显式拆成三层。

### 4.1 OrbitSpec

`OrbitSpec` 是版本化、可协作、进 Git 的定义对象。

它回答的问题是：

- Orbit 是谁？
- Orbit 包含哪些文件成员？
- 每个文件成员是什么角色？
- 各类 scope 如何围绕成员角色导出？
- 哪些说明需要进入 orchestration / `AGENTS.md` lane？

建议放在：

- `.orbit/config.yaml`
- `.orbit/orbits/<orbit-id>.yaml`

建议内容包括：

- `id`
- `name`
- `description`
- `members`
- `rules`

其中：

- `description`
  - 作为结构化元信息存在于 orbit yaml
  - 不应只依赖 `AGENTS.md`
- `members`
  - 是 Orbit 的第一性对象
  - 每个 member 对应一组文件，而不是抽象语义节点
  - scope 不再先从一份模糊 path list 推导，而是先从 member role 推导
- `rules`
  - 主要负责声明 role 到各类 scope 的映射规则
  - 必要时允许 member 做局部覆盖

### 4.2 OrbitSharedState

`OrbitSharedState` 是版本化、可协作、进 Git，但属于 runtime 层的持久协作态。

它回答的问题是：

- 当前 runtime 中是否包含该 orbit？
- 该 orbit 的来源是什么？
- 当前 runtime 使用哪些 bindings？
- 这个 runtime 里对 orbit 的安装 / provenance 是什么？

建议放在：

- `.harness/runtime.yaml`
- `.harness/vars.yaml`
- `.harness/installs/<orbit-id>.yaml`

注意：

- 这层不适合存 live execution flags
- 不应直接写入“是否正在执行”“当前阶段”这类高频易变状态
- 它存的是共享事实，不是瞬时运行态

### 4.3 OrbitLocalState

`OrbitLocalState` 是 repo-local、不进 Git 的运行态。

它回答的问题是：

- 当前工作区进入了哪个 orbit？
- 当前 projection 是什么？
- 最近一次状态分类是什么？
- 当前本地运行是否正在执行？
- 当前本地 execution 处于哪个阶段？

建议放在：

- `.git/orbit/state/current_orbit.json`
- `.git/orbit/state/resolved_scope/*.txt`
- `.git/orbit/state/warnings.json`
- `.git/orbit/state/last_status.json`
- `.git/orbit/state/runs/<orbit-id>.json`（新增建议）

---

## 5. Orbit 内容模型建议

这里把 Orbit 明确定义为：

**`一个 orbit = 元信息 + 一组文件成员（members）+ 基于成员角色推导出的 scope 规则`。**

也就是说：

- Orbit 的一等对象不是一份扁平 `include/exclude` path list
- Orbit 的一等对象是“文件成员”
- 每个成员都有明确角色
- scope 的存在目的，就是把这些角色转成不同命令面会消费的路径集合

这个模型与 harness runtime / harness template 中已有的 `members` 不冲突：

- `.harness/runtime.yaml` 的 `members`
  - 表示 runtime 里声明了哪些 orbit
- `.harness/template.yaml` 的 `members`
  - 表示一个 harness template 由哪些 orbit 组成
- `.orbit/orbits/<orbit-id>.yaml` 的 `members`
  - 表示这条 orbit 内部由哪些文件成员构成

三者同名，但层级不同：

- harness-level member = orbit membership
- orbit-level member = file membership

### 5.1 Orbit 元信息

Orbit 自身仍保留结构化元信息：

- `id`
- `name`
- `description`

推荐冻结为：

- `id`
  - 机器标识
  - 继续保持 path / ref-safe
- `name`
  - 人类展示名
- `description`
  - 面向人和 agent 的结构化说明
  - 可进入 future orchestration / `AGENTS.md` lane
  - 但不以 `AGENTS.md` 作为唯一真相源

### 5.2 Orbit 文件成员

`members` 是 orbit 的核心内容层。

每个 member 代表：

- 一组被同等对待的文件
- 这些文件在 orbit 中承担同一种角色
- 这些文件因角色不同，进入不同 scope

推荐最小 member 合同：

```yaml
members:
  - key: docs-content
    name: Docs Content
    role: subject
    paths:
      include:
        - docs/**
        - README.md
      exclude:
        - docs/generated/**
```

建议每个 member 具备以下属性：

- `key`
  - orbit 内稳定标识
- `name`
  - 人类展示名
- `description`
  - 对该 member 职责的说明
- `role`
  - 成员角色
- `paths.include`
  - 该成员覆盖的路径集合
- `paths.exclude`
  - 对该成员的局部剔除规则
- `lane`
  - 可选，声明该成员是否进入特殊处理通道，例如 `agents`
- `scopes`
  - 可选，对默认 role-scope 映射做局部覆盖

### 5.3 成员角色

第一阶段建议冻结四种写死角色：

- `meta`
  - orbit 自身的元信息控制面
  - 载体固定为 `.orbit/orbits/<orbit-id>.yaml`
  - 这个文件本身就是一个成员
- `subject`
  - 处理对象本身
  - 即 orbit 主要要修改、比对、导出的目标文件
- `rule`
  - 约束 subject 如何被处理的规则文件
  - 例如 harness rules、lint/config/schema/policy/template rule
- `process`
  - 描述“如何处理 subject”的过程文件
  - 例如说明文档、playbook、prompt、checklist、rolling docs、plans、issues

这样定义后，scope 的语义就变清楚了：

- `meta`
  - 是 orbit 的自描述与控制成员
- `subject`
  - 是主工作对象
- `rule`
  - 是约束内容
- `process`
  - 是操作流程与编排内容

这里建议再冻结一个结构规则：

- `.orbit/orbits/<orbit-id>.yaml`
  - 既是 OrbitSpec 宿主
  - 也是固定的 `meta member`
- `members[]`
  - 只列显式的 `subject / rule / process`
  - 不再重复声明 `meta`

这样可以避免“元信息文件要在自己的 members[] 里再声明一次”的递归问题。

### 5.4 Scope 不再先定义路径，而是先定义角色映射

在这个模型里，scope 不是第一性对象。

第一性对象是：

1. member
2. member role
3. role 到 scope 的映射

推荐把 orbit scope 收口为四个正式消费面：

- `projection`
  - 当前 view 里应显示哪些成员
- `orbit_write`
  - scoped `diff/log/commit/restore` 应作用哪些成员
- `export`
  - template save / publish 应消费哪些成员
- `orchestration`
  - 哪些成员与 orbit 元信息一起进入 `AGENTS.md` / future orchestration lane

也就是说：

- 先决定一个路径属于哪个 member
- 再由 member 的 `role` 决定它进入哪些 scope

### 5.5 推荐的默认 role-scope 映射

这里建议把映射分成两层：

1. 目标语义层
2. 当前 v0.3 兼容层

因为你刚刚补充的约束非常关键：

- `AGENTS.md` 更适合作为派生产物，而不是基础 authored member
- `subject` 不应简单等于当前 `OwnedPaths`
- orbit-local commit 更像是“迭代和优化 orbit 规则”的提交面
- 全局暂存/提交仍应保持为原生 Git 能力，而不是 Orbit scope

#### 目标语义层

目标上，建议冻结如下角色消费矩阵：

| role | projection | orbit_write | export | orchestration |
| --- | --- | --- | --- | --- |
| `meta` | yes | yes | yes | structured-source |
| `subject` | yes | no | no | no |
| `rule` | yes | yes | yes | yes |
| `process` | yes | no | no | yes |

解释：

- `meta`
  - 是 orbit 自身的控制面
  - 仍应进入 orbit-local write / export
  - 同时作为 orchestration 的结构化输入源
- `subject`
  - 默认只负责“看见什么”
  - 不再默认进入 orbit-local commit / restore / template export
- `rule`
  - 是 orbit 规则迭代的主工作对象
  - 默认进入 orbit-local write / export / orchestration
- `process`
  - 默认只提供上下文和编排素材
  - 不进入 orbit-local write，也不进入普通 template export

#### 当前 v0.3 兼容层

当前代码里 Orbit 自己只有：

- `OwnedPaths`
- `ProjectionOnlyPaths`
- `CompanionPaths`
- `ScopedOperationPaths`
- `ProjectionPaths`

因此在 member-role 模型真正落地之前，只能先用一个兼容近似：

| role | 当前近似落点 |
| --- | --- |
| `meta` | `CompanionPaths`，并随当前 companion 逻辑进入 projection/write/export |
| `rule` | `OwnedPaths` |
| `subject` | 临时仍落在 `OwnedPaths`，但这是兼容行为，不是目标语义 |
| `process` | `ProjectionOnlyPaths` |

这意味着：

- `subject -> OwnedPaths`
  - 只是一条过渡映射
  - 不是建议冻结的最终语义
- 后续应优先把 `subject` 迁出 orbit-local write/export
  - 改挂到仅 `projection`

若某个 orbit 确实需要让 `process` 文件参与普通 scoped write，可以通过 member-local `scopes.write: true` 覆盖默认值。

### 5.6 规则层的职责

在这个模型里，`rules` 不再主要负责直接列路径。

`rules` 的职责改为：

- 定义 role 到 scope 的映射
- 定义是否允许局部 member 覆盖默认行为
- 定义 orchestration 的输入策略

推荐最小规则合同：

```yaml
rules:
  scope:
    projection_roles: [meta, subject, rule, process]
    write_roles: [meta, rule]
    export_roles: [meta, rule]
    orchestration_roles: [meta, rule, process]
  orchestration:
    include_orbit_description: true
    materialize_agents_from_meta: true
```

如果 member 需要例外处理，再在 member 上写局部 `scopes.*` 覆盖。

这里建议额外冻结一条：

- `AGENTS.md`
  - 默认不是第一性 authored member
  - 而是 orchestration lane 的派生产物
  - 默认由 `meta.description` + 选中的 `rule/process` 成员共同生成

也就是说：

- `meta`
  - 是 AGENTS 编排的基础输入
- `rule/process`
  - 是 AGENTS 编排的补充素材
- `AGENTS.md`
  - 是 materialization result
  - 可以晚一点落地，不必现在把基础模型绑死在 `AGENTS.md` 文件本体上

### 5.7 过程文件的正式落点

你提到的“过程文件”在这个模型里不再单独做一套平行结构，而是直接作为：

- `members[].role = process`

来表达。

这意味着：

- 过程文件首先是 orbit 的文件成员
- 它们不是 runtime state
- 也不是 `meta` 成员本身
- 它们是有明确角色的版本化文件

同时要区分两类“过程相关内容”：

- `process members`
  - 进 Git
  - 是 OrbitSpec 的一部分
  - 例如说明文档、步骤清单、lane 文档、plans、issues、rolling docs
- `execution artifacts`
  - 不进 Git
  - 属于 OrbitLocalState
  - 例如运行日志、checkpoint、阶段状态

不要把两类内容混到同一个 `process files` 概念里。

### 5.8 建议冻结的 OrbitSpec 形态

综合上面的定义，建议 OrbitSpec 采用如下 clean-break 形态：

```yaml
id: docs
name: Documentation
description: User-facing docs orbit.

meta:
  file: .orbit/orbits/docs.yaml
  include_in_projection: true
  include_in_write: true
  include_in_export: true
  include_description_in_orchestration: true

members:
  - key: docs-content
    name: Docs Content
    role: subject
    paths:
      include:
        - docs/**
        - README.md
      exclude:
        - docs/generated/**

  - key: docs-rules
    name: Docs Rules
    role: rule
    paths:
      include:
        - .markdownlint.yaml
        - vale.ini

  - key: docs-process
    name: Docs Process
    description: Writer and agent workflow for documentation changes.
    role: process
    paths:
      include:
        - docs/process/**
    scopes:
      write: false
      orchestration: true

rules:
  scope:
    projection_roles: [meta, subject, rule, process]
    write_roles: [meta, rule]
    export_roles: [meta, rule]
    orchestration_roles: [meta, rule, process]
  orchestration:
    include_orbit_description: true
    materialize_agents_from_meta: true
```

这份结构和你的意图是一一对应的：

- 元信息
  - `name`
  - `description`
  - `meta`
- Subjects
  - `members[role=subject]`
- 规则
  - `members[role=rule]` + `rules.scope`
- 过程文件
  - `members[role=process]`

### 5.9 对当前设计的优化点

这个模型相对当前 path-list 心智的直接收益是：

1. Orbit 的心智模型从“一个 pattern 集合”变成“一个有角色分工的文件成员集合”。
2. scope 不再需要靠 `owned / projection_only / companion` 这些结果态名字反推意图。
3. 流程文档、规则文件、业务对象可以统一落进一个 member 模型里；`AGENTS.md` 则建立在这些结构化输入之上派生。
4. template save / harness template save / future orchestration 都可以围绕 member role 做更稳定的消费，而不是继续围绕散落 path list 特判。
5. `shared_scope` / `projection_visible` 这类概念，在 clean-break 版本中可以逐步退化为 role-scope 派生规则，而不是用户主 authoring model。

---

## 6. Orbit 状态模型建议

可以把当前讨论中的 Orbit 状态稳定成四类。

### 6.1 永久信息

这对应 `OrbitSpec`。

适合包含：

- orbit id / name / description
- members
- role-scope rules
- orchestration generation rules

落点：

- `.orbit/`

### 6.2 永久运行过程信息

这不是 live execution。
它更准确地说是：

- 与 runtime 持久协作有关的 Orbit metadata
- 安装来源
- runtime membership
- versioned bindings

适合包含：

- current runtime members
- install provenance
- shared bindings
- runtime-level generated artifacts provenance（若未来需要版本化）

落点：

- `.harness/`

不建议包含：

- 正在执行中
- 当前第几步
- 当前锁定阶段

这类内容太易变，不适合进 Git。

### 6.3 Orbit 本身运行过程信息

这对应 `OrbitLocalState`。

适合包含：

- 当前进入的 orbit
- 当前 projection generation / hash
- 当前解析 plan 的本地 cache
- 最近一次 status / warnings
- 当前本地 execution 是否在跑
- execution phase / stage / timestamps
- 当前 orbit 的本地 runtime 清单
- 当前 orbit 的 Git 分层状态梳理

落点：

- `.git/orbit/state/`

这层建议至少明确回答：

1. 是否在执行
2. 处于什么阶段
3. 当前 orbit 的文件成员清单是什么
4. 当前 orbit 的 projection / orbit-local write 视图分别是什么
5. 当前 Git 状态在 orbit / global 两个维度上如何划分

### 6.4 Git 状态分层

除了 OrbitSpec / runtime metadata / local execution state 之外，还需要把 Git 状态的观察面显式拆开。

建议冻结两组状态：

#### Orbit 视角

- `orbit_projection_state`
  - 当前 orbit 视图里哪些文件可见
- `orbit_stage_state`
  - 当前 orbit-local 暂存面包含哪些文件
- `orbit_commit_state`
  - 当前 orbit-local 提交面理论上会提交哪些文件

#### 普通 Git 视角

- `global_stage_state`
  - 当前整个 repo 的暂存态
- `global_commit_state`
  - 当前整个 repo 的普通提交面

这两组状态不应互相混写。

Orbit 的职责更适合是：

- 清晰展示 orbit-local 视角
- 明确指出哪些变化不属于 orbit-local write 面
- 但不重新定义原生 Git 的全局暂存/提交语义

---

## 7. Projection 内核优化建议

### 7.1 继续保留多层 scope，但收口成正式对象

当前不需要让用户显式管理“多 scope 系统”，但内部必须保留多 scope 语义。

在新的 member-first 模型里，建议把它收口成一个正式 `ProjectionPlan`：

```go
type ProjectionPlan struct {
    OrbitID              string
    ControlPaths         []string
    MetaPaths            []string
    SubjectPaths         []string
    RulePaths            []string
    ProcessPaths         []string
    ProjectionPaths      []string
    OrbitWritePaths      []string
    ExportPaths          []string
    OrchestrationPaths   []string
    PlanHash             string
}
```

它应统一回答：

- Orbit 需要读哪些控制文件？
- 哪些路径属于 `meta`？
- 哪些路径属于 `subject` / `rule` / `process`？
- 当前 view 应显示哪些文件？
- 当前 orbit-local write 应作用哪些文件？
- 当前 export/template 流程应消费哪些文件？
- 当前 orchestration / `AGENTS.md` lane 应消费哪些文件？

### 7.2 建议正式引入 Role-Aware Path Classification

后续状态与 status classification 不应再只保留 `InScope bool`。

建议每个 path 都携带两类信息：

1. `member_role`
2. `scope_flags`

其中 `member_role` 建议固定为：

- `meta`
- `subject`
- `rule`
- `process`
- `outside`

而 `scope_flags` 则表达：

- `projection`
- `orbit_write`
- `export`
- `orchestration`

这样：

- `status`
  - 不只能说“in-scope / out-of-scope”
  - 还能说“这个文件是 process member，可见但默认不可写”
- `commit`
  - 可以稳定只消费 `orbit_write=true` 的成员路径
- `template save`
  - 可以稳定只消费 `export=true` 的成员路径
- `AGENTS.md`
  - 可以稳定只消费 `orchestration=true` 的成员路径与 orbit 元信息

### 7.3 命令消费矩阵

建议冻结如下消费矩阵：

| 命令 | 输入集合 |
| --- | --- |
| `orbit files` | `ProjectionPaths` |
| `orbit enter` | `ProjectionPaths` |
| `orbit status` | `ProjectionPaths` + role-aware path classification |
| `orbit diff` | `OrbitWritePaths` |
| `orbit log` | `OrbitWritePaths` |
| `orbit commit` | `OrbitWritePaths` + orbit-writable-untracked |
| `orbit restore` | `OrbitWritePaths` |
| `orbit template save` | `ExportPaths` |
| `harness template save` | runtime orbit `ExportPaths` merge |
| `AGENTS` / orchestration lane | orbit metadata + `OrchestrationPaths` |

### 7.4 当前第一优先问题

当前第一优先优化项应是：

- 把 Orbit 从 path-list 驱动切到 member-role 驱动
- 把“可见内容”和“orbit-local write 内容”拆开

根因不是 Git，而是当前设计先有 path list、后有 scope 结果，缺少“文件为什么属于这个 orbit”的第一性表达。

因此第一阶段应先把 member role 和 role-scope 映射收稳，再谈更多状态文件。

---

## 8. 存储落点建议

### 8.1 `.orbit/`

继续作为：

- versioned control plane
- OrbitSpec

建议未来允许扩展但仍保持保守：

```text
.orbit/
  config.yaml
  orbits/
    <orbit-id>.yaml
```

orbit yaml 后续建议直接演化为：

```yaml
id: docs
name: Docs
description: User-facing docs orbit
members:
  - key: docs-content
    role: subject
    paths:
      include:
        - docs/**
  - key: docs-process
    role: process
    paths:
      include:
        - docs/process/**
rules:
  scope:
    projection_roles: [meta, subject, process]
    write_roles: [meta]
    export_roles: [meta]
    orchestration_roles: [meta, process]
  orchestration:
    materialize_agents_from_meta: true
```

### 8.2 `.harness/`

继续作为：

- OrbitSharedState / runtime metadata host

建议保持当前合同，不在这里塞 live execution state。

### 8.3 `.git/orbit/state/`

继续作为：

- OrbitLocalState / local projection runtime

建议未来演化为：

```text
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
  runs/
    <orbit-id>.json
  orbit.lock
```

其中：

- `orbits/<orbit-id>/file_inventory.json`
  - 落盘一份当前 orbit 的全部文件说明
  - 例如每个路径属于哪个 member、哪个 role、进入哪些 scope
- `orbits/<orbit-id>/runtime_state.json`
  - 本地运行态
  - 例如 `running`、`phase`、`timestamps`、`plan_hash`
- `orbits/<orbit-id>/git_state.json`
  - Git 状态梳理
  - 例如 `orbit_projection_state`、`orbit_stage_state`、`orbit_commit_state`
  - 以及 `global_stage_state`、`global_commit_state` 的对应摘要

- `runs/<orbit-id>.json`
  - 可选
  - 仅本地执行态
  - 不进入 Git

---

## 9. 分阶段优化路线

### Phase 1：Projection Kernel 收口

目标：

- 让 projection / write / export / orchestration 四种 scope 边界显式化
- 让 scope 明确建立在 member role 之上
- 让 orbit-local commit 明确区别于“普通 Git 的全局提交”

建议改动：

1. 引入 `OrbitMember`
2. 引入 `OrbitMemberRole`
3. 引入 `ProjectionPlan`
4. 引入 role -> scope 映射
5. 把 `status` 改成 role-aware classification
6. 把 untracked path 的 projection / write / export / orchestration 判定拆开

完成标准：

- 每个 path 都能稳定归属到某个 member role
- `status` 可以稳定显示 role 和 scope flags
- `orbit commit / template save / orchestration` 不再依赖单个 `InScope bool`

### Phase 2：OrbitSpec 增强

目标：

- 让 Orbit authoring model 从 path-list 转成 member-list

建议改动：

1. 给 orbit yaml 正式引入 `members`
2. 给 member 增加 `role`
3. 给 member 增加 `paths.include/exclude`
4. 给 member 增加可选 `lane` / `scopes` 覆盖
5. 让 `description` 正式进入 orchestration 输入
6. 明确 `AGENTS.md` 默认由 meta 驱动生成，而不是作为第一性 member

完成标准：

- Orbit 不再只有 `id + include/exclude`
- subject / rule / process 三类文件不再混在同一 path list 里
- `AGENTS.md` 变成基于 meta/rule/process 的派生产物
- process docs / rule files 都能以 member role 稳定归类

### Phase 3：Local Execution State

目标：

- 给 Orbit 本地运行过程信息一个正式宿主

建议改动：

1. 在 `.git/orbit/state/` 下新增 execution state 文件
2. 新增 orbit file inventory / git state ledger
3. 定义 `running / phase / timestamps / plan_hash`
4. 保持 repo-local，不进入 Git

完成标准：

- “是否在执行”“处于什么阶段”不再只能临时推断
- 每条 orbit 都能落盘一份完整文件说明
- orbit/local/global 三套 Git 观察面不再混在一个 snapshot 里
- 不污染 `.orbit/` 和 `.harness/`

### Phase 4：可选的 Shared Workflow State

目标：

- 仅在真实协作场景出现后，再考虑是否需要共享 workflow metadata

约束：

- 不把 Orbit 直接扩成任务系统
- 不把高频 live execution flags 写进 `.harness/`
- 只有当确实存在跨机器协作需求时，才新增 schema-backed shared workflow object

---

## 10. 非目标

本方案明确不做：

1. 不引入多 worktree 并行 Orbit；
2. 不引入数据库、后台服务、远端状态中心；
3. 不把 Orbit 升级为 block-level / semantic workspace system；
4. 不把 `AGENTS.md` 变成 OrbitSpec 的唯一真相源；
5. 不在第一阶段就引入共享 execution state；
6. 不让 `.git/orbit/state/` 成为跨分支协作元数据层。

---

## 11. 一句话总结

后续 Orbit 优化应围绕一个核心方向展开：

**把 Orbit 从“靠 path list 间接推导意图”的系统，收口成“元信息 + 文件成员 + 成员角色 + 角色驱动 scope”的系统；`.orbit/` 管定义，`.harness/` 管共享运行态，`.git/orbit/state/` 管本地运行态。**
