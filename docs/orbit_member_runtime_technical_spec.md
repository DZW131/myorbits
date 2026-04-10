# Orbit Member Runtime Technical Spec

版本：v0.1
状态：specialized supplement under v0.4 baseline
阶段：post-v0.3 enhancement
关联文档：
- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_v0_4_development_plan.md`
- `docs/context/orbit_member_filesystem_behavior.md`
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/orbit_scope_visibility_technical_spec.md`
- `docs/testing-strategy.md`

---

## 1. 文档目标

本文档定义 Orbit 成员化模型的技术合同，重点回答：

1. OrbitSpec 应如何表达固定角色成员；
2. role -> scope 应如何落到当前 projection kernel；
3. `AGENTS.md` 如何改为派生产物；
4. `.git/orbit/state/` 应如何升级为更清晰的本地运行态结构；
5. 当前代码与目标语义之间需要怎样的兼容层。

若本文与 `docs/orbit_v0_4_technical_spec.md` 冲突，以 v0.4 主文档为准。本文只保留 member model、orchestration、ledger 相关专项细节。

---

## 2. 已冻结结论

本方案冻结以下结论：

1. Orbit 固定只有 `meta / subject / rule / process` 四种角色。
2. OrbitSpec 的固定 `meta` 载体是 `<orbit-id>.yaml` companion definition；steady-state runtime host 为 `.harness/orbits/<orbit-id>.yaml`，template / source branch 仍可使用 `.orbit/orbits/<orbit-id>.yaml`。
3. Orbit 固定只有 `projection / orbit_write / export / orchestration` 四个行为面。
4. 原生 Git 的全局暂存与全局提交不属于 Orbit scope。
5. `AGENTS.md` 默认是 orchestration 派生产物，不是基础 authored member。
6. 运行态根 `AGENTS.md` 永远只是整个 harness 的 block 容器，也是 agent 默认读取的入口，不是任一 orbit 的第一性真相源。
7. 每个 orbit 只保存一个会出现在运行态 `AGENTS.md` 中的精简 brief；该 brief 必须可变量化、可物化、可显式回填。
8. 若运行态当前 orbit block 需要回填为真相源，只能通过显式命令 `orbit brief backfill` 写回元信息；不做自动反向同步。
9. orbit template branch 不再携带根 `AGENTS.md`；orbit 级 AGENTS 内容只从 OrbitSpec / orchestration 输入物化。
10. future member-generated block 的手工编辑默认只影响运行态容器；只有显式回填后才改变结构化真相源。
11. `.git/orbit/state/` 仍然不进 Git，但应升级为更明确的 orbit-local ledger。

---

## 3. 对象模型

### 3.1 OrbitSpec

建议目标结构：

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
  agents_template: |
    # Docs orbit
    Work on $project_name documentation using the docs rules and process.

members:
  - key: docs-content
    role: subject
    paths:
      include:
        - docs/**
      exclude: []

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
  orchestration:
    include_orbit_description: true
    materialize_agents_from_meta: true
```

### 3.2 OrbitMember

建议技术结构：

```go
type OrbitMemberRole string

const (
    OrbitMemberMeta    OrbitMemberRole = "meta"
    OrbitMemberSubject OrbitMemberRole = "subject"
    OrbitMemberRule    OrbitMemberRole = "rule"
    OrbitMemberProcess OrbitMemberRole = "process"
)

type OrbitMember struct {
    Key         string                 `yaml:"key"`
    Name        string                 `yaml:"name,omitempty"`
    Description string                 `yaml:"description,omitempty"`
    Role        OrbitMemberRole        `yaml:"role"`
    Paths       OrbitMemberPaths       `yaml:"paths"`
    Lane        string                 `yaml:"lane,omitempty"`
    Scopes      *OrbitMemberScopePatch `yaml:"scopes,omitempty"`
}
```

### 3.3 Meta Member

`meta` 不建议作为普通 `members[]` 条目重复声明。

固定规则：

- `meta.file = 当前 OrbitSpec 宿主路径`
- steady-state runtime host = `.harness/orbits/<orbit-id>.yaml`
- template / source branch host = `.orbit/orbits/<orbit-id>.yaml`
- 它是单例
- 它同时承担：
  - Orbit 元信息宿主
  - companion definition
  - orchestration 的结构化输入源
- `meta.agents_template` 可选：
  - 它保存 variableized 的 orbit brief / AGENTS payload
  - 内容应足够精简，适合被复制进运行态根 `AGENTS.md` 的当前 orbit block
  - 仅在显式回填 / 兼容迁移路径下写入
  - 若存在，则优先于运行态容器中的手工文本

---

## 4. 行为系统

### 4.1 默认 role -> scope

目标语义如下：

| role | projection | orbit_write | export | orchestration |
| --- | --- | --- | --- | --- |
| `meta` | yes | yes | yes | structured-source |
| `subject` | yes | no | no | no |
| `rule` | yes | yes | yes | yes |
| `process` | yes | no | no | yes |

### 4.2 局部覆盖

`members[].scopes.*` 只允许做局部覆盖。

第一版建议只允许：

- `scopes.write`
- `scopes.export`
- `scopes.orchestration`

不建议允许：

- 覆盖 `meta` 的固定 companion 语义
- 改变原生 Git 的 global stage/commit

### 4.3 AGENTS Orchestration

建议默认管线：

1. 若 `meta.agents_template` 非空，优先把它作为 AGENTS body 真相源；
2. 否则读取 `meta.description`；
3. 读取启用 orchestration 的 `rule` 成员；
4. 读取启用 orchestration 的 `process` 成员；
5. 生成当前 orbit 对应的 runtime block payload 或 template payload。

补充约束：

- 运行态根 `AGENTS.md` 是整个 harness 的入口，体积预算有限；
- 因此单个 orbit 的 `meta.agents_template` 应被视为“brief”，而不是无限增长的全文容器；
- 更长的流程、规则或背景材料应继续留在 `rule` / `process` 成员路径中，由 orchestration 选择性消费。

因此：

- `AGENTS.md` 是 materialized artifact
- runtime 根 `AGENTS.md` 只是容器
- `meta.agents_template` / `meta.description` / `rule` / `process` 是 source inputs
- materializer 只拥有“当前 orbit block”，不拥有整个 `AGENTS.md` 文件
- future member-generated block 的手工编辑不会自动改写真相源

### 4.4 显式回填命令

为兼容当前“先编辑 runtime block，再决定是否收编为真相源”的工作流，建议新增显式命令：

- 固定命令名：`orbit brief backfill`

该命令的目标不是修改运行态容器，而是把当前 orbit block 显式回填到结构化元信息。

建议合同：

1. 目标 orbit 默认为 current orbit，也允许 `--orbit <id>`。
2. 命令读取运行态根 `AGENTS.md`，并按现有 parser 验证 marker 合法性。
3. 只提取当前 orbit block 的内部内容；不吸收 unmarked prose，也不吸收其他 orbit block。
4. 对提取内容执行“反向变量化”：
   - 用当前 `.harness/vars.yaml` 中已解析的 runtime value 做 reverse replacement；
   - 成功匹配的值回写为变量名占位，例如 `$project_name`；
   - 未声明的普通文本保持原样；
   - 若 reverse replacement 出现歧义，则 fail-closed，不写入真相源。
5. 写入目标是当前 repo 中该 orbit 的结构化宿主文件；steady-state runtime host 为 `.harness/orbits/<orbit-id>.yaml`，并更新其中的 `meta.agents_template`。
6. 成功写入应以 companion definition 原子改写为准；若无法稳定保留无关字段、注释或 YAML 结构，则 fail-closed。
7. 该命令绝不直接修改运行态 `AGENTS.md`。
8. 若当前 orbit block 缺失、运行态文件非法，或 companion definition 无法稳定改写，则命令 fail-closed。

这条命令是“把运行态编辑收编为真相源”的唯一入口；系统不做自动回填。

建议把行为边界再冻结为以下场景表：

1. 根 `AGENTS.md` 不存在：命令失败，提示当前 runtime 容器缺失。
2. 根文件存在但 marker 非法：命令失败，保持 fail-closed，不尝试容错抽取。
3. 根文件合法但不存在当前 orbit block：命令失败，不从 unmarked prose 猜测内容。
4. 同时存在 unmarked prose、其他 orbit block 和当前 orbit block：只读取当前 orbit block 内部内容，其余全部忽略。
5. block 内容中出现某个已声明变量的唯一 runtime value：回填为对应变量名占位，例如 `$project_name`。
6. block 内容中出现未声明值：按字面量保留，不强行抽象成变量。
7. 两个变量拥有相同 runtime value：reverse replacement 歧义，命令失败。
8. 两个候选变量值发生重叠，导致替换顺序会改变结果：命令失败，不做“最长优先”的隐式猜测。
9. 回填成功后，只更新 `meta.agents_template`；不修改 `meta.description`、`rule`、`process`，也不改 runtime 容器。

### 4.4.1 运行态手工编辑生命周期

future member-generated block 上的手工编辑按下面的合同处理：

1. 开发者直接修改运行态当前 orbit block 后，这些文本立即只存在于 runtime 容器。
2. 只要没有显式执行 `orbit brief backfill`，结构化真相源仍保持旧值。
3. 任何重新 materialize 当前 orbit block 的操作，都可以用结构化真相源覆盖这些运行态手工编辑。
4. 若开发者希望保留这些编辑，必须先执行 `orbit brief backfill`，再进行后续 materialize / save / install 流程。
5. validate / check 可以报告 drift，但不能自动把运行态编辑升级为真相源。

### 4.5 Orbit Template AGENTS Contract

在 member runtime 收敛阶段，orbit template lane 应调整为：

1. orbit template branch 不再写根 `AGENTS.md`。
2. orbit template save / publish 直接消费 `meta.agents_template` 或 orchestration output，而不是再把模板态 `AGENTS.md` 作为文件保存。
3. unmarked prose 视为容器级文本，不默认视为“当前 orbit 所有”。
4. 若运行态当前 orbit block 缺失，则 `orbit brief backfill` 与任何依赖该 block 的兼容抽取都应 fail-closed。
5. 不提供对“legacy orbit template branch 携带根 `AGENTS.md` payload”格式的 backward compatibility。
6. harness template 的根 `AGENTS.md` lane 仍是独立合同，不能因为 orbit template 的收敛而一起删除。

---

## 5. 当前代码兼容层

### 5.1 当前稳定 ScopeSet

当前代码稳定存在的内部结构见现有实现：

- `OwnedPaths`
- `ProjectionOnlyPaths`
- `CompanionPaths`
- `ScopedOperationPaths`
- `ProjectionPaths`

### 5.2 兼容映射

在不立刻重写所有命令的前提下，建议先用如下过渡映射：

| 目标角色 | 当前近似落点 |
| --- | --- |
| `meta` | `CompanionPaths` |
| `rule` | `OwnedPaths` |
| `subject` | 临时 `OwnedPaths` |
| `process` | `ProjectionOnlyPaths` |

重要说明：

- `subject -> OwnedPaths` 只是一条兼容路径
- 它不是最终语义
- 最终语义应把 `subject` 从 orbit-local write/export 中迁出

### 5.3 当前命令消费面

当前代码语义大致是：

- `orbit enter/files/status`
  - 消费 `ProjectionPaths`
- `orbit diff/log/commit/restore`
  - 消费 `ScopedOperationPaths`
- `orbit template save`
  - 消费 `OwnedPaths`，并额外注入 companion definition
- `harness template save`
  - 消费 member `OwnedPaths`

因此下一阶段的改造重点不是推翻命令，而是先把内部 plan 改成 role-aware。

---

## 6. ProjectionPlan

建议目标结构：

```go
type ProjectionPlan struct {
    OrbitID            string
    ControlPaths       []string
    MetaPaths          []string
    SubjectPaths       []string
    RulePaths          []string
    ProcessPaths       []string
    ProjectionPaths    []string
    OrbitWritePaths    []string
    ExportPaths        []string
    OrchestrationPaths []string
    PlanHash           string
}
```

它统一回答：

1. 哪些是控制文件；
2. 哪些路径属于哪个角色；
3. 当前 orbit view 看什么；
4. orbit-local write 作用什么；
5. template/export 消费什么；
6. orchestration 消费什么。

---

## 7. `.git/orbit/state/` 升级

### 7.1 保留文件

继续保留：

- `current_orbit.json`
- `resolved_scope/*.txt`
- `warnings.json`
- `last_status.json`
- `orbit.lock`

### 7.2 新增文件

建议新增：

```text
.git/orbit/state/orbits/<orbit-id>/
  file_inventory.json
  runtime_state.json
  git_state.json
```

### 7.3 `file_inventory.json`

用途：

- 落盘当前 orbit 的完整文件说明

建议最小结构：

```json
{
  "orbit": "docs",
  "generated_at": "2026-04-04T12:00:00Z",
  "files": [
    {
      "path": "docs/guide.md",
      "member_key": "docs-content",
      "role": "subject",
      "projection": true,
      "orbit_write": false,
      "export": false,
      "orchestration": false
    }
  ]
}
```

### 7.4 `runtime_state.json`

用途：

- 本地运行状态，不进 Git

建议最小结构：

```json
{
  "orbit": "docs",
  "running": true,
  "phase": "planning",
  "entered_at": "2026-04-04T12:00:00Z",
  "updated_at": "2026-04-04T12:05:00Z",
  "plan_hash": "..."
}
```

### 7.5 `git_state.json`

用途：

- 把 orbit 视角和普通 Git 视角并列整理

建议最小结构：

```json
{
  "orbit": "docs",
  "orbit_projection_state": {},
  "orbit_stage_state": {},
  "orbit_commit_state": {},
  "global_stage_state": {},
  "global_commit_state": {}
}
```

---

## 8. 系统层 Orbit 文件

### 8.1 版本化控制文件

- `.orbit/config.yaml`
- `.harness/orbits/<orbit-id>.yaml`（steady-state runtime host）
- `.orbit/orbits/<orbit-id>.yaml`（template / source branch host）
- `.orbit/template.yaml`
- `.orbit/source.yaml`

### 8.2 repo-local 运行态文件

- `.git/orbit/state/current_orbit.json`
- `.git/orbit/state/resolved_scope/<orbit-id>.txt`
- `.git/orbit/state/warnings.json`
- `.git/orbit/state/last_status.json`
- `.git/orbit/state/orbit.lock`

### 8.3 优化空间

最明显的优化空间有三处：

1. `resolved_scope/*.txt`
   - 只能表达路径列表
   - 不能表达 role 与行为面
2. `warnings.json` 与 `last_status.json`
   - 适合后续并入更统一的 `git_state.json` 观察面
3. `AGENTS.md`
   - 适合从基础 authored file 退到 orchestration 派生产物

---

## 9. 测试要求

新增实现至少应覆盖：

1. role -> scope 映射单测
2. `meta` 固定 companion 行为单测
3. `subject` 不进入 orbit_write/export 的语义测试
4. `process` 只进入 projection/orchestration 的语义测试
5. `file_inventory.json` / `runtime_state.json` / `git_state.json` 读写测试
6. AGENTS 派生输入选择测试
