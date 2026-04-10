# Worker Guide

版本：v0.4
状态：worker-facing usage guide
适用范围：已存在的 runtime / harness 环境中的任务执行者

关联文档：
- `docs/orbit_positioning_and_personas.md`
- `docs/quickstart.md`
- `docs/orbit_brief_lane_v0_4_technical_spec.md`

---

## 1. 这份文档给谁看

这份文档给正在执行任务的人或 agent 看。

你的目标不是理解 Orbit 的全部内核，而是快速搞清楚：

1. 现在在哪个 orbit；
2. 当前子目标是什么；
3. 当前能改什么、不能改什么；
4. 什么算完成；
5. 结果应该记到哪里。

---

## 2. 你只需要知道的最小心智模型

- `harness`
  - 整个工作环境
- `orbit`
  - 一个局部工作单元
- 根 `AGENTS.md`
  - 宏观任务编排入口
- orbit brief
  - 当前 orbit 的局部入口
- orbit rule / process 文件
  - 详细规则与过程材料

你通常**不需要**理解：

- manifest kind
- install provenance
- export / publish 细节
- authoring lane 细节

---

## 3. 一个好 orbit 应该告诉你的事

进入一个 orbit 后，理想情况下你应该马上知道：

1. 当前子目标是什么。
2. 当前有哪些环境限制。
3. 当前允许修改哪些文件或区域。
4. 当前有哪些必须遵守的规则。
5. 当前最小完成探针是什么。
6. 当前结果应该记录到哪里，最少记录什么。

如果这些信息缺失，不是 worker 的问题，而是 orbit 还需要继续打磨。

---

## 4. 当前推荐的最小执行流程

如果你已经在一个准备好的 runtime 里，推荐按下面的顺序工作。

### 4.1 先确认环境

```bash
harness inspect
orbit list
```

`harness inspect` 用来确认当前 repo 是一个正常 runtime。  
`orbit list` 用来查看当前有哪些 orbit 可进入。

### 4.2 进入一个 orbit

```bash
orbit show docs
orbit enter docs
orbit current
orbit status
```

推荐顺序：

- 先用 `orbit show <orbit-id>` 看 orbit 定义和基础边界
- 再用 `orbit enter <orbit-id>` 进入当前工作单元
- 再用 `orbit current` / `orbit status` 确认当前状态

### 4.3 再读入口，而不是自己猜

执行前，优先读这几类内容：

1. 根 `AGENTS.md`
   - 看宏观任务编排、orbit 关系、切换原则
2. 当前 orbit brief
   - 看当前局部目标、限制、规则、探针
3. 当前 orbit 的 rule / process 文件
   - 看详细要求

Orbit 只负责把这些入口外显出来，不规定你必须怎样“理解”它们。

### 4.4 执行工作

工作时，继续使用你本来就会的工具：

- 编辑器
- shell
- Git
- 测试命令
- repo 内脚本

Orbit 不接管你的内部执行方式，它只负责约束和记录外部边界。

### 4.5 进行最小检查

完成后，不要先做整套重检查，先跑当前 orbit 约定的最小 done probe。

这个 probe 可能是：

- 一个脚本
- 一条测试命令
- 一个状态检查
- 一个文档或记录文件是否已更新

Orbit 负责说明“检查什么”，不规定你必须如何组织内部验证过程。

### 4.6 留痕并收尾

当 orbit 要求留下记录时，按 orbit 的要求把结果写到指定位置。

Orbit 适合约束的是：

- 记录到哪里
- 至少记录什么

Orbit 不适合约束的是：

- 你必须用 session 总结
- 你必须用哪种摘要格式生成方法

最后再看变更范围：

```bash
orbit diff
orbit commit -m "update docs orbit"
orbit leave
```

---

## 5. Worker 最常用的命令

| 命令 | 用途 |
| --- | --- |
| `orbit list` | 看有哪些 orbit |
| `orbit show <orbit-id>` | 看某个 orbit 的定义与边界 |
| `orbit enter <orbit-id>` | 进入当前工作单元 |
| `orbit current` | 看当前 orbit |
| `orbit status` | 看 in-scope / out-of-scope 状态 |
| `orbit diff` | 只看当前 orbit 范围的改动 |
| `orbit log` | 看当前 orbit 范围历史 |
| `orbit commit` | 只提交当前 orbit 范围改动 |
| `orbit restore` | 恢复当前 orbit 范围内容 |
| `orbit leave` | 回到完整工作区视图 |

---

## 6. 你不应该被迫关心的东西

如果你只是 worker，下列内容不应该成为你的常规负担：

- `.harness/manifest.yaml` 的具体字段
- `.harness/installs/*.yaml` 的 provenance 细节
- template save / publish 的区别
- brief backfill 的内部工作方式

如果你频繁需要关心这些，说明作者侧还应该继续把体验做轻。

---

## 7. 什么时候该找 orbit 或 harness 作者

下面这些问题，不应靠 worker 临时发明解决方案：

- 当前 orbit 没讲清楚目标和限制
- 当前 orbit 没讲清楚记录目标
- 当前 orbit 没给最小 done probe
- 根 `AGENTS.md` 与 orbit brief 明显冲突
- orbit 需要频繁手工解释才能用

这些都属于 authoring 质量问题。

---

## 8. 下一步

如果你要设计或优化一个 orbit，而不是执行任务：

- 看 [docs/orbit_author_guide.md](./orbit_author_guide.md)

如果你要搭建整个 harness：

- 看 [docs/harness_author_guide.md](./harness_author_guide.md)
