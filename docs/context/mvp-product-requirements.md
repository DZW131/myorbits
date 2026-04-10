# Orbit Git MVP PRD

## 1. Executive Summary

Orbit Git MVP 是一个运行在标准 Git 仓库之上的本地 CLI，用来在同一个仓库、同一个 branch、同一个工作区里定义多个文件级 `orbit`，并把它们作为可切换的工作视图使用。

Orbit 不替代 Git，也不创建新的历史系统。所有真实版本历史仍然由 Git commit DAG 承载；Orbit 提供的是文件集合视图、作用域内操作，以及围绕单工作区切换风险的一套安全护栏。

MVP 的目标是打通一个最小但完整的闭环：

- 定义 orbit
- 解析 orbit scope
- 切换 orbit view
- 查看 orbit files / status / diff / log
- 只提交 orbit scope 内改动
- 将 orbit scope 恢复到历史 revision，并落成正常 Git commit
- 对 orbit 外改动采用 warning mode，而不是隐式忽略

## 2. Mission

### 2.1 产品使命

让用户在不改变 Git 基础工作流的前提下，按命名文件作用域聚焦仓库中的不同工作维度。

### 2.2 核心原则

- Git 仍是唯一真相源。文件版本、提交历史、merge、rebase、push、pull 都仍由 Git 管理。
- Orbit 不是 branch、tag 或额外 DAG。Orbit 只是同一历史上的文件作用域投影视图。
- 单工作区优先。MVP 不使用 worktree，不创建多工作区并行副本。
- 写操作必须落成普通 Git commit。Orbit 不能生成脱离主历史的隐式平行历史。
- 文件级优先。MVP 只做文件级 orbit，不做 block 级、行级或语义级 orbit。
- 失败保守。任何可能把脏改动静默隐藏、或可能误提交 scope 外改动的场景，都必须显式警告或 fail-closed。
- 本地轻量。MVP 只依赖版本化配置 `.orbit/` 与 repo-local 运行态 `.git/orbit/state/`，不引入数据库、远程服务或后台进程。

## 3. Background and Problem

标准 Git 擅长管理整个仓库的历史，但不擅长表达“我现在只想围绕某一组文件工作”。

在真实仓库中，往往天然存在多个工作维度：

- 文档维度
- CLI 入口维度
- 构建 / 发布维度
- 仓库规则维度

用户希望：

- 在同一个 branch 背景下，只看到某个维度相关文件
- 只对该维度做状态查看、diff、log、commit、restore
- orbit 外改动不要被悄悄吞掉
- 整个仓库仍然保持标准 Git 历史和普通协作方式

Orbit Git MVP 解决的正是这个“单仓库内多文件视图治理”问题。

## 4. MVP Scope

### 4.1 In Scope

#### Core Functionality

- 单仓库、单项目、单工作区、单 branch 背景。
- 多个 orbit 共存，每个 orbit 由命名规则定义。
- `orbit` 由 `id`、`description`、`include`、`exclude` 构成。
- 支持 `shared_scope` 与 `projection_visible` 业务路径规则。
- 支持解析 tracked files 上的 user view scope。
- 支持进入与离开 orbit view。
- 支持 orbit 范围内的 `files`、`status`、`diff`、`log`、`commit`、`restore`。
- 支持 outside changes warning mode。
- 支持在切换时阻止会被隐藏的 dirty tracked paths。
- 所有写操作最终落成普通 Git commit。

#### Technical

- 使用 `.orbit/config.yaml` 与 `.orbit/orbits/*.yaml` 作为版本化配置。
- 使用 `.git/orbit/state/` 作为 repo-local 运行态目录。
- 控制平面读取不依赖当前 sparse view 是否可见。
- 使用 Git sparse-checkout 做视图投影。
- 使用 Git pathspec 做作用域内 diff、log、add、commit、restore。
- 使用系统 `git` CLI 完成 repo root、git dir、status、diff、log、sparse-checkout、pathspec、ref 更新等操作。

#### Integration

- 与普通 Git 命令共存。
- 不改变现有 push / pull / merge / rebase 流程。
- 机器可读命令支持 `--json`。

### 4.2 Out of Scope

#### Core Functionality

- 多 worktree 并行 orbit。
- 多仓库或多项目模型。
- block 级、行级或语义级 orbit。
- 任务、测试、agent run 元数据治理。
- orbit 权限系统。
- 图形界面。

#### Technical

- 对 Git core 做任何修改。
- 自建版本数据库或服务端状态中心。
- 自动 runtime-to-template promotion。
- 自动 push `refs/orbits/*`。
- 背景守护进程、自动 hook、checkpoint 或 metadata branch 工作流。

### 4.3 Success Criteria

用户可以在一个普通 Git 仓库里：

- 定义 `docs` orbit 和 `cmd` orbit
- 进入 `docs` orbit 并只看到 `docs/**`、`README.md` 等相关 tracked files
- 只查看 `docs` orbit 的状态、diff 和历史
- 只提交 `docs` orbit 范围内改动
- 将 `docs` orbit 恢复到旧 revision，并生成新的普通 Git commit
- 在切换会隐藏 dirty tracked paths 时收到警告并被阻止
- 全程仍能正常使用普通 Git 命令

## 5. Target Users

### 5.1 主要用户画像

1. 单人开发者
   希望在复杂仓库中按文件视图聚焦，减少上下文噪音。
2. 人 + agent 协同开发者
   希望明确约束 agent 只围绕某个文件集合工作。
3. 单仓多域代码库维护者
   希望不拆仓、不强行分 branch，也能按领域进行局部操作。

### 5.2 技术熟悉度

- 默认熟悉 Git、CLI 和仓库级工程约束。
- 可以接受 sparse-checkout 与 pathspec 这类 Git 原生能力。

### 5.3 关键需求与痛点

- 整仓 `git status` 与 `git diff` 噪音过大。
- 同一 branch 上不同领域改动容易混入一个提交。
- 单工作区切换文件视图时，最担心脏改动被隐藏。
- 希望 orbit 是 Git 上层编排，而不是新的版本系统。

## 6. Core Concepts

- `Repo`
  标准 Git 仓库。
- `Branch`
  标准 Git branch，仍是主历史线。
- `Orbit`
  一个命名的文件作用域定义，由 include / exclude 规则组成。
- `Orbit View`
  当前工作区应用某个 orbit 后的可见视图。
- `Control Scope`
  Orbit CLI 内部读取控制平面所需的配置范围，至少包含 `.orbit/config.yaml` 与 `.orbit/orbits/*.yaml`。
- `User View Scope`
  某个 orbit 在当前仓库上对用户呈现和对 scoped operations 生效的路径集合。
- `Current Orbit`
  当前工作区已激活的 orbit；MVP 同一时刻只允许一个。
- `Outside Changes`
  当前工作区中不属于 current user view scope 的 modified / deleted / untracked paths。
- `Warning Mode`
  outside changes 默认仅警告，不直接失败；但当切换会把 dirty tracked paths 隐藏时，系统必须阻止。

## 7. User Stories

1. 作为用户，我想进入 `docs` orbit，只围绕文档文件工作，避免被代码与构建文件干扰。
2. 作为用户，我想只查看 `docs` orbit 的历史，而不是整个仓库的历史噪音。
3. 作为用户，我想在 `docs` orbit 中提交文档改动时，不把 `cmd/` 下的临时修改一起带进 commit。
4. 作为用户，我想把 `docs` orbit 恢复到 `HEAD~2` 的内容，但仓库历史仍然是一条普通 Git 历史。
5. 作为用户，我允许 orbit 外改动存在，但希望 Orbit 明确提醒我，不要悄悄忽略。
6. 作为用户，我希望在切换 orbit 时，如果有 dirty tracked files 会被隐藏，系统直接阻止切换。

## 8. Core Architecture & Patterns

### 8.1 高层架构

Orbit Git MVP 采用四层结构：

1. Git Core
   Git 负责 object store、commit DAG、merge、rebase、push、pull、diff、restore。
2. Orbit Spec Layer
   负责 orbit 配置、control scope 读取、include / exclude、always-visible 与行为选项。
3. Orbit Projection Runtime
   负责 user view scope 计算、sparse-checkout 投影、current orbit 运行态、切换安全检查。
4. Orbit Scoped Operations
   负责 `status`、`diff`、`log`、`commit`、`restore` 这些 orbit 范围内操作。

### 8.2 Scope 计算

MVP 对外采用两层 scope 语义：

- `Control Scope`
  - CLI 内部读取完整控制平面所需的范围
  - 至少包含 `.orbit/config.yaml` 与 `.orbit/orbits/*.yaml`
  - 不进入 sparse-checkout，也不默认进入 scoped write
- `User View Scope`
  - 进入 orbit 后用户看到的范围
  - 也是 `status / diff / log / commit / restore` 默认作用的范围

`User View Scope` 的输入：

- `git ls-files` 的 tracked files
- orbit `include`
- orbit `exclude`
- repo 级 `shared_scope`
- repo 级 `projection_visible`

规则：

1. CLI 先读取完整 control scope。
2. 从 tracked files 开始计算用户业务文件集合。
3. 路径匹配至少一个 `include` 的 tracked path 进入 owned 候选集合。
4. `shared_scope` 命中的 tracked path 也并入 owned 候选集合。
5. 匹配任一 `exclude` 的路径从 owned 候选集合剔除。
6. `projection_visible` 命中的 tracked path 只补入 projection-only 集合，不提升为 owned scope。
7. 系统自动把当前 orbit 的定义文件 `.orbit/orbits/<orbit-id>.yaml` 注入 user view。
8. `.orbit/config.yaml` 不进入任何 orbit 的 user view。
9. 其它 orbit yaml 不进入当前 orbit 的 user view。
10. 得到稳定排序后的 projection scope；scoped operations 默认只作用于 owned scope。

MVP 的 sparse-checkout 投影以当前 projection scope 为准；scoped read/write 以 owned scope 为准，而不是 control scope。

### 8.3 单工作区投影

- Orbit view 切换发生在当前工作区本身。
- MVP 统一使用 non-cone sparse-checkout。
- sparse-checkout 写入的是 current user view scope 对应的具体 tracked paths，而不是 control scope，也不是未解析的 orbit 配置模式。
- `orbit leave` 恢复完整 tracked working tree 视图。

### 8.4 Outside Changes Warning Mode

在不同命令中的语义：

- `orbit status`
  - 列出 outside changes
  - 不阻止查看
- `orbit diff`
  - 默认只看 current user view scope
  - `--outside` 时查看 scope 外改动
- `orbit commit`
  - 对 outside changes 发 warning
  - 默认只提交 current user view scope 内改动
  - outside changes 留在工作区
- `orbit enter <target>`
  - 若 dirty tracked paths 在切换后会被隐藏，则 warning + 默认阻止
- `orbit leave`
  - 恢复全视图
  - 不因 outside changes 阻止离开

### 8.5 Scoped Write Model

- `orbit commit` 只提交 current user view scope 内改动。
- `orbit restore --to <rev>` 的语义是把当前 user view scope 内容恢复到 `<rev>`，然后生成新的普通 Git commit。
- 若 `<rev>` 早于当前 orbit 定义文件的存在时间，`restore` 默认 fail-closed；仅显式允许时才可删除当前 orbit 并在成功后自动离开该 orbit view。
- Orbit 不维护独立历史链；orbit history 始终是主提交 DAG 上的路径投影。
- 可选辅助 refs 只作为锚点，不作为真相源。

## 9. Tools / Features

### 9.1 仓库初始化与配置

#### `orbit init`

- 初始化 `.orbit/`、`.orbit/orbits/`、`.git/orbit/state/`
- 生成最小默认配置
- 幂等；再次执行不能悄悄覆盖已有配置
- 只初始化 state 目录，不创建伪造的 `current_orbit.json`
- 不改写仓库历史

#### `orbit add <orbit-id>`

- 创建新的 orbit 定义文件 `.orbit/orbits/<orbit-id>.yaml`
- 生成最小 skeleton，便于用户补充 include / exclude
- skeleton 至少包含 `id`、`description`、`include`、`exclude`
- 文件名 `<orbit-id>.yaml` 与 YAML 内的 `id` 必须一致
- 同名定义文件已存在时稳定失败，不做隐式覆盖

#### `orbit validate`

- 校验 orbit id 唯一
- 校验 pattern 合法
- 校验 include 非空
- 校验 scope 可解析
- 必须实际执行 scope resolution，而不只是做 YAML schema 检查
- scope 解析为空时给出 warning，而不是直接 hard error
- `files` 与 `validate` 必须共享同一套 resolver 语义
- 输出稳定错误

### 9.2 查看与解析

#### `orbit list`

- 枚举 `.orbit/orbits/*.yaml` 中可发现的 orbit
- 输出顺序稳定
- 不依赖 cache；详细配置错误统一交给 `orbit validate`

#### `orbit show <orbit-id>`

- 显示某个 orbit 的规范化定义模型，而不是原始文本文件

#### `orbit files <orbit-id>`

- 输出该 orbit 当前解析后的 user view scope
- 只基于 tracked files 解析业务路径，不把 untracked files 计入 scope
- 当前 orbit 的定义文件由系统自动并入 user view
- `.orbit/config.yaml` 与其它 orbit yaml 不进入该输出
- 可回写 `resolved_scope/<orbit-id>.txt` 作为 cache，但 cache 不是权威定义

#### `orbit current`

- 显示当前运行态中的 current orbit
- current 只能来自 `.git/orbit/state/current_orbit.json`，不能从配置推断
- state 指向的 orbit 已被删除时给出 stale warning
- state 文件损坏时返回错误，不能静默退化成 `none`

### 9.3 视图切换

#### `orbit enter <orbit-id>`

执行逻辑：

1. 读取并校验完整 control scope
2. 解析 target user view scope
3. 检查当前工作区 dirty paths
4. 如果存在切换后会被隐藏的 dirty tracked paths：
   - 输出 warning
   - 默认阻止切换
5. 否则应用 sparse-checkout
6. 写入 current orbit 运行态

说明：

- CLI 必须能在任意当前 view 下读取目标 orbit 定义，而不要求用户先 `leave`

#### `orbit leave`

- 关闭当前 orbit 状态
- `git sparse-checkout disable`
- 恢复完整 tracked working tree

### 9.4 状态与历史

#### `orbit status`

必须展示：

- current orbit
- in-scope modified / deleted / untracked
- out-of-scope modified / deleted / untracked
- hidden-dirty risk
- 是否可安全切换
- 是否存在 commit warning

#### `orbit diff`

- 默认只看 current user view scope diff
- 支持 `orbit diff --outside`

#### `orbit log`

- 基于 Git path-limited log 展示 current user view scope 历史
- 支持 `orbit log -- --oneline -20` 这类参数透传

### 9.5 写操作

#### `orbit commit -m "<message>"`

执行逻辑：

1. 解析 current user view scope
2. 分类 dirty paths
3. 若存在 outside changes，输出 warning
4. 使用 user view scope pathspec 只 stage 并提交 in-scope 改动
5. 提交信息按配置追加 trailer：

```text
Orbit: <orbit-id>
```

6. 可选更新本地辅助 ref：

```text
refs/orbits/<orbit-id>/last-scoped
```

关键语义：

- 只提交 current user view scope 内改动
- outside changes 仍保留在工作区
- 不允许 silently 带上 scope 外变更

#### `orbit restore --to <rev> [--allow-delete-current-orbit]`

执行逻辑：

1. 解析 current user view scope
2. 检查 `<rev>` 下当前 orbit 定义文件是否存在
3. 若不存在：
   - 默认 fail-closed
   - 仅 `--allow-delete-current-orbit` 时允许继续
4. 从 `<rev>` 恢复 current user view scope 内容到当前工作区
5. 生成新的普通 Git commit
6. 提交信息默认带：

```text
restore <orbit-id> orbit to <rev>

Orbit: <orbit-id>
Orbit-Restore-From: <rev>
```

7. 可选更新本地辅助 ref：

```text
refs/orbits/<orbit-id>/last-restore
```

8. 若显式允许删除当前 orbit 且 restore 结果删除了该 orbit：
   - 自动 `leave`
   - 清理 current orbit 运行态
   - 输出 warning

### 9.6 配置与运行态布局

版本化配置：

```text
.orbit/
  config.yaml
  orbits/
    docs.yaml
    cmd.yaml
```

repo-local 运行态：

```text
.git/orbit/
  state/
    current_orbit.json
    resolved_scope/
      docs.txt
      cmd.txt
    warnings.json
    last_status.json
```

### 9.7 示例配置

`.orbit/config.yaml`

```yaml
version: 1

shared_scope: []
projection_visible: []

behavior:
  outside_changes_mode: warn
  block_switch_if_hidden_dirty: true
  commit_append_trailer: true
  sparse_checkout_mode: no-cone
```

`.orbit/orbits/docs.yaml`

```yaml
id: docs
description: Documentation orbit

include:
  - docs/**
  - README.md
  - AGENTS.md
  - CONTRIBUTING.md

exclude:
  - docs/archive/**
```

## 10. Non-Functional Requirements

- 安全性：不能因为视图切换而悄悄隐藏未提交改动。
- 兼容性：应兼容普通 Git 仓库与常规命令。
- 可恢复性：所有 orbit 写操作必须最终体现在正常 Git commit 中。
- 轻量性：MVP 不依赖数据库、远程服务、后台进程。
- 可扩展性：后续可扩展辅助 refs、任务记录、更细粒度 orbit，但不影响当前 Git core 边界。

## 11. Risks and Limitations

1. 单工作区 + sparse-checkout 天然存在切换风险。
   dirty tracked paths 如果被隐藏，会让用户误判改动已消失，因此必须在切换前做严格检查。
2. untracked files 无法像 tracked files 一样被完美投影。
   sparse-checkout 主要作用于 tracked files，因此 orbit view 对 untracked 文件不是绝对纯净视图。
3. 文件级投影没有语义级隔离。
   一个文件可同时属于多个 orbit，多个 orbit 会共享该文件。
4. orbit history 不是独立历史。
   它依赖当前 scope 在主提交 DAG 上的路径投影，因此 scope 变化会影响 log 视图解释。

## 12. MVP Acceptance Criteria

对一个包含 `docs/`、`cmd/`、`README.md` 的普通 Git 仓库：

1. 能定义 `docs` orbit 与 `cmd` orbit。
2. 能进入 `docs` orbit。
3. 能查看 `docs` orbit 的 files、status、diff、log。
4. 能只提交 `docs` orbit 改动。
5. 能把 `docs` orbit 恢复到旧 revision。
6. 在存在将被隐藏的 dirty tracked paths 时，切换 orbit 会警告并阻止。
7. 普通 Git 命令仍可正常使用。
