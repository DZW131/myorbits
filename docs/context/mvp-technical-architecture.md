# Orbit Technical Architecture

本文承接 [product-requirements.md](./product-requirements.md)，描述 Orbit Git MVP 在“单仓库、单工作区、单 branch、文件级 orbit 视图”约束下的技术实现方式。

本文为 sparse-checkout + pathspec 方案提供清晰、可落地、可测试的工程边界。

## 1. Design Intent

- Orbit 是 Git wrapper CLI，不改 Git 本体。
- Orbit 是“当前工作区中的一个可切换视图”，不是 branch、tag 或平行历史。
- 同一时刻只允许一个 current orbit。
- 单工作区方案必须优先处理“脏改动被隐藏”的风险。
- 所有真实写操作必须最终表现为普通 Git commit。
- 版本化配置放在 `.orbit/`；repo-local 运行态放在 `.git/orbit/state/`。
- 命令层保持薄，Git 调用、scope 解析和状态写入分别下沉到同级领域包。
- 实现应围绕 orbit 定义、视图投影与 scoped operations 组织代码。

## 2. Technology Stack

### 2.1 Language and Runtime

- 语言：Go 1.26+
- CLI：`github.com/spf13/cobra`
- 标准库优先：`context`、`encoding/json`、`os`、`os/exec`、`path/filepath`、`strings`
- 构建目标：本地单二进制 CLI，优先支持 macOS / Linux

### 2.2 Configuration and Persistence

- YAML：`gopkg.in/yaml.v3`
  - `.orbit/config.yaml`
  - `.orbit/orbits/*.yaml`
- JSON：标准库 `encoding/json`
  - `.git/orbit/state/current_orbit.json`
  - `.git/orbit/state/warnings.json`
  - `.git/orbit/state/last_status.json`
- 文件写入：原子写 + rename
- 并发保护：repo-local lock file，避免多个 Orbit 命令同时改 sparse-checkout 或运行态状态文件

### 2.3 Pattern Matching

- Orbit pattern 使用 repo-root 相对、斜杠分隔的 glob 语法。
- 推荐使用 `github.com/bmatcuk/doublestar/v4` 解析 `*`、`**`、`?`。
- Orbit pattern 的职责是“把 tracked files 解析成 user-facing data scope”，而不是直接充当 sparse-checkout pattern 语言。
- 控制平面读取范围不通过 orbit pattern 推导。
- sparse-checkout 投影使用的是“resolved projection file list”，不是未解析的 include / exclude 模式。

这一点是刻意设计：

- scope 解析语义由 Orbit 自己稳定控制
- sparse-checkout 只负责把已解析 scope 投影到工作区
- 避免 Git sparse pattern 语义与 Orbit pattern 语义漂移后出现不可解释行为

### 2.4 Git Interaction Model

所有关键 Git 语义都通过系统 `git` 调用完成，不使用 `sh -c`。

必须直接使用 `git` 的能力包括：

- 仓库定位
  - `git rev-parse --show-toplevel`
  - `git rev-parse --absolute-git-dir`
- tracked files
  - `git ls-files -z`
- 控制平面读取回退
  - `git show HEAD:.orbit/config.yaml`
  - `git show HEAD:.orbit/orbits/<orbit-id>.yaml`
  - `git cat-file -e HEAD:<path>`
- 工作区状态
  - `git status --porcelain=v1 -z -uall`
- 视图投影
  - `git sparse-checkout init --no-cone`
  - `git sparse-checkout set --stdin`
  - `git sparse-checkout disable`
- 历史与差异
  - `git diff`
  - `git log`
- 写操作
  - `git add -A --pathspec-from-file=...`
  - `git commit --pathspec-from-file=...`
  - `git restore --source=<rev> --pathspec-from-file=...`
- 可选辅助 refs
  - `git update-ref refs/orbits/<orbit-id>/...`

规则：

- 所有 Git 命令必须使用显式参数列表构造。
- repo-relative path 统一相对于 repo root，而不是当前 cwd。
- 能用 `-z` / NUL 分隔的地方一律优先使用，避免特殊文件名导致歧义。
- 控制平面读取优先读工作树；仅当路径因 sparse view 不可见时，才回退到 `HEAD` 内容。
- Orbit 对 path-limited Git 操作统一走 pathspec helper / temp file abstraction，但 `--pathspec-from-file` 的支持必须按命令能力判断，不能按 Git 大版本统一假定。
- `git add`、`git commit`、`git restore` 在 MVP 中依赖 `--pathspec-from-file` + `--pathspec-file-nul` 作为安全边界。
- `git diff`、`git log` 不应作为文档契约去假定支持 `--pathspec-from-file`；Orbit 可先尝试 temp file 路径，但命令不支持时必须回退到显式 `-- <path>...` 参数列表。

### 2.5 Testing and Quality

- 测试框架：Go `testing`
- 断言库：`github.com/stretchr/testify`
- 所有涉及 Git 状态的测试都在隔离 temp repo 中进行
- 质量基线：
  - scope 解析稳定、幂等
  - sparse-checkout 切换失败保守
  - scoped commit 不得误提交 scope 外文件
  - 状态文件写入原子且可恢复

## 3. Technical Architecture

### 3.1 Command-Centric Structure

建议采用 command-centric 结构：

1. `cmd/orbit/cli/commands`
   - Cobra 命令注册
   - flag 解析
   - stdout / stderr / `--json`
2. `cmd/orbit/cli/orbit`
   - 配置加载
   - orbit 定义校验
   - include / exclude 匹配
   - scope 解析
3. `cmd/orbit/cli/view`
   - enter / leave
   - current orbit 运行态
   - dirty-path safety gate
   - status 分类
4. `cmd/orbit/cli/scoped`
   - files
   - diff
   - log
   - commit
   - restore
5. `cmd/orbit/cli/git`
   - repo 发现
   - tracked files
   - porcelain 解析
   - sparse-checkout
   - pathspec 辅助
   - optional ref updates
6. `cmd/orbit/cli/state`
   - `.git/orbit/state/` 的读写
   - current orbit
   - resolved scope cache
   - warnings / last status / lock
7. `cmd/orbit/cli/ids`
   - orbit id 校验
   - path 规范化帮助函数

命令层不应直接：

- 读写 `.git/info/sparse-checkout`
- 拼 Git 命令细节
- 直接改 `.git/orbit/state/*`

### 3.2 Core Data Model

### Versioned Configuration

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

建议数据结构：

- `GlobalConfig`
  - `Version int`
  - `AlwaysVisible []string`
  - `Behavior BehaviorConfig`
- `BehaviorConfig`
  - `OutsideChangesMode string`
  - `BlockSwitchIfHiddenDirty bool`
  - `CommitAppendTrailer bool`
  - `SparseCheckoutMode string`
- `OrbitDefinition`
  - `ID string`
  - `Description string`
  - `Include []string`
  - `Exclude []string`
- `ScopeSet`
  - `ControlReadPaths []string`
  - `UserDataPaths []string`
  - `CompanionPaths []string`
  - `ProjectionPaths []string`

Orbit 定义文件约束：

- 只识别 `.orbit/orbits/*.yaml`
- 文件名 `<orbit-id>.yaml` 与 YAML 内 `id` 必须一致
- `orbit add` 默认不覆盖已存在的定义文件；若需要覆盖，必须是显式用户动作而不是隐式行为

### Repo-Local Runtime State

`.git/orbit/state/current_orbit.json`

```json
{
  "orbit": "docs",
  "entered_at": "2026-03-19T12:00:00Z",
  "sparse_enabled": true
}
```

`.git/orbit/state/resolved_scope/<orbit-id>.txt`

- 以稳定排序缓存最近一次 projection scope 解析结果
- 主要用于可观察性、调试和状态复用
- 不是权威定义来源，命令正确性不能依赖它存在且有效

`.git/orbit/state/warnings.json`

- 最近一次 `status`、`enter`、`commit` 产生的 warning 摘要

`.git/orbit/state/last_status.json`

- 最近一次状态分类结果

建议数据结构：

- `CurrentOrbitState`
  - `Orbit string`
  - `EnteredAt time.Time`
  - `SparseEnabled bool`
- `ResolvedScope`
  - `Orbit string`
  - `Paths []string`
  - `ResolvedAt time.Time`
- `WarningSnapshot`
  - `CurrentOrbit string`
  - `Messages []string`
  - `CreatedAt time.Time`
- `StatusSnapshot`
  - `CurrentOrbit string`
  - `InScope []PathChange`
  - `OutOfScope []PathChange`
  - `HiddenDirtyRisk []string`
  - `SafeToSwitch bool`
  - `CommitWarnings []string`

### Internal Change Model

- `PathChange`
  - `Path string`
  - `Code string`
  - `Tracked bool`
  - `InScope bool`

`Code` 建议保留 Git porcelain 语义，例如：

- `M`
- `D`
- `A`
- `??`
- `R`

### 3.3 State and Data Boundaries

Orbit MVP 只有三类真相源：

1. Git commit DAG
2. `.orbit/` 下的版本化 orbit 配置
3. `.git/orbit/state/` 下的本地运行态

约束：

- `.orbit/` 是配置，不是缓存。
- `.git/orbit/state/` 是运行态，不是历史系统。
- control scope 读取属于 `.orbit/` 配置层访问能力，不意味着这些路径必须进入 sparse view 或 scoped write set。
- `resolved_scope/*.txt` 只是缓存，不是权威定义。
- `refs/orbits/*` 只是辅助锚点，不是权威历史。
- `sparse-checkout` 只是当前工作区投影，不是 orbit 的真相源。

### 3.4 Control Loading and User View Resolution

Orbit MVP 在实现层建议拆成两段：

1. Control Loading Pipeline
2. User View Resolution Pipeline

#### Control Loading Pipeline

职责：

- 在任何当前 sparse view 下都能读取完整控制平面
- 服务 `init / add / validate / list / show`
- 为 user view resolver 提供全量配置输入

步骤：

1. 优先读取工作树中的 `.orbit/config.yaml`
2. 优先读取工作树中的 `.orbit/orbits/*.yaml`
3. 若某个控制文件当前因 sparse view 不可见，则回退到 `HEAD` 版本

#### User View Resolution Pipeline

`orbit files <orbit-id>`、`orbit enter <orbit-id>`、`orbit status`、`orbit commit`、`orbit restore` 都依赖统一的 user view resolver。

`orbit validate` 与 `orbit files` 必须共享同一个 matcher / resolver 语义，避免“validate 可解析但 files 结果不同”的结构性分裂。

步骤：

1. 通过 control loader 读取 global config
2. 通过 control loader 读取目标 orbit definition
3. 获取 `git ls-files -z` 结果
4. 规范化为 repo-root 相对、斜杠分隔路径
5. 用 include pattern 选出 owned candidate paths
6. 将 `shared_scope` 命中的 tracked paths 并入 owned candidate paths
7. 用 exclude pattern 剔除路径，得到 `OwnedPaths`
8. 计算 `ProjectionOnlyPaths = projection_visible - exclude - OwnedPaths`
9. 将 `.orbit/orbits/<orbit-id>.yaml` 作为 `CompanionPaths` 自动注入
10. 显式排除 `.orbit/config.yaml` 与其它 orbit yaml
11. 生成 `ScopedOperationPaths = OwnedPaths + CompanionPaths`
12. 生成 `ProjectionPaths = OwnedPaths + ProjectionOnlyPaths + CompanionPaths`
13. 去重并稳定排序
14. 回写 `resolved_scope/<orbit-id>.txt`

规则：

- `include` 为空直接校验失败
- `exclude` 允许为空
- `shared_scope` / `projection_visible` 只允许匹配业务文件，不允许命中控制平面路径
- projection scope 允许解析为空，但 `validate` 应给出 warning；是否允许空 view 进入由命令层决定
- `resolved_scope/*.txt` 只能作为 projection scope cache；正确性不能依赖 cache
- sparse-checkout 投影以 `ProjectionPaths` 为准；scoped read/write 以 `ScopedOperationPaths` 为准

### 3.5 View Projection Flow

### `orbit enter <orbit-id>`

执行步骤：

1. 获取 repo root 和 git dir
2. 获取互斥锁
3. 通过 control loader 读取并 `validate` 目标 orbit
4. 解析目标 `ProjectionPaths`
5. 获取工作区状态：`git status --porcelain=v1 -z -uall`
6. 找出 dirty tracked paths
7. 计算“切换后会被隐藏的 dirty tracked paths”
8. 若集合非空且 `block_switch_if_hidden_dirty=true`：
   - 写 warning snapshot
   - 返回非零退出
9. 否则：
   - `git sparse-checkout init --no-cone`
   - `git sparse-checkout set --stdin`
   - stdin 使用 `ProjectionPaths` 的具体路径列表
10. 写入 `current_orbit.json`

说明：

- 这里写入 sparse-checkout 的不是原始 orbit glob，而是 resolved projection file list。
- 由于 MVP 是单工作区，切换 gate 必须发生在真正写 sparse-checkout 之前。
- dirty tracked control-plane files 若会被隐藏，也必须触发 hidden-dirty gate。

### `orbit leave`

执行步骤：

1. 获取互斥锁
2. `git sparse-checkout disable`
3. 删除或清空 `current_orbit.json`
4. 保留 scope cache 与 warnings 作为最近运行痕迹

### 3.6 Status Classification

`orbit status` 依赖统一的 porcelain 解析器。

输入：

- current orbit
- current projection scope
- `git status --porcelain=v1 -z -uall`

输出：

- in-scope modified / deleted / untracked
- out-of-scope modified / deleted / untracked
- hidden dirty risk
- safe-to-switch
- commit warnings

分类规则：

- path 在 current projection scope 内即为 in-scope，否则为 out-of-scope
- untracked path 即使不在 tracked user data scope 内，也要列入 outside changes
- `safe-to-switch=false` 的条件是：存在 dirty tracked paths，且切换到目标 orbit 后会被隐藏

### 3.7 Scoped Read Operations

### `orbit diff`

- 要求 current orbit 存在
- 默认只看 current projection scope
- 通过 Orbit pathspec helper 生成稳定路径列表
- 若 `git diff` 不支持 `--pathspec-from-file`，回退到 `git diff -- <path>...`
- `--outside` 时对 outside changes 生成补集 pathspec

### `orbit log`

- 要求 current orbit 存在
- 通过 Orbit pathspec helper 实现 path-limited history
- 若 `git log` 不支持 `--pathspec-from-file`，回退到 `git log [args...] -- <path>...`
- `--` 之后的参数按原样透传给 Git

### `orbit files`

- 不要求 current orbit
- 直接输出指定 orbit 的 current projection scope
- 只基于 tracked files 解析业务路径，再自动注入当前 orbit definition path
- 输出 repo-root 相对、斜杠分隔、稳定排序路径
- 可回写 `resolved_scope/<orbit-id>.txt`，但 cache 损坏或缺失不能改变命令结果

### `orbit current`

- 只读取 `.git/orbit/state/current_orbit.json`
- 不允许因为“只有一个 orbit”之类的配置事实去推断 current orbit
- state 文件缺失时返回稳定的 `none` 语义
- state 文件损坏时返回错误
- state 引用的 orbit 已不存在时返回 stale warning；这仍然是 state 语义，不应被静默清空

### 3.8 Scoped Write Operations

### `orbit commit -m "<message>"`

这是 MVP 最需要 fail-closed 的命令。

执行步骤：

1. 要求 current orbit 存在
2. 获取互斥锁
3. 解析 current `ProjectionPaths`
4. 读取 porcelain 状态并分类
5. 若存在 outside changes，生成 warning snapshot
6. 若 current projection scope 与 in-scope untracked 内没有可提交改动，返回 `nothing to commit`
7. 为 current `ProjectionPaths` 与 in-scope untracked 生成 NUL-delimited temp pathspec file
8. 执行：
   - `git add -A --pathspec-from-file=<temp> --pathspec-file-nul`
   - `git commit -m <message> --pathspec-from-file=<temp> --pathspec-file-nul`
9. 若配置要求 trailer，则在提交消息末尾追加：

```text
Orbit: <orbit-id>
```

10. 可选 best-effort 更新：

```text
refs/orbits/<orbit-id>/last-scoped
```

关键约束：

- 绝不能把 scope 外改动 silently 带进 commit。
- `git commit --pathspec-from-file=...` 是 scoped commit 的关键安全边界。
- 若可选 ref 更新失败，主 commit 不回滚；仅输出 warning。

### `orbit restore --to <rev>`

执行步骤：

1. 要求 current orbit 存在
2. 获取互斥锁
3. 解析 current `ProjectionPaths`
4. 检查 current projection scope 内是否已有未提交改动
   - MVP 建议默认 fail-closed
   - 后续可再引入 `--force`
5. 检查 `<rev>` 下当前 orbit definition path 是否存在
   - 默认 fail-closed
   - 仅显式 `--allow-delete-current-orbit` 时允许继续
6. 生成 NUL-delimited temp pathspec file
7. 执行：
   - `git restore --source=<rev> --worktree --staged --pathspec-from-file=<temp> --pathspec-file-nul`
   - `git commit -m "<restore message>" --pathspec-from-file=<temp> --pathspec-file-nul`
8. 可选 best-effort 更新：

```text
refs/orbits/<orbit-id>/last-restore
```

9. 若显式允许删除当前 orbit 且目标 revision 中已不存在该 orbit definition：
   - 自动 `git sparse-checkout disable`
   - 清空 current orbit state
   - 输出 warning

恢复语义：

- restore 不是改写 orbit 历史
- restore 是“把某次历史里的 current projection scope 内容重新应用到当前工作区，再提交一个新的普通 Git commit”

### 3.9 Validation Rules

`orbit validate` 必须覆盖：

- 全局配置可解析
- orbit 文件可解析
- orbit id 唯一
- orbit id 合法，可安全用于文件名与 ref 名
- orbit 文件名与 `id` 一致
- include 非空
- glob pattern 合法
- always-visible pattern 合法，且不得命中 `.orbit/config.yaml`、`.orbit/**` 或 `.orbit/orbits/*.yaml`
- scope 可解析，不发生 matcher 错误
- empty scope 给 warning，而不是 hard error
- 行为配置值合法：
  - `outside_changes_mode=warn`
  - `sparse_checkout_mode=no-cone`

输出要求：

- 默认文本输出可读
- `--json` 输出稳定结构
- 多个 orbit 的结果保持稳定排序

### 3.10 Optional Refs Strategy

辅助 refs 不是 MVP 的真相源，但允许作为本地索引：

```text
refs/orbits/<orbit-id>/last-scoped
refs/orbits/<orbit-id>/last-restore
refs/orbits/<orbit-id>/anchors/<name>
```

约束：

- Orbit 必须在没有这些 refs 的情况下仍能完整工作
- 默认不自动 push `refs/orbits/*`
- 这些 refs 失败不会破坏主命令结果

## 4. Code File Structure

建议的代码结构如下：

```text
cmd/
  orbit/
    main.go
    cli/
      root.go
      commands/
        init.go
        add.go
        validate.go
        list.go
        show.go
        files.go
        current.go
        enter.go
        leave.go
        status.go
        diff.go
        log.go
        commit.go
        restore.go
      orbit/
        config.go
        definition.go
        load.go
        match.go
        resolve.go
        validate.go
      view/
        current.go
        enter.go
        leave.go
        status.go
        warnings.go
      scoped/
        files.go
        diff.go
        log.go
        commit.go
        restore.go
      git/
        repo.go
        tracked_files.go
        status.go
        sparse_checkout.go
        diff.go
        log.go
        add_commit.go
        restore.go
        refs.go
        pathspec.go
      state/
        fsstore.go
        lock.go
        current_orbit.go
        projection_cache.go
        warnings.go
        last_status.go
      ids/
        validate.go
      testutil/
        repo.go
```

组织原则：

- 命令层只做参数和输出
- orbit 包只做定义与解析
- view 包只做视图切换和状态分类
- scoped 包只做 orbit 范围内操作
- git 包负责所有 Git 交互细节
- state 包负责 `.git/orbit/state/` 的持久化

## 5. Security & Configuration

### 5.1 Identifier and Path Safety

- `orbit-id` 只能使用安全字符集
- 任何 repo-relative path 都必须：
  - 规范化
  - 阻止 `..`
  - 阻止绝对路径逃逸
- 写入 `.git/orbit/state/resolved_scope/<orbit-id>.txt` 前必须先校验 id

### 5.2 Sparse-Checkout Safety

- 永远在写 sparse-checkout 前做 hidden-dirty 检查
- 不允许命令层跳过该检查直接调用 `git sparse-checkout set`
- `orbit leave` 是唯一允许恢复全视图的正式入口

### 5.3 Scoped Commit Safety

- 所有 scoped Git 操作优先使用 NUL-delimited pathspec files
- `orbit commit` 不能依赖“当前 index 恰好干净”这种隐式假设
- 使用 `git commit --pathspec-from-file=...` 避免 scope 外 staged changes 被误提交

### 5.4 Logging and Output

- 用户输出只展示 orbit id、路径、数量、warning 摘要
- 不打印完整文件内容
- 调试日志与 stdout / stderr 分离
- `--json` 命令输出结构必须稳定，便于后续脚本与 agent 消费

### 5.5 Locking and Recovery

- `enter`、`leave`、`commit`、`restore` 必须持有同一把 repo-local lock
- 异常退出时，lock 的恢复策略必须明确
- 状态写入采用原子替换，避免半写入 JSON

## 6. Testing Alignment

测试应围绕 Orbit Git MVP。

必须优先覆盖：

- 配置解析与 orbit id 校验
- include / exclude / always-visible 的 scope 解析
- empty scope / duplicate id / invalid pattern
- `orbit enter` 的 hidden-dirty gate
- `orbit leave` 恢复完整视图
- `orbit status` 的 in-scope / out-of-scope 分类
- `orbit diff` / `orbit log` 的 pathspec 正确性
- `orbit commit` 只提交 scope 内改动
- 已有 scope 外 staged changes 时，`orbit commit` 仍不能误提交它们
- `orbit restore` 生成新的普通 commit
- 运行态状态文件原子写和锁语义

测试应优先服务当前 MVP 的 orbit 定义、视图切换、状态分类与 scoped operations。
