# Orbit Extraction Guide

状态：working guide
适用范围：把某个 workflow 拆成严格符合 v0.4 定义的 orbit

若本文与下列文档冲突，以它们为准：
- `docs/orbit_author_guide.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_positioning_and_personas.md`

---

## 1. 目标

本指南只解决一件事：

**把一个 workflow 拆成若干个“可复用工作单元” orbit，并为每个 orbit 产出最小合规文件。**

orbit 不是一步命令，不是一堆文件，也不是一个大而全流程。

---

## 2. 合规底线

每个 orbit 必须同时满足：

1. 它是一个可复用工作单元，而不是一次性的步骤切片。
2. 它能独立回答这 6 个问题：
   - 负责什么，不负责什么
   - 进入后先看什么
   - 哪些内容只需可见，哪些内容可写
   - 哪些内容应进入导出
   - done probe 是什么
   - 结果记录到哪里
3. 它写清这 8 项 authored contract：
   - `objective`
   - `scope boundary`
   - `rules`
   - `done probe`
   - `failure condition`
   - `abnormal exit hint`
   - `record target`
   - `record minimum`
4. 它使用正式角色模型：
   - `meta`
   - `subject`
   - `rule`
   - `process`
5. 它不混淆 4 个 surface：
   - `projection`
   - `orbit_write`
   - `export`
   - `orchestration`

---

## 3. 何时拆 orbit

当 workflow 进入下一段后，以下任一项发生明显变化，就应考虑切 orbit：

- `objective`
- `rules`
- `projection`
- `orbit_write`
- `done probe`
- `record target`
- `failure / abnormal exit`

不要按“步骤数量”拆。
要按“工作模式变化”拆。

---

## 4. 何时不要拆

以下情况不要单独做 orbit：

- 只是同一目标下的一个小步骤
- 只是一个命令或工具动作
- 只是为了让 prompt 变短
- 没有独立 done probe
- 没有独立 record target
- 只能解释成“实现细节”，不能解释成“工作单元”

---

## 5. 每个 orbit 的最小文件

每个 orbit 至少产出 4 个文件：

```text
.harness/orbits/<orbit-id>.yaml
docs/orbits/<orbit-id>/subject.md
docs/orbits/<orbit-id>/rules.md
docs/orbits/<orbit-id>/process.md
```

说明：

- `<orbit-id>.yaml`
  - 宿主配置
  - `meta.agents_template` 是 brief 的结构化真相源
- `subject.md`
  - 工作对象或本 orbit 的主要产物
- `rules.md`
  - 规则、边界、done probe、failure、record contract
- `process.md`
  - 进入条件、步骤、handoff 条件

如果 `subject.md` 是该 orbit 的主要产物，通常应显式加：

```yaml
scopes:
  write: true
  export: true
```

因为 `subject` 默认不进入 `orbit_write` / `export`。

---

## 6. 抽取流程

### Step 1：先把 workflow 正规化

先把原 workflow 写成若干段，每段只描述：

- 当前目标
- 当前输入
- 当前输出
- 当前规则
- 当前完成信号

不要先写 orbit 名字。

### Step 2：按工作模式切段

用第 3 节的变化项检查每一段。

若两段的目标、规则、可写范围、完成标准都差不多，就不要拆。

### Step 3：为每段做“可复用工作单元”判断

只有满足以下条件，才能升级成 orbit：

- 能复用到同类任务
- 有稳定输入
- 有稳定输出
- 有独立 done probe
- 有独立 record target

### Step 4：为每个 orbit 定义 4 个角色

- `meta`
  - orbit 描述与 brief
- `subject`
  - 被处理对象或主要产物
- `rule`
  - 规则、约束、probe
- `process`
  - 操作步骤与 handoff

### Step 5：写 4 个文件

先写最小版，不要贪多。

### Step 6：做合规检查

逐个 orbit 检查：

- 是否回答了 6 个问题
- 是否写齐 8 项 contract
- 是否区分了 4 个 surface
- 是否只是工作单元，而不是步骤
- 是否把根 `AGENTS.md` 当成 authored truth

---

## 7. YAML 约束

宿主文件必须满足：

- 路径固定为 `.harness/orbits/<orbit-id>.yaml`
- `meta.file` 必须等于该路径
- `meta` 不重复写进 `members[]`
- `members[]` 只使用：
  - `subject`
  - `rule`
  - `process`
- `meta.agents_template` 只写短而稳的 brief

---

## 8. 精炼模板

### 8.1 YAML

```yaml
id: <orbit-id>
name: <Orbit Name>
description: >
  <One-sentence objective>

meta:
  file: .harness/orbits/<orbit-id>.yaml
  include_in_projection: true
  include_in_write: true
  include_in_export: true
  include_description_in_orchestration: true
  agents_template: |
    # <orbit-id>
    You are responsible for <objective>.
    Read `docs/orbits/<orbit-id>/rules.md`.
    Follow `docs/orbits/<orbit-id>/process.md`.
    Write results in `docs/orbits/<orbit-id>/subject.md`.

members:
  - key: main-subject
    name: Main Subject
    description: Main work object or artifact.
    role: subject
    paths:
      include:
        - docs/orbits/<orbit-id>/subject.md
    scopes:
      write: true
      export: true

  - key: main-rules
    name: Main Rules
    description: Rules and guardrails.
    role: rule
    paths:
      include:
        - docs/orbits/<orbit-id>/rules.md

  - key: main-process
    name: Main Process
    description: Workflow and handoff steps.
    role: process
    paths:
      include:
        - docs/orbits/<orbit-id>/process.md
    scopes:
      export: true

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

### 8.2 `subject.md`

```md
# Subject

## Objective

## Scope
- In scope:
- Out of scope:

## Main Output

## Done

## Handoff
```

### 8.3 `rules.md`

```md
# Rules

## Objective

## Must Do

## Must Not Do

## Done Probe

## Failure Condition

## Abnormal Exit Hint

## Record Target

## Record Minimum
```

### 8.4 `process.md`

```md
# Process

## Trigger

## Inputs

## Steps
1.
2.
3.

## Handoff
```

---

## 9. AI 提示模板

把下面这段直接给 AI：

```text
请根据 `docs/orbit_extraction_guide.md`，从下面这个 workflow 中抽取 orbit。

要求：
1. 只抽取严格符合 v0.4 定义的 orbit。
2. orbit 必须是“可复用工作单元”，不是步骤。
3. 每个 orbit 都要回答 6 个问题，并写齐 8 项 authored contract。
4. 每个 orbit 都要输出 4 个文件：
   - .harness/orbits/<orbit-id>.yaml
   - docs/orbits/<orbit-id>/subject.md
   - docs/orbits/<orbit-id>/rules.md
   - docs/orbits/<orbit-id>/process.md
5. 使用正式角色：meta / subject / rule / process。
6. 不要把根 AGENTS.md 当 authored truth。
7. 输出尽量精炼，不要写成长篇手册。

workflow:
<在这里粘贴 workflow>
```

---

## 10. 最后检查

如果一个候选 orbit 说不清下面任何一项，就不要通过：

- 它到底负责什么
- 它写什么，不写什么
- 什么算完成
- 结果记到哪里
- 它为什么值得复用

通过标准只有一个：

**它是一个边界清楚、可复用、可进入、可退出、可记录的工作单元。**
