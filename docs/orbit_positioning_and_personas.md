# Orbit Positioning And Personas

版本：v0.4
状态：positioning / experience guide
适用范围：v0.4 unified control plane 的上层定位与用户分层

关联文档：

- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_v0_4_development_plan.md`
- `docs/orbit_brief_lane_v0_4_technical_spec.md`
- `docs/orbit_template_authoring_guide.md`
- `docs/quickstart.md`

---

## 1. 文档目标

这份文档不定义新的状态宿主、命令合同或实现细节。

它只回答 4 个问题：

1. Orbit 到底是给谁用的；
2. Orbit 应该解决什么，不该解决什么；
3. 不同用户在什么场景下应当看到什么体验；
4. 怎样让 Orbit 保持轻量，同时把价值发挥到最大。

若本文与 v0.4 PRD、technical spec、testing strategy 冲突，以正式主文档为准。

---

## 2. 一句话定义

**Orbit 是一个 Git-native 的工作环境控制平面。它负责把 worker 的任务环境、作用域、规则、探针、记录目标、留痕与模板复用方式外部化，让人或 agent 可以在更低认知负担下执行任务；但它不负责 worker 的内部认知、推理、总结方法或实现路径。**

## 2.1 Orbit 的本质

Orbit 的本质不是“任务本身”，也不是“agent 的思考引擎”。

它更准确的定位是：

- 一类常见工作的**可复用执行合同**
- 一个最小工作模块的**外部控制骨架**
- 一个把“这一类工作该如何进入、执行、退出、记录、复用”外显出来的控制平面

因此：

- 复杂任务可以由多个 orbit 组成
- 一个 orbit 也可能不是复杂任务里的子任务，而只是一个被广泛复用的工作环节
- Orbit 的价值，不在于替 worker 思考，而在于让某一类高维相似工作的执行效率、可靠性和可审计性持续提升

## 2.2 Orbit 服务谁

Orbit 终极上服务 3 类用户：

1. `worker`
   - 在现成环境中执行任务的人或 agent
2. `orbit 作者`
   - 设计和迭代单个 orbit template 的人
3. `harness 作者`
   - 组合多个 orbit、搭建整套工作环境的人

这三类用户共享同一套内核，但不应承受同样的认知负担。

## 2.3 Orbit 有几个主场景

Orbit 最核心的主场景也是 3 个：

1. `worker 执行`
   - 进入一个 orbit，围绕当前工作单元执行任务
2. `orbit 作者迭代`
   - 设计和优化一个可复用的工作模块，再导出或发布
3. `harness 作者编排`
   - 把多个 orbit 组合成完整 runtime，并导出为 harness template

---

## 3. 主场景下，用户做什么，Orbit 背后做什么

| 用户         | 场景               | 用户动作                                   | 常见命令                                                                                               | Orbit 背后做什么 / 达成什么效果                                                                                               |
| ---------- | ---------------- | -------------------------------------- | -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| worker     | 进入并执行一个工作单元      | 查看环境，进入 orbit，确认当前边界                   | `harness inspect` `orbit list` `orbit show <id>` `orbit enter <id>` `orbit current` `orbit status` | 读取 runtime 与 OrbitSpec，解析当前 orbit 的 `projection`，把工作区切到低噪音视图，并显示当前目标、边界与风险                                         |
| worker     | 只在当前工作单元范围内查看与提交 | 查看改动、历史，提交 scope 内改动                   | `orbit diff` `orbit log` `orbit commit` `orbit leave`                                              | 用 `orbit_write` surface 约束 Git pathspec，避免把 scope 外改动误带进提交；离开时恢复完整工作区视图                                            |
| orbit 作者   | 新建一个工作模块骨架       | 创建 orbit skeleton 并补 authored contract | `harness init` `orbit add <id>`                                                                    | 创建 hosted OrbitSpec 宿主，给一个新 orbit 提供最小骨架                                                                           |
| orbit 作者   | 编辑入口与导出内容        | 维护 brief、规则、过程与导出边界                    | 编辑 `.harness/orbits/<id>.yaml` `orbit brief materialize` `orbit brief backfill`                    | 让结构化 brief 与容器 block 互相同步，稳定 orbit 的入口与 authored truth                                                             |
| orbit 作者   | 导出或发布单个 orbit    | 保存为可安装 template 或从 source 发布           | `orbit template save` `orbit template publish` `orbit template init-source`                        | 按 `export` surface 生成 installable orbit template，支持 source authoring、runtime writeback 与 direct template authoring |
| harness 作者 | 搭建整个工作环境         | 创建 runtime、安装 orbit、检查组合               | `harness create` `harness install` `harness inspect` `harness check`                               | 建立 runtime 控制平面、安装 orbit template、写入 provenance，并验证整体一致性                                                           |
| harness 作者 | 导出整套环境           | 把成熟 runtime 导出为 harness template       | `harness template save`                                                                            | 把多个 orbit 的组合、宏观编排与必要文件导出为可复用 harness template                                                                     |

---

## 4. 终极形态下，Orbit 只做 6 件事

为了保持体量轻、边界清楚，Orbit 终极上只应稳定做下面 6 件事：

### 4.1 支撑最小工作模块骨架

把一类工作定义成最小可执行单元，至少具备：

- 目标
- 条件 / 约束
- 规则
- 成功 / 失败 / 异常退出语义
- 记录要求
- 文件系统级可观测性

### 4.2 界定文件边界

明确：

- `projection`
  - 看见什么
- `orbit_write`
  - 允许改什么
- `export`
  - 能带走什么
- `orchestration`
  - 哪些内容进入入口编排

### 4.3 物化执行入口

把 orbit authored contract 物化成 worker 真正能消费的入口环境，例如：

- 当前 orbit brief
- 根 `AGENTS.md` 中的当前 orbit block
- 必要的 rule / process 入口

### 4.4 统一退出语义

提供稳定的退出词汇与记录语义，例如：

- `success`
- `failure`
- `abnormal_exit`
- `external_stop`

Orbit 统一的是语义，不一定统一检测实现。

### 4.5 支撑模板物化与复用

让工作模块可以：

- 使用变量替换进行物化
- 安装为 runtime
- 保存 / 发布为 orbit template
- 从 runtime 显式写回模板

变量替换属于这一层能力，而不是 Orbit 的全部定义。

### 4.6 记录控制面元信息

记录与控制、退出、复用、审计直接相关的最小元信息，例如：

- authored truth
- revision identity
- runtime provenance
- current orbit state
- 必要 ledger / cache
- 结果记录落点

这不是全量过程数据库，而是控制面元信息。

---

## 5. 顶层设计原则

### 5.1 外部控制，不碰内部实现

Orbit 只负责定义和物化下面这些“外部控制”：

- 现在处于哪个 orbit
- 当前看见什么
- 当前允许写什么
- 当前要遵守哪些规则
- 当前的完成探针是什么
- 当前结果要记录到哪里、至少记录什么
- 当前内容如何导出、发布、复用

Orbit **不负责**：

- agent 的内部推理过程
- 任务拆分是否使用什么方法
- session 总结到底怎么生成
- model 内部的置信度或策略实现
- 多 agent 内部调度细节

换句话说，Orbit 只落实：

- 结果记录到哪里
- 结果最少要包含什么
- 哪些规则/探针/边界必须被看见

它不要求 worker 必须用哪种内部方法得到这些内容。

### 5.1A Orbit Framework 与 Orbit Template 不是一回事

为了让边界保持清楚，必须把下面两层分开：

- `Orbit framework`
  - 通用控制平面
  - 提供统一宿主、surface、命令入口、物化与回填机制
- `Orbit template`
  - 某一类工作的 authored contract
  - 定义这一类工作该如何进入、如何完成、如何退出、如何留痕

简单说：

- framework 负责“能力底座”
- template 负责“某类工作的具体说明”

如果把两者混成一层，Orbit 很快就会变成又重又难演化的半平台。

### 5.1B 责任分层

| 层                         | 负责什么                                                                       | 不负责什么                                                  |
| ------------------------- | -------------------------------------------------------------------------- | ------------------------------------------------------ |
| `Orbit framework`         | 宿主、surface、入口容器、brief lane、template/writeback/publish lane、统一退出语义词汇、记录落点机制 | 具体任务定义、预算检测实现、probe 实现、agent 内部推理                      |
| `Orbit template`          | 某类工作的目标、边界、规则、done probe、记录要求、成功/失败/异常退出语义                                 | Git/manifest/provenance 机制实现、通用 budget 计量、runtime 调度实现 |
| `Harness / agent runtime` | turns/tokens/time/cost 等预算来源、外部中断、重试调度、异常检测、记录写入执行                         | Orbit authored truth 本身、模板内容定义                         |
| `Worker`                  | 消费入口、执行任务、按合同留下结果                                                          | 重新定义模板合同、绕过边界自行发明控制面                                   |

最重要的判断标准是：

- 只要是“所有 orbit 通用且稳定”的能力，优先放 framework
- 只要是“某一类工作才成立”的说明，放 template
- 只要涉及具体 runtime 预算、调度、监控，放 harness / agent runtime

### 5.2 让 worker 负担最小

worker 不应该被迫理解：

- manifest taxonomy
- install provenance
- export surface 细节
- branch identity 细节

worker 只需要在执行时快速知道：

1. 现在该做什么；
2. 允许改什么；
3. 什么算做完；
4. 做完后把结果放哪里。

### 5.3 让作者有清晰的演化路径

Orbit 的价值不仅在执行时，也在于：

- orbit 作者可以稳定迭代单个工作单元
- harness 作者可以组合多个 orbit 并导出完整工作环境
- runtime 优化可以显式写回模板，而不是丢失在一次 session 里

### 5.4 让系统小于它控制的任务

Orbit 应该是一层小而稳的控制平面，而不是一个比业务更重的系统。

因此：

- 优先定义稳定外部合同
- 优先复用 Git
- 优先使用文件与显式命令
- 不引入服务端、数据库、后台守护进程
- 不把“元认知”做成一套持续高负担的旁白系统

### 5.5 退出语义必须统一，但退出判别不必统一实现

Orbit 作为控制平面，适合统一的是“退出语义”，不适合强行统一的是“退出检测实现”。

因此更合理的分工是：

- framework 提供统一退出词汇，例如：
  - `success`
  - `failure`
  - `abnormal_exit`
  - `external_stop`
- template 定义在该类工作里，什么情况属于这些退出
- harness / agent runtime 决定是否真的有能力检测：
  - time limit
  - retry limit
  - token / turn budget
  - cost budget

这能让 Orbit 保持轻量：

- 有 budget 监控能力的 runtime，可以真实执行这些约束
- 没有 budget 监控能力的 runtime，也可以把它们先当 authored hint 使用

---

## 6. 三类用户壳

同一个内核，对外应当呈现为三类用户壳。

| 用户         | 核心问题            | 主要对象                 | 应暴露的东西                                                       | 不应强迫理解的东西                         |
| ---------- | --------------- | -------------------- | ------------------------------------------------------------ | --------------------------------- |
| worker     | 我现在该做什么         | 当前 orbit             | 子目标、边界、规则、探针、记录目标                                            | manifest / provenance / export 细节 |
| orbit 作者   | 我怎样迭代一个可复用的工作单元 | 单个 orbit             | authored truth、brief、template save/publish、runtime writeback | harness 组合细节                      |
| harness 作者 | 我怎样搭建和演化整个工作环境  | 整个 runtime / harness | install、inspect、check、root orchestration、harness template    | 单 orbit 内部实现细节                    |

---

## 7. Worker 场景

## 7.1 执行状态

这是 Orbit 最重要的场景。

当 worker 进入一个 orbit 时，Orbit 应当尽量一次回答下面 6 个问题：

1. 当前 orbit 的子目标是什么；
2. 当前环境有哪些限制；
3. 当前允许修改哪些内容；
4. 当前必须遵守哪些规则；
5. 当前最小完成探针是什么；
6. 当前结果要记录到哪里、最少记录什么。

如果一个 orbit 不能把这 6 件事讲清楚，它对 worker 来说就仍然偏重。

## 7.2 初始化

从产品体验看，worker 会消费“初始化后的入口环境”，但这个初始化动作本身更接近作者工作，而不是执行工作。

边界建议：

- 根 `AGENTS.md`
  - 由 harness 作者负责
  - 只承载宏观任务编排、orbit 地图、切换原则、全局安全规则
- orbit brief
  - 由 orbit 作者负责
  - 只承载当前 orbit 的局部入口信息
- rule / process 文件
  - 承载较长、较细的规则与过程材料

因此：

- worker 负责消费入口
- harness / orbit 作者负责搭建入口

## 7.3 动态编排

这是后续可扩展能力，但边界必须很稳。

Orbit 可以承担的，是**显式的外部编排信号**，例如：

- 当前建议切换到哪个 orbit
- 为什么要切换
- 是否需要验证、重试、升级或人工介入

Orbit 不应该先变成：

- 自动多 agent 调度器
- 自主 planner 引擎
- 内部推理与反思系统

更合适的方向是：

- 事件驱动
- 输出控制信号
- 保留显式留痕

而不是持续代替 worker 思考。

---

## 8. 作者场景

## 8.1 Orbit 作者

Orbit 作者面对的是“一个可复用工作单元”的设计与迭代。

他们需要能：

- 直接开发 orbit template
- 从 source 分支开发再发布
- 在 runtime 中优化一个已安装 orbit，并显式写回模板

Orbit 作者最重要的职责不是把提示词写得越来越长，而是把下面这些东西设计清楚：

- orbit 负责什么，不负责什么
- 哪些文件应可见
- 哪些文件应可写
- 哪些文件应可导出
- brief 要暴露什么
- done probe 与结果记录目标是什么

## 8.2 Harness 作者

Harness 作者面对的是“整个工作环境”的搭建和演化。

他们需要能：

- 创建 runtime
- 安装多个 orbit
- 定义根 `AGENTS.md` 的宏观编排
- 检查 runtime 一致性
- 把优化后的 runtime 导出为 harness template

Harness 作者负责的是“系统组合”和“宏观控制”，而不是替某个 orbit 写所有局部规则。

---

## 9. Orbit 真正拥有的内容

Orbit 适合拥有的，应该都是稳定、外显、可审计、可复用的控制信息。

### 9.1 环境与边界

- 当前 orbit 是什么
- `projection / orbit_write / export / orchestration` 四个 surface
- 当前 worker 可见范围
- 当前 worker 可写范围

### 9.2 入口与规则

- 当前 orbit brief
- 当前 orbit 的 rule / process 输入
- 当前 orbit 的最小完成探针
- 当前 orbit 的记录目标与记录内容要求

### 9.3 留痕与复用

- authored truth
- runtime provenance
- brief materialize / backfill
- orbit template / harness template 的导出与发布

### 9.4 审计与复盘的最小合同

Orbit 可以要求：

- 结果写到哪里
- 结果至少记录哪些字段
- 什么情况下要留 warning 或 handoff

Orbit 不要求：

- worker 一定用 session 总结
- 一定用某种反思模板
- 一定用哪种 agent 内部策略

### 9.5 统一退出模型

从控制平面角度，一个 orbit 至少应支持下面几类退出：

- `success`
  - 达到 done probe，正常完成
- `failure`
  - 正常结束，但结果明确不满足该 orbit 的目标
- `abnormal_exit`
  - 未得到明确成功/失败，就因重试上限、时长上限、资源浪费、环境不匹配等停止
- `external_stop`
  - 被父级 harness、人工或全局预算中断

这里最重要的边界是：

- Orbit 适合统一这些退出语义
- Orbit 不一定要自己统一实现“如何检测超限”

---

## 10. Orbit Template 应当具备的 authored contract

一个成熟的 orbit template，至少应清楚写出下面 8 项：

1. `objective`
   - 这一类工作要完成什么
2. `scope boundary`
   - 这一类工作看什么、改什么、不改什么
3. `rules`
   - 必须遵守哪些规则
4. `done probe`
   - 什么算成功完成
5. `failure condition`
   - 什么情况应明确判定为失败
6. `abnormal exit hint`
   - 哪些情况应停止并转入异常退出
7. `record target`
   - 结果记录到哪里
8. `record minimum`
   - 最少记录什么

这 8 项是 template 的 authored contract，不是 framework 的自动能力。

例如：

- framework 可以提供 `abnormal_exit` 这个出口语义
- template 才能定义“连续 3 次无效重试”在该类任务里意味着什么
- runtime 才决定自己能不能检测“连续 3 次无效重试”

---

## 11. Orbit 不应该拥有的内容

下面这些能力，不适合进 Orbit 内核：

- 内部 chain-of-thought
- 推理策略实现
- 长篇自我反思文本
- 自动多 agent 编排引擎
- 后台状态中心
- 通用 memory 数据库
- block-level / semantic orbit
- 与 Git 并列的另一套历史系统
- 通用的 tokens / turns / time / cost 预算计量实现

预算相关更适合作为：

- template 的 authored hint
- harness / agent runtime 的检测能力
- 外层编排的退出决策输入

这些东西要么会被模型能力持续吞并，要么会让 Orbit 变重、变脆、变难维护。

---

## 12. 一个好 orbit 的最小 authored contract

每个 orbit 至少应该稳定提供下面 6 项信息：

1. `objective`
   - 这个 orbit 当前负责什么
2. `constraints`
   - 当前限制是什么
3. `rules`
   - 必须遵守哪些规则
4. `done probe`
   - 最小完成探针是什么
5. `record target`
   - 结果记录到哪里
6. `record content`
   - 最少要记录什么

这 6 项比“要求 worker 使用什么内部思维方法”更稳定，也更适合作为控制面合同。

---

## 13. 现有命令面应该怎样被理解

现有 CLI 可以被理解成三类用户壳的命令入口：

### 13.1 Worker shell

- `orbit list`
- `orbit show <orbit-id>`
- `orbit enter <orbit-id>`
- `orbit current`
- `orbit status`
- `orbit diff`
- `orbit log`
- `orbit commit`
- `orbit restore`

### 13.2 Orbit author shell

- `orbit add`
- `orbit brief materialize`
- `orbit brief backfill`
- `orbit template save`
- `orbit template publish`
- `orbit template init-source`

### 13.3 Harness author shell

- `harness create`
- `harness init`
- `harness install`
- `harness inspect`
- `harness check`
- `harness template save`

---

## 14. 文档地图

如果你要执行任务：

- 先看 [docs/worker_guide.md](./worker_guide.md)

如果你要设计或迭代单个 orbit：

- 先看 [docs/orbit_author_guide.md](./orbit_author_guide.md)
- 需要更细的模板作者细节时，再看 [docs/orbit_template_authoring_guide.md](./orbit_template_authoring_guide.md)

如果你要搭建整个 harness：

- 先看 [docs/harness_author_guide.md](./harness_author_guide.md)

如果你想看当前正式入口与可执行示例：

- 看 [docs/quickstart.md](./quickstart.md)
