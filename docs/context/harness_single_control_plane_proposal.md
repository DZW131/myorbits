# Harness Single Control Plane Proposal

版本：v0.1
状态：proposal
阶段：post-v0.3 exploration
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/context/orbit_member_filesystem_behavior.md`
- `docs/orbit_member_runtime_technical_spec.md`

---

## 1. 文档目标

这份文档用于把“`.orbit/` 与 `.harness/` 合并成单一版本化控制面”落成一份可讨论的提案。

它要回答的问题是：

1. 是否可以把当前 `.orbit/` 的版本化定义迁入 `.harness/`；
2. 若采用单目录控制面，运行态、orbit template、harness template 三种形态各自最少需要哪些文件；
3. `shared_scope`、`projection_visible`、`behavior` 这类 repo 级配置在成员化模型下是否还应继续保留；
4. 如何在减少用户心智负担的同时，仍保留清晰的状态边界。

本提案不是当前 v0.3 的正式 source of truth。
它与现有 v0.3 文档存在边界差异，只有在后续正式更新 PRD / technical spec 后，才应进入实现阶段。

---

## 2. 问题定义

当前模型里，用户需要同时理解两套版本化隐藏目录：

- `.orbit/`
  - orbit 定义
  - projection control plane
  - orbit template manifest
- `.harness/`
  - runtime identity
  - runtime vars
  - install records
  - harness template manifest

这在工程边界上是清楚的，但在用户心智上有三个问题：

1. “版本化控制文件到底该先看哪个目录”不够直观；
2. 运行态、模板态、orbit 定义被拆在两处，增加查找成本；
3. `.orbit/config.yaml` 中的 `shared_scope` / `projection_visible` / `behavior` 仍然让用户保留 path-list 思维，而不是转向成员化 OrbitSpec。

如果后续目标是把 Orbit 收口为：

- 元信息
- 文件成员
- 成员角色
- 角色驱动 scope

那么继续保留 `.orbit/` 与 `.harness/` 两套版本化控制目录，收益已经开始低于成本。

---

## 3. 一句话提案

**把版本化控制面统一收口到 `.harness/`；用一个根清单文件表达“当前这条分支是什么”，用 `.harness/orbits/<orbit-id>.yaml` 表达每个 OrbitSpec；`.git/orbit/state/` 继续保留为唯一 repo-local 运行态。**

这意味着：

- `.orbit/` 退出版本化控制面；
- `.harness/` 成为唯一版本化控制根；
- OrbitSpec 的 authored 宿主统一为 `.harness/orbits/<orbit-id>.yaml`；
- 运行态与模板态都由 `.harness/manifest.yaml` 的 `kind` 决定；
- `shared_scope` / `projection_visible` / `behavior` 不再作为主 authoring model。

---

## 4. 设计原则

### 4.1 一个版本化控制根

用户只需要记住：

- `.harness/`
  - 版本化控制面
- `.git/orbit/state/`
  - 本地运行态

不再需要在 `.orbit/` 与 `.harness/` 之间做第一次判断。

### 4.2 一个根清单文件

统一引入：

```text
.harness/manifest.yaml
```

它回答的不是“某个 orbit 是什么”，而是：

- 当前这条 revision / branch 是 runtime、orbit template、还是 harness template；
- 当前分支对应的顶层元信息是什么；
- 当前需要怎样解释 `.harness/orbits/*.yaml`。

### 4.3 Orbit 定义统一放一处

所有 OrbitSpec 都统一放在：

```text
.harness/orbits/<orbit-id>.yaml
```

不再区分：

- orbit runtime 用 `.orbit/orbits/*`
- harness template 又额外带一份 orbit definitions

这样可以把“Orbit 自己是什么”与“当前 branch 是哪种形态”拆开。

### 4.4 不把所有内容塞进一个大 YAML

本提案主张“一个目录”，不主张“一个文件”。

原因：

1. runtime identity、orbit definitions、bindings、install provenance 的变化频率不同；
2. 单个 orbit 的定义天然应独立演化，不适合内联进一个巨大根文件；
3. install record 按 orbit 拆文件，更容易局部更新、审查与冲突合并；
4. 用户只需要记住一个目录，不需要牺牲对象边界。

---

## 5. 建议文件系统

统一后的版本化控制面：

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml
  vars.yaml              # 仅 runtime 按需出现
  installs/
    <orbit-id>.yaml      # 仅 runtime 按需出现
```

repo-local 运行态保持：

```text
.git/orbit/state/
  current_orbit.json
  resolved_scope/
  warnings.json
  last_status.json
  orbit.lock
```

本提案下：

- `.harness/manifest.yaml`
  - 当前 branch / revision 的顶层身份
- `.harness/orbits/<orbit-id>.yaml`
  - OrbitSpec
- `.harness/vars.yaml`
  - runtime 级共享 bindings
- `.harness/installs/<orbit-id>.yaml`
  - runtime 级安装来源记录

---

## 6. 三种形态的最小文件合同

### 6.1 运行态

运行态仓库建议最小结构：

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml
  vars.yaml              # 可选
  installs/
    <orbit-id>.yaml      # 仅 install-backed orbit 才有

.git/orbit/state/
  ...
```

建议 `manifest.yaml`：

```yaml
schema_version: 1
kind: runtime
runtime:
  id: project-a
  name: Project A
  created_at: 2026-04-05T10:00:00Z
  updated_at: 2026-04-05T10:30:00Z
members:
  - orbit_id: docs
    source: manual
    added_at: 2026-04-05T10:05:00Z
  - orbit_id: cli
    source: install
    added_at: 2026-04-05T10:10:00Z
```

运行态下：

- `.harness/orbits/*.yaml`
  - 当前 runtime 中可用的 OrbitSpec 定义
- `.harness/vars.yaml`
  - 只在需要变量绑定时存在
- `.harness/installs/*.yaml`
  - 只为 install-backed member 保留 provenance
- `.git/orbit/state/*`
  - 继续负责 current orbit、projection cache、warnings、status、本地锁

运行态下明确不应出现：

- `.orbit/config.yaml`
- `.orbit/template.yaml`
- `.harness/template.yaml`

### 6.2 Orbit Template

单 orbit 模板建议最小结构：

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml

<owned template files...>
AGENTS.md            # 仅当该 orbit 模板确实导出它时出现
```

建议 `manifest.yaml`：

```yaml
schema_version: 1
kind: orbit_template
template:
  orbit_id: docs
  created_from_branch: main
  created_from_commit: abc123
  created_at: 2026-04-05T11:00:00Z
```

Orbit template 下：

- 只允许一个 OrbitSpec
- `vars.yaml` 不进入模板
- `installs/` 不进入模板
- 不包含 repo-local `.git/orbit/state/*`

这样用户只需看：

1. `manifest.kind=orbit_template`
2. `template.orbit_id`
3. `.harness/orbits/<orbit-id>.yaml`

### 6.3 Harness Template

harness 模板建议最小结构：

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml

<combined template files...>
AGENTS.md            # 仅当组合结果包含它时出现
```

建议 `manifest.yaml`：

```yaml
schema_version: 1
kind: harness_template
template:
  harness_id: project-a
  created_from_branch: main
  created_from_commit: abc123
  created_at: 2026-04-05T12:00:00Z
members:
  - orbit_id: docs
  - orbit_id: cli
includes_root_agents: true
```

Harness template 下：

- `vars.yaml` 不进入模板
- `installs/` 不进入模板
- 不再需要单独的 `.orbit/template.yaml`
- 模板的所有 member orbit 定义仍统一落在 `.harness/orbits/*.yaml`

---

## 7. OrbitSpec 的新宿主

本提案下，OrbitSpec 不再放在 `.orbit/orbits/<orbit-id>.yaml`，而是改为：

```text
.harness/orbits/<orbit-id>.yaml
```

其核心结构保持成员化模型：

```yaml
id: docs
name: Documentation
description: User-facing docs orbit.

meta:
  file: .harness/orbits/docs.yaml
  include_in_projection: true
  include_in_write: true
  include_in_export: true
  include_description_in_orchestration: true

members:
  - key: docs-content
    role: subject
    paths:
      include:
        - docs/**
        - README.md

  - key: docs-rules
    role: rule
    paths:
      include:
        - .markdownlint.yaml
        - vale.ini

  - key: docs-process
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
```

这里的关键变化是：

- `meta.file` 也统一指向 `.harness/orbits/<orbit-id>.yaml`
- Orbit 的 authored 定义与 runtime/template 的 branch kind 解耦
- OrbitSpec 自己不再依赖 repo-level `.orbit/config.yaml`

---

## 8. 配置文件精简建议

### 8.1 删除 `.orbit/config.yaml`

本提案建议移除整个 `.orbit/config.yaml`。

原因：

1. 它会把用户拉回“repo 级 overlay + path-list”思维；
2. 它让 OrbitSpec 之外又多一份必须理解的配置入口；
3. 在成员化模型下，scope 更适合从 member role 推导，而不是从 repo 级 pattern overlay 推导。

### 8.2 移除 `shared_scope`

`shared_scope` 的语义是：

- repo 级共享 owned scope

但在成员化模型下，更直接的表达方式是：

- 把该文件显式放进某个 orbit 的 `rule` 或 `process` member；
- 如果多个 orbit 都需要它，就在多个 OrbitSpec 里显式声明；
- overlap 成为显式建模结果，而不是 repo 级隐式叠加。

因此本提案建议：

- 不再保留 repo 级 `shared_scope`
- 不再把它作为用户 authoring model

### 8.3 移除 `projection_visible`

`projection_visible` 的存在，本质上是在旧模型里补出“只可见、不属于 write / export”的文件层。

在成员化模型下，这个语义天然对应：

- `process`
  - visible
  - orchestration
  - 默认不进入 orbit_write / export

因此本提案建议：

- 不再保留 repo 级 `projection_visible`
- 让“只投影可见”的语义改由 `process` 成员承载

### 8.4 移除 repo 级 `behavior`

当前 `behavior` 中的大部分字段已经接近冻结常量，而不是高价值配置：

- `outside_changes_mode`
- `block_switch_if_hidden_dirty`
- `commit_append_trailer`
- `sparse_checkout_mode`

本提案建议：

1. 这些行为先收成内核固定策略；
2. 若未来确实需要可调，再优先考虑显式 CLI flag；
3. 不把它们继续保留为 repo 级版本化配置。

这样可以避免用户为了使用 Orbit 还要先理解一组“看起来可配、实际上几乎不该改”的参数。

---

## 9. 分支识别简化

当前模型下，branch classifier 需要组合判断：

- `.harness/template.yaml`
- `.orbit/template.yaml`
- `.harness/runtime.yaml`
- `.orbit/config.yaml`

本提案建议简化为只看一个文件：

```text
.harness/manifest.yaml
```

分类规则变成：

1. valid `.harness/manifest.yaml` with `kind=runtime`
   - `kind=runtime`
2. valid `.harness/manifest.yaml` with `kind=orbit_template`
   - `kind=template`, `template_kind=orbit`
3. valid `.harness/manifest.yaml` with `kind=harness_template`
   - `kind=template`, `template_kind=harness`
4. 其它
   - `kind=plain`

直接收益：

- 不再需要靠多个 marker 文件组合判断；
- 不再需要同时理解 `.orbit/template.yaml` 与 `.harness/template.yaml` 的差异；
- runtime / template 的顶层身份与 orbit definitions 解耦。

---

## 10. 为什么这套方案更省心

对用户来说，需要记住的内容从：

1. `.orbit/config.yaml`
2. `.orbit/orbits/*.yaml`
3. `.orbit/template.yaml`
4. `.harness/runtime.yaml`
5. `.harness/vars.yaml`
6. `.harness/installs/*.yaml`
7. `.harness/template.yaml`

收成：

1. `.harness/manifest.yaml`
2. `.harness/orbits/*.yaml`
3. `.harness/vars.yaml`（仅 runtime 按需）
4. `.harness/installs/*.yaml`（仅 runtime 按需）
5. `.git/orbit/state/*`

也就是说：

- 版本化控制面只有一个根目录；
- 顶层身份只有一个入口文件；
- Orbit 定义永远在一个固定位置；
- runtime 附加文件只在 runtime 里按需出现；
- 模板态天然比运行态更瘦。

---

## 11. 非目标

本提案不做：

1. 不把 `.git/orbit/state/*` 并入 `.harness/`；
2. 不把所有内容合成一个巨型 YAML；
3. 不引入共享 live execution state；
4. 不把 Orbit 重新扩成 repo 级 path overlay 系统；
5. 不在本提案中定义完整迁移步骤与兼容窗口。

---

## 12. 一句话总结

如果后续优化目标是“减少用户心智负担”，比起继续维护 `.orbit/` 与 `.harness/` 两套版本化控制目录，更合理的方向是：

**让 `.harness/` 成为唯一版本化控制根，用 `.harness/manifest.yaml` 表达 branch kind，用 `.harness/orbits/<orbit-id>.yaml` 表达 OrbitSpec，用角色驱动 scope 取代 repo 级 `shared_scope` / `projection_visible` / `behavior`。**
