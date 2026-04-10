# Harness Single Control Plane Technical Spec

版本：v0.1
状态：proposal
阶段：post-v0.3 exploration
对应提案：`docs/context/harness_single_control_plane_proposal.md`
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/orbit_member_runtime_technical_spec.md`
- `docs/testing-strategy.md`

---

## 1. 文档目标

本文档把单目录控制面方案收口为可实现的技术合同，重点回答：

1. 版本化控制面如何从 `.orbit/ + .harness/` 收口为单一 `.harness/`；
2. 运行态、orbit template、harness template 三种形态如何共享同一根目录与同一顶层清单；
3. OrbitSpec、branch classification、命令入口和模板/安装链路需要怎样切换；
4. 哪些兼容策略是显式不做的。

本文档不是当前 v0.3 正式 source of truth。
若后续采纳此方向，应先用本文件更新对应 PRD / technical spec，再进入代码实现。

---

## 2. 默认实现决策

本方案没有阻塞性的前置决策点；实现阶段默认采用以下选择：

1. 采用 clean-break，不做 `.orbit/*` 与 `.harness/*` 双写。
2. 保留 `.harness/vars.yaml` 与 `.harness/installs/*.yaml`，不再改动其宿主。
3. 用 `.harness/manifest.yaml` 的 `kind` 作为唯一 branch 顶层识别入口。

非阻塞但明确延后：

1. 不在第一阶段提供自动迁移器；
2. 不在第一阶段兼容读取旧 `.orbit/config.yaml`；
3. 不在第一阶段为旧 layout 提供长期读兼容窗口。

也就是说，本方案把“兼容”视为显式迁移任务，而不是长期双布局共存。

---

## 3. 已冻结结论

### 3.1 版本化控制根

唯一版本化控制根为：

```text
.harness/
```

唯一 repo-local 运行态根继续为：

```text
.git/orbit/state/
```

不再保留 `.orbit/` 作为版本化控制目录。

### 3.2 顶层清单

唯一顶层清单文件为：

```text
.harness/manifest.yaml
```

它负责表达：

- 当前 branch / revision 的形态
- 顶层 runtime 或 template 元信息
- 当前清单声明的 members

### 3.3 OrbitSpec 宿主

所有 OrbitSpec 统一落在：

```text
.harness/orbits/<orbit-id>.yaml
```

其 `meta.file` 固定等于对应 repo-relative 路径。

### 3.4 repo 级 overlay 配置移除

以下 repo 级 authoring 配置不再保留：

- `shared_scope`
- `projection_visible`
- `behavior`
- `.orbit/config.yaml`

对应语义改为：

- `shared_scope`
  - 由多个 OrbitSpec 的显式 member overlap 表达
- `projection_visible`
  - 由 `process` role 表达
- `behavior`
  - 先收成内核固定策略，必要时再升为显式 CLI flag

### 3.5 本地状态边界不变

`.git/orbit/state/*` 继续只承载：

- current orbit
- projection cache
- warnings / status
- orbit-local runtime / git ledger
- lock

本方案不把任何 live execution state 合并进 `.harness/`。

---

## 4. 目录与文件合同

### 4.1 运行态

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml
  vars.yaml
  installs/
    <orbit-id>.yaml

.git/orbit/state/
  ...
```

规则：

1. `manifest.yaml` 必须存在且 `kind=runtime`。
2. `orbits/` 可为空；zero-member runtime 合法。
3. `vars.yaml` 可缺失。
4. `installs/` 可缺失；仅 install-backed orbit 需要对应 record。

### 4.2 Orbit Template

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml

<owned template files...>
AGENTS.md
```

规则：

1. `manifest.yaml` 必须存在且 `kind=orbit_template`。
2. `orbits/` 中必须且只能有一个 OrbitSpec。
3. 不允许存在 `vars.yaml`。
4. 不允许存在 `installs/`。
5. 不允许存在 `.git/orbit/state/*`。

### 4.3 Harness Template

```text
.harness/
  manifest.yaml
  orbits/
    <orbit-id>.yaml

<combined template files...>
AGENTS.md
```

规则：

1. `manifest.yaml` 必须存在且 `kind=harness_template`。
2. `orbits/` 中必须至少包含一个 OrbitSpec。
3. 不允许存在 `vars.yaml`。
4. 不允许存在 `installs/`。
5. 不允许存在 `.git/orbit/state/*`。

---

## 5. `.harness/manifest.yaml` 合同

### 5.1 共同字段

建议顶层结构：

```yaml
schema_version: 1
kind: runtime | orbit_template | harness_template
```

规则：

1. `schema_version` 固定为 `1`。
2. `kind` 必填。
3. 除 `kind` 对应分支外，不允许混用其它分支字段。
4. 解析必须 fail-closed。

### 5.2 Runtime Manifest

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

规则：

1. `runtime.id` 必填，沿用 `orbit-id` 字符安全规则。
2. `members` 必须显式存在，可以为空数组。
3. `members[].orbit_id` 必须唯一。
4. `members[].source` 首阶段只允许：
   - `manual`
   - `install`

### 5.3 Orbit Template Manifest

```yaml
schema_version: 1
kind: orbit_template
template:
  orbit_id: docs
  created_from_branch: main
  created_from_commit: abc123
  created_at: 2026-04-05T11:00:00Z
```

规则：

1. `template.orbit_id` 必填。
2. `orbits/` 中必须存在且只存在 `<orbit_id>.yaml`。
3. `created_from_*` 为模板 provenance，不参与 runtime 行为判断。

### 5.4 Harness Template Manifest

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

规则：

1. `template.harness_id` 必填。
2. `members` 必须显式存在且至少包含一个 member。
3. `members[].orbit_id` 必须唯一。
4. `includes_root_agents` 缺省为 `false`。

---

## 6. OrbitSpec 合同

OrbitSpec 延续成员化模型，但固定宿主变更为：

```text
.harness/orbits/<orbit-id>.yaml
```

其 `meta.file` 固定为：

```text
.harness/orbits/<orbit-id>.yaml
```

规则：

1. `meta.file` 必须与 `id` 对应。
2. `meta` 仍是固定 singleton member，不进入 `members[]`。
3. `subject / rule / process` 继续为显式 `members[]`。
4. `rules.scope.*` 继续是 role -> behavior 的唯一规则入口。

兼容策略：

1. 第一阶段不支持 legacy `include / exclude` schema。
2. 第一阶段不支持 `.orbit/orbits/*.yaml` 读兼容。
3. 若需要迁移旧仓库，应通过显式迁移任务完成，而不是在 steady-state loader 内混读。

---

## 7. Branch Classification 合同

branch classifier 只读取：

```text
.harness/manifest.yaml
```

分类规则：

1. valid manifest with `kind=runtime`
   - `kind=runtime`
2. valid manifest with `kind=orbit_template`
   - `kind=template`, `template_kind=orbit`
3. valid manifest with `kind=harness_template`
   - `kind=template`, `template_kind=harness`
4. 其它
   - `kind=plain`

不再依赖：

- `.orbit/config.yaml`
- `.orbit/template.yaml`
- `.harness/runtime.yaml`
- `.harness/template.yaml`

直接收益：

1. 顶层身份只依赖一个文件；
2. template kind 不再靠多个 marker 文件组合判断；
3. branchinfo / inspect / status / list 的解释口径统一。

---

## 8. 命令层切换合同

### 8.1 Runtime Bootstrap

首阶段正式入口应为：

- `harness init`
- `harness create`

它们负责：

1. 初始化 `.harness/manifest.yaml`
2. 初始化 `.harness/orbits/`
3. 不再创建 `.orbit/config.yaml`

`orbit init` 不再是正式入口。
若保留，应仅作为兼容 wrapper，并输出迁移引导。

### 8.2 Orbit Authoring Commands

以下命令改读写：

- `orbit add`
- `orbit show`
- `orbit list`
- `orbit validate`

切换为：

- 读取 `.harness/orbits/*.yaml`
- 写入 `.harness/orbits/<orbit-id>.yaml`

不再读取 repo 级 global config。

### 8.3 Runtime Projection Commands

以下命令仅允许在 `kind=runtime` 下运行：

- `orbit enter`
- `orbit leave`
- `orbit current`
- `orbit files`
- `orbit status`
- `orbit diff`
- `orbit log`
- `orbit commit`
- `orbit restore`

它们的 scope 解析不再依赖：

- `shared_scope`
- `projection_visible`
- `behavior`

而改为：

1. 读取目标 OrbitSpec
2. 基于 member role 生成 `ProjectionPlan`
3. 由 `ProjectionPlan` 派生 projection / orbit_write / export / orchestration

### 8.4 Template Save / Install

`orbit template save`：

1. 写入 `manifest.kind=orbit_template`
2. 写入目标 OrbitSpec 到 `.harness/orbits/<orbit-id>.yaml`
3. 不写 `vars.yaml`
4. 不写 `installs/`

`harness template save`：

1. 写入 `manifest.kind=harness_template`
2. 写入所有 member OrbitSpec 到 `.harness/orbits/*.yaml`
3. 不写 `vars.yaml`
4. 不写 `installs/`

`harness install`：

1. 从 `manifest.kind=orbit_template` 的源读取单 orbit 模板
2. 写入 runtime manifest members
3. 必要时写入 `.harness/vars.yaml`
4. 必要时写入 `.harness/installs/<orbit-id>.yaml`

---

## 9. 测试合同

### 9.1 Schema / Loader

至少覆盖：

1. runtime manifest codec
2. orbit template manifest codec
3. harness template manifest codec
4. `.harness/orbits/*.yaml` path / id 一致性
5. 非法 kind / 混合字段 fail-closed

### 9.2 Classification

至少覆盖：

1. valid runtime manifest -> runtime
2. valid orbit template manifest -> template/orbit
3. valid harness template manifest -> template/harness
4. 缺失 manifest -> plain
5. manifest 非法 -> invalid / fail-closed

### 9.3 Command Behavior

至少覆盖：

1. `harness init` 不再创建 `.orbit/config.yaml`
2. `orbit add/show/list/validate` 只读写 `.harness/orbits/*`
3. runtime projection commands 在非 runtime branch 下稳定失败
4. `orbit template save` / `harness template save` 写出正确 `kind`
5. `harness install` 从新模板合同成功读取并写回 runtime

### 9.4 Behavior Simplification

至少覆盖：

1. 无 `shared_scope` 时 role overlap 仍可表达共享文件
2. `process` member 只进入 projection / orchestration
3. 无 `projection_visible` 时 projection-only 语义由 `process` 承载
4. 无 repo-level `behavior` 文件时命令行为稳定

---

## 10. 非目标

本方案不做：

1. `.orbit/*` 与 `.harness/*` 双写；
2. 长期 legacy loader 混读；
3. 自动后台迁移旧仓库布局；
4. 共享 live execution state；
5. 把 `.git/orbit/state/*` 合并进版本化控制面。

---

## 11. 一句话总结

单目录控制面方案的技术收口点是：

**用 `.harness/manifest.yaml` 表达 branch 顶层身份，用 `.harness/orbits/*.yaml` 表达所有 OrbitSpec，用角色驱动 scope 取代 repo 级 overlay，并让 `.git/orbit/state/*` 继续单独承载本地运行态。**
