# Orbit Member Filesystem And Behavior

版本：v0.1
状态：proposal
阶段：post-v0.3 enhancement
关联文档：
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/orbit_scope_visibility_technical_spec.md`
- `docs/testing-strategy.md`

---

## 1. 文档目标

这份文档只做一件事：

- 用最少但清晰的内容，固定 Orbit 下一阶段的文件系统与行为系统。

它不展开实现细节，不替代技术方案，也不展开开发排期。

---

## 2. 一句话定义

**一个 orbit = `.orbit/` 中的结构化元信息 + 固定角色的文件成员集合；Orbit 只定义自己的 view / orbit-local write / export / orchestration 行为，不重定义原生 Git 的全局暂存与全局提交。**

---

## 3. 核心结论

### 3.1 Orbit 的固定角色

Orbit 固定只有四种角色：

- `meta`
  - orbit 自身的元信息控制面
  - 固定载体是 `.orbit/orbits/<orbit-id>.yaml`
- `subject`
  - orbit 关注和观察的业务对象
- `rule`
  - orbit 自身迭代和约束用的规则文件
- `process`
  - orbit 的过程文件，例如 plan、issue、playbook、rolling docs

### 3.2 Orbit 的固定行为面

Orbit 固定只有四个行为面：

- `projection`
  - 当前 orbit view 中显示什么
- `orbit_write`
  - `orbit diff / log / commit / restore` 作用什么
- `export`
  - `orbit template save / publish` 与 `harness template save` 消费什么
- `orchestration`
  - 哪些结构化输入用于生成 `AGENTS.md` 或后续编排产物

### 3.3 AGENTS.md 的定位

- `AGENTS.md` 默认不是第一性 authored member。
- `AGENTS.md` 默认是 orchestration lane 的派生产物。
- 它建立在：
  - `meta.description`
  - 选中的 `rule` 成员
  - 选中的 `process` 成员
  之上。

### 3.4 Orbit 不接管原生 Git

- 全局暂存态与全局提交态继续是原生 Git 的能力。
- Orbit 只定义 orbit-local 行为。
- Orbit 可以观察和提示全局 Git 状态，但不重新定义它。

---

## 4. 文件系统

### 4.1 `.orbit/`

`.orbit/` 是版本化控制面。

```text
.orbit/
  config.yaml
  orbits/
    <orbit-id>.yaml
  template.yaml      # 仅 orbit template branch
  source.yaml        # 仅 source branch
```

职责：

- 定义 orbit 元信息
- 定义 orbit 成员与角色
- 定义 role -> scope 规则
- 定义 orchestration 输入规则

### 4.2 `.harness/`

`.harness/` 是版本化运行态协作面。

```text
.harness/
  runtime.yaml
  vars.yaml
  installs/
    <orbit-id>.yaml
  bundles/
    <harness-id>.yaml
  template.yaml
```

职责：

- runtime identity
- runtime members
- install provenance
- shared vars

不负责：

- live execution state
- orbit 本地运行阶段

### 4.3 `.git/orbit/state/`

`.git/orbit/state/` 是 repo-local 运行态。

建议落点：

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
  orbit.lock
```

职责：

- 当前进入的 orbit
- projection cache
- 本地运行状态
- 本地 Git 状态整理
- orbit 文件清单落盘

---

## 5. 默认行为

### 5.1 默认 role -> behavior 映射

| role | projection | orbit_write | export | orchestration |
| --- | --- | --- | --- | --- |
| `meta` | yes | yes | yes | structured-source |
| `subject` | yes | no | no | no |
| `rule` | yes | yes | yes | yes |
| `process` | yes | no | no | yes |

解释：

- `meta`
  - 进入 orbit 自身的控制、导出与编排
- `subject`
  - 默认只进入 orbit view
- `rule`
  - 是 orbit-local 迭代的主对象
- `process`
  - 默认只提供上下文和编排输入

### 5.2 Git 状态观察面

Orbit 需要区分两组状态：

#### Orbit 视角

- `orbit_projection_state`
- `orbit_stage_state`
- `orbit_commit_state`

#### 普通 Git 视角

- `global_stage_state`
- `global_commit_state`

Orbit 的职责是：

- 展示 orbit 视角
- 对 global 视角做提示和梳理
- 不改变原生 Git 语义

---

## 6. 系统层 Orbit 文件

当前系统层 Orbit 文件建议明确为：

### 6.1 版本化控制文件

- `.orbit/config.yaml`
- `.orbit/orbits/<orbit-id>.yaml`
- `.orbit/template.yaml`
- `.orbit/source.yaml`

### 6.2 repo-local 运行态文件

- `.git/orbit/state/current_orbit.json`
- `.git/orbit/state/resolved_scope/<orbit-id>.txt`
- `.git/orbit/state/warnings.json`
- `.git/orbit/state/last_status.json`
- `.git/orbit/state/orbit.lock`

### 6.3 下一阶段建议新增

- `.git/orbit/state/orbits/<orbit-id>/file_inventory.json`
- `.git/orbit/state/orbits/<orbit-id>/runtime_state.json`
- `.git/orbit/state/orbits/<orbit-id>/git_state.json`

---

## 7. 优化方向

最重要的优化点只有三个：

1. 把 Orbit 从 path-list 心智收口为 member-role 心智。
2. 把 `AGENTS.md` 从基础 authored file 改成 orchestration 派生产物。
3. 把 `.git/orbit/state/` 从零散状态文件升级为“current orbit + 文件清单 + 本地运行态 + Git 状态梳理”的清晰结构。
