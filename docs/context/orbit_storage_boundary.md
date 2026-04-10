# Orbit Storage Boundary

版本：V0.2
状态：作为 Orbit V0.2 的边界基线
适用范围：Orbit MVP 与 V0.2 Phase 2 设计
关联文档：

- `docs/context/mvp-product-requirements.md`
- `docs/context/mvp-technical-architecture.md`
- `docs/context/two-scope-refactor.md`
- `docs/context/orbit_phase2_prd.md`

---

## 1. 文档目的

本文件用于把 Orbit 当前仓库里的几类状态边界正式写清楚，尤其是：

- `.orbit/`
- `.git/orbit/state/`
- Git DAG
- sparse-checkout projection
- `refs/orbits/*`

目标不是重复 PRD，而是避免后续阶段把配置、运行态、缓存和历史混写。

---

## 2. 一句话原则

`.orbit/` 放版本化配置与版本化元数据；`.git/orbit/state/` 放 repo-local 运行态与缓存；Git 历史仍然只以 Git DAG 为准。  
任何时候都不允许把这三者混成同一个“状态层”。

---

## 3. 五类状态的角色

### 3.1 Git DAG

Git commit DAG 是 canonical history。

它负责：

- 已提交文件内容
- 历史 revision
- 普通 commit / restore 结果
- 模板 branch 内容

它不负责：

- 当前本地工作区进入了哪个 orbit
- 最近一次 warning/status 快照
- projection cache

### 3.2 `.orbit/`

`.orbit/` 是版本化配置层，也是后续阶段允许承载版本化 Orbit 元数据的地方。

它负责：

- 全局 Orbit 配置
- orbit definitions
- 后续可版本化共享的模板/bindings/install metadata

它不负责：

- 当前本地工作区运行态
- 调试 cache
- 最近一次命令执行快照

### 3.3 `.git/orbit/state/`

`.git/orbit/state/` 是 repo-local runtime state。

它负责：

- 当前工作区是否处于某个 orbit
- 最近一次 projection cache
- 最近一次 warning/status snapshot
- repo-local lock

它不负责：

- 作为历史来源
- 作为跨分支共享元数据
- 作为跨机器同步机制
- 作为模板安装记录的权威来源

### 3.4 sparse-checkout projection

sparse-checkout 只是当前 working tree 的投影视图。

它负责：

- 当前工作区“看见哪些 tracked paths”

它不负责：

- 定义 orbit 本身
- 决定当前 orbit 是谁
- 替代 `.orbit/` 或 `.git/orbit/state/`

### 3.5 `refs/orbits/*`

`refs/orbits/*` 只是 best-effort auxiliary refs。

它可以用于：

- 记录最近一次 scoped commit / restore 的辅助引用

它不能用于：

- 保证命令正确性
- 成为恢复或分类的唯一依据

---

## 4. `.orbit/` 中应该放什么

当前 MVP 已正式定义：

```text
.orbit/
  config.yaml
  orbits/
    <orbit-id>.yaml
```

### 4.1 当前已稳定的文件

`.orbit/config.yaml`

- 全局行为配置
- 例如 `shared_scope`、`projection_visible`、`behavior`
- 属于 control plane
- 不进入任何 orbit 的默认 user view

`.orbit/orbits/<orbit-id>.yaml`

- 单个 orbit 定义
- 文件名与 YAML 中的 `id` 必须一致
- 当前 orbit 自己的 definition path 会作为 companion path 自动进入 projection

### 4.2 Phase 2 建议新增的版本化文件

本仓库若进入 Phase 2，建议只在 `.orbit/` 下新增“版本化、可共享”的文件：

```text
.orbit/
  vars.yaml
  installs/
    <orbit-id>.yaml
```

推荐含义：

`.orbit/vars.yaml`

- 仓库级变量 bindings
- 进入版本控制
- 保存模板应用/复用所需的变量值和可选描述

`.orbit/installs/<orbit-id>.yaml`

- 记录当前仓库或当前分支上，该 orbit 是从哪个模板源安装来的
- 属于版本化元数据，不属于 repo-local runtime state

### 4.3 模板 branch 中的额外文件

对于模板 branch，建议允许存在：

```text
.orbit/
  template.yaml
```

它只用于模板分支中的模板 manifest，不应被当前普通 runtime branch 默认依赖。

---

## 5. `.git/orbit/state/` 中应该放什么

当前 MVP 已正式定义并已有实现的运行态文件包括：

```text
.git/orbit/state/
  current_orbit.json
  resolved_scope/
    <orbit-id>.txt
  warnings.json
  last_status.json
  orbit.lock
```

### 5.1 已稳定的运行态文件

`current_orbit.json`

- 当前本地工作区进入的是哪个 orbit
- 进入时间
- 是否处于 sparse projection

`resolved_scope/<orbit-id>.txt`

- 最近一次 projection scope 的稳定排序 cache
- 当前代码里已经转向 projection-cache 术语
- 但磁盘目录名为了兼容仍保留 `resolved_scope/`

`warnings.json`

- 最近一次 `enter` / `status` / `commit` 等命令的 warning 摘要

`last_status.json`

- 最近一次状态分类快照

`orbit.lock`

- 保护 `enter` / `leave` / `commit` / `restore` 等命令间的串行化

### 5.2 不应该放入这里的内容

以下内容不应进入 `.git/orbit/state/`：

- `.orbit/config.yaml` 的镜像副本
- orbit definitions 的镜像副本
- 模板 branch manifest
- 可共享的 bindings
- 跨分支可见的安装记录
- 任何必须进入 Git 历史才有意义的元数据

---

## 6. 设计判断表

| 问题 | 应放哪里 | 原因 |
| --- | --- | --- |
| 全局 Orbit 行为配置 | `.orbit/config.yaml` | 版本化配置 |
| orbit definition | `.orbit/orbits/*.yaml` | 版本化配置 |
| 当前工作区已进入哪个 orbit | `.git/orbit/state/current_orbit.json` | repo-local runtime state |
| 最近一次 projection paths | `.git/orbit/state/resolved_scope/*.txt` | 可重建 cache |
| 最近一次 warning/status | `.git/orbit/state/*.json` | repo-local snapshot |
| 模板变量 bindings | `.orbit/vars.yaml` | 需要共享与复用 |
| 模板安装来源记录 | `.orbit/installs/*.yaml` | 需要跨分支/跨协作保留 |
| 最近一次 scoped commit 辅助 ref | `refs/orbits/*` | best-effort auxiliary ref |

---

## 7. 写入责任边界

### 7.1 command 层

`cmd/orbit/cli/commands`

- 只负责 CLI 参数、stdout/stderr、`--json`
- 不直接写 `.git/orbit/state/*`
- 不直接操作 sparse-checkout 细节

### 7.2 orbit 层

`cmd/orbit/cli/orbit`

- 负责读取 `.orbit/` 配置与定义
- 负责 scope resolution
- 后续可扩展到读取 `.orbit/vars.yaml` 与 `.orbit/installs/*.yaml`

### 7.3 state 层

`cmd/orbit/cli/state`

- 只负责 `.git/orbit/state/` 的读写
- 负责原子写、lock、snapshot、projection cache

### 7.4 git 层

`cmd/orbit/cli/git`

- 只负责 Git 访问适配
- 不直接决定哪些内容属于配置、运行态或模板语义

---

## 8. 反模式

以下做法都应视为设计错误：

1. 把 `.orbit/` 当 cache 层使用。
2. 把 `.git/orbit/state/` 当跨分支共享元数据层使用。
3. 把 sparse-checkout 可见性当成 source of truth。
4. 让 `refs/orbits/*` 成为命令正确性的必需前提。
5. 在 `.orbit/` 下引入一个泛化的 `state.yaml`，把配置、安装记录和本地运行态混在一起。

---

## 9. 关于当前仓库里出现的额外文件

如果当前本地仓库里出现未被本文件与 MVP 文档正式定义的 `.git/orbit/state/*` 文件，例如实验性 manifest 文件，应按以下原则处理：

- 它们不自动成为正式 contract
- 当前命令正确性不能依赖它们
- 若后续要正式引入，必须先更新技术文档与测试矩阵

---

## 10. 对 Phase 2 的直接约束

进入 Phase 2 时，以下边界应继续成立：

1. 不引入 `.orbit/state.yaml` 作为泛化状态文件。
2. `template apply` 不因为需要记录安装来源，就把版本化元数据写入 `.git/orbit/state/`。
3. `template save` / `template apply` 不改变 `enter` / `leave` 作为 projection 控制入口的既有职责。
4. 模板相关的共享信息优先落在 `.orbit/`，本地执行快照仍留在 `.git/orbit/state/`。
