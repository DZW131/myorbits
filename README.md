# Orbit

Orbit 是一个 Git-native 的工作环境控制平面。

它的目标不是替 worker 思考，而是把 worker 执行任务时真正稳定有价值的外部控制信息外显出来：

- 当前在哪个 orbit
- 当前看见什么
- 当前允许改什么
- 当前规则、探针、记录目标是什么
- 当前内容如何导出、发布、复用

如果再压缩一层，Orbit 的本质是：

- 把一类常见工作的执行合同外部化
- 把它变成可进入、可退出、可记录、可复用的最小工作模块
- 让这类工作在 Git 仓库中获得稳定的文件边界和低负担执行入口

当前仓库提供两个 CLI：

- `orbit`
  - 面向单个 orbit 的定义、projection、scoped operations、brief 与 template authoring
- `harness`
  - 面向整个 runtime 的创建、安装、检查、组合与 harness template 导出

Git 仍然负责：

- commit
- branch
- merge
- rebase
- push / pull

Orbit / Harness 不替代 Git，也不引入第二套历史系统。

## Orbit 服务谁

- `worker`
  - 在现成 runtime 中执行任务
- `orbit 作者`
  - 设计和迭代单个 orbit template
- `harness 作者`
  - 组合多个 orbit 并演化整个工作环境

## Orbit 的 3 个主场景

1. `worker 执行`
   - 进入一个 orbit，围绕当前工作单元执行任务
2. `orbit 作者迭代`
   - 设计、优化、导出或发布单个 orbit
3. `harness 作者编排`
   - 安装 orbit、组合 runtime、导出 harness template

## 终极形态下，Orbit 只做 6 件事

1. 支撑最小工作模块骨架
2. 界定文件边界
3. 物化执行入口
4. 统一退出语义
5. 支撑模板物化与复用
6. 记录控制面元信息

完整说明见：
[docs/orbit_positioning_and_personas.md](./docs/orbit_positioning_and_personas.md)

## 当前边界

当前 v0.4 主线保持下面这些硬边界：

- 单仓库
- 单项目
- 单工作区
- 单 branch 背景
- `projection` 基于 `git sparse-checkout`
- scoped operations 基于 Git pathspec
- 版本化控制平面进入 `.harness/`
- repo-local 运行态进入 `.git/orbit/state/`

正式宿主分工：

```text
.harness/manifest.yaml
.harness/orbits/<orbit-id>.yaml
.harness/vars.yaml
.harness/installs/*.yaml
.harness/bundles/*.yaml
.git/orbit/state/*
AGENTS.md                 # orchestration artifact
```

## Orbit 负责什么

- 外部环境控制
- orbit / harness authored truth
- projection / orbit_write / export / orchestration 四个 surface
- brief materialize / backfill
- template / writeback / publish lanes
- provenance、留痕与最小审计合同

## Orbit 不负责什么

- agent 内部推理
- session 总结方法
- chain-of-thought
- 后台守护进程
- 远程服务
- 数据库式 canonical state
- block-level / semantic orbit

## 从哪里开始

如果你想先理解定位和边界：

1. [docs/orbit_positioning_and_personas.md](./docs/orbit_positioning_and_personas.md)

如果你要执行任务：

1. [docs/worker_guide.md](./docs/worker_guide.md)
2. [docs/quickstart.md](./docs/quickstart.md)

如果你要设计或迭代单个 orbit：

1. [docs/orbit_author_guide.md](./docs/orbit_author_guide.md)
2. [docs/orbit_template_authoring_guide.md](./docs/orbit_template_authoring_guide.md)
3. [docs/quickstart.md](./docs/quickstart.md)

如果你要搭建整个 harness：

1. [docs/harness_author_guide.md](./docs/harness_author_guide.md)
2. [docs/quickstart.md](./docs/quickstart.md)

如果你要看正式 v0.4 source of truth：

1. [docs/orbit_v0_4_prd.md](./docs/orbit_v0_4_prd.md)
2. [docs/orbit_v0_4_technical_spec.md](./docs/orbit_v0_4_technical_spec.md)
3. [docs/orbit_v0_4_development_plan.md](./docs/orbit_v0_4_development_plan.md)
4. [docs/testing-strategy.md](./docs/testing-strategy.md)

## 常见主路径

执行任务：

```bash
harness inspect
orbit list
orbit enter <orbit-id>
orbit current
orbit status
```

安装与检查 runtime：

```bash
harness create demo-repo
harness install <template-source> --bindings .harness/vars.yaml
harness check --json
```

导出单个 orbit：

```bash
orbit template save <orbit-id> --to orbit-template/<orbit-id>
```

导出整个 harness：

```bash
harness template save --to harness-template/workspace
```

## Development Rules

- 命令层保持薄
- Git 交互统一下沉到 `cmd/orbit/cli/git`
- repo-local 状态统一下沉到 `cmd/orbit/cli/state`
- 不使用 worktree
- 不引入数据库、远程服务、后台守护进程

## Validation

日常校验：

```bash
mise run fmt
mise run lint
mise run test:ci
```

本地构建双二进制：

```bash
mise run build
./.dist/bin/orbit --help
./.dist/bin/harness --help
```

如需生成 shell completion：

```bash
./.dist/bin/orbit completion zsh > _orbit
./.dist/bin/harness completion zsh > _harness
```

## Release

GoReleaser 负责双二进制发布：

```bash
goreleaser check
goreleaser release --snapshot --clean
```

正式发布流程见：[docs/release.md](./docs/release.md)
