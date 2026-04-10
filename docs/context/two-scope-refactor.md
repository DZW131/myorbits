# Orbit 两层 Scope 重构设计

本文是 Orbit Git MVP 的一份独立重构设计提案，用于解决当前单层 scope 模型在 Phase 4 之后暴露出的控制平面误伤问题。

本文本身不是新的 source of truth；若采纳，后续应把对应结论同步回写到：

1. `docs/context/mvp-product-requirements.md`
2. `docs/context/mvp-technical-architecture.md`
3. `docs/testing-strategy.md`
4. `docs/context/mvp-development-plan.md`

## 1. 背景与问题定义

当前实现和现有文档基本采用单层 scope 模型：

- scope 由 `tracked files + include + exclude + shared_scope` 解析得到
- `orbit files / enter / status / diff / log / commit / restore` 共用这一份 resolved scope
- 历史示例里曾通过 `always_visible` 并入额外可见路径；当前术语已收口为 `shared_scope`

这在 Phase 1 到 Phase 3 基本可用，但到了 Phase 4 会出现一个真实风险：

- `commit` 和 `restore` 的 pathspec 默认沿用同一份 current scope
- 若 `.orbit/config.yaml` 或其它控制平面路径进入 current scope，scoped write 就会把它们当成普通业务文件处理
- 尤其是 `restore --to <old-rev>`，当目标 revision 早于 Orbit 配置出现时间时，控制平面文件会被恢复成“不存在”

问题本质不是“restore 实现错了”，而是：

**控制平面读取范围、用户视图范围、scoped read/write 范围，被压成了同一个 scope。**

这与现有边界要求是冲突的：

- Git DAG 是历史真相源
- `.orbit/` 是版本化配置
- `.git/orbit/state/` 是 repo-local 运行态
- sparse-checkout 是投影，不是真相源

## 2. 重构目标

这次重构的目标不是扩大 MVP，而是在不破坏现有架构边界的前提下，把“控制平面可读”与“用户可见/可写作用域”拆开。

必须同时满足：

1. 普通 `diff / log / commit / restore` 不会误伤 `.orbit/config.yaml`
2. 当前 orbit 仍然可以带着自己的定义文件工作
3. CLI 在任何 sparse view 下都能读取完整控制平面，支持直接从一个 orbit 切到另一个 orbit
4. `.git/orbit/state/` 仍然只是运行态，不升级成配置镜像层

不在本次设计中引入：

- worktree
- daemon
- 配置数据库
- 语义级/block 级 scope
- 依赖 `refs/orbits/*` 的主路径正确性

## 3. 总体方案

对外语义采用“两层 scope”：

- `Control Scope`
- `User View Scope`

但在代码内部，不建议只保留两个扁平列表。为了避免未来再次拆分，建议内部 resolver 直接返回一个结构化结果：

```go
type ScopeSet struct {
    ControlReadPaths []string
    UserDataPaths    []string
    CompanionPaths   []string
    ProjectionPaths  []string
}
```

语义说明：

- `ControlReadPaths`
  - CLI 内部读取控制平面所需路径
  - 不属于用户视图，也不属于默认 scoped write set
- `UserDataPaths`
  - orbit 真正的业务文件边界
  - 来自 `include / exclude` 解析结果与业务型 `shared_scope`
- `CompanionPaths`
  - 当前 orbit 的伴随控制文件
  - MVP 中仅包含 `.orbit/orbits/<current-orbit>.yaml`
- `ProjectionPaths`
  - 实际写入 sparse-checkout、构成“用户进入 orbit 后看到什么”的路径集合
  - `ProjectionPaths = UserDataPaths + CompanionPaths`

对外文档仍然可以继续只讲两层：

- `Control Scope`
- `User View Scope`

其中：

- 对外的 `User View Scope` 对应内部的 `ProjectionPaths`
- 但代码内部保留 `UserDataPaths` 与 `CompanionPaths` 的区别，避免未来对 `files / diff / commit / restore` 的策略调整时再次推倒 resolver

## 4. 新语义定义

### 4.1 Control Scope

`Control Scope` 是 Orbit CLI 内部使用的控制平面读取范围，不进入 sparse-checkout，不默认进入 scoped read/write。

至少固定包含：

- `.orbit/config.yaml`
- `.orbit/orbits/*.yaml`

它只服务于：

- 配置加载
- orbit 枚举
- orbit 校验
- user view resolver 的输入
- 当前/目标 orbit 的控制决策

它不是用户视图，也不是 cache。

### 4.2 User Data Scope

`UserDataPaths` 是 orbit 真正的业务作用域，只由业务规则构成：

- tracked(include - exclude)
- tracked(shared_scope_user)

其中：

- `shared_scope` 的语义在本次重构中收窄
- 只允许匹配跨 orbit 共享显示的业务文件
- 不再承担控制平面注入职责

### 4.3 Companion Paths

`CompanionPaths` 是当前 orbit 的伴随控制文件。

MVP 中仅自动注入：

- `.orbit/orbits/<current-orbit>.yaml`

显式排除：

- `.orbit/config.yaml`
- 其它 orbit yaml

这样做的原因是：

- 当前 orbit 的定义文件应能随 orbit 一起工作
- 但全局配置与其它 orbit 定义不应进入当前 orbit 的默认 scoped write 集合

### 4.4 Projection Scope

`ProjectionPaths` 是真正写入 sparse-checkout 的集合，也是面向用户的“进入此 orbit 后会看到什么”。

MVP 中定义为：

```text
ProjectionPaths = UserDataPaths + CompanionPaths
```

本次重构不再允许通过 `shared_scope` 把整个 `.orbit/**` 注入任何 orbit 的用户视图。

## 5. `shared_scope` / `projection_visible` 语义收窄

当前示例配置把 `.orbit/**` 放进历史术语 `always_visible`。

重构后：

- `shared_scope` 仅表示会进入 owned scope 的跨 orbit 共享业务文件
- `projection_visible` 仅表示投影可见但不属于 owned scope 的业务文件
- 默认示例删除 `.orbit/**`，且 `shared_scope` 默认为空
- 当前 orbit yaml 的可见性由系统自动注入，不再依赖 `always_visible`

因此，以下配置应在新模型中被视为非法：

- `.orbit/config.yaml`
- `.orbit/**`
- `.orbit/orbits/*.yaml`

## 6. 控制平面读取策略

控制平面读取必须从“当前工作树是否可见”中解耦。

推荐读取优先级：

1. 工作树优先
   - 若文件当前在工作树中可见且存在，直接读工作树
   - 这样在 full view 或当前 orbit yaml 可见时，CLI 可以感知未提交配置改动
2. Git 当前提交回退
   - 若文件当前因 sparse view 不可见，则从 Git `HEAD` 读取其内容
   - 推荐使用 blob 读取，不把配置镜像到 `.git/orbit/state/`

这个策略依赖一个必须写进文档的不变量：

**任何会把 dirty tracked control-plane files 隐藏掉的 view switch，都必须被 hidden-dirty gate 阻止。**

否则，CLI 就可能在用户不知情的情况下对隐藏的工作树改动退回到 `HEAD` 内容。

## 7. 命令级语义矩阵

### 7.1 `orbit init`

语义基本不变：

- 初始化 `.orbit/`
- 初始化 `.orbit/orbits/`
- 初始化 `.git/orbit/state/`
- 不伪造 current orbit state

变化：

- 默认配置不再写 `.orbit/**` 或 `README.md` 到 `shared_scope`

### 7.2 `orbit add / validate / list / show`

这些属于控制平面命令，应基于 `Control Scope` 工作，而不是依赖当前用户视图可见性。

要求：

- `list` 在任何当前 view 下都能枚举所有 orbit
- `show <id>` 在任何当前 view 下都能读取目标 orbit yaml
- `validate` 校验完整控制平面，不只校验当前 view 可见的定义文件

### 7.3 `orbit files <id>`

建议重定义为：

- 输出目标 orbit 的 `ProjectionPaths`
- 这是“进入此 orbit 后用户会看到的路径集合”
- 不是控制平面读取范围

若后续需要更细粒度展示，可再扩展：

- `orbit files --data`
- `orbit files --projection`

但 MVP 首版不要求立刻加 flag。

### 7.4 `orbit enter <target>`

新流程：

1. 通过 control loader 读取完整控制平面
2. 校验目标 orbit
3. 解析 target `ProjectionPaths`
4. 读取 porcelain 状态
5. 计算 hidden-dirty gate
6. 通过后写入 sparse-checkout
7. 写入 `current_orbit.json`

关键效果：

- 在 `docs` orbit 中可以直接 `orbit enter cmd`
- 即使 `cmd.yaml` 当前不在用户视图中，CLI 仍能读取它

### 7.5 `orbit leave`

语义不变：

- 获取锁
- `git sparse-checkout disable`
- 清理 `current_orbit.json`
- 恢复 full tracked view

### 7.6 `orbit status`

`status` 的 current scope 改为 current `ProjectionPaths`。

效果：

- `.orbit/config.yaml` 不再属于 in-scope
- 当前 orbit yaml 属于 in-scope
- 其它 orbit yaml 不属于 current view scope

tracked 路径按 `ProjectionPaths` 判定；untracked 路径继续按 orbit matcher 判定。

### 7.7 `orbit diff / orbit log`

默认作用于 current `ProjectionPaths`。

理由：

- 这与“用户当前进入 orbit 后看到什么”一致
- 当前 orbit yaml 若可见，也可被一起 diff/log

### 7.8 `orbit commit`

默认提交：

- `ProjectionPaths`
- 加上 in-scope untracked paths

这样可以保证：

- `.orbit/config.yaml` 不会再被 scoped commit 默认带上
- 当前 orbit yaml 仍可作为 companion path 与业务文件一起提交
- 当前实现中“commit 需要补上 in-scope untracked”的要求仍成立

### 7.9 `orbit restore --to <rev>`

默认恢复 current `ProjectionPaths`，但必须增加 birth-boundary preflight。

预检查：

- 检查 `<rev>` 下是否存在 `.orbit/orbits/<current>.yaml`

若不存在：

- 默认 fail-closed
- 仅在显式传入 `--allow-delete-current-orbit` 时允许继续

若允许继续：

1. 使用当前 `ProjectionPaths` 完成 restore
2. 生成普通 Git commit
3. 自动 `leave`
4. 清空 `current_orbit.json`
5. 输出 warning：当前 orbit 已因 restore 消失，系统已自动退出 orbit view

这样可以避免 restore 成功后留下 stale current state。

## 8. 包级改造方案

### 8.1 `cmd/orbit/cli/orbit`

这是本次重构的核心。

建议新增两组职责。

#### A. Control Loader

建议接口：

```go
type ControlLoader interface {
    LoadGlobalConfig(ctx context.Context, repoRoot string) (GlobalConfig, error)
    ListDefinitions(ctx context.Context, repoRoot string) ([]Definition, error)
    LoadDefinition(ctx context.Context, repoRoot string, orbitID string) (Definition, error)
}
```

职责：

- 工作树可见时读工作树
- 工作树不可见时回退到 `HEAD`

#### B. Scope Resolver

建议替换当前单一 `ResolveScope(...) []string` 风格接口，改成显式语义接口：

```go
func ResolveScopeSet(
    ctx context.Context,
    repo git.Repo,
    loader ControlLoader,
    orbitID string,
) (ScopeSet, error)
```

或更窄化的接口：

```go
func ResolveProjectionScope(...)
func ResolveUserDataScope(...)
func ResolveCurrentScopeSet(...)
func ResolveTargetScopeSet(...)
```

不要继续复用旧的“resolved scope”命名去承载新语义。

### 8.2 `cmd/orbit/cli/git`

新增控制文件读取与 revision 探测 helper：

```go
ReadFileWorktreeOrHEAD(ctx, repoRoot, path) ([]byte, error)
ReadFileAtRev(ctx, repoRoot, rev, path) ([]byte, error)
PathExistsAtRev(ctx, repoRoot, rev, path) (bool, error)
```

用途：

- control loader 读隐藏控制文件
- restore birth-boundary preflight

保留现有原则：

- 继续使用系统 `git`
- 继续使用显式参数列表
- 不使用 `sh -c`

### 8.3 `cmd/orbit/cli/view`

#### `enter.go`

- 用 control loader 读配置
- 用新的 resolver 解析 target `ProjectionPaths`
- hidden-dirty gate 继续按“切换后会被隐藏的 dirty tracked paths”判断

这意味着：

- `.orbit/config.yaml` 在 full view 中若 dirty，进入任何 orbit 都会被阻止
- `docs.yaml` dirty 时，从 `docs` 切到 `cmd` 也会被阻止

#### `status.go`

- 改用 current `ProjectionPaths`
- 分类器本身不必改变“tracked vs untracked”的总规则

#### `leave.go`

- 不因两层 scope 设计本身改变
- 保持当前“以恢复 full tracked view 为中心”的修复结果

### 8.4 `cmd/orbit/cli/scoped`

建议把 scoped 层的 scope 入口统一起来：

```go
ResolveCurrentProjectionScope(...)
ResolveTargetProjectionScope(...)
ResolveOutsidePaths(...)
```

`diff / log / commit / restore` 都只从这里拿 scope，不再各自碰 orbit 解析细节。

#### `diff.go / log.go`

- 改为依赖 current `ProjectionPaths`
- 现有 pathspec / fallback 逻辑可基本保留

#### `commit.go`

- 改为依赖 current `ProjectionPaths`
- scoped write set = `ProjectionPaths + in-scope untracked`

#### `restore.go`

- 在调用 `git restore` 前增加 birth-boundary preflight
- 支持 `--allow-delete-current-orbit`
- 若允许删除当前 orbit，则 restore 完成后自动 leave

### 8.5 `cmd/orbit/cli/commands`

命令层继续保持薄。

只做：

- 参数读取
- 错误转译
- 输出与 `--json`

新要求：

- `restore` 增加 `--allow-delete-current-orbit`
- `files` 输出 `ProjectionPaths`
- 其它命令接到新的 scope resolver 接口

### 8.6 `cmd/orbit/cli/state`

状态层不升级为配置镜像层。

继续只保留：

- `current_orbit.json`
- `resolved_scope/*.txt`
- `warnings.json`
- `last_status.json`
- lock

若允许 `restore --allow-delete-current-orbit`：

- restore 成功后应清空 `current_orbit.json`
- 可选写一条 warning snapshot，说明 orbit 已被该 restore 删除并已自动 leave

## 9. 迁移与校验规则

### 9.1 默认配置迁移

默认配置不再生成：

```yaml
shared_scope:
  - .orbit/**
```

推荐示例应收缩为业务文件，例如：

```yaml
shared_scope: []
projection_visible:
  - README.md
```

### 9.2 旧仓库兼容策略

`validate` 应新增 hard error 或明确迁移提示：

- `shared_scope` 命中 `.orbit/config.yaml`
- `shared_scope` 命中 `.orbit/orbits/*.yaml`
- `shared_scope` 使用 `.orbit/**`

推荐策略：

- 首版直接 hard error
- 不允许系统默默沿用旧行为

理由：

- 继续放行就无法保证 restore/commit 的安全边界

## 10. 测试计划

所有涉及 Git 状态的测试继续使用隔离 temp repo。

### 10.1 Unit Tests

#### `orbit`

1. control loader：工作树可见时读工作树
2. control loader：工作树不可见时回退到 `HEAD`
3. resolver：自动注入当前 orbit yaml
4. resolver：排除 `.orbit/config.yaml`
5. resolver：排除其它 orbit yaml
6. validate：拒绝 `shared_scope` / `projection_visible` 命中控制平面路径

#### `git`

1. `ReadFileAtRev`
2. `PathExistsAtRev`
3. hidden path 下的 `HEAD` 读取能力

#### `scoped`

1. current `ProjectionPaths` 解析
2. outside path 补集计算仍正确
3. restore preflight：目标 rev 缺少当前 orbit yaml 时失败
4. restore preflight：`--allow-delete-current-orbit` 时允许继续

### 10.2 CLI Integration Tests

1. 进入 `docs` orbit 后：
   - `.orbit/config.yaml` 不在工作树可见范围
   - `.orbit/orbits/docs.yaml` 可见
   - `.orbit/orbits/cmd.yaml` 不可见
2. 在 `docs` orbit 中直接 `orbit enter cmd`：
   - 命令成功
   - 证明 CLI 不依赖当前工作树可见性读取 `cmd.yaml`
3. `orbit files docs` 输出 user view，而非 control scope
4. `orbit diff / log / commit` 默认不涉及 `.orbit/config.yaml`
5. `orbit restore --to <rev-with-config>`：
   - 恢复成功
   - 不误删全局配置
6. `orbit restore --to <rev-before-orbit-birth>`：
   - 默认失败
   - 带 `--allow-delete-current-orbit` 成功
   - 成功后自动 leave
7. hidden-dirty gate：
   - `.orbit/config.yaml` dirty 时进入任意 orbit 被阻止
   - `docs.yaml` dirty 时从 `docs` 切到 `cmd` 被阻止

## 11. 文档修改清单

采纳此方案后，至少需要同步修改：

### `docs/context/mvp-product-requirements.md`

- `shared_scope` / `projection_visible` 的职责
- scope 计算章节
- `enter / files / status / diff / log / commit / restore` 的“current scope”表述
- 示例配置，删除 `.orbit/**`

### `docs/context/mvp-technical-architecture.md`

- data boundary 中补充 control-plane read 与 user view 的区别
- scope pipeline 从单层改成 control loading + projection resolution
- `enter` / `status` / `diff` / `log` / `commit` / `restore` 全部改成基于 current `ProjectionPaths`
- restore 增加 birth-boundary preflight 与 `--allow-delete-current-orbit`

### `docs/testing-strategy.md`

- 增加 control loader、projection resolver、restore birth-boundary 的测试矩阵
- 更新 `files`、`commit`、`restore` 对 scope 的定义

### `docs/context/mvp-development-plan.md`

- 把后续 scope 重构工作拆成独立阶段或独立 PR 计划
- 不要继续把单层 resolved scope 当成未来正确模型

## 12. 推荐实施顺序

### Phase A：文档先行

- 更新 `product-requirements.md`
- 更新 `technical-architecture.md`
- 更新 `testing-strategy.md`

### Phase B：Control Loader 落地

- 不改外部 CLI 语义
- 先让 CLI 在任何 view 下都能读取完整控制平面

### Phase C：Projection Resolver 落地

- 把旧的 resolved scope 改成 `ScopeSet`
- 更新 `files / enter / status / diff / log`

### Phase D：Scoped Write 收口

- 更新 `commit / restore`
- 加上 restore birth-boundary preflight
- 增加 `--allow-delete-current-orbit`

### Phase E：迁移与清理

- validate 拒绝旧式 `.orbit/**` 注入
- 更新默认配置示例
- 清理旧命名与兼容层

## 13. 最终结论

这次“两层 scope 重构”的核心不是新增一个 state 镜像层，也不是把控制平面塞进 `.git/orbit/state/`。

它解决问题的方式是：

1. 把控制平面读取与用户 scoped read/write 范围分开
2. 让 `.orbit/config.yaml` 退出所有默认 user view
3. 让当前 orbit yaml 以 companion path 身份进入当前 orbit view
4. 让 `diff / log / commit / restore` 默认只作用于 user view 的 projection 集合
5. 对“restore 到 orbit 诞生前”增加显式边界与 fail-closed 规则

这样可以同时满足：

- 普通 restore/commit 不误伤全局配置
- 当前 orbit 仍能带着自己的定义文件工作
- CLI 仍能直接跨 orbit 切换
- `.orbit/`、`.git/orbit/state/`、sparse-checkout 三类状态边界不被打乱
