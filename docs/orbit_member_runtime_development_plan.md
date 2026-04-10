# Orbit Member Runtime Development Plan

版本：v0.1
状态：specialized supplement under v0.4 baseline
阶段：post-v0.3 enhancement
关联文档：
- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_v0_4_development_plan.md`
- `docs/context/orbit_member_filesystem_behavior.md`
- `docs/orbit_member_runtime_technical_spec.md`
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/testing-strategy.md`

---

## 1. 文档目标

本文档把 Orbit 成员化模型的实现拆成可执行阶段，回答：

1. 先做什么，后做什么；
2. 哪些内容可以兼容落地；
3. 哪些行为要延后切换；
4. 每阶段的测试和完成标准是什么。

若本文与 `docs/orbit_v0_4_development_plan.md` 冲突，以 v0.4 主文档为准。本文只保留 member runtime / orchestration lane 的专项排期。

---

## 2. 开发原则

整个推进过程遵守以下原则：

1. 先收口内部对象模型，再切命令行为。
2. 先建立兼容层，再切换 authoring model。
3. 不把原生 Git 的全局暂存/提交重新抽象进 Orbit scope。
4. `AGENTS.md` 先改成派生模型，再考虑 materialization 细节。
5. `.git/orbit/state/` 的新增文件先做本地 ledger，不进入 Git。

---

## 3. 阶段总览

建议分五个阶段：

1. 文档冻结与对象模型引入
2. OrbitSpec 兼容解析与校验
3. ProjectionPlan 与 role-aware classification
4. 本地状态文件升级
5. AGENTS orchestration 与行为切换

---

## 4. Phase 1：文档冻结与对象模型引入

### 4.1 目标

先把正式文档与代码内数据结构引入好，但不改变 CLI 对外行为。

### 4.2 任务

1. 新增正式总览文档。
2. 新增正式技术方案文档。
3. 新增正式开发计划文档。
4. 在代码中引入：
   - `OrbitMemberRole`
   - `OrbitMember`
   - `ProjectionPlan`
5. 暂不切换现有命令消费面。

### 4.3 测试

1. 纯结构体与枚举单测。
2. YAML codec 单测。
3. 未使用新字段时，现有解析不回归。

### 4.4 完成标准

- 文档齐备。
- 代码内已有目标对象模型。
- 现有命令行为无变化。

---

## 5. Phase 2：OrbitSpec 兼容解析与校验

### 5.1 目标

让 Orbit 同时接受：

- 旧模型：`id + description + include + exclude`
- 新模型：`meta + members + rules`

### 5.2 任务

1. 扩展 OrbitSpec 解析器：
   - steady-state runtime host = `.harness/orbits/*.yaml`
   - template / source branch host = `.orbit/orbits/*.yaml`
2. 引入兼容层：
   - 旧模型可自动投影到新内部模型
   - 新模型可直接进入新内部模型
3. 增加校验：
   - 固定角色合法性
   - member key 唯一性
   - meta file 与 orbit id 一致性
   - role-scope 配置合法性
4. `orbit add`
   - 先保持旧 skeleton 或提供新 skeleton 开关
5. `orbit show`
   - 能稳定展示新结构

### 5.3 测试

1. 旧 schema 兼容测试。
2. 新 schema 解析测试。
3. 冲突与非法字段校验测试。
4. add/show 回归测试。

### 5.4 完成标准

- 新旧 schema 都能稳定解析。
- 旧仓库不需要立刻迁移。
- 新文档中的 OrbitSpec 已能被代码理解。

---

## 6. Phase 3：ProjectionPlan 与 role-aware classification

### 6.1 目标

把当前 path-list 内核升级成 role-aware plan，但尽量不打断命令主路径。

### 6.2 任务

1. 基于成员角色构建：
   - `MetaPaths`
   - `SubjectPaths`
   - `RulePaths`
   - `ProcessPaths`
2. 生成：
   - `ProjectionPaths`
   - `OrbitWritePaths`
   - `ExportPaths`
   - `OrchestrationPaths`
3. `status` 改为 role-aware classification。
4. tracked/untracked 分类都输出 role 与 scope flags。
5. 保持当前命令外壳不变：
   - `enter/files/status` 继续消费 projection
   - `diff/log/commit/restore` 改消费 orbit_write
   - `template save` 改消费 export

### 6.3 测试

1. role -> scope 解析矩阵测试。
2. tracked role classification 测试。
3. untracked role classification 测试。
4. `subject` 不再进入 orbit_write/export 的行为测试。
5. `rule` 进入 orbit_write/export 的行为测试。
6. `process` 仅进入 projection/orchestration 的行为测试。

### 6.4 完成标准

- 每个 path 都能稳定归属角色。
- orbit-local 行为不再依赖单个 `InScope bool`。
- `subject` 与 `rule/process` 的行为差异被正式落地。

---

## 7. Phase 4：本地状态文件升级

### 7.1 目标

把 `.git/orbit/state/` 从零散状态文件升级为清晰 ledger。

### 7.2 任务

1. 新增：
   - `orbits/<orbit-id>/file_inventory.json`
   - `orbits/<orbit-id>/runtime_state.json`
   - `orbits/<orbit-id>/git_state.json`
2. 保留现有：
   - `current_orbit.json`
   - `resolved_scope/*.txt`
   - `warnings.json`
   - `last_status.json`
3. 让新文件优先用于观察与调试。
4. 暂不删除旧文件。

### 7.3 测试

1. 三个新文件的读写单测。
2. 原子写测试。
3. 与现有状态文件并存测试。
4. 读取缺失/损坏文件的恢复测试。

### 7.4 完成标准

- 每条 orbit 都能落盘完整文件说明。
- 本地运行状态与 Git 状态梳理可单独读取。
- 不破坏现有状态文件读取链路。

---

## 8. Phase 5：AGENTS Orchestration 与行为切换

### 8.1 目标

把 `AGENTS.md` 从 authored file 模型切到 orchestration 派生模型，并明确：

- runtime 根 `AGENTS.md` 只是 block 容器
- 结构化元信息才是第一性真相源
- 运行态手工编辑若要升级为真相源，必须走显式回填命令
- orbit template 不再携带模板态根 `AGENTS.md`

### 8.2 任务

1. 在 schema 中为 `meta` 增加可选 `agents_template` 字段：
   - 保存 variableized 的 orbit brief / AGENTS payload
   - 作为进入运行态当前 orbit block 的首选真相源
2. 定义 orchestration 输入优先级：
   - 先读 `meta.agents_template`
   - 若为空，再读 `meta.description`
   - 再读选中的 `rule`
   - 再读选中的 `process`
3. 实现 AGENTS payload builder。
4. 新增显式回填命令 `orbit brief backfill`：
   - 从运行态根 `AGENTS.md` 中只提取当前 orbit block
   - 反向变量化后写回 `meta.agents_template`
   - 不自动吸收 unmarked prose
5. 保持与当前运行态 block lane 兼容。
6. `template save` / `harness template save`
   - 不再把 `AGENTS.md` 作为基础 authored process member
   - orbit template 不再写根 `AGENTS.md`
   - orbit template 只消费 `meta.agents_template` 或 orchestration output
   - 不提供对 legacy orbit template 根 `AGENTS.md` payload 的 backward compatibility
   - harness template 继续把运行态根文件当容器导出输入，但先剥离 runtime markers
7. future member-generated block 的手工编辑：
   - 默认只影响运行态容器
   - 只有显式执行 `orbit brief backfill` 才能覆盖真相源

### 8.3 测试

1. orchestration 输入选择测试。
2. `meta.agents_template` 优先级测试。
3. AGENTS 生成测试。
4. `orbit brief backfill` 成功回填测试。
5. `orbit brief backfill` 的缺失 block / malformed runtime / reverse replacement 歧义 fail-closed 测试。
6. 与现有 runtime block lane 的兼容测试。
7. `orbit template save` 不再导出根 `AGENTS.md`，并只消费结构化 brief / orchestration output 的回归测试。
8. `harness template save` 的 runtime marker stripping / payload-only 导出回归测试。
9. “运行态手工编辑不会自动改写真相源，且下次 materialize 可被覆盖”测试。

### 8.4 完成标准

- `AGENTS.md` 建立在结构化输入之上。
- runtime 根 `AGENTS.md` 明确只是容器，不再被当成第一性输入。
- Orbit 基础模型不再依赖 `AGENTS.md` 文件本体。
- 运行态手工编辑只有在显式执行 `orbit brief backfill` 后才会改写真相源。
- orbit template lane 不再默认吸收 unmarked prose，也不再导出模板态根 `AGENTS.md`。
- 现有 AGENTS lane 不回归。

---

## 9. 推荐顺序

推荐严格按以下顺序推进：

1. 先冻结文档。
2. 再做 schema 兼容。
3. 再做 ProjectionPlan。
4. 再做本地状态升级。
5. 最后切 AGENTS orchestration。

原因：

- 如果先切 AGENTS，会把基础对象模型继续绑死在文件特判上；
- 如果先切命令而不做兼容 schema，会把现有仓库直接打断；
- 如果不先做 ProjectionPlan，本地状态文件就没有稳定输入源。

---

## 10. 最小交付切片

如果要用最小风险方式推进，建议先做两个切片：

### Slice A

- 文档冻结
- `OrbitMemberRole` / `OrbitMember` / `ProjectionPlan`
- 新旧 schema 兼容解析

### Slice B

- role-aware `ProjectionPlan`
- `file_inventory.json`
- `git_state.json`

这两个切片完成后，再决定是否进入 AGENTS orchestration 切换。
