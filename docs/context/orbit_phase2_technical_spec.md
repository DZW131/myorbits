# Orbit Phase 2 Technical Spec

版本：V0.2
状态：作为 Orbit V0.2 Phase 2 的实现基线
对应 PRD：`docs/context/orbit_phase2_prd.md`
关联边界文档：`docs/context/orbit_storage_boundary.md`

---

## 1. 文档定位

本文件是 Phase 2 的实现规范，不是新的需求文档。

它承接：

- `docs/context/orbit_phase2_prd.md`
- `docs/context/orbit_storage_boundary.md`
- 当前 Orbit MVP 的稳定边界与代码结构

它回答的问题是：

- Phase 2 的文件应该落在哪里；
- 模板保存与模板应用各自的 source of truth 是什么；
- 模板编辑如何做到“不污染当前 worktree”；
- 新命令如何接入现有 command-centric 结构；
- Orbit V0.2 主线应该先做什么，以及哪些内容暂不纳入主线交付。

本文件继续使用 `orbit_phase2_technical_spec.md` 作为正式文件名，避免和 PRD 中的入口分叉。

---

## 2. 稳定前提与直接约束

Phase 2 必须建立在当前 MVP 已稳定的边界之上：

1. `.orbit/` 是版本化配置层，也是 Phase 2 可版本化共享元数据的落点。
2. `.git/orbit/state/` 仍然只承载 repo-local runtime state、snapshot 与 cache。
3. `enter` / `leave` 仍然是唯一正式的 projection 控制入口。
4. 模板 branch 属于 Git DAG，不是 `.git/orbit/state/` 的延伸。
5. Phase 2 继续沿用当前 command-centric 结构，不引入 worktree、daemon、服务端或数据库。

直接约束：

1. 不引入泛化的 `.orbit/state.yaml`。
2. `template apply` 不得把版本化安装记录写进 `.git/orbit/state/`。
3. `template save` 的最终模板编辑必须发生在 temp dir 中，而不是当前 worktree 中。
4. `template save` 和 `template apply` 都不能改变 `enter` / `leave` 的职责边界。
5. `template apply` 默认不自动 `enter`。
6. branch 分类不能依赖 branch name。
7. 模板 branch 不能依赖 `.git/orbit/state/*` 的任何文件。

---

## 3. Phase 2 的真实目标

Phase 2 不是“再加几个 branch 子命令”，而是建立 Orbit 的模板生命周期闭环：

1. 从运行态仓库导出模板态；
2. 把模板态沉淀为本地或远程模板 branch；
3. 在其他仓库把模板再次应用为运行态；
4. 尽可能复用已有 bindings，减少重复输入；
5. 能识别 template / runtime / plain branch。

按交付优先级排序：

1. `orbit template save`
2. `orbit template apply`
3. `orbit bindings init`
4. `orbit branch inspect`
5. `orbit branch list`
6. `orbit branch status`

`orbit repo status` 不纳入 Orbit V0.2 Phase 2 主线交付。它与现有 `orbit status` 易重叠，且不是模板闭环的关键路径。

---

## 4. 关键术语

### 4.1 Runtime Repo

当前用户实际工作的仓库。它包含：

- `.orbit/config.yaml`
- `.orbit/orbits/*.yaml`
- `.orbit/vars.yaml`
- `.orbit/installs/*.yaml`
- 当前工作区内的运行态文件内容

### 4.2 Template Branch

一个保存了某个 orbit 模板快照的 Git branch。  
它至少包含：

- `.orbit/template.yaml`
- `.orbit/orbits/<orbit-id>.yaml`
- 模板化后的 orbit 用户文件

它不依赖 `.git/orbit/state/*` 中的任何文件。

### 4.3 Bindings

运行态变量绑定表，位于 `.orbit/vars.yaml`。  
它保存：

- 变量名
- 当前值
- 可选说明

Bindings 是版本化内容，不是本地 cache。

### 4.4 Install Record

安装记录，位于 `.orbit/installs/<orbit-id>.yaml`。  
它记录：

- 当前 orbit 来自哪个模板源
- 使用了哪个模板 ref / commit
- 何时安装

Install record 是版本化元数据，不是 runtime state。

### 4.5 Branch Kind

Phase 2 统一只区分三类 branch：

- `template`
- `runtime`
- `plain`

### 4.6 Orbit User Scope

Phase 2 中“模板内容”的权威来源是 orbit 的 user scope，而不是当前 sparse-checkout 可见性，也不是 `.git/orbit/state/resolved_scope/*` 的 cache。

Orbit definition 文件是 companion control file，应一并进入模板 branch，但 `.orbit/config.yaml` 不进入模板 branch。

---

## 5. 存储边界与落盘合同

### 5.1 Runtime Repo 中允许的文件

```text
.orbit/
  config.yaml
  vars.yaml
  orbits/
    <orbit-id>.yaml
  installs/
    <orbit-id>.yaml

.git/orbit/state/
  current_orbit.json
  resolved_scope/
  warnings.json
  last_status.json
  orbit.lock
```

边界解释：

- `.orbit/config.yaml`：全局 Orbit 配置，仍属于 control plane。
- `.orbit/orbits/<orbit-id>.yaml`：orbit definition。
- `.orbit/vars.yaml`：版本化 bindings。
- `.orbit/installs/<orbit-id>.yaml`：版本化安装来源记录。
- `.git/orbit/state/*`：本地执行状态、projection cache、warnings、status snapshot、lock。

### 5.2 Template Branch 中允许的文件

```text
.orbit/
  template.yaml
  orbits/
    <orbit-id>.yaml

<orbit user files...>
```

Orbit V0.2 中，模板 branch 不应包含：

- `.git/orbit/state/*`
- `.orbit/config.yaml`
- `.orbit/vars.yaml`
- `.orbit/installs/*.yaml`

原因：

- `.orbit/config.yaml` 是 repo-level control plane，不属于模板本体；
- `.orbit/vars.yaml` 是运行态 concrete bindings，不应连同具体值一起进入模板；
- `.orbit/installs/*.yaml` 是 runtime repo 的安装历史，不属于模板内容。

### 5.3 为什么不使用 `.orbit/state.yaml`

禁止引入 `.orbit/state.yaml` 作为泛化状态文件。

原因：

1. 它会和 `.git/orbit/state/` 的 repo-local runtime state 混淆。
2. 它会把配置、安装记录和本地运行态写进同一层。
3. 这违反了 `docs/context/orbit_storage_boundary.md` 已明确写下的边界。

Phase 2 需要的版本化元数据，应拆分为：

- `.orbit/vars.yaml`
- `.orbit/installs/<orbit-id>.yaml`
- `.orbit/template.yaml`

### 5.4 远程缓存边界

若后续需要为远程模板读取做缓存：

- 应放在 `.git/orbit/` 下的 repo-local cache，或 temp dir；
- 不应写进 `.orbit/`。

`.orbit/` 不是 cache 层。

---

## 6. Schema 合同

### 6.1 `.orbit/vars.yaml`

```yaml
schema_version: 1
variables:
  project_name:
    value: Orbit
    description: 项目名称，用于文档标题与说明文字
  service_url:
    value: http://localhost:3000
    description: 服务访问地址
```

约束：

- key 必须是稳定变量名；
- `value` 必须是字符串；
- `description` 可选；
- Orbit V0.2 不支持嵌套对象、数组或非字符串类型。

### 6.2 `.orbit/template.yaml`

```yaml
schema_version: 1
kind: template
template:
  orbit_id: docs
  default_template: false
  created_from_branch: main
  created_from_commit: abc123
  created_at: 2026-03-21T10:00:00Z
variables:
  project_name:
    description: 项目名称，用于文档标题与说明文字
    required: true
  service_url:
    description: 服务访问地址
    required: true
```

职责：

- 标识该 branch 为模板 branch；
- 给出模板对应的 `orbit_id`；
- 支持远程模板源“默认模板”判定；
- 保存模板声明的变量集合与说明；
- 保存模板来源分支与 commit。

说明：

- `variables` 中只保存变量元信息，不保存运行态 concrete values；
- `required` 的判定规则为：模板内容中存在对该变量的引用；
- `default_template` 仅用于模板源解析，不改变模板内容；
- Orbit V0.2 中允许仓库里存在多个 default template，但解析时只有“恰好一个 default”才自动命中，多于一个则视为歧义。

### 6.3 `.orbit/installs/<orbit-id>.yaml`

```yaml
schema_version: 1
orbit_id: docs
template:
  source_kind: local_branch
  source_repo: ""
  source_ref: orbit-template/docs
  template_commit: abc123
applied_at: 2026-03-21T10:30:00Z
```

字段说明：

- `source_kind`：
  - `local_branch`
  - `remote_git`
- `source_repo`：
  - 本地 branch 安装时可为空；
  - 远程模板安装时保存规范化 Git URL。
- `source_ref`：
  - 本地 branch 名或远程模板 branch/ref。
- `template_commit`：
  - 实际应用的模板 commit。
- `applied_at`：
  - 安装时间。

说明：

- install record 只记录模板安装来源，不记录当前 projection 状态；
- 它是版本化元数据，因此应进入 Git 历史。

---

## 7. 命令与 package 结构

Phase 2 继续使用 command-centric 分层，不把模板逻辑塞进现有 `view` 或 `state` 包。

建议新增：

```text
cmd/orbit/cli/template/
cmd/orbit/cli/bindings/
cmd/orbit/cli/branchinfo/
```

### 7.1 `template`

负责：

- 模板 manifest 读写与校验；
- 模板保存；
- 模板应用；
- 模板渲染与变量扫描；
- 模板源解析；
- 最终模板编辑流程。

建议文件：

- `manifest.go`
- `save.go`
- `apply.go`
- `render.go`
- `scan.go`
- `source.go`
- `edit.go`

### 7.2 `bindings`

负责：

- `.orbit/vars.yaml` 读写；
- bindings merge；
- 缺失变量 skeleton 生成；
- prompt / editor 抽象。

建议文件：

- `vars.go`
- `merge.go`
- `skeleton.go`
- `prompt.go`
- `editor.go`

### 7.3 `branchinfo`

负责：

- branch 分类；
- branch inspect；
- branch list；
- branch status。

建议文件：

- `classify.go`
- `inspect.go`
- `list.go`
- `status.go`

### 7.4 `git` 扩展点

`git` 层仍然只做 Git 适配，不做业务编排。

建议新增 helper：

- `ls-remote`
- fetch temp ref
- read file / tree snapshot at revision
- temp index write
- `commit-tree`
- `update-ref`

明确不做：

- 不引入 Git worktree 作为模板编辑或模板写入机制；
- 不依赖 shell 展开；
- 不以 branch 命名规则替代 manifest 校验。

### 7.5 命令树

建议 Phase 2 增加以下命令：

```text
orbit template save <orbit-id> --to <template-branch>
orbit template apply <template-branch|git-url>
orbit bindings init <template-source> [--out <path>]
orbit branch status
orbit branch inspect <branch>
orbit branch list
```

Orbit V0.2 建议 flags：

- `orbit template save`
  - `--to`
  - `--dry-run`
  - `--edit-template`
  - `--overwrite`
  - `--default`
  - `--json`
- `orbit template apply`
  - `--ref`
  - `--bindings`
  - `--interactive`
  - `--editor`
  - `--overwrite-existing`
  - `--dry-run`
  - `--json`
- `orbit bindings init`
  - `--out`
  - `--json`

新命令应继续支持稳定 stdout/stderr；适合机器消费的输出应支持 `--json`。

这里的 flags 列表表示 V0.2 目标命令面，而不是要求每个命令在首次落地阶段同时完成全部 JSON 输出。

- 各阶段可以先交付稳定的人类可读 stdout/stderr；
- Phase 2C-2 负责统一补齐或硬化跨命令的 `--json` 契约；
- 若某个命令在更早阶段提供了 `--json`，该结构就应被视为正式契约并保持稳定。

### 7.6 与现有代码结构的责任边界

继续复用现有 `orbit` 包来负责：

- `.orbit/config.yaml` 读取
- `.orbit/orbits/*.yaml` 读取
- scope resolution

Phase 2 新增的 `template` / `bindings` / `branchinfo` 包负责各自的业务编排，不把模板渲染逻辑塞回 `orbit` 包。

`state` 包继续只负责 `.git/orbit/state/`，不承担：

- `.orbit/vars.yaml` 读写
- `.orbit/installs/*.yaml` 读写
- `.orbit/template.yaml` 读写
- branch 分类

`commands` 层继续保持薄：

- 参数解析
- stdout / stderr
- `--json`

业务编排应放在 `template` / `bindings` / `branchinfo` 领域包中。

---

## 8. Shared 实现原语

在进入具体命令前，先抽出 6 个共享原语。

### 8.1 Template Content Builder

输入：

- runtime repo root
- orbit id
- resolved user scope
- orbit definition path
- `.orbit/vars.yaml`

输出：

- 模板候选文件树
- 变量替换摘要
- ambiguity 列表

说明：

- 它只负责“从 runtime repo 构建模板候选内容”；
- 不负责打开编辑器；
- 不负责写 branch。

### 8.2 Template Variable Scanner

输入：

- 模板候选文件树

输出：

- 被引用的变量集合
- 未声明变量
- 未使用变量

说明：

- 编辑后必须重新跑一次；
- control files 与用户文件可以分开校验。

### 8.3 Template Source Resolver

输入：

- 本地 branch 名，或
- Git URL，或
- Git URL + ref

输出：

- 模板源类型
- 规范化 source repo
- 具体 template ref
- 可读取的 template commit

### 8.4 Bindings Merge Engine

输入：

- template manifest 中的变量声明
- `--bindings` 文件
- 当前仓库 `.orbit/vars.yaml`
- 交互式补填结果

输出：

- 完整 bindings map
- 仍缺失的变量列表
- 变量来源摘要

说明：

- `description` 优先使用 template manifest 中的变量声明；
- 若 manifest 未提供 `description`，则按绑定值来源优先级回退：
  `--bindings` > `.orbit/vars.yaml` > interactive / editor；
- 该回退规则只影响补填提示、dry-run 摘要和调试输出，不改变值来源本身的优先级判定。

### 8.5 Template Branch Writer

输入：

- 最终模板文件树
- 目标 branch
- commit message / metadata

输出：

- 创建或更新后的模板 commit

说明：

- 基于 temp index + `git commit-tree` + `git update-ref`；
- 不污染当前 worktree；
- 不要求当前用户先切到模板 branch。

### 8.6 Conflict Analyzer

输入：

- 当前 runtime repo 内容
- 目标 orbit id
- 待写入文件树
- install record 候选

输出：

- orbit-id 冲突
- path 冲突
- install-source 冲突
- branch 已存在冲突

---

## 9. 状态机

### 9.1 `template save` 状态机

```text
LOAD_ORBIT_DEFINITION
  -> RESOLVE_USER_SCOPE
  -> LOAD_RUNTIME_BINDINGS
  -> BUILD_TEMPLATE_CANDIDATE
  -> SUBSTITUTE_VALUES_TO_VARS
  -> REVIEW_SUBSTITUTIONS
  -> OPTIONAL_FINAL_TEMPLATE_EDIT
  -> RESCAN_VARIABLES
  -> VALIDATE_TEMPLATE_MANIFEST
  -> WRITE_TEMPLATE_BRANCH
  -> DONE
```

### 9.2 `template apply` 状态机

```text
RESOLVE_TEMPLATE_SOURCE
  -> LOAD_TEMPLATE_MANIFEST
  -> LOAD_TEMPLATE_TREE
  -> RESOLVE_BINDINGS
  -> RENDER_RUNTIME_FILES
  -> CHECK_CONFLICTS
  -> PREVIEW_OR_CONFIRM
  -> WRITE_RUNTIME_FILES
  -> WRITE_INSTALL_RECORD
  -> UPDATE_RUNTIME_BINDINGS_IF_NEEDED
  -> DONE
```

---

## 10. 核心算法

### 10.1 值替换算法（运行态 -> 模板态）

基本规则：

- 对 `.orbit/vars.yaml` 中的每个 `var_name -> literal_value`，在 orbit user scope 的文本文件中查找 `literal_value`；
- 命中后替换成 `$var_name`；
- 只做字面量精确替换，不做模糊替换和语义推理。

文件类型规则：

- 仅 Markdown 文件（`.md`）参与变量替换；
- 其它文本文件原样保留，不扫描 `$var_name`，也不参与渲染；
- 二进制文件原样复制；
- 第一版不让控制文件参与变量替换。

替换顺序：

- 按 `literal_value` 长度倒序替换；
- 先替换更长的值，再替换更短的值。

原因：

- 避免值包含关系导致错误替换，例如：
  - `http://localhost:3000/api`
  - `http://localhost:3000`

歧义处理：

- 若多个变量映射到同一个 literal，不自动猜；
- 进入 ambiguity 集合；
- 默认 fail-closed；
- 后续如需要交互式消歧，再单独扩展。

### 10.2 模板变量扫描算法

匹配规则：

- 模式：`$[A-Za-z_][A-Za-z0-9_]*`
- 仅对 Markdown 文件执行扫描
- 收集所有唯一变量名
- 与 manifest 的变量集合比较

校验结果：

- 模板引用有、manifest 没有：失败或要求补齐
- manifest 有、模板未引用：允许保留

### 10.3 bindings 合并优先级

最终 bindings 生成规则：

1. 先加载 `--bindings <file>`；
2. 再加载当前仓库 `.orbit/vars.yaml`；
3. 对模板所需变量：
   - 若 `--bindings` 中有值，用它；
   - 否则若 `.orbit/vars.yaml` 中有同名值，用它；
   - 否则加入缺失集合；
4. 对缺失集合：
   - 若 `--interactive`，逐个询问；
   - 若 `--editor`，生成 skeleton 后打开编辑器；
   - 否则报缺失错误。
5. 对每个变量的 `description`：
   - 优先取 template manifest 中的声明；
   - 若 manifest 未提供，则按 `--bindings` > `.orbit/vars.yaml` > interactive / editor 的顺序回退到第一个可用说明；
   - 该步骤只决定说明文字展示，不改变变量值的最终来源。

设计理由：

- 显式输入优先于仓库默认；
- 仓库默认优先于交互；
- 尽量减少重复填写。

### 10.4 模板渲染算法（模板态 -> 运行态）

对于每个模板文件：

1. 若文件为 Markdown（`.md`），扫描 `$var_name`
2. Markdown 文件用最终 bindings 中的值替换
3. 非 Markdown 文件原样保留，不参与模板变量渲染
4. 若 Markdown 文件有变量缺失，拒绝渲染并报错

首版不支持：

- 条件块
- 循环
- 函数式模板

首版只做纯文本变量替换。

### 10.5 远程模板源解析算法

输入：

- Git URL
- 可选 `--ref`

输出：

- 一个确定的模板 branch，或
- 一个候选列表供用户选择，或
- 明确错误

算法：

1. `git ls-remote --heads` 枚举候选 heads；
2. 对每个 candidate 检查是否存在合法 `.orbit/template.yaml`；
3. 收集所有合法 template branches；
4. 按以下规则决策：
   - 显式 `--ref` 优先；
   - 唯一合法 template branch 直接命中；
   - 多个合法 template branches 中恰好一个 `default_template: true`，则自动命中；
   - 多个合法 template branches 且没有唯一 default，则返回候选列表；
   - 没有合法 template branch，则明确报错。

说明：

- 远程模板源解析只使用 Git，不依赖 GitHub API；
- branch name 可以是入口，但不是模板身份的权威依据。

### 10.6 冲突判定算法

至少识别以下 4 类冲突：

1. Orbit 已存在冲突
   - `.orbit/orbits/<orbit-id>.yaml` 已存在
   - 同 orbit id 已注册
2. 路径冲突
   - 目标写入路径已存在且内容不同
3. 安装来源冲突
   - `.orbit/installs/<orbit-id>.yaml` 已存在且来源不同
4. 模板 branch 已存在冲突
   - `template save` 的目标 branch 已存在

默认策略：

- 不静默覆盖；
- 非交互模式无显式覆盖参数时失败；
- 交互模式下由用户显式确认。

---

## 11. `orbit template save`

### 11.1 输入与 flags

最小输入：

- `<orbit-id>`
- `--to <template-branch>`

建议 flags：

- `--dry-run`
- `--edit-template`
- `--overwrite`
- `--default`
- `--json`

### 11.2 内容来源规则

`template save` 的输入内容来自当前 runtime repo，而不是当前是否已 `enter` 到该 orbit。

权威输入应为：

1. 当前仓库 `.orbit/config.yaml`
2. 当前仓库 `.orbit/orbits/<orbit-id>.yaml`
3. 基于 repo tracked files 重新解析出的该 orbit user scope
4. scope 内文件的 runtime 内容：
   - 当前在 worktree 可见时，读取 worktree；
   - 当前因 sparse view 不可见但仍属于 tracked scope 时，回退到 index / `HEAD` snapshot；
   - 不能因为当前不可见就把该路径从模板候选中漏掉
5. 当前仓库 `.orbit/vars.yaml`

说明：

- MVP 已禁止把 dirty tracked paths 静默隐藏，因此这里的回退只面向当前被 sparse 隐藏的 clean tracked paths；
- `template save` 的内容来源规则必须独立于当前是否已 `enter` 到目标 orbit。

不应依赖：

- `.git/orbit/state/resolved_scope/*`
- 当前 sparse-checkout 可见性
- 当前是否存在 `current_orbit.json`

### 11.3 保存范围

Orbit V0.2 中，`template save` 写入模板 branch 的内容应是：

1. 模板化后的 orbit 用户文件；
2. `.orbit/orbits/<orbit-id>.yaml`；
3. 生成后的 `.orbit/template.yaml`。

它不应把以下文件带进模板 branch：

- `.orbit/config.yaml`
- `.orbit/vars.yaml`
- `.orbit/installs/*.yaml`
- `.git/orbit/state/*`

### 11.4 最终模板编辑

`--edit-template` 的实现要求：

1. 在 temp dir 中 materialize 模板候选文件树；
2. 打开编辑器只编辑 temp dir 中的候选文件；
3. 编辑结束后重新扫描变量引用与文件集合；
4. 再生成最终 `.orbit/template.yaml`；
5. 最终写入模板 branch。

Orbit V0.2 中：

- manifest 由系统生成，不要求用户手工编辑；
- 用户编辑的是模板文件树本身；
- 模板编辑结果不能回写当前 runtime repo。

### 11.5 `--default` 语义

`--default` 的职责是：

- 在目标 template branch 的 manifest 中写入 `default_template: true`

Orbit V0.2 中：

- 不要求 `template save --default` 自动回写并清理其他 template branches 的 default 标记；
- 若一个源仓库存在多个 default templates，远程解析时视为歧义，不自动选。

### 11.6 写入算法

`template save` 不使用 worktree。推荐算法：

1. 组装最终模板文件树；
2. 使用 temp index 计算 tree object；
3. 通过 `git commit-tree` 生成模板 commit；
4. 通过 `git update-ref` 创建或更新目标 branch。

优势：

- 不污染当前工作区；
- 不要求切换 branch；
- 更容易在测试中隔离。

### 11.7 `--dry-run`

`--dry-run` 不写 branch，只输出：

- 解析到的模板文件列表；
- 变量替换摘要；
- ambiguity；
- 将写入的 manifest 摘要。

### 11.8 失败策略

以下场景 fail-closed：

- orbit id 无效或不存在；
- `.orbit/orbits/<orbit-id>.yaml` 非法；
- `.orbit/vars.yaml` 非法；
- 替换后出现 ambiguity 且未明确处理；
- 编辑后出现未声明变量；
- 目标 branch 已存在且未显式 `--overwrite`；
- 目标 branch 写入失败。

---

## 12. `orbit template apply`

### 12.1 输入与 flags

最小输入：

- `<template-branch>`，或
- `<git-url>`

建议 flags：

- `--ref <branch>`
- `--bindings <path>`
- `--interactive`
- `--editor`
- `--overwrite-existing`
- `--dry-run`
- `--json`

### 12.2 模板源类型

Phase 2 按 source kind 只区分两类：

- `local_branch`
- `remote_git`

本地 branch 与远程 Git URL 的后续处理可以共享同一套 manifest / tree 读取逻辑。

### 12.3 应用结果

`template apply` 成功后应写入：

1. 渲染后的 runtime 用户文件；
2. `.orbit/orbits/<orbit-id>.yaml`；
3. `.orbit/installs/<orbit-id>.yaml`；
4. 若需要复用或补齐变量，则更新 `.orbit/vars.yaml`。

`template apply` 不应写入：

- `.git/orbit/state/*`
- `current_orbit.json`
- projection cache

### 12.4 与 `enter` 的边界

`template apply` 只负责安装，不负责 projection。

也就是说：

- apply 成功后，仓库已具备运行该 orbit 的配置与文件；
- 当前工作区是否切换到该 orbit，仍由 `orbit enter <orbit-id>` 控制。

### 12.5 冲突处理

至少检查以下冲突：

1. 当前仓库已存在同名 orbit id；
2. 将写入的目标路径存在未提交改动；
3. 已存在 `.orbit/installs/<orbit-id>.yaml` 且来源不同；
4. 将写入的 orbit definition 路径已存在且内容不同。

默认行为：

- 不静默覆盖；
- 返回明确冲突摘要；
- 由用户显式决定是否继续；
- 非交互模式无 `--overwrite-existing` 时失败。

### 12.6 `--dry-run`

`--dry-run` 不写工作区，只输出：

- source summary
- manifest summary
- resolved bindings 来源摘要
- 待写入文件列表
- 冲突摘要

### 12.7 失败策略

以下场景 fail-closed：

- 模板源没有合法 `.orbit/template.yaml`
- 模板缺少合法 `.orbit/orbits/<orbit-id>.yaml`
- manifest 非法
- 缺失变量
- 目标路径冲突且未确认覆盖
- 写 install record 失败

---

## 13. `orbit bindings init`

`bindings init` 的职责不是应用模板，而是生成可填写的 bindings skeleton。

最小行为：

1. 读取模板 manifest；
2. 取出变量声明与 description；
3. 生成 `.yaml` skeleton；
4. 支持 `--out <path>`。

建议输出形态：

```yaml
schema_version: 1
variables:
  project_name:
    value: ""
    description: 项目名称，用于文档标题与说明文字
```

Orbit V0.2 中：

- `bindings init` 不修改当前仓库；
- 它只是为了给 `template apply --bindings` 提供输入模板。

---

## 14. Branch 分类与 inspect

### 14.1 分类规则

`template`：

- 目标 branch/revision 中存在合法 `.orbit/template.yaml`

`runtime`：

- 不是 template branch；
- 但 branch/revision 中存在合法 `.orbit/config.yaml`；
- 且至少存在一个合法 `.orbit/orbits/*.yaml`

`plain`：

- 不满足以上任一条件

说明：

- runtime branch 的判定不依赖 `.git/orbit/state/*`；
- 也不要求 `.orbit/installs/*.yaml` 必须存在；
- 否则手工创建 orbit 的运行态仓库会被误判成 plain。

### 14.2 `branch inspect`

应至少输出：

- branch kind
- orbit id
- template manifest 摘要
- install record 摘要
- default template 标记

### 14.3 `branch list`

第一轮建议只列出本地 branches，并给出分类结果。

远程 branches 的枚举后置到 Phase 2C。

### 14.4 `branch status`

第一轮建议只回答“当前 branch 是 template / runtime / plain 哪一种”，以及为什么。

---

## 15. Git 与写入策略

Phase 2 继续遵守当前 Orbit 的 Git 适配原则：

1. 优先使用系统 `git`；
2. 优先使用显式参数列表；
3. 不使用 `sh -c` 执行普通 Git 操作；
4. 能用 temp index / temp ref 的地方，不用 worktree。

命令级策略：

- `template save`
  - 读取当前 repo 文件内容
  - 使用 temp index + `commit-tree` + `update-ref`
- `template apply`
  - 读取本地或远程 template tree
  - 渲染后写回 runtime repo
- `branchinfo`
  - 主要依赖 `rev-parse`、`show`、`cat-file`、`ls-remote`

远程模板读取建议采用：

- temp ref fetch
- 读取完成后清理 temp ref

不能让临时 fetch 结果成为运行正确性的长期前提。

---

## 16. 非功能要求

### 16.1 可审计

每次 `template save`、`template apply` 都应输出清晰摘要：

- 来源
- 目标
- 冲突数
- 自动填充变量数
- 需要补填变量数

### 16.2 幂等性

同一模板源 + 同一 bindings，多次 apply 结果应一致。

### 16.3 安全性

默认不覆盖已有 orbit 和已有文件。

### 16.4 可恢复性

失败时不应留下半写入状态；至少要做到：

- 先渲染到内存或 temp dir
- 最后统一写入工作区
- branch 写入失败不污染当前 runtime repo

---

## 17. 测试要求

### 17.1 Unit Tests

需要覆盖：

- `.orbit/vars.yaml` schema 与 validation
- `.orbit/template.yaml` schema 与 validation
- `.orbit/installs/*.yaml` schema 与 validation
- replacement engine
- 长度倒序替换
- ambiguity detection
- render logic
- bindings merge precedence
- branch classification

### 17.2 Temp Repo Integration Tests

使用 isolated temp repo，覆盖：

- `template save`
- `template save --dry-run`
- `template save --edit-template`
- `template save --default`
- `template apply <local-branch>`
- install record 写入
- 冲突处理
- apply 不自动 `enter`

### 17.3 Remote Integration Tests

使用本地 bare repo 模拟远程模板源，覆盖：

- `ls-remote` + template source selection
- temp ref fetch
- `template apply <git-url>`

### 17.4 必须钉死的不变量

以下行为必须有明确测试：

1. `template apply` 不自动 `enter`
2. `.git/orbit/state/` 不承载模板安装元数据
3. 模板编辑不污染当前 worktree
4. branch 分类不依赖 branch name
5. `.orbit/config.yaml` 不进入模板 branch
6. `.orbit/vars.yaml` 不连同 concrete values 一起进入模板 branch
7. 同 literal 多变量时不会静默替换
8. 目标 branch 已存在且无 `--overwrite` 时失败

---

## 18. 当前冻结的结论

当前阶段先冻结以下结论，作为 Orbit V0.2 的实现前提：

1. `.orbit/` / `.git/orbit/state/` / Git DAG 的边界不变。
2. 不引入 `.orbit/state.yaml`。
3. install record 固定写在 `.orbit/installs/<orbit-id>.yaml`。
4. 模板 branch 以 `.orbit/template.yaml` 为唯一权威标识。
5. `template save` 的最终编辑必须在 temp dir 中完成。
6. `template save` 的 branch 写入走 temp index + `commit-tree` + `update-ref`。
7. `template apply` 默认不自动 `enter`。
8. branch 分类靠 manifest 与 runtime config，不靠 branch name。
9. Orbit V0.2 只做字符串变量与文本文件替换。
10. 覆盖行为默认都要显式参数或显式确认。

---

## 19. 明确暂缓的内容

以下内容不纳入 Orbit V0.2 Phase 2 主线交付：

- 模板市场 / registry
- 权限与审批流
- 复杂变量类型系统
- 智能变量抽取
- `AGENTS.md` 的复杂 merge/append 方案
  这部分若在 V0.2 收尾阶段落地，应遵守独立方案文档，并继续保持当前模板 branch contract 不变。
- 泛化的 `repo status` 子系统
- 让控制文件本身参与模板变量替换
- 多 orbit 组合安装事务
- 模板签名与权限控制
