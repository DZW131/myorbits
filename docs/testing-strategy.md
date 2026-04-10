# Orbit / Harness Testing Strategy

这份文档定义当前 Orbit / Harness 主线的测试策略。

继承下来的 MVP 测试仍围绕“单工作区 orbit 视图系统”展开：

- orbit 定义在 `.orbit/`
- 控制平面读取不依赖当前 sparse view 是否可见
- 当前视图投影使用 sparse-checkout
- 作用域操作使用 current scoped-operation scope 与 pathspec
- 所有真实写操作都落成普通 Git commit

当前 v0.4 主线在此基础上再加一层统一控制面约束：

- branch 顶层身份统一落在 `.harness/manifest.yaml`
- runtime OrbitSpec steady-state host 落在 `.harness/orbits/*.yaml`
- runtime versioned metadata 落在 `.harness/vars.yaml` 与 `.harness/installs/*.yaml`
- orbit template / harness template branch 继续保留各自模板合同文件，但不能回退成旧 `.orbit/*` 顶层身份判断

如果本文件与 [context/mvp-product-requirements.md](./context/mvp-product-requirements.md) 或 [context/mvp-technical-architecture.md](./context/mvp-technical-architecture.md) 冲突，以那两份设计文档为准，并先更新设计。

## 1. Testing Goals

当前主线测试至少要保证七件事：

1. Orbit 不会破坏 Git 的基本语义。
2. Orbit 不会在单工作区里把脏改动静默隐藏。
3. Orbit 不会把 scope 外改动误提交到 scoped commit。
4. `.orbit/` 配置、current user view scope 与 `.git/orbit/state/` 运行态不会混写。
5. 从真实命令入口执行时，用户能得到稳定结果和稳定输出。
6. v0.4 统一控制面 branch identity 与 host paths 不会悄悄回退到旧 `.orbit/*` 顶层 contract。
7. 对外 quickstart 主路径能被一条自动化 acceptance smoke 反复重放。

## 2. Current Test Pyramid

Orbit / Harness 当前需要三层主测试，再加一个可选的后续层。

### 2.1 Layer A: Unit Tests

作用：验证纯逻辑、纯数据转换、局部 Git 适配和本地状态读写。

优先覆盖：

- `cmd/orbit/cli/ids`
  - `orbit-id`
  - repo-relative path 规范化
- `cmd/orbit/cli/orbit`
  - legacy `.orbit/config.yaml` 兼容读取
  - `.harness/manifest.yaml`
  - `.harness/orbits/*.yaml`
  - template / source branch 的 hosted OrbitSpec
  - control loader
  - include / exclude / shared_scope / projection_visible 规则
  - owned scope / projection-only scope / projection paths 解析
  - current orbit definition companion path 注入
  - control-plane 路径排除
- `cmd/orbit/cli/view`
  - current orbit state
  - hidden-dirty gate
  - 基于 current projection scope 的 in-scope / out-of-scope 分类
- `cmd/orbit/cli/scoped`
  - files
  - diff pathspec 组装
  - log pathspec 组装
  - commit trailer
  - restore message
- `cmd/orbit/cli/state`
  - current orbit 读写
  - resolved projection cache
  - warnings / last status
  - lock
  - 原子写
- `cmd/orbit/cli/git`
  - repo root / git dir 解析
  - tracked files 获取
  - porcelain 解析
  - sparse-checkout 调用封装
  - NUL-delimited pathspec helpers
  - optional ref 更新

要求：

- 默认使用 `t.Parallel()`
- 不修改进程级全局状态时必须并行
- 只测一个职责，不把命令拼装、文件系统写入、Git 调用和业务判断混成一个用例

### 2.2 Layer B: CLI Integration Tests

作用：从真实命令边界验证 Orbit Git MVP 的核心工作流。

必须覆盖：

- `orbit init`
  - 创建 `.orbit/`、`.orbit/orbits/`、`.git/orbit/state/`
  - 幂等
  - 不创建伪造的 current orbit state
- `orbit add`
  - 创建 orbit skeleton
  - 非法 `orbit-id` fail-closed
  - 已存在同名定义文件时稳定失败
- `orbit validate`
  - 结构合法时成功
  - duplicate id、filename/id 不一致、invalid pattern、empty include 时失败
  - `shared_scope` / `projection_visible` 命中 control-plane 路径时 fail-closed
  - empty scope 返回 warning
  - 多个 orbit 结果顺序稳定
- `orbit list` / `orbit show`
  - 输出稳定
  - `list` 不依赖 cache，坏定义错误交给 `validate`
- `orbit files`
  - 输出目标 orbit 的 projection paths
  - 自动注入当前 orbit definition path
  - 排除 `.orbit/config.yaml` 与其它 orbit definition paths
  - cache 损坏或缺失时重新解析结果一致
- `orbit current`
  - 无 current orbit 时输出稳定
  - stale state 给 warning
  - damaged state 返回错误
  - enter / leave 后状态正确
- `orbit enter`
  - 通过 control loader 读取 target orbit，即使其 definition 当前被隐藏
  - 解析 target user view scope / projection paths 并应用 sparse-checkout
  - dirty tracked paths 会被隐藏时阻止切换
  - dirty control-plane tracked paths 会被隐藏时同样阻止切换
  - outside untracked files 仅警告
- `orbit leave`
  - 恢复完整 tracked view
  - 不破坏普通工作区
- `orbit status`
  - 同时展示 in-scope / out-of-scope modified、deleted、untracked
  - 给出 hidden-dirty risk 和 commit warnings
- `orbit diff`
  - 默认只看 current scoped-operation scope
  - `--outside` 只看 current projection scope 外改动
- `orbit log`
  - 只看 current scoped-operation scope 的 path-limited history
  - Git 参数透传稳定
- `orbit commit`
  - 只提交 current scoped-operation scope
  - outside changes 保留在工作区
  - 追加 `Orbit: <orbit-id>` trailer
  - 可选辅助 ref 更新失败不回滚主 commit
  - 即使 repo index 已有 scope 外 staged changes，也不能把它们误提交进去
- `orbit restore`
  - 将 current scoped-operation scope 恢复到指定 revision
  - target revision 缺失 current orbit definition path 时默认 fail-closed
  - `--allow-delete-current-orbit` 时可继续并自动 leave
  - 生成新的普通 Git commit
  - 可选辅助 ref 更新失败不回滚主 commit
- `harness init` / `harness create`
  - 创建 runtime `.harness/manifest.yaml` 与 `.harness/orbits/`
  - 已存在非 runtime manifest 时 fail-closed
  - 幂等或已初始化提示稳定
- `harness install`
  - 写入 `.harness/manifest.yaml`、`.harness/orbits/*.yaml`、`.harness/installs/*.yaml`、`.harness/vars.yaml`
  - runtime steady-state 流程不依赖 `.orbit/config.yaml`
  - duplicate orbit / duplicate member 默认 fail-closed
- `harness inspect` / `harness check`
  - member 计数、branch identity、schema 与 install consistency 输出稳定
- `harness template save`
  - 导出 `harness_template` branch
  - branch 内包含 `.harness/manifest.yaml`、`.harness/orbits/*.yaml`

执行方式：

- 每个测试使用独立 temp repo
- 测试里显式初始化 Git 用户配置
- 隔离用户全局 Git 配置
- 通过真实 CLI 命令或 Cobra 根命令入口执行

### 2.3 Layer C: Quickstart Acceptance Smoke

作用：把当前 v0.4 quickstart 主路径固化成一条 doc-derived smoke，覆盖双二进制与四类 branch identity。

必须覆盖：

- build `orbit` / `harness` 双二进制
- template source repo：
  - `harness init`
  - `orbit add`
  - `orbit template save`
  - `orbit branch inspect orbit-template/<orbit-id> --json`
- source authoring repo：
  - `orbit template init-source`
  - `orbit branch inspect HEAD --json`
- runtime repo：
  - `harness create`
  - `harness install`
  - `harness inspect`
  - `harness check --json`
  - 安装完成后先提交首个 runtime commit
  - `orbit enter` / `current` / `status` / `diff` / `leave`
  - `orbit branch inspect HEAD --json`
  - clean runtime writeback
  - migrated runtime writeback
- harness template export：
  - `harness template save`
  - `orbit branch inspect harness-template/<name> --json`
- runtime repo 断言：
  - `.harness/manifest.yaml` 存在
  - `.harness/orbits/*.yaml` 存在
  - `.orbit/config.yaml` 不作为正式 runtime host 出现
  - migrated runtime 里的 legacy `.orbit/config.yaml` 不会被带回 template payload

执行方式：

- 入口统一使用 `mise run acceptance:quickstart`
- shell 脚本应输出稳定步骤名，失败时能定位到具体阶段
- 文档主路径若发生 drift，这层 smoke 必须先红灯

### 2.4 Layer D: Deferred Tests

这些测试不是 MVP 起步阶段必须项，只有在复杂度明显上升后再加：

- benchmark / perf regression
- build-tag 分层集成测试
- 大 scope 仓库压力测试
- 并发竞争专项测试
- 除 quickstart smoke 外更大、更慢的真 E2E 脚本测试

触发条件：

- orbit 数量明显增多
- 仓库文件数量明显增多
- sparse-checkout 切换性能成为实际问题
- CI 耗时明显上升
- 并发执行 Orbit 命令成为真实场景

## 3. Minimum Coverage Matrix

下面这些场景在当前主线必须有测试。MVP 继承项仍是基线，v0.4 统一控制面覆盖是增量要求。

### 3.1 Git and Path Safety

- 非 Git 仓库中运行命令应失败
- 子目录运行命令时仍能正确解析 repo root
- 非法 `orbit-id` 不能进入状态文件路径
- 非法路径片段不能逃逸 repo root
- NUL-delimited pathspec 文件能正确处理带空格路径

### 3.2 Init and Config

- 首次初始化创建最小目录结构
- 再次执行不会破坏已有配置和状态
- 缺失父目录时能自动创建
- 默认配置内容稳定
- `init` 不会伪造 `current_orbit.json`

### 3.3 Validate

- `.orbit/config.yaml` 缺失
- `.orbit/orbits/` 缺失
- orbit 定义 YAML 非法
- `id` 缺失
- duplicate id
- 文件名与 `id` 不一致
- `include` 为空
- invalid include glob
- invalid exclude glob
- invalid shared-scope glob
- invalid projection-visible glob
- `shared_scope` 命中 `.orbit/config.yaml`
- `shared_scope` 命中 `.orbit/orbits/*.yaml`
- `projection_visible` 命中 `.orbit/config.yaml`
- `projection_visible` 命中 `.orbit/orbits/*.yaml`
- 行为配置值非法
- scope 解析为空时给出 warning

### 3.4 Control Loading and User View Resolution

- control loader 在工作树可见时读取工作树配置
- control loader 在控制文件因 sparse view 不可见时回退到 `HEAD`
- `include` 命中目录与根文件
- `shared_scope` 正确并入 owned user data paths
- `projection_visible` 只并入 projection-only paths
- `exclude` 正确覆盖 `include` / `shared_scope` / `projection_visible`
- 当前 orbit definition path 自动注入 projection scope
- `.orbit/config.yaml` 不进入任何 user view scope / projection scope
- 其它 orbit definition paths 不进入当前 orbit 的 user view scope / projection scope
- tracked files 稳定排序
- scope cache 写入幂等
- scope cache 损坏不影响重新解析

### 3.4A Template-Owned Scope

- `orbit template save` 只消费 owned scope
- `orbit template publish` 只消费 owned scope
- `projection_visible` 文件不会进入 orbit template
- `harness template save` member candidate 只消费 owned scope
- 模板变量扫描 / 渲染 / runtime-to-template replacement 仅对 `.md` 文件生效
- `AGENTS.md` 仅在属于 owned scope 时才进入 template special lane

### 3.4B Single-Control-Plane Branch Identity

- runtime branch 存在 `.harness/manifest.yaml` 且 `kind=runtime`
- runtime branch 的 OrbitSpec steady-state host 在 `.harness/orbits/*.yaml`
- runtime steady-state 命令在没有 `.orbit/config.yaml` 时仍能工作
- source branch 存在 `.harness/manifest.yaml` 且 `kind=source`
- orbit template branch 存在 `.harness/manifest.yaml` 且 `kind=orbit_template`
- orbit template payload 导出 `.harness/orbits/<orbit-id>.yaml`
- harness template branch 存在 `.harness/manifest.yaml` 且 `kind=harness_template`
- harness template payload 导出 `.harness/orbits/*.yaml`
- `branch inspect` 对 runtime / source / orbit_template / harness_template 四类 branch 的 `kind` / `template_kind` 输出稳定
- install / apply / publish 的 source loading 允许 branch manifest，并对缺失或不一致的 branch manifest fail-closed

### 3.5 Enter and Leave

- 进入 orbit 时写入 sparse-checkout
- 进入 orbit 时写入 current orbit state
- dirty tracked paths 会被目标 view 隐藏时阻止进入
- dirty `.orbit/config.yaml` 会被目标 view 隐藏时阻止进入
- dirty 当前 orbit definition path 在切换到其它 orbit 后会被隐藏时阻止进入
- dirty tracked paths 若仍在目标 projection scope 内，则允许进入
- outside untracked path 只警告，不作为切换阻塞条件
- 在一个 orbit 内直接 `enter` 另一个 orbit 时，target definition 即使当前不可见也能成功读取
- `leave` 后 sparse-checkout disable
- `leave` 在 current state 缺失时仍能恢复完整 tracked view
- `leave` 后 current orbit state 清空

### 3.6 Status

- current orbit 为空时提示稳定
- modified / deleted / untracked 基于 current projection scope 正确分类
- hidden-dirty risk 正确计算
- last status snapshot 可重复写入
- warnings snapshot 可重复写入

### 3.7 Diff and Log

- `diff` 仅包含 current user view scope / projection paths
- `diff --outside` 仅包含 current projection scope 外路径
- `log` 只展示 current user view scope / projection paths 历史
- `log -- --oneline -20` 参数透传正确
- `diff` / `log` 在 current orbit 缺失时 fail-closed
- `diff` / `log` 的 `--json` 输出结构稳定

### 3.8 Commit

- current orbit 缺失时失败
- 无 in-scope 改动时返回 `nothing to commit`
- 只提交 current projection scope 内 modified / deleted / untracked
- `.orbit/config.yaml` 不会被普通 scoped commit 静默带入
- scope 外 unstaged 改动保留
- scope 外 staged 改动不被误提交
- 提交信息追加 orbit trailer
- `refs/orbits/<orbit-id>/last-scoped` 更新为 best-effort

### 3.9 Restore

- current orbit 缺失时失败
- target revision 缺失 current orbit definition path 时默认失败
- `--allow-delete-current-orbit` 时允许继续 restore
- `--allow-delete-current-orbit` 成功后自动 leave 并清理 `current_orbit.json`
- restore 到历史 revision 后生成新 commit
- `.orbit/config.yaml` 不会被普通 scoped restore 误恢复
- restore message 包含 `Orbit:` 与 `Orbit-Restore-From:`
- `refs/orbits/<orbit-id>/last-restore` 更新为 best-effort
- restore 默认在 current projection scope 内已有脏改动时 fail-closed

### 3.10 State and Locking

- current orbit JSON 原子写
- resolved projection cache 原子写
- warning / status snapshot 原子写
- lock 能阻止并发写命令互相踩踏
- 异常后 lock 恢复策略稳定

### 3.11 Quickstart Smoke

- `docs/quickstart.md` 的正式命令序列能被脚本逐步重放
- smoke 会同时覆盖 `runtime`、`source`、`orbit_template`、`harness_template` 四类 branch
- 安装完成后先提交首个 runtime commit，再执行 Orbit projection 命令
- runtime writeback smoke 会覆盖 clean runtime 与 migrated runtime 两条基础路径
- 双二进制构建、help 路径和文档主路径发生 drift 时，这层 smoke 能优先暴露问题

## 4. Test Harness Rules

- 所有涉及 Git 的测试必须使用独立 temp repo。
- 所有 repo 都必须显式设置：
  - `user.name`
  - `user.email`
- 默认禁用用户全局 Git 配置对测试的影响。
- 优先使用真实 `git` 命令准备测试数据，不手写伪造 `.git/` 内容。
- 默认使用 `t.Parallel()`；只有修改 cwd、env 或共享临时资源时才禁用。
- 对用户输出的断言优先验证稳定字段和关键 warning，不依赖无关空白细节。
- 当测试需要比较文件列表时，统一使用 repo-root 相对、斜杠分隔、稳定排序。
- 触达 quickstart / help / release 文档，或触达单控制面 branch kind / host path 路由时，至少重跑一次 `mise run acceptance:quickstart`。
- shell-based acceptance 应优先复用现有 `scripts/` 和 `mise` 入口，不为单个 smoke 重复发明新的测试 harness。
