# Orbit Phase 2 开发计划

版本：V0.2
状态：作为 Orbit V0.2 Phase 2 的执行基线
对应 PRD：`docs/context/orbit_phase2_prd.md`
对应技术方案：`docs/context/orbit_phase2_technical_spec.md`
关联边界文档：`docs/context/orbit_storage_boundary.md`

---

## 1. 文档目标

本文档基于以下 source of truth 制定：

1. `docs/context/mvp-product-requirements.md`
2. `docs/context/mvp-technical-architecture.md`
3. `docs/testing-strategy.md`
4. `docs/context/orbit_phase2_prd.md`
5. `docs/context/orbit_phase2_technical_spec.md`
6. `docs/context/orbit_storage_boundary.md`
7. `CONTRIBUTING.md`
8. `AGENTS.md`

目标是为 Orbit Phase 2 提供一份可执行的开发推进计划，明确：

- 分阶段目标与交付物
- 每阶段的开发内容与技术依赖
- 包结构与责任边界
- 测试、验证与收口要求
- issue 拆分顺序与阶段完成标准

本文档只覆盖 Phase 2 主线，不扩展到模板市场、权限系统、复杂模板语言、远程服务或 GUI。

---

## 2. 总体开发策略

### 2.1 开发原则

整个实现必须持续满足以下原则：

- Git DAG 仍是唯一历史真相源。
- `.orbit/` 只放版本化配置与版本化元数据。
- `.git/orbit/state/` 只放 repo-local runtime state、snapshot 与 cache。
- `template save` 和 `template apply` 都不能改变 `enter` / `leave` 的职责边界。
- 模板编辑只能作用于 temp dir 中的模板候选内容，不得回写当前 runtime worktree。
- 模板 branch 的身份只能由 `.orbit/template.yaml` 判定，不能靠 branch name。
- 远程模板解析只使用 Git，不依赖 GitHub API。
- 默认不静默覆盖已有 orbit、已有路径和已有模板 branch。
- `AGENTS.md` 先不进入 Phase 2 主线闭环，只在设计上预留扩展位；若在 V0.2 收尾阶段落地，按独立方案推进。

### 2.2 开发排序原则

本阶段推荐严格按下面顺序开发：

1. 先冻结底层合同，再写命令。
2. 先做本地闭环，再做远程。
3. 先做 save/apply 主链路，再做体验增强。
4. 先做普通文件模板，再做 `AGENTS.md` 特殊文件。
5. 所有 branch 判定都基于文件合同，不基于 branch name。
6. 所有模板写 branch 的逻辑都绕开当前 worktree。

这样排序的原因：

- `template save` / `template apply` 依赖同一套 schema、source resolver、bindings merge、conflict analyzer。
- 远程模板应用只是“模板源解析 + 本地 apply”的上层包装，不应该先做。
- `AGENTS.md` 会引入 shared file / append merge / fragment 等特殊语义，过早实现会把主链路复杂化。
- `template save` 的 temp dir 编辑和 temp index 写入是 Phase 2 的实现基石，应先稳定。

### 2.3 命令批次视图

从命令交付视角，建议按下面四批推进：

1. 第一批：`orbit template save`
2. 第二批：`orbit template apply <local-branch>`
3. 第三批：`orbit branch status / inspect / list`
4. 第四批：`orbit template apply <git-url>`、`orbit bindings init`、`--editor` / `--json` / dry-run 强化

这与工程阶段是一致的：

- Phase 2A-0 打稳共享原语
- Phase 2A-1 打通 runtime -> template branch
- Phase 2B-1 打通 local template branch -> runtime repo
- Phase 2B-2 做 branchinfo 可观察性
- Phase 2C-1 打通 remote source
- Phase 2C-2 做体验增强与收尾

关于 `--json` 的阶段口径：

- V0.2 的命令面以支持 `--json` 为目标口径。
- 但阶段推进允许先交付稳定的人类可读 stdout/stderr，再在 Phase 2C-2 统一补齐或硬化 JSON 契约。
- 若某阶段某命令的 `--json` 自然形成且成本低，可以提前交付；一旦交付，就应视为正式契约并保持稳定。

### 2.4 分支与 PR 策略

遵循 `CONTRIBUTING.md`，从 `main` 拉分支开发，通过 PR 合并回 `main`。

建议：

- 每个阶段一个主分支，一个主 PR。
- 阶段内部如果改动过大，可再拆 2 到 4 个小 PR，但不要跨阶段混合。
- 功能实现使用 `feature/phase2-<topic>`。
- 文档与硬化使用 `docs/phase2-<topic>` 或 `chore/phase2-<topic>`。

建议分支名：

- `feature/phase2-foundations`
- `feature/phase2-template-save`
- `feature/phase2-local-apply`
- `feature/phase2-branchinfo`
- `feature/phase2-remote-source`
- `feature/phase2-hardening`

---

## 3. 技术依赖与责任边界

### 3.1 已有基础能力

Phase 2 默认依赖当前 MVP 已经稳定的能力：

- `orbit` 包中的配置加载与 scope resolution
- `git` 包中的 repo 发现、tree/file 读取、pathspec 与 ref 操作
- `state` 包中的 lock、snapshot 与 repo-local runtime state
- `ids` 包中的 `orbit-id` 和路径规范化

Phase 2 不应重做这些能力，只应在现有基础上扩展：

- bindings 读写
- template manifest
- template source resolution
- render / scan / conflict analysis
- branch classification

### 3.2 新增包

沿用当前 command-centric 结构，并新增 3 个领域包：

- `cmd/orbit/cli/template`
  - manifest
  - save
  - apply
  - render
  - scan
  - source
  - edit
- `cmd/orbit/cli/bindings`
  - vars
  - merge
  - skeleton
  - prompt
  - editor
- `cmd/orbit/cli/branchinfo`
  - classify
  - inspect
  - list
  - status

### 3.3 责任边界

责任边界要求：

- `commands` 只负责 flags、stdout/stderr、`--json`
- `orbit` 继续负责 `.orbit/config.yaml` 与 `.orbit/orbits/*.yaml`
- `state` 不负责 `.orbit/vars.yaml` / `.orbit/installs/*.yaml` / `.orbit/template.yaml`
- `git` 只做 Git 适配，不写业务策略
- `template` / `bindings` / `branchinfo` 负责各自的领域编排

### 3.4 共享原语依赖顺序

建议先抽出以下共享原语，再上命令：

1. `bindings` schema 与读写
2. `template` manifest schema 与读写
3. branch classifier
4. replacement engine
5. variable scanner
6. bindings merge engine
7. template content builder
8. template branch writer
9. template source resolver
10. render engine
11. conflict analyzer

依赖顺序建议如下：

```text
bindings schema
  -> replacement / render / merge
template manifest
  -> source resolver / classifier / bindings init
scanner + replacement + content builder
  -> template save
source resolver + render + conflict analyzer
  -> template apply
manifest + install record + classifier
  -> branch inspect/list/status
```

---

## 4. 阶段划分总览

建议分为 7 个阶段：

- `Phase 2A-0`：底层合同冻结与共享原语落地
- `Phase 2A-1`：`orbit template save` 最小闭环
- `Phase 2B-1`：`orbit template apply` 本地闭环
- `Phase 2B-2`：branch info 能力
- `Phase 2C-1`：远程模板源解析与 apply
- `Phase 2C-2`：体验增强与收尾
- `V0.2 收尾扩展`：`AGENTS.md` 特殊机制

下面按阶段展开。

---

## 5. Phase 2A-0：底层合同冻结与共享原语落地

这一阶段不追求用户可用命令闭环，而是把后面所有实现依赖的“地基”先定住。

### 5.1 目标

冻结 schema、branch 分类、变量处理与模板候选构建逻辑，使后续 save/apply/branchinfo 都建立在统一合同上。

### 5.2 建议分支名

- `feature/phase2-foundations`

### 5.3 开发内容

#### 任务 1：冻结文件合同与 schema

实现并冻结以下 schema：

1. `.orbit/vars.yaml`
2. `.orbit/template.yaml`
3. `.orbit/installs/<orbit-id>.yaml`

最小要求：

- `.orbit/vars.yaml`
  - `schema_version`
  - `variables.<var>.value`
  - `variables.<var>.description?`
- `.orbit/template.yaml`
  - `schema_version`
  - `kind: template`
  - `template.orbit_id`
  - `template.default_template`
  - `template.created_from_branch`
  - `template.created_from_commit`
  - `template.created_at`
  - `variables.<var>.description?`
  - `variables.<var>.required`
- `.orbit/installs/<orbit-id>.yaml`
  - `schema_version`
  - `orbit_id`
  - `template.source_kind`
  - `template.source_repo`
  - `template.source_ref`
  - `template.template_commit`
  - `applied_at`

产出物：

- schema struct / model
- validation 逻辑
- YAML 编解码器
- 错误信息规范

#### 任务 2：实现 branch 分类内核

实现 `branchinfo/classify.go` 的底层能力，支持对任意 revision 判断：

- `template`
- `runtime`
- `plain`

分类规则：

- `template`
  - 存在合法 `.orbit/template.yaml`
- `runtime`
  - 不是 template
  - 存在合法 `.orbit/config.yaml`
  - 至少存在一个合法 `.orbit/orbits/*.yaml`
- `plain`
  - 以上都不满足

产出物：

- `ClassifyRevision(repo, rev)` 一类接口
- 统一的分类结果结构体
- 原因解释字段

#### 任务 3：实现 Bindings Merge Engine

实现 `bindings/merge.go`，处理以下输入：

1. template manifest 中的变量声明
2. `--bindings <file>`
3. 当前仓库 `.orbit/vars.yaml`
4. 交互式 / editor 填写结果

合并优先级：

1. `--bindings`
2. `.orbit/vars.yaml`
3. 交互式补填 / editor 补填

产出物：

- bindings merge 结果结构
- unresolved 变量集合
- 每个变量来源摘要

#### 任务 4：实现 Template Variable Scanner

实现对模板文件树的变量扫描：

- 匹配 `$[A-Za-z_][A-Za-z0-9_]*`
- 汇总唯一变量集合
- 对比 manifest 中声明的变量集合

输出：

- 引用到的变量
- 未声明变量
- 未使用变量

#### 任务 5：实现 Template Content Builder

输入：

- runtime repo root
- orbit id
- resolved user scope
- `.orbit/orbits/<orbit-id>.yaml`
- `.orbit/vars.yaml`

输出：

- 模板候选文件树
- 替换摘要
- ambiguity 集合

内容规则：

- 只处理 orbit user scope 下的用户文件
- `.orbit/orbits/<orbit-id>.yaml` 作为 companion control file 一并进入模板
- 不让 `.orbit/config.yaml`、`.orbit/vars.yaml`、`.orbit/installs/*` 进入模板
- 不依赖 `.git/orbit/state/resolved_scope/*`
- scope 内文件的读取规则必须与 technical spec 一致：可见路径读 worktree，当前被 sparse 隐藏的 clean tracked path 回退到 index / `HEAD` snapshot，不能因不可见而漏掉模板文件

#### 任务 6：实现值替换与 ambiguity detection

实现运行态到模板态的值替换规则：

- 字面量精确替换
- 按 literal 长度倒序替换
- 同 literal 对应多个变量时 fail-closed

### 5.4 交付物

- schema / validation / YAML codecs
- branch classifier
- bindings merge engine
- variable scanner
- template content builder
- replacement / ambiguity engine

### 5.5 验收点

- 非法字段、缺少必要字段能稳定报错
- description 可选但能正常序列化 / 反序列化
- install record 只记录模板来源，不记录 projection 状态
- 当前 branch 和未 checkout revision 判定结果一致
- branch name 不影响分类结果
- scope 外文件不会进入模板
- `.orbit/config.yaml` / `.orbit/vars.yaml` 不会进入模板
- 二进制文件原样复制
- 文本文件才执行变量替换
- 长 literal 先替换
- 相同值映射多个变量时不静默替换

### 5.6 完成定义

满足以下条件即可进入 Phase 2A-1：

- schema 已冻结并通过单测
- branch 分类已稳定
- 变量替换与 scanner 已稳定
- template content builder 可稳定产出候选模板树

---

## 6. Phase 2A-1：`orbit template save` 最小闭环

这一阶段的目标是把 `runtime -> template branch` 跑通。

### 6.1 建议分支名

- `feature/phase2-template-save`

### 6.2 开发内容

#### 任务 1：实现 Template Branch Writer

实现不依赖 worktree 的模板 branch 写入：

1. 组装最终模板文件树
2. 使用 temp index 生成 tree
3. 使用 `git commit-tree`
4. 使用 `git update-ref` 写 branch

#### 任务 2：实现 `--dry-run` 模板导出预览

`orbit template save --dry-run` 输出：

- 模板文件列表
- 替换摘要
- ambiguity 摘要
- 将生成的 `.orbit/template.yaml` 摘要

#### 任务 3：实现 `orbit template save`

完整状态机：

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

支持 flags：

- `--to`
- `--dry-run`
- `--edit-template`
- `--overwrite`
- `--default`
- `--json`

#### 任务 4：实现最终模板编辑

`--edit-template` 行为：

1. 在 temp dir materialize 模板候选文件树
2. 打开编辑器编辑 temp dir 内容
3. 编辑完成后重新扫描变量
4. 重建 `.orbit/template.yaml`
5. 再写 branch

#### 任务 5：`template save` 集成测试

测试覆盖：

- 普通保存
- `--dry-run`
- `--edit-template`
- `--default`
- branch 已存在但无 `--overwrite`
- ambiguity 场景 fail
- `.orbit/config.yaml` 不进入模板
- `.orbit/vars.yaml` 不进入模板

### 6.3 交付物

- template branch writer
- save dry-run
- `orbit template save`
- final template edit
- save integration tests

### 6.4 验收点

- 不切换当前 branch
- 不污染当前工作区
- 能把某个 orbit 保存成 template branch
- template branch 中只有：
  - `.orbit/template.yaml`
  - `.orbit/orbits/<orbit-id>.yaml`
  - 模板化用户文件
- `--default` 仅写 manifest，不负责回收其他 default 标记
- 编辑不会回写 runtime worktree
- 编辑后新增变量能被重新识别
- 编辑后未声明变量时失败

### 6.5 完成定义

满足以下条件即可进入 Phase 2B-1：

- `template save` 可生成合法 template branch
- `--edit-template` 不污染当前 worktree
- `--dry-run` 输出稳定摘要

---

## 7. Phase 2B-1：`orbit template apply` 本地闭环

这一阶段的目标是把 `local template branch -> runtime repo` 跑通。

### 7.1 建议分支名

- `feature/phase2-local-apply`

### 7.2 开发内容

#### 任务 1：实现 Template Source Resolver（本地部分）

输入本地 branch，输出：

- source kind = `local_branch`
- template ref
- template commit
- manifest 内容
- template tree 读取接口

#### 任务 2：实现模板渲染

对模板文件树执行：

- 变量扫描
- 使用 resolved bindings 渲染
- 缺失变量时报错

#### 任务 3：实现 Conflict Analyzer

最少识别 4 类冲突：

1. orbit 已存在冲突
2. 路径冲突
3. 安装来源冲突
4. 模板来源相关冲突摘要

#### 任务 4：实现 `orbit template apply <local-branch>`

完整状态机：

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

支持 flags：

- `--bindings`
- `--interactive`
- `--overwrite-existing`
- `--dry-run`
- `--json`

#### 任务 5：实现缺失变量交互式补填

在 apply 中支持：

- `--interactive`
- 基于 description 的逐项提示

#### 任务 6：本地 apply 集成测试

测试覆盖：

- 本地模板 apply
- apply 复用 `.orbit/vars.yaml`
- 缺失变量交互补填
- install record 写入
- apply 不自动 enter
- 非交互模式缺失变量失败
- 冲突默认失败
- `--overwrite-existing` 成功

### 7.3 交付物

- local template source resolver
- render engine
- conflict analyzer
- `orbit template apply <local-branch>`
- interactive bindings prompt
- local apply integration tests

### 7.4 验收点

- 非 template branch 会失败
- 能读取 `.orbit/template.yaml`
- 能读取 `.orbit/orbits/<orbit-id>.yaml`
- 同一模板 + 同一 bindings 结果稳定一致
- 缺失变量直接失败
- apply 后不自动 `enter`
- install record 写入正确
- 必要时 `.orbit/vars.yaml` 更新正确
- overwrite 行为必须显式

### 7.5 完成定义

满足以下条件即可进入 Phase 2B-2：

- 能从本地 template branch apply
- apply 会写 orbit definition / install record
- apply 默认不自动 enter
- bindings 复用与缺失交互已可用

---

## 8. Phase 2B-2：branch info 能力

这一阶段做可观察性和诊断能力。

### 8.1 建议分支名

- `feature/phase2-branchinfo`

### 8.2 开发内容

#### 任务 1：实现 `orbit branch status`

输出当前 branch 是：

- `template`
- `runtime`
- `plain`

并说明原因。

#### 任务 2：实现 `orbit branch inspect <branch>`

输出至少包括：

- branch kind
- orbit id
- template manifest 摘要
- install record 摘要
- default template 标记

#### 任务 3：实现 `orbit branch list`

第一轮只列本地 branches，并给出分类结果。

第一轮输出要求：

- 先确保稳定的人类可读 stdout/stderr。
- 若 `--json` 在这一阶段低成本自然落地，可以一并实现，但它不是进入 Phase 2C-1 的唯一硬门槛。

### 8.3 交付物

- `orbit branch status`
- `orbit branch inspect`
- `orbit branch list`

### 8.4 验收点

- 解释信息清晰
- 不依赖 branch name
- template branch 能读出 manifest 摘要
- runtime branch 能读出 runtime repo 特征
- plain branch 也能返回分类结果而非直接报错

### 8.5 完成定义

满足以下条件即可进入 Phase 2C-1：

- branch inspect/status/list 可用
- template / runtime / plain 分类稳定

---

## 9. Phase 2C-1：远程模板源解析与 apply

这一阶段才引入远程 Git 仓库。

### 9.1 建议分支名

- `feature/phase2-remote-source`

### 9.2 开发内容

#### 任务 1：实现远程模板源枚举

仅使用 Git，不依赖 GitHub API：

1. `git ls-remote --heads`
2. 枚举 heads
3. 检查每个 candidate 是否存在合法 `.orbit/template.yaml`
4. 收集所有合法 template branches

#### 任务 2：实现默认模板解析规则

规则：

1. 显式 `--ref` 优先
2. 唯一合法 template branch 自动命中
3. 多个合法 template branches 中恰好一个 `default_template: true` 自动命中
4. 多个合法 templates 且无唯一 default，返回候选列表
5. 无合法 templates，明确报错

#### 任务 3：实现 temp ref fetch 与远程文件读取

- fetch 到临时 ref
- 按 revision 读取 `.orbit/template.yaml` 与模板树
- 读完后清理 temp ref

#### 任务 4：实现 `orbit template apply <git-url>`

把远程模板源解析结果接入本地 apply 后半段。

支持 flags：

- `--ref`
- `--bindings`
- `--overwrite-existing`
- `--interactive`
- `--dry-run`
- `--json`

#### 任务 5：远程集成测试

使用本地 bare repo 模拟 remote，覆盖：

- `ls-remote`
- template source selection
- temp ref fetch
- remote apply
- 无模板 / 多模板 / default 模板场景

### 9.3 交付物

- remote heads discover
- default template resolution
- temp ref fetch
- remote template apply
- remote integration tests

### 9.4 验收点

- 非 template branch 不会误入候选
- branch name 不是唯一依据
- 唯一模板分支场景自动命中
- 恰好一个 default 自动命中
- 多 default 视为歧义
- 无模板时报错
- 不污染正常 refs
- 临时 ref 可清理
- 远程读取失败时不会污染本地 repo 状态
- install record 中 `source_kind = remote_git`

### 9.5 完成定义

满足以下条件即可进入 Phase 2C-2：

- 能从远程 Git URL 自动发现模板源并 apply
- remote source selection 已稳定
- temp ref fetch / cleanup 已稳定

---

## 10. Phase 2C-2：体验增强与收尾

这部分不改变主链路正确性，但会影响可用性、联调效率和自动化质量。

### 10.1 建议分支名

- `feature/phase2-hardening`

### 10.2 开发内容

#### 任务 1：实现 `orbit bindings init`

行为：

- 读取模板 manifest
- 生成 skeleton
- 支持 `--out`
- 支持 `--json`

#### 任务 2：实现 `--editor` 填写模式

对于 apply 缺失变量：

- 生成 skeleton 到 temp file
- 打开编辑器
- 读取结果
- 合并回 bindings

#### 任务 3：统一并补全 `--json` 输出

目标：

- 把前面阶段已经自然提供的 JSON 结构收口为稳定契约；
- 对尚未提供 `--json` 的命令在这一阶段补齐；
- 统一 stdout/stderr 与 `--json` 的职责边界。

命令覆盖：

- `template save`
- `template apply`
- `bindings init`
- `branch status`
- `branch inspect`
- `branch list`

#### 任务 4：增强 dry-run 摘要

save dry-run 至少输出：

- 文件列表
- 替换摘要
- ambiguity
- manifest 摘要

apply dry-run 至少输出：

- source summary
- manifest summary
- bindings 来源摘要
- 将写入的文件列表
- 冲突摘要

#### 任务 5：文档与命令帮助同步

更新内容：

- CLI help
- examples
- error message 文案
- README / docs 入口链接

### 10.3 交付物

- `bindings init`
- editor mode
- stable `--json`
- dry-run polish
- docs / help / examples 同步

### 10.4 验收点

- `bindings init` 不修改当前仓库
- description 正常带出
- editor 模式仅补缺失变量
- 非法 YAML 返回清晰错误
- stdout / stderr 职责清晰
- `--json` 输出稳定
- dry-run 足够支撑联调和测试

### 10.5 完成定义

满足以下条件即可宣告 Phase 2 主线完成：

- 能从远程 Git 仓库自动发现模板源并 apply
- `bindings init` / `editor` / `json` / `dry-run` 完整可用
- 关键不变量全部被测试覆盖

---

## 11. V0.2 收尾扩展：`AGENTS.md` 特殊机制

这一部分不进入 Phase 2 主线闭环，但可以作为 V0.2 的最后一段扩展开发；前提是主链路先稳定，并保持模板 branch contract 不被破坏。

### 11.1 为什么后置

因为 `AGENTS.md` 不是普通文件，它会引入：

- shared file
- append merge
- fragment
- marker block
- orbit 级 section 更新

这会改变“模板文件等于用户文件树”的简单模型。

如果现在就做，会明显干扰：

- template content builder
- conflict analyzer
- template apply 写入逻辑

### 11.2 后置阶段最小任务

1. 定义 `AGENTS.md` fragment 与 shared-file manifest 扩展合同
2. 定义 marker 规范
3. template save 时从 runtime `AGENTS.md` 抽取当前 orbit block
4. template apply 时 append / replace 目标 block
5. 增加 AGENTS 冲突检测与交互

---

## 12. 每阶段验证要求

### 12.1 Unit Tests

Phase 2 需要的 unit tests 包括：

- `.orbit/vars.yaml` schema / validation
- `.orbit/template.yaml` schema / validation
- `.orbit/installs/*.yaml` schema / validation
- replacement engine
- 长度倒序替换
- ambiguity detection
- variable scanner
- bindings merge precedence
- render logic
- branch classification

### 12.2 Temp Repo Integration Tests

所有涉及 Git 状态和工作区写入的测试都应使用 isolated temp repos。

至少覆盖：

- `template save`
- `template save --dry-run`
- `template save --edit-template`
- `template save --default`
- `template apply <local-branch>`
- 路径冲突 / orbit 冲突 / branch 已存在冲突
- install record 写入
- apply 不自动 `enter`

### 12.3 Remote Integration Tests

使用本地 bare repo 模拟远程模板源，覆盖：

- `git ls-remote --heads`
- remote template branch 自动选择
- temp ref fetch 与清理
- `template apply <git-url>`

### 12.4 文档与契约验证

在每一阶段结束时，都应验证：

- PRD 与技术 spec 仍然一致
- 命令 flags 与文档一致
- `--json` 输出契约没有漂移
- 错误语义与 fail-closed 行为没有和文档冲突

---

## 13. 共同质量门槛

每个阶段在合并前都应满足：

1. 代码通过格式化、lint 和测试
2. 新增命令至少有一条 temp repo integration test
3. 新增 schema 至少有 validation 单测
4. 关键 fail-closed 行为被单独覆盖
5. 文档同步更新

在提交或合并前，至少执行：

```bash
mise run fmt
mise run lint
mise run test:ci
```

如果阶段中新增了远程模板相关逻辑，还应补跑对应 integration tests，确认：

- temp bare repo 测试通过
- remote source selection 测试通过
- temp ref 清理行为通过

---

## 14. 必须钉死的不变量测试清单

这些测试必须有，不能省：

1. `template apply` 不自动 `enter`
2. `.git/orbit/state/` 不承载模板安装元数据
3. 模板编辑不污染当前 worktree
4. branch 分类不依赖 branch name
5. `.orbit/config.yaml` 不进入模板 branch
6. `.orbit/vars.yaml` 不连 concrete values 一起进入模板 branch
7. 同 literal 多变量时不会静默替换
8. 目标 template branch 已存在且无 `--overwrite` 时失败
9. 远程模板解析只用 Git，不依赖 GitHub API
10. install record 进入 Git 历史，而不是写进 repo-local state

---

## 15. 推荐 issue 拆分顺序

如果你要开始真正创建开发任务，建议顺序如下：

### 第一批：基础能力

1. schema: vars / template / install
2. branch classifier
3. variable scanner
4. replacement engine
5. bindings merge engine
6. template content builder

### 第二批：template save

1. temp index + commit-tree writer
2. save dry-run
3. template save 主命令
4. final template edit
5. save integration tests

### 第三批：template apply

1. local template source resolver
2. render engine
3. conflict analyzer
4. template apply 主命令
5. interactive bindings prompt
6. apply integration tests

### 第四批：branch info

1. branch status
2. branch inspect
3. branch list

### 第五批：远程模板源

1. remote heads discover
2. default template resolution
3. temp ref fetch
4. remote template apply
5. remote integration tests

### 第六批：体验增强

1. bindings init
2. editor mode
3. json outputs
4. dry-run enhancement
5. docs / help / examples sync

### 第七批：后置扩展

1. `AGENTS.md` fragment model
2. `AGENTS.md` save/apply support
3. `AGENTS.md` conflict handling

---

## 16. 每阶段的最小完成定义

### Phase 2A 完成定义

满足以下条件即可进入 Phase 2B：

- schema 已冻结并通过单测
- branch 分类已稳定
- 变量替换与 scanner 已稳定
- `template save` 可生成合法 template branch
- `--edit-template` 不污染当前 worktree

### Phase 2B 完成定义

满足以下条件即可进入 Phase 2C：

- 能从本地 template branch apply
- apply 会写 orbit definition / install record
- apply 默认不自动 enter
- bindings 复用与缺失交互已可用
- branch inspect/status/list 可用

### Phase 2C 完成定义

满足以下条件即可宣告 Phase 2 主线完成：

- 能从远程 Git 仓库自动发现模板源并 apply
- bindings init / editor / json / dry-run 完整可用
- 关键不变量全部被测试覆盖

---

## 17. 可选的实际执行节奏

如果你需要一版更实操的节奏，可以参考：

### 第 1 周

- schema
- classifier
- scanner
- replacement engine
- bindings merge

### 第 2 周

- template content builder
- template branch writer
- save dry-run
- template save

### 第 3 周

- `--edit-template`
- save integration tests
- render engine
- conflict analyzer

### 第 4 周

- local template apply
- interactive bindings
- apply integration tests
- branch status / inspect / list

### 第 5 周

- remote discover
- default template selection
- temp ref fetch
- remote apply

### 第 6 周

- bindings init
- editor mode
- json outputs
- dry-run polish
- docs / help / test stabilization

### V0.2 收尾阶段

- `AGENTS.md` shared file 机制

---

## 18. 阶段完成判定

当以下条件全部满足时，可认为 Phase 2 开发完成：

1. `orbit template save` 稳定可用
2. `orbit template apply <local-branch>` 稳定可用
3. `orbit template apply <git-url>` 稳定可用
4. `orbit bindings init` 稳定可用
5. `orbit branch inspect/list/status` 稳定可用
6. bindings、manifest、install record 三类 schema 已被单测锁定
7. temp repo 和 bare remote integration tests 已覆盖主链路
8. 文档、flags、`--json` 契约保持一致

如果只达到 `template save + local apply`，则应视为 Phase 2 的中间里程碑，而不是最终完成。

---

## 19. 一句话总结

这版开发计划的核心策略是：

**先把“模板内容如何形成、如何写 branch、如何再应用”这条主链路做稳；
再补远程来源、交互增强和特殊文件；
`AGENTS.md` 单列后置，避免主线模型过早复杂化。**
