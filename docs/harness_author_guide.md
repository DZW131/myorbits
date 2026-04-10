# Harness Author Guide

版本：v0.4
状态：harness-author-facing usage guide
适用范围：runtime 组合、共享变量绑定、宏观编排、harness template 导出

关联文档：

- `docs/orbit_positioning_and_personas.md`
- `docs/quickstart.md`
- `docs/orbit_author_guide.md`
- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_brief_lane_v0_4_technical_spec.md`

---

## 1. 这份文档给谁看

这份文档给“整个 harness 工程的开发者”看。

你的关注点不是单个 orbit 的局部体验，而是：

- 这个 runtime 里要装哪些 orbit
- 它们怎样组合
- 共享变量怎样统一收敛
- 根 `AGENTS.md` 怎样表达宏观编排
- 如何检查整个 runtime
- 如何把成熟的 runtime 导出成 harness template

---

## 2. Harness 作者真正拥有的内容

Harness 作者负责的，是**整套工作环境的外部控制**。

主要包括：

- runtime 的创建与安装
- 多个 orbit 的选择与组合
- 共享 `.harness/vars.yaml` 的维护
- 根 `AGENTS.md` 的宏观任务编排
- runtime 一致性检查
- harness template 导出

Harness 作者不需要替每个 orbit 写完所有局部规则；那是 orbit 作者的职责。

---

## 3. 根 `AGENTS.md` 应该写什么

根 `AGENTS.md` 更适合承载宏观内容：

- 整体目标
- orbit 地图
- 哪些任务应该进入哪个 orbit
- orbit 之间的切换规则
- 全局安全规则
- 全局结果与审计要求
- 升级、求助、停止条件

它**不适合**承担：

- 每个 orbit 的完整细节说明
- 每个 orbit 的全部 rules / process 原文
- 过长的局部执行材料

局部细节应留给 orbit brief 与 orbit 自身的 rule / process 文件。

---

## 4. Harness 作者的主场景

Harness 作者最典型的场景是：

- 根据当前任务，选择多个合适的 orbit template
- 把它们安装进同一个 runtime
- 统一变量绑定
- 完成根 `AGENTS.md` 的宏观编排
- 检查 worker 进入时的真实体验

这条主路径里，其实有 3 条不同 lane：

- `install lane`
  - 把 orbit template 安装进 runtime
- `bindings lane`
  - 统一收敛变量，形成共享 `.harness/vars.yaml`
- `orchestration lane`
  - 把 orbit brief 放进根 `AGENTS.md`，再补上 harness 级宏观编排

不要把这 3 条 lane 混成一个黑箱动作。

---

## 5. 当前推荐的最优搭建流程

## 5.1 创建或初始化 runtime

新建 runtime：

```bash
harness create demo-repo
```

在已有 Git 仓库中初始化：

```bash
harness init
```

这一步会建立 runtime 控制面的正式宿主。

## 5.2 先规划多个 orbit 的共享变量绑定

如果你要安装多个 orbit，最稳的做法不是直接一个个装，而是先把变量入口规划出来。

先为每个候选 orbit template 生成 bindings skeleton：

```bash
orbit bindings init <template-a> > /tmp/template-a.vars.yaml
orbit bindings init <template-b> > /tmp/template-b.vars.yaml
```

然后把这些 skeleton 合并成**一个共享的** `.harness/vars.yaml`。

推荐把这一步视为 harness 作者的正式动作，而不是 install 过程里的附属细节。

原因是：

- `.harness/vars.yaml` 是 runtime 级共享宿主
- 后续多个 install 会复用这里已有的值
- 根 `AGENTS.md` brief materialize 也会读这里的值
- 后续 `harness template save` 也会基于这里做变量替换

如果你每个模板都直接单独 `--out .harness/vars.yaml`，很容易互相覆盖。

所以当前仓库下更推荐：

- 先输出到临时文件或 stdout
- 再合并成一份 `.harness/vars.yaml`
- 再把这份共享 vars 作为后续 install 的统一输入

## 5.3 先 dry-run，再正式安装多个 orbit

对每个模板，先做 dry-run：

```bash
harness install <template-a> --bindings .harness/vars.yaml --dry-run
harness install <template-b> --bindings .harness/vars.yaml --dry-run
```

这一步的价值不是“走个形式”，而是提前确认：

- 当前 bindings 是否足够
- 会写入哪些文件
- 是否有冲突
- 是否需要 `--overwrite-existing`

如果有缺失变量，你有 3 种补法：

- 直接完善 `.harness/vars.yaml`
- 安装时加 `--interactive`
- 安装时加 `--editor`

当 preview 没问题后，再正式安装：

```bash
harness install <template-a> --bindings .harness/vars.yaml
harness install <template-b> --bindings .harness/vars.yaml
```

安装完成后，当前 runtime 会拥有：

- `.harness/manifest.yaml`
- `.harness/orbits/*.yaml`
- `.harness/installs/*.yaml`
- `.harness/vars.yaml`
- 对应模板渲染出的 runtime 文件

## 5.4 物化 orbit brief，再编排根 `AGENTS.md`

安装完成后，不要立刻把根 `AGENTS.md` 当成手写大文档。

更推荐先把每个 orbit 的 brief block 物化出来：

```bash
orbit brief materialize --orbit <orbit-a>
orbit brief materialize --orbit <orbit-b>
```

然后再编辑根 `AGENTS.md` 的 harness 级宏观内容。

根 `AGENTS.md` 更适合写：

- 整体任务目标
- orbit 地图
- 哪些情况进入哪个 orbit
- orbit 之间的切换规则
- 全局安全规则
- 全局升级、求助、停止条件
- 全局结果与审计要求

它不应该承担：

- 每个 orbit 的完整局部说明
- 每个 orbit 的全部 rules / process 原文
- 过长的局部执行材料

如果你手工改了某个 orbit block，并且希望把这份修改沉淀回结构化 truth，再显式执行：

```bash
orbit brief backfill --orbit <orbit-id>
```

如果你只是在根 `AGENTS.md` 里补 harness 级宏观编排，不需要 backfill。

## 5.5 检查 runtime 与 worker 入口

```bash
harness inspect
harness check --json
```

Harness 作者不应只看安装是否成功，还应从 worker 视角验证实际体验：

```bash
orbit list
orbit enter <orbit-id>
orbit current
orbit status
orbit leave
```

如果 worker 进入后仍然需要大量口头解释，说明 harness 或 orbit 还没有收口好。

## 5.6 提交 runtime 基线

Orbit projection 只对 tracked files 工作，所以运行态搭好后建议先提交一次正常 Git commit：

```bash
git add -A
git commit -m "compose harness runtime"
```

## 5.7 导出 harness template

当 runtime 组合已经稳定时：

```bash
harness template save --to harness-template/workspace
```

---

## 6. 当前仓库里，变量替换和 AGENTS 编排究竟是怎样工作的

### 6.1 变量替换

当前变量替换的正式宿主是：

```text
.harness/vars.yaml
```

它不是一个临时文件，而是 runtime 级共享 provenance。

当前仓库里，至少下面几条路径会消费它：

- `harness install`
- `orbit brief materialize`
- `harness template save`

这意味着，最优流程不是“装完再集中替换一次变量”，而是：

- 先收敛共享 vars
- 再让 install / materialize / save 自动消费它

### 6.2 根 `AGENTS.md`

根 `AGENTS.md` 在 v0.4 里是：

- orchestration artifact
- runtime 容器文件
- worker 默认读取的入口

它不是：

- orbit authored truth
- 任意单 orbit 的第一性真相源
- 通用回写宿主

对单 orbit 来说：

- 结构化 truth 在 `.harness/orbits/<orbit-id>.yaml -> meta.agents_template`
- `orbit brief materialize` 是 `truth -> container block`
- `orbit brief backfill` 是 `container block -> truth`

所以 harness 作者在编排根 `AGENTS.md` 时，应该把它当成“宏观入口 + orbit brief block 容器”，而不是把所有 authored 内容都塞进去。

---

## 7. AI / Agent 最适合介入的地方

最值得让 AI 帮忙的，不是直接替你盲装模板，而是下面两件事：

### 7.1 帮你合并和补全 `.harness/vars.yaml`

这是最适合 AI 介入的结构化入口，因为：

- 输入稳定
- 输出结构稳定
- 变量有 description
- 多个 orbit 之间容易出现共享值

所以更推荐：

- 先用 `orbit bindings init` 生成多个 skeleton
- 再让 AI 帮你合并成一份共享 `.harness/vars.yaml`
- 最后由你校正少量关键值

### 7.2 帮你起草根 `AGENTS.md` 的宏观编排

AI 适合帮你生成：

- 整体目标
- orbit 地图
- 进入条件
- 切换条件
- 升级条件
- 全局记录要求

AI 不适合在这个场景里负责：

- 猜测 orbit 的文件边界
- 替代 orbit 作者写完整局部规则
- 直接把根 `AGENTS.md` 当 authored truth 反向替换一切

---

## 8. 当前最值得补的 3 个薄封装

当前仓库已经有足够的 primitive，但对 harness 作者来说，最值得补的是更薄的高意图命令：

- `harness bindings plan <template...>`
  - 一次扫描多个 orbit template，产出共享 `.harness/vars.yaml` skeleton 与缺失报告
- `harness install batch <template...>`
  - 基于同一份 `.harness/vars.yaml` 做统一 preview 和批量安装
- `harness agents compose`
  - 把 harness 级宏观模板和已安装 orbit brief blocks 组合成当前根 `AGENTS.md`

这 3 个方向都很轻，因为它们本质上只是对现有 primitive 的组合，不会把 Orbit 变成重平台。

---

## 9. Harness 作者与 Orbit 作者的边界

Harness 作者负责：

- 宏观编排
- 安装与组合
- 共享 vars 的维护
- runtime 级检查
- 整体模板导出

Orbit 作者负责：

- 单 orbit 的边界
- 单 orbit 的 brief / rules / process
- 单 orbit 的 done probe
- 单 orbit 的 record target

如果这两层混在一起，系统会很快变重。

---

## 10. 运行态优化后的两条演化路径

## 10.1 只优化某个 orbit

如果你在 runtime 里主要优化的是单个 orbit：

```bash
orbit template save <orbit-id>
```

这条路适合把局部优化写回 orbit template。

## 10.2 优化的是整个工作环境

如果你优化的是：

- orbit 组合
- 宏观编排
- runtime 级体验

那么更适合：

```bash
harness template save --to harness-template/workspace
```

---

## 11. 动态编排的未来边界

后续如果要做动态编排，Harness 作者应优先考虑：

- 如何定义切换条件
- 如何定义升级条件
- 如何定义全局控制信号
- 如何留下显式审计记录

而不是先做：

- 自动调度器
- 持续后台服务
- 内部推理系统

更稳的方向是：先把“什么时候切换、切到哪里、为什么切”外显出来。

---

## 12. 下一步

如果你要设计单个 orbit：

- 看 [docs/orbit_author_guide.md](./orbit_author_guide.md)

如果你只是进入现有 runtime 执行任务：

- 看 [docs/worker_guide.md](./worker_guide.md)
