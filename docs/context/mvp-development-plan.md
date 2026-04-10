# Orbit Git MVP 开发计划

## 1. 文档目标

本文档基于以下 source of truth 制定：

1. `docs/context/mvp-product-requirements.md`
2. `docs/context/mvp-technical-architecture.md`
3. `docs/testing-strategy.md`
4. `CONTRIBUTING.md`
5. `AGENTS.md`

目标是为 Orbit Git MVP 提供一份可执行的开发推进计划，明确：

- 分阶段目标与交付物
- 每阶段的开发内容与注意事项
- 测试与质量保证要求
- 建议分支命名与合并节奏

本文档只覆盖当前 MVP，不扩展到 worktree、多仓库、远程服务、数据库、GUI 或 block-level orbit。

## 2. 总体开发策略

### 2.1 开发原则

整个实现必须持续满足 PRD 的核心原则：

- Git 是唯一真相源，Orbit 只是文件作用域视图和 scoped 操作层。
- MVP 只支持单仓库、单项目、单工作区、单 branch 背景。
- 所有写操作必须落成普通 Git commit。
- Orbit 只处理文件级 scope，不进入 block、line 或语义级别。
- 遇到可能隐藏脏改动或误提交 scope 外改动的场景，必须 fail-closed 或显式 warning。
- `.orbit/` 是版本化配置，`.git/orbit/state/` 是本地运行态，二者不能混用。

### 2.2 开发顺序

建议按以下顺序推进，而不是并行铺开全部命令：

1. 先打稳基础设施与数据边界。
2. 再完成 control loading、配置校验与 user view resolution。
3. 再做 view projection 与 status classification。
4. 最后做 scoped read/write operations。
5. 收尾阶段集中做稳定性、JSON 输出、文档与验收闭环。

原因很直接：

- `enter`、`status`、`commit`、`restore` 都依赖 control loader 与统一的 user view resolver。
- `commit` 和 `restore` 是风险最高的 fail-closed 区域，必须建立在前置能力稳定的前提下。
- sparse-checkout、state、warning snapshot、lock 都属于全局共享能力，越早收敛越好。

### 2.3 命令批次视图

从命令交付视角，建议按下面四批推进：

1. 第一批：`init / add / validate / list / show / files / current`
2. 第二批：`enter / leave / status`
3. 第三批：`diff / log`
4. 第四批：`commit / restore`

这与工程分阶段并不冲突：

- Phase 0 负责把基础骨架、Git 适配、state、ids、lock、原子写先打稳。
- Phase 1 对应第一批命令中的“定义与解析闭环”。
- Phase 2 对应第二批命令中的“视图切换与状态分类”。
- Phase 3 对应第三批命令中的 scoped read operations。
- Phase 4 对应第四批命令中的 scoped write operations。
- Phase 5 负责行为稳定化、回归与发布准备。

### 2.4 系统分层与包职责

建议始终从“四层系统 + 七个同级包”的视角审视实现边界。

四层系统：

1. Git Core
   - Git 负责 object store、commit DAG、merge、rebase、push、pull、diff、restore。
2. Orbit Spec Layer
   - `.orbit/config.yaml`
   - `.orbit/orbits/*.yaml`
   - include / exclude / shared_scope / projection_visible / behavior
   - control scope / control loader
3. Projection Runtime
   - current orbit
   - user view scope resolution
   - projection paths
   - sparse-checkout projection
   - hidden-dirty gate
   - repo-local runtime state
4. Scoped Operations
   - `status / diff / log / commit / restore`

七个同级包：

- `commands`
  - 命令注册、flag 解析、stdout/stderr、`--json`
- `orbit`
  - 配置加载、control loader、校验、匹配、user view 解析
- `view`
  - `current / enter / leave / status / warnings`
- `scoped`
  - `files / diff / log / commit / restore`
- `git`
  - repo 发现、tracked files、porcelain、sparse-checkout、pathspec、refs
- `state`
  - `.git/orbit/state/` 的读写、projection cache、snapshot、lock
- `ids`
  - `orbit-id` 与 repo-relative path 安全校验

任何阶段都应坚持：

- 命令层保持薄。
- 不从命令层直接写 sparse-checkout。
- 不从命令层直接写 `.git/orbit/state/*`。
- 不把 cache、projection 或辅助 refs 当作真相源。

### 2.5 两层 Scope 重构落地顺序

当实现切到两层 scope 模型时，建议按下面顺序推进，而不是直接从写命令开始改：

1. 先更新 PRD、技术架构、测试策略与开发计划，明确 control scope、user view scope 和 projection paths 的语义。
2. 再落地 control loader，让 CLI 在任何 sparse view 下都能读取完整控制平面。
3. 再落地 user view resolver / projection resolver，让 `files / enter / status / diff / log` 统一切到新模型。
4. 最后改 `commit / restore` 的 scoped write 边界，补齐 restore birth-boundary preflight 与 `--allow-delete-current-orbit`。

原因：

- 这是一次边界重构，不只是匹配规则微调。
- 若不先稳定 control loader，`enter` 与跨 orbit 切换会依赖当前工作树可见性，行为不可靠。
- 若不先收敛 user view scope，`commit / restore` 仍可能把控制平面路径带进 scoped write set。

### 2.6 分支与 PR 策略

遵循 `CONTRIBUTING.md`，从 `main` 拉分支开发，通过 PR 合并回 `main`。

建议：

- 一个阶段对应一个主分支，一个主 PR。
- 若阶段过大，可在主分支下继续拆小 PR，但不要跨阶段混合。
- 文档类调整单独使用 `docs/<topic>`，功能实现使用 `feature/<topic>`，整理和硬化使用 `chore/<topic>`。

## 3. 阶段计划

### 3.1 Phase 0: 基础骨架与运行边界

#### 目标

搭好 Orbit MVP 的命令骨架、包边界和基础仓库能力，确保后续开发都落在正确结构上。

#### 建议分支名

- `feature/mvp-foundation`

#### 开发内容

- 建立并确认 command-centric 目录边界：
  - `commands`
  - `orbit`
  - `view`
  - `scoped`
  - `git`
  - `state`
  - `ids`
- 接通 Cobra root command 与各子命令注册。
- 完成 Git repo 发现能力：
  - repo root
  - absolute git dir
- 建立 repo-local state 根目录与 lock 机制。
- 封装原子写能力，用于 JSON 与 scope cache。
- 落地 `orbit-id` 校验与 repo-relative path 规范化帮助函数。

#### 需要注意的细节

- 命令层只能做参数解析、输出和错误转译，不能直接操作 sparse-checkout 或状态文件。
- 任何 Git 调用都使用显式参数列表，不使用 `sh -c`。
- 所有 repo-relative path 都以 repo root 为基准，不依赖当前 cwd。
- 状态写入必须使用原子替换，避免半写入 JSON。
- `enter`、`leave`、`commit`、`restore` 后续都要共用同一把 repo-local lock，这一层不要临时设计两套机制。

#### 交付物

- 可编译的 CLI 基础骨架
- `git` / `state` / `ids` 基础能力
- 统一错误模型和基础输出约定

#### 质量保证

- Unit tests:
  - `ids` 校验
  - path 规范化
  - state 原子写
  - lock 基本语义
  - repo root / git dir 解析
- 集成验证：
  - 非 Git 仓库 fail-closed
  - 子目录执行时仍能正确定位 repo root

### 3.2 Phase 1: 配置模型、初始化与校验

#### 目标

完成 `.orbit/` 配置模型、control loader、`init` / `add` / `validate` / `list` / `show` / `files` / `current` 的基础能力，并稳定 control scope 与 user view resolver 输入输出。

#### 建议分支名

- `feature/mvp-config-resolver`

#### 开发内容

- 实现全局配置和 orbit 定义模型：
  - `.orbit/config.yaml`
  - `.orbit/orbits/*.yaml`
- 实现 `orbit init`
  - 初始化 `.orbit/`
  - 初始化 `.orbit/orbits/`
  - 初始化 `.git/orbit/state/`
  - 幂等，不覆盖已有配置
- 实现 `orbit add <orbit-id>`
  - 生成 orbit skeleton
  - 文件名与 `id` 强一致
  - 已存在时稳定失败
- 实现 `orbit validate`
  - 校验配置结构
  - 校验 `orbit-id`
  - 校验 include/exclude/shared_scope/projection_visible pattern
  - 拒绝 `shared_scope` / `projection_visible` 命中 control-plane 路径
  - 校验 behavior 配置值
  - 校验 scope 可解析
  - empty scope 输出 warning
- 实现 `orbit list` 和 `orbit show <orbit-id>`
  - 稳定排序
  - `show` 输出规范化结构，不回显原始文本
- 实现 control loader
  - 工作树可见时优先读工作树
  - 控制文件因 sparse view 不可见时回退到 `HEAD`
- 实现统一 user view resolver，供 `validate` 和 `files` 共用。
- 实现 `orbit files <orbit-id>` 及 `resolved_scope/<orbit-id>.txt` projection cache 回写。
  - 输出目标 orbit 的 user view scope / projection paths
  - 自动注入 `.orbit/orbits/<orbit-id>.yaml`
  - 排除 `.orbit/config.yaml` 与其它 orbit definition paths
- 实现 `orbit current` 的基础只读语义
  - 只从 `.git/orbit/state/current_orbit.json` 读取
  - state 缺失返回稳定 `none`
  - state 损坏返回错误
  - 先不依赖 enter/leave 完整闭环即可交付只读能力

#### 需要注意的细节

- control scope 的权威输入只能是 `.orbit/` 配置与 control loader，不能依赖 cache 或 state。
- user view scope 的权威输入只能是 control loader 结果 + `git ls-files -z` + resolver，不能依赖 cache。
- `files` 和 `validate` 必须共用完全相同的 control loader 与匹配语义，避免结构性分裂。
- user data scope 只基于 tracked files 解析，不能把 untracked files 算进来。
- `shared_scope` 只并入命中的业务 tracked paths，`projection_visible` 只影响 projection；二者都不能用来注入 `.orbit/config.yaml` 或 `.orbit/orbits/*.yaml`。
- `.orbit/config.yaml` 永远不进入 user view scope；当前 orbit definition path 由系统自动注入，其它 orbit definition paths 必须排除。
- empty scope 是 warning，不是 hard error；但 warning 文案要稳定。

#### 交付物

- 可用的配置目录与 orbit 定义 skeleton
- 统一的 control loading、匹配、解析、校验链路
- 稳定的 orbit 枚举、展示、user view scope / projection 输出
- 基础可用的 `current` 只读命令

#### 质量保证

- Unit tests:
  - config load / validate
  - duplicate id
  - filename 与 `id` 不一致
  - invalid pattern
  - `shared_scope` / `projection_visible` 命中 control-plane 路径时失败
  - control loader 在工作树 / `HEAD` 两种来源下读取一致
  - include / exclude / shared_scope / projection_visible 解析
  - user view scope / projection 排序、去重、cache 幂等
- CLI integration tests:
  - `init` 幂等
  - `add` 非法 id fail-closed
  - `validate` 多 orbit 稳定排序
  - `files` 输出当前 orbit definition path，且排除 `.orbit/config.yaml` 与其它 orbit definitions
  - `files` 在 cache 缺失或损坏时仍输出一致结果
  - `current` 在 state 缺失和损坏时输出稳定

### 3.3 Phase 2: Current Orbit、View Projection 与状态分类

#### 目标

完成单工作区 orbit 进入/退出流程、current orbit 运行态、control-backed projection 切换、hidden-dirty gate 和状态分类。

这一阶段是 MVP 的风险中心。这里决定 Orbit 是否真的能在单工作区下“安全切换而不静默隐藏脏改动”。

#### 建议分支名

- `feature/mvp-view-runtime`

#### 开发内容

- 补齐 `current_orbit.json` 生命周期
  - 由 `enter` 写入
  - 由 `leave` 清理
  - `current` 对 stale orbit 给 warning
- 实现 `orbit enter <orbit-id>`
  - 加锁
  - 通过 control loader 读取并 validate target orbit
  - resolve target user view scope / projection paths
  - 读取 `git status --porcelain=v1 -z -uall`
  - 计算 dirty tracked paths
  - 识别切换后会被隐藏的 dirty tracked paths
  - 覆盖 dirty control-plane tracked paths
  - 触发 hidden-dirty gate
  - 应用 `git sparse-checkout init --no-cone`
  - 应用 `git sparse-checkout set --stdin`
  - 写 current orbit state
- 实现 `orbit leave`
  - 加锁
  - `git sparse-checkout disable`
  - 清理 current orbit state
  - 保留 warning/status/scope cache
- 实现 `orbit status`
  - current orbit
  - 基于 current projection scope 的 in-scope changes
  - 基于 current projection scope 的 out-of-scope changes
  - hidden-dirty risk
  - safe-to-switch
  - commit warnings
- 实现 `warnings.json` 和 `last_status.json` snapshot。

推荐自底向上实现顺序：

1. `git/status.go`
2. `git/sparse_checkout.go`
3. `view/current.go`
4. `view/warnings.go`
5. `view/enter.go`
6. `view/leave.go`
7. `view/status.go`
8. `commands/enter.go` / `commands/leave.go` / `commands/status.go`

#### 需要注意的细节

- sparse-checkout 使用的是 resolved projection file list，不是 orbit glob。
- 必须使用 non-cone sparse-checkout。
- hidden-dirty gate 只针对会被隐藏的 dirty tracked paths；dirty control-plane tracked paths 也在其中；outside untracked files 只能 warning，不能作为 enter 阻塞条件。
- `leave` 不因 outside changes 阻止执行。
- status 分类时，untracked path 即使不在 tracked scope 中，也要列入 outside changes。
- control loader 必须允许从一个 orbit 直接切到另一个 orbit，即使 target definition 当前不在工作树视图里。

#### 交付物

- 可切换的 orbit 工作区视图
- 可观察的 current orbit 运行态
- 稳定的状态分类与 warning snapshot

#### 质量保证

- Unit tests:
  - porcelain 解析
  - hidden-dirty 计算
  - in-scope / out-of-scope 分类
  - current orbit state 读写与损坏处理
- CLI integration tests:
  - `enter` 成功应用 sparse-checkout
  - dirty `.orbit/config.yaml` 会被隐藏时阻止切换
  - dirty 当前 orbit definition path 会被隐藏时阻止切换
  - dirty tracked path 被隐藏时阻止切换
  - outside untracked path 只警告
  - 在一个 orbit 中直接 `enter` 另一个 orbit 时仍能成功读取 target definition
  - `leave` 恢复完整 tracked view
  - `leave` 在 current state 缺失时仍恢复完整 tracked view
  - `current` 在 enter / leave / stale / damaged 情况下输出稳定
  - `status` 同时展示 scope 内外变更

### 3.4 Phase 3: Scoped Read Operations

#### 目标

在 current orbit 基础上完成只读 scoped 操作，让用户可以在 orbit 内查看当前 user view scope / projection paths 的文件、差异和历史。

#### 建议分支名

- `feature/mvp-scoped-read`

#### 开发内容

- 完成 `orbit files <orbit-id>` 的 CLI 体验与输出收敛。
  - 明确输出的是 target user view scope / projection paths，不是 control scope
- 实现 `orbit diff`
  - 默认只看 current user view scope / projection paths
  - `--outside` 查看 current projection scope 外变更
  - 通过 Orbit pathspec helper 组装稳定路径列表
  - 命令支持 `--pathspec-from-file` 时可使用 temp file；不支持时回退到显式 `-- <path>...`
- 实现 `orbit log`
  - 基于 current user view scope / projection paths 实现 path-limited history
  - 命令支持 `--pathspec-from-file` 时可使用 temp file；不支持时回退到显式 `-- <path>...`
  - 支持 `--` 后 Git 参数透传
- 统一 scoped read 命令的错误与 `--json` 输出结构。

推荐自底向上实现顺序：

1. `git/pathspec.go`
2. `git/diff.go`
3. `git/log.go`
4. `scoped/diff.go`
5. `scoped/log.go`
6. `commands/diff.go` / `commands/log.go`

#### 需要注意的细节

- `diff` 和 `log` 都要求 current orbit 存在，不能从“只有一个 orbit”推断。
- `diff --outside` 应基于当前 repo 变更与 current projection scope 补集生成 pathspec，而不是重新定义第二套匹配规则。
- Git 参数透传必须保持原样，不要在 Orbit 层偷偷改写用户的 log 参数。
- pathspec 临时文件必须使用 NUL 分隔，以保证带空格路径安全。
- `--pathspec-from-file` 的支持必须按命令能力判断，不能按 Git 大版本统一假定。
- 若走显式 `-- <path>...` fallback，MVP 接受参数长度受平台限制这一已知边界；命令必须直接暴露明确错误，不能静默截断路径列表。

#### 交付物

- current orbit user view / projection 范围内的稳定 diff/log 能力
- 可复用的 pathspec temp file 工具链

#### 质量保证

- Unit tests:
  - diff pathspec 组装
  - log pathspec 组装
  - outside path 计算
- CLI integration tests:
  - `diff` 仅包含 current user view scope / projection paths
  - `diff --outside` 仅包含 current projection scope 外路径
  - `log` 只展示 path-limited history
  - `log -- --oneline -20` 参数透传正确
  - `diff` / `log` 的 `--json` 输出结构稳定
  - current orbit 缺失时 `diff` / `log` fail-closed
  - 带空格或前导 `-` 的路径在 `diff` / `log` 中仍安全工作

### 3.5 Phase 4: Scoped Write Operations

#### 目标

完成 MVP 中风险最高的写路径：`orbit commit` 与 `orbit restore`，确保 fail-closed 与普通 Git commit 语义成立。

这一阶段是 MVP 的安全中心。这里决定 Orbit 是否真的不会误提交 scope 外改动，也不会把 restore 做成隐式历史改写。

#### 建议分支名

- `feature/mvp-scoped-write`

#### 开发内容

- 实现 `orbit commit -m "<message>"`
  - 要求 current orbit 存在
  - 加锁
  - resolve current user view scope / projection paths
  - 读取并分类工作区状态
  - outside changes 生成 warning snapshot
  - current projection scope 与 in-scope untracked 内无改动时返回 `nothing to commit`
  - `git add -A --pathspec-from-file=... --pathspec-file-nul`
  - `git commit -m ... --pathspec-from-file=... --pathspec-file-nul`
  - 按配置追加 `Orbit: <orbit-id>` trailer
  - best-effort 更新 `refs/orbits/<orbit-id>/last-scoped`
- 实现 `orbit restore --to <rev>`
  - 要求 current orbit 存在
  - 加锁
  - resolve current user view scope / projection paths
  - 检查 target revision 中是否存在 current orbit definition path
  - definition 缺失时默认 fail-closed
  - `--allow-delete-current-orbit` 时允许继续
  - current projection scope 内存在未提交改动时默认 fail-closed
  - `git restore --source=<rev> --worktree --staged --pathspec-from-file=... --pathspec-file-nul`
  - 创建新的普通 Git commit
  - 写入 `Orbit:` 和 `Orbit-Restore-From:` trailer
  - best-effort 更新 `refs/orbits/<orbit-id>/last-restore`
  - 若显式允许删除 current orbit 且 definition 在 target revision 不存在，则 restore 后自动 leave

推荐自底向上实现顺序：

1. `git/add_commit.go`
2. `git/restore.go`
3. `git/refs.go`
4. `scoped/commit.go`
5. `scoped/restore.go`
6. `commands/commit.go` / `commands/restore.go`

#### 需要注意的细节

- `git commit --pathspec-from-file=...` 是 scoped commit 的安全边界，不能退化成“先 add，再普通 commit”。
- 必须显式防止 current projection scope 外 staged changes 被误提交。
- outside changes 可以 warning，但不能被 silently 带入 commit。
- `.orbit/config.yaml` 不得进入普通 scoped write set；当前 orbit definition path 仅作为 companion path 参与 current projection scope。
- `restore` 的语义是重新应用历史 current projection scope 内容并提交新 commit，不是改写历史。
- `restore` 必须有 birth-boundary preflight，避免默认把当前 orbit definition 恢复成删除状态后留下 stale current state。
- optional refs 失败时主命令不能回滚，但必须有 warning。

#### 交付物

- 可安全提交 orbit 范围内改动的命令
- 可安全恢复 orbit 到历史 revision 的命令
- 标准化 trailer 和 best-effort refs 能力

#### 质量保证

- Unit tests:
  - commit trailer 拼装
  - restore message 生成
  - ref 更新失败路径
- CLI integration tests:
  - `commit` 只提交 current projection scope 内 modified / deleted / untracked
  - `.orbit/config.yaml` 不会被普通 scoped commit 带入
  - scope 外 unstaged 改动保留
  - scope 外 staged 改动不被误提交
  - 提交消息追加 `Orbit: <orbit-id>`
  - `restore` 生成新的普通 commit
  - `restore` 默认在 current projection scope 内已有脏改动时 fail-closed
  - target revision 缺失 current orbit definition path 时默认失败
  - `--allow-delete-current-orbit` 时 restore 成功并自动 leave
  - `.orbit/config.yaml` 不会被普通 scoped restore 误恢复

### 3.6 Phase 5: 稳定性收口与发布准备

#### 目标

把 MVP 从“功能可运行”收口到“行为稳定、输出稳定、可合并可发布”的状态。

#### 建议分支名

- `chore/mvp-hardening`

#### 开发内容

- 审查所有命令的 stdout/stderr 文案与 `--json` 结构稳定性。
- 检查 warning snapshot、status snapshot、resolved projection cache 的一致性与幂等。
- 补齐 README、命令帮助、示例配置与用户使用路径。
- 补齐遗漏测试，清理重复逻辑和临时实现。
- 做一次基于 PRD Acceptance Criteria 的端到端验收演练。

#### 需要注意的细节

- 不要在收尾阶段引入超出 MVP 边界的新能力。
- `refs/orbits/*` 必须保持 optional，不能被任何主路径依赖。
- 输出层优先展示 orbit id、路径、数量、warning 摘要，不打印完整文件内容。
- 若发现实现与 PRD/架构文档冲突，应先修文档或停下来对齐，不要直接带着偏差收口。

#### 交付物

- 行为稳定的 MVP 命令集
- 完整测试矩阵
- 可用于 PR 合并和后续演示的文档与验收记录

#### 质量保证

- 运行统一校验：
  - `mise run fmt`
  - `mise run lint`
  - `mise run test:ci`
- 对照 `docs/context/mvp-product-requirements.md` 第 4、8、9、12 章逐条验收。
- 做一次 temp repo 场景回归：
  - `init`
  - `add`
  - `validate`
  - `enter`
  - `status`
  - `diff`
  - `log`
  - `commit`
  - `restore`
  - `leave`

## 4. 质量保证策略

### 4.1 测试分层

严格采用 `docs/testing-strategy.md` 的 MVP Test Pyramid：

- Layer A: Unit Tests
- Layer B: CLI Integration Tests
- Layer C: Deferred Tests

MVP 期间不优先投入 benchmark、压力测试和大规模并发专项，除非仓库规模和复杂度明显上升。

### 4.2 测试推进顺序

测试不应等到所有命令完成后再补，而应随阶段推进同步收敛。

推荐顺序：

1. 先补 control loader、配置校验、`orbit-id` 和 path 安全，以及 control-plane 路径约束。
2. 再补 user view / projection 解析、`enter` hidden-dirty gate、`leave` 恢复完整视图、`current` 生命周期。
3. 再补 `status` 基于 current projection scope 的分类、warning snapshot、last status snapshot。
4. 再补 `diff / log` pathspec 正确性、参数透传、`--json` 输出与显式参数 fallback 安全性。
5. 最后补 `commit` 不误提交 current projection scope 外改动、`restore` birth-boundary preflight、`--allow-delete-current-orbit`、optional refs best-effort、state 原子写与锁恢复。

每个阶段结束时，至少要有覆盖本阶段主风险的 acceptance tests。

### 4.3 每阶段测试要求

- 新增逻辑默认配套 unit tests。
- 涉及真实 Git 状态、sparse-checkout、pathspec、commit、restore 的能力，必须配套 temp repo integration tests。
- 默认使用 `t.Parallel()`；只有修改 cwd、env 或共享资源时才禁用。
- 测试断言优先验证稳定字段和关键 warning，不依赖无关空白或输出布局细节。

### 4.4 合并前检查

每个阶段 PR 合并前至少满足：

1. `mise run fmt`
2. `mise run lint`
3. `mise run test:ci`
4. 相关命令有对应 unit/integration tests
5. PR 描述包含变更目的、主要修改点、本地验证方式、是否影响命令/文档/状态文件

## 5. 风险与控制点

### 5.1 高风险区域

- sparse-checkout 切换前 hidden-dirty 检查
- control scope / user view scope / projection cache 语义漂移
- 控制平面路径泄漏进 scoped write set
- current projection scope 外 staged changes 被误提交
- damaged state 被静默吞掉
- `.orbit/`、`.git/orbit/state/`、`refs/orbits/*` 三类状态边界混淆

### 5.2 控制策略

- 所有核心命令复用统一 control loader、user view resolver、porcelain parser、state store、pathspec helper。
- 对写操作全部加锁，并坚持原子写。
- 对 `current`、`status`、`commit`、`restore` 保持 fail-closed。
- 不允许命令层直接改 sparse-checkout 和 state 文件。
- 每一阶段结束后做一次针对本阶段命令集的 temp repo 回归。

### 5.3 必守主线

整个实现过程中，持续检查以下四条主线是否被破坏：

1. 统一 control loader 与 user view resolver
   - `files / enter / status / commit / restore` 不能各算各的 user view / projection scope。
2. 命令层保持薄
   - Git 细节与 state 持久化不能从领域包泄漏到命令层。
3. 真相源边界清楚
   - Git DAG、`.orbit/`、`.git/orbit/state/` 必须各司其职。
4. 单工作区风险 fail-closed
   - 尤其是 hidden-dirty gate 与 scoped commit safety。

## 6. 建议里程碑

### Milestone A

- 完成 Phase 0 + Phase 1
- 结果：可以初始化、定义、校验、列出和解析 orbit

### Milestone B

- 完成 Phase 2 + Phase 3
- 结果：可以进入/退出 orbit，并在 orbit 内查看状态、diff、log

### Milestone C

- 完成 Phase 4 + Phase 5
- 结果：可以安全 commit/restore，并通过 MVP 验收与质量门槛

## 7. 建议执行节奏

如果团队希望降低风险，建议采用以下 PR 顺序：

1. `feature/mvp-foundation`
2. `feature/mvp-config-resolver`
3. `feature/mvp-view-runtime`
4. `feature/mvp-scoped-read`
5. `feature/mvp-scoped-write`
6. `chore/mvp-hardening`
7. `docs/mvp-usage-and-release-notes`

其中前五个 PR 对应核心实现，第六个 PR 做收口与回归，第七个 PR 可选，用于补充面向使用者的文档整理。

## 8. 完成定义

Orbit Git MVP 可视为完成，当且仅当同时满足：

- PRD 第 4 章 In Scope 能力均已落地
- PRD 第 9 章命令语义均已实现
- PRD 第 12 章 Acceptance Criteria 可在真实 temp repo 中复现
- `docs/testing-strategy.md` 的 Minimum Coverage Matrix 已有对应测试
- `mise run fmt`、`mise run lint`、`mise run test:ci` 全部通过
- 没有引入任何超出 MVP 边界的架构负担
