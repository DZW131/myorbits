# Orbit Author Guide

版本：v0.4
状态：orbit-author-facing usage guide
适用范围：单个 orbit 的设计、迭代、保存与发布

关联文档：

- `docs/orbit_positioning_and_personas.md`
- `docs/orbit_template_authoring_guide.md`
- `docs/orbit_brief_lane_v0_4_technical_spec.md`
- `docs/quickstart.md`

---

## 1. 这份文档给谁看

这份文档给“单个 orbit 的开发者”看。

你的目标不是搭建整个 harness，而是把一个 orbit 打磨成：

- 边界清楚
- worker 易用
- 可复用
- 可导出
- 可继续迭代

---

## 2. 先把 orbit 的职责定清楚

一个好的 orbit，首先不是“很多文件”，而是一个清晰的工作单元。

设计时先回答下面 6 个问题：

1. 这个 orbit 负责什么，不负责什么。
2. worker 进入后应该先看到什么。
3. 哪些内容只需可见，哪些内容可写。
4. 哪些内容应进入导出模板。
5. 当前最小 done probe 是什么。
6. 当前结果应记录到哪里、至少记录什么。

如果这些问题还答不清楚，不要急着堆更多模板文件。

---

## 3. Orbit 作者最重要的边界

Orbit 作者负责的是**外部控制合同**，不是 worker 的内部实现。

你应当定义：

- 入口 brief
- rule / process 输入
- 可见 / 可写 / 可导出范围
- done probe
- record target

你不应当强行定义：

- agent 的内部推理方法
- session 总结必须如何生成
- 必须使用哪种思维链或计划格式

Orbit 只需要规定“结果落在哪里、最少包含什么”，不需要接管内部过程。

### 3.1 不要把 framework 责任写进 template

作为 orbit 作者，你写的是某类工作的 authored contract，不是在重写 Orbit 框架本身。

你要定义的是：

- 这类工作是什么
- 这类工作何时成功、何时失败、何时应异常退出
- 这类工作要把结果落到哪里

你不需要定义的是：

- manifest / state / provenance 的底层机制
- Git 交互方式
- turns / tokens / time / cost 的通用计量实现
- agent 内部如何做计划、反思、总结

如果一个模板越来越像“内部 runtime 平台设计”，通常说明边界放错了。

### 3.2 模板的 authored contract 应至少包含什么

推荐每个 orbit template 至少写清下面 8 项：

1. `objective`
2. `scope boundary`
3. `rules`
4. `done probe`
5. `failure condition`
6. `abnormal exit hint`
7. `record target`
8. `record minimum`

这 8 项里：

- `done probe`
  - 说明什么算成功
- `failure condition`
  - 说明什么情况应明确判定为失败
- `abnormal exit hint`
  - 说明什么情况不该继续消耗资源，应中止、升级或切换

后两项尤其重要，因为它们能让 orbit 不只是“完成路径清楚”，也让“退出路径清楚”。

---

## 4. 三种作者路径怎么选

| 路径                  | 适合场景                                | 核心命令                                      |
| ------------------- | ----------------------------------- | ----------------------------------------- |
| 直接维护 orbit template | orbit 小而稳定，直接维护 installable payload | `orbit template save` / `publish`         |
| source 分支开发再发布      | orbit 复杂，需要 author-only 文件、脚本、实验材料  | `orbit template init-source` / `publish`  |
| runtime 优化后写回       | 先在真实 runtime 里调通，再反写模板              | `harness install` + `orbit template save` |

简单选择法：

- 小而稳定的 orbit，用 direct template
- 复杂且长期演化的 orbit，用 source
- 真实跑通后的沉淀，用 runtime writeback

---

## 5. 你真正要写好的内容

## 5.1 Brief 要短而稳

brief 只应该承载 worker 进入当前 orbit 时必须先知道的内容，例如：

- 当前子目标
- 当前限制
- 当前关键规则
- 当前最小 done probe
- 当前记录目标

不要把所有知识都塞进 brief。

更长的内容应继续放在：

- `rule`
- `process`
- 相关业务文件

## 5.2 Surface 要清楚

你需要分清 4 个 surface：

- `projection`
  - worker 看见什么
- `orbit_write`
  - Orbit scoped commands 允许改什么
- `export`
  - 模板导出能带走什么
- `orchestration`
  - brief / `AGENTS.md` lane 吃什么

如果这 4 个面混在一起，worker 体验和模板质量都会变重。

## 5.3 Probe 要便宜

done probe 的目标是提供“最小完成信号”，而不是把整个质量系统复制进每个 orbit。

优先选择：

- 最小脚本
- 最小测试
- 最小状态检查
- 最小文档更新检查

### 5.3A Failure 与 abnormal exit 也要写清楚

很多 orbit 只写了“怎么成功”，但没有写“什么时候该停”。

更稳的写法是同时定义：

- `success`
  - done probe 满足
- `failure`
  - 明确不满足目标，且继续本 orbit 没意义
- `abnormal_exit`
  - 还不能下成功/失败结论，但继续消耗已经不划算

常见的 `abnormal_exit hint` 可以包括：

- 超过建议重试次数
- 超过建议时长
- 超过建议 turns / tokens / cost
- 证据持续冲突
- 当前情况明显不匹配这个 orbit

### 5.3B 预算检测不是模板必须实现的

这里要刻意保持轻量：

- 模板可以声明 budget hint
- 模板可以声明超过预算后的推荐动作
- 模板不需要自己实现 budget 检测

更具体地说：

- Orbit framework 负责提供统一退出语义
- Harness / agent runtime 负责决定哪些 budget 能被真实检测
- 模板只负责把“应当何时停止继续消耗”写清楚

## 5.4 Record 要外显

Orbit 最值得拥有的，不是“如何思考”，而是：

- 结果写到哪里
- 至少写什么
- 哪些 warning / handoff 必须留下

建议把异常退出时的最小记录也一起写清楚，例如：

- 退出类型
- 当前卡住的原因
- 已尝试过什么
- 建议下一步动作

---

## 6. 当前推荐的作者循环

## 6.1 直接维护 orbit template

适合快速起一个小而清楚的 orbit。

```bash
harness init
orbit add docs
```

然后：

1. 编辑 `.harness/orbits/<orbit-id>.yaml`
2. 编辑 orbit 相关规则、过程与业务文件
3. 如需更新 brief 容器，使用 `orbit brief materialize`
4. 如需把容器编辑回填到结构化 truth，使用 `orbit brief backfill`
5. 用 `orbit template save` / `orbit template publish` 导出或发布

## 6.2 source 分支开发再发布

适合复杂 orbit。

```bash
orbit template init-source
```

然后：

1. 在 source 分支维护 hosted OrbitSpec 与 source-only 文件
2. 用 source-only 文件支持开发、实验、发布检查
3. 用 `orbit template publish` 生成 installable `orbit_template`

## 6.3 runtime 中优化后写回

适合“先在真实环境里调通，再沉淀模板”。

```bash
harness install <template-source>
orbit enter <orbit-id>
```

在 runtime 中验证 orbit 的 worker 体验后，再用：

```bash
orbit template save <orbit-id>
```

把当前 export surface 反写回模板。

---

## 7. Orbit 作者最容易犯的 5 个错误

1. 把 brief 写成整本手册，而不是入口。
2. 不区分 projection、write、export、orchestration。
3. 不告诉 worker 结果应记录到哪里。
4. 把沉重的全量检查塞进每个 orbit。
5. 试图规定 agent 内部如何思考，而不是规定外部合同。
6. 只写成功路径，不写 failure / abnormal exit 路径。
7. 把 budget hint 当成模板自己必须实现的 runtime 能力。

---

## 8. 什么时候该升级成 harness 设计问题

如果你发现自己在解决下面这些问题，它们更像 harness 作者问题，而不是单 orbit 作者问题：

- 多个 orbit 的切换顺序
- 全局安全规则
- 根 `AGENTS.md` 的宏观编排
- 多 orbit 的组合导出
- runtime 级安装与检查策略

这时应该转到 harness author 视角。

---

## 9. 下一步

如果你要看更完整的 orbit template 作者工作流：

- 看 [docs/orbit_template_authoring_guide.md](./orbit_template_authoring_guide.md)

如果你要搭建整个 harness：

- 看 [docs/harness_author_guide.md](./harness_author_guide.md)
