# Harness-Centric Runtime Model PRD（产品需求文档）

版本：v0.3
状态：clean-break 基线，可进入技术设计与实现拆分
关联文档：
- `docs/harness_cli_full_recommendations.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/harness_centric_runtime_development_plan.md`
- `docs/harness_install_progress_prd.md`
- `docs/harness_install_source_repo_resolution_technical_spec.md`
- `docs/harness_mixed_install_technical_spec.md`
- `docs/context/orbit_phase2_prd.md`（历史背景）
- `docs/context/orbit_phase2_technical_spec.md`（历史背景）
- `docs/context/orbit_agents_md_development.md`（后置扩展）

---

# 1. 文档目的

本文档定义 Orbit 下一阶段的正式产品模型。

这一次不是“在旧 Orbit Runtime 上继续叠加功能”，也不是“先兼容旧模型、再慢慢迁移”。
本文件直接把新的世界观一次性说清楚：

- `orbit`：定义单元、模板单元、projection 目标
- `harness`：唯一正式运行态对象、唯一工程组合层
- `template`：分为 orbit template 和 harness template
- `runtime`：若无特别说明，始终指 harness runtime

本文档回答 4 个问题：

1. 哪些概念被保留；
2. 哪些概念被正式废弃；
3. 新模型下数据应该放在哪里；
4. 下一阶段的正式命令与流程是什么。

---

# 2. 一句话定义

**新模型 = orbit 只负责定义、模板与 projection；harness 负责唯一正式运行态；单个 repo 只承载一个 harness runtime，运行态版本化元数据统一进入 `.harness/`。**

---

# 3. 核心结论

## 3.1 一仓库一 Harness

下一阶段采用单 repo、单 harness runtime 模型：

- 一个 repo 只对应一个 harness runtime；
- 不支持“一个 repo 内多个 harness 并存”；
- 不引入 `default harness` 概念；
- 也不需要 `.harness/harnesses/<harness-id>.yaml` 这一层目录。

`harness` 仍然可以有稳定的 `id` / `name`，但它是“当前 repo 的 harness 身份”，不是“repo 中多个 harness 之一”。

## 3.2 Runtime 只有 Harness Runtime

以下口径从本阶段开始正式废弃：

- “独立的 orbit runtime”
- “runtime branch = orbit runtime branch”
- “单 orbit runtime 和多 orbit runtime 并存”

保留的只有一种正式运行态：

- `harness runtime`

`orbit runtime` 最多只允许作为一个解释性短语，表示“某条 orbit 在当前 harness runtime 中的局部运行态内容”，它不是正式对象、不是 branch 类型、不是落盘实体。

## 3.3 Orbit 不再承载运行态元数据

`.orbit/` 只保留 orbit 自身的定义与 orbit template 相关元数据。

原本 Phase 2 中放在 `.orbit/` 里的运行态版本化信息，本阶段统一收敛到 `.harness/`，包括：

- repo 级 bindings / vars
- repo 级 install records
- harness runtime 身份与 members

## 3.4 Projection 继续属于 Orbit

`enter / leave / current / status / diff / log / commit / restore` 继续围绕 orbit projection 运作。

也就是说：

- runtime 是 harness
- projection 是 orbit

Harness 的引入不改变 projection 的对象，只改变 projection 所在的正式宿主。

## 3.5 正式安装入口改为 `harness install`

从产品定义上，安装 orbit template 到运行态工程的正式入口是：

```bash
harness install <orbit-template-source>
```

`orbit template apply` 不再是新阶段主路径，不应继续出现在下一阶段的主文档、主示例和命令推荐中。

---

# 4. 对象模型

## 4.1 Orbit

Orbit 是最小定义单元。

它负责：

- 定义“哪些文件属于它”
- 作为 projection 目标
- 作为单 orbit 模板导出的源对象

它不负责：

- 作为正式运行态对象存在
- 记录 repo 级 bindings
- 记录安装来源
- 记录组合关系

## 4.2 Orbit Template

Orbit template 是单个 orbit 的模板态快照。

它至少包含：

- 该 orbit 的模板化文件树
- 变量占位符
- `.orbit/template.yaml`
- `.orbit/orbits/<orbit-id>.yaml`

## 4.3 Harness Runtime

Harness runtime 是系统中唯一正式的 runtime 对象。

它表示：

- 当前 repo 已启用运行态工程
- 当前 repo 的 harness 身份
- 当前 repo 包含哪些 orbit members
- 当前 repo 的 repo-level bindings 与 install records

## 4.4 Harness Template

Harness template 是多个 member orbit 模板快照组合而成的模板态对象。

关键点：

1. Harness template 是快照型，不是引用型；
2. 它的输入是 member orbit 的模板快照，不是对 runtime repo 做“代码 vs 内容”猜测；
3. 它是独立于 orbit template 的另一种模板类型。

## 4.5 Projection

Projection 指用户当前进入的 orbit 视图。

Projection 的职责保持不变：

- 进入某条 orbit
- 只看该 orbit 的 user view / scoped operations
- 离开后恢复完整工作区视图

---

# 5. 存储模型

## 5.1 `.orbit/` 的职责

`.orbit/` 只承载 orbit 自身的版本化定义与 orbit template 元数据：

```text
.orbit/
  orbits/
    <orbit-id>.yaml
  template.yaml          # 仅 orbit template branch 使用
```

`.orbit/` 负责：

- orbit definitions
- orbit template manifest

`.orbit/` 不再负责：

- runtime bindings
- install records
- harness runtime identity
- harness members

## 5.2 `.harness/` 的职责

`.harness/` 是新的运行态版本化元数据层：

```text
.harness/
  runtime.yaml
  vars.yaml
  installs/
    <orbit-id>.yaml
  template.yaml          # 仅 harness template branch 使用
```

`.harness/runtime.yaml` 至少记录：

- harness `id`
- harness `name`（可选）
- members
- created_at / updated_at 等稳定元数据

`.harness/vars.yaml`：

- 保存 runtime repo 级 bindings
- 供 `harness install`、`orbit template save` 等流程复用

`.harness/installs/<orbit-id>.yaml`：

- 记录该 orbit 当前安装来源
- 记录模板 ref / commit / source 摘要

`.harness/template.yaml`：

- 只出现在 harness template branch
- 作为 harness template manifest

## 5.3 `.git/orbit/state/` 的职责

`.git/orbit/state/` 继续只承载 repo-local 的 projection state 与 cache：

- 当前进入的 orbit
- projection cache
- warnings / status snapshot
- repo-local lock

这里不放版本化内容，也不放 harness runtime metadata。

## 5.4 Git DAG 的职责

Git DAG 继续承载：

- runtime branch 历史
- orbit template branch 历史
- harness template branch 历史

它仍然是唯一历史真相源。

---

# 6. 产品目标

## 6.1 主目标

建立一套真正 clean-break 的 Harness-Centric 运行模型，使系统具备：

1. runtime 语义彻底统一为 harness runtime；
2. orbit 与 harness 的职责彻底分开；
3. 运行态版本化元数据从 `.orbit/` 收敛到 `.harness/`；
4. 单 orbit 安装、多 orbit 组合、harness template 导出都落到同一模型里；
5. orbit projection 体验保持稳定。

## 6.2 子目标

1. 明确一仓库一 harness；
2. 用 `harness install` 取代旧的 orbit 安装主路径；
3. 用 `.harness/vars.yaml` 取代旧的 runtime `.orbit/vars.yaml`；
4. 用 `.harness/installs/*.yaml` 取代旧的 runtime `.orbit/installs/*.yaml`；
5. 用 harness members 明确表达组合关系；
6. 用 harness template 表达多 orbit 模板组合。

---

# 7. 非目标

本阶段不处理以下内容：

1. 多 harness 共存于同一 repo；
2. `default harness` 与 related fallback 机制；
3. 旧版 repo 的兼容迁移方案；
4. 同一 repo 中同一 orbit 的多版本并存；
5. orbit dependency / transitive dependency 系统；
6. registry / marketplace；
7. 智能代码语义分析或“代码 vs 内容”自动判定；
8. `AGENTS.md` 与 harness 组合逻辑的泛化 shared-file 系统。

---

# 8. 范围

本阶段包含 5 个能力域：

1. Harness Runtime 建模
2. Orbit Template 安装到 Harness Runtime
3. Harness Members 管理
4. Harness Template 导出
5. Harness 冲突检测与 branch classification 更新

---

# 9. 功能需求

## 9.1 Harness Runtime 建模

系统应允许在 repo 中显式记录 harness runtime 元信息。

最小能力：

- repo 是否启用 harness runtime
- harness 的 `id`
- harness 的 members
- runtime 级 vars / installs 的稳定落点

本阶段不做：

- repo 内多个 harness 的建模
- `default harness`
- harness 间切换

## 9.2 Orbit Template 安装

正式安装命令：

```bash
harness install <orbit-template-source>
```

它的语义是：

> 把一个 orbit template 安装到当前 repo 对应的 harness runtime 中。

安装结果至少包括：

1. 写入运行态文件；
2. 写入 `.orbit/orbits/<orbit-id>.yaml`；
3. 写入 `.harness/installs/<orbit-id>.yaml`；
4. 必要时更新 `.harness/vars.yaml`；
5. 更新 `.harness/runtime.yaml` 中的 members；
6. 不自动 `enter`。

补充安装合同：

- 同一 `orbit-id` 的再次安装默认失败；
- 只有显式 `--overwrite-existing` 才允许覆盖既有安装实例；
- 首阶段不支持 `--as`，安装 identity 固定使用模板中的 `orbit_id`。

本阶段不再把“安装一个 orbit”表述为“创建独立 orbit runtime”。

## 9.3 Orbit Template 导出

正式单 orbit 模板导出命令继续为：

```bash
orbit template save <orbit-id> --to <template-branch>
```

但它在运行态读取 bindings 的来源改为：

- 当前 repo 的 `.harness/vars.yaml`
- 或显式传入的 bindings 文件

导出的 orbit template 仍然不包含：

- `.harness/runtime.yaml`
- `.harness/vars.yaml`
- `.harness/installs/*.yaml`

## 9.4 Harness Members 管理

系统应允许显式维护当前 harness 的 members。

至少支持：

- 添加 orbit member
- 移除 orbit member
- 查看当前 harness 的 members 与摘要

member 的加入必须显式发生，不做隐式猜测。

## 9.5 Harness Template 导出

系统应允许把当前 harness 导出为 harness template。

正式导出逻辑不是扫描 runtime repo 再猜“哪些文件属于模板”，而是：

1. 读取当前 harness 的 members；
2. 对每个 member orbit 生成 orbit template candidate / export；
3. 合并多个 candidate；
4. 合并变量声明；
5. 检测变量与路径冲突；
6. 生成 harness template branch。

导出结果至少包含：

- `.harness/template.yaml`
- 组合后的模板文件树
- 必要的 member orbit definitions

## 9.6 Harness 冲突检测

本阶段至少检测以下冲突：

1. 重复 member 冲突
2. member 身份冲突
3. 变量名冲突
4. 内容路径冲突
5. template branch 写入冲突

系统不得静默覆盖。

---

# 10. 关键流程

## 10.1 创建 Harness Runtime

1. 用户执行 `harness create <path>` 或 `harness init`
2. 系统初始化 `.harness/runtime.yaml`
3. 当前 repo 成为 harness runtime repo
4. 用户可继续定义 orbit 或安装 orbit template

## 10.2 安装单 Orbit Template

1. 用户执行 `harness install <orbit-template-source>`
2. 系统解析 orbit template
3. 系统读取模板内 `orbit_id`
4. 若同一 `orbit-id` 已存在，则默认失败；只有显式 `--overwrite-existing` 才进入覆盖安装
5. 系统渲染运行态文件
6. 系统写入 orbit definition
7. 系统写入 `.harness/installs/<orbit-id>.yaml`
8. 必要时更新 `.harness/vars.yaml`
9. 系统将该 orbit 写入 `.harness/runtime.yaml` 的 members
10. apply 完成，但不自动 `enter`

## 10.3 管理 Harness Members

1. 用户执行 `harness add <orbit-id>` 或 `harness remove <orbit-id>`
2. 系统校验 orbit definition 是否存在
3. 系统检查重复 / 身份冲突
4. 系统更新 `.harness/runtime.yaml`

## 10.4 导出 Harness Template

1. 用户执行 `harness template save --to <template-branch>`
2. 系统读取当前 harness members
3. 对每个 member 构建 orbit template candidate
4. 系统合并多个 candidate
5. 系统检测变量与路径冲突
6. 可选进入最终模板编辑
7. 系统写出 harness template branch

---

# 11. CLI 口径（产品层）

## 11.1 Orbit 命令

Orbit 继续负责定义、projection 与单 orbit 模板：

```bash
orbit add <orbit-id>
orbit validate
orbit list
orbit show <orbit-id>
orbit files <orbit-id>
orbit current
orbit enter <orbit-id>
orbit leave
orbit status
orbit diff
orbit log
orbit commit
orbit restore
orbit template save <orbit-id> --to <template-branch>
orbit bindings init <template-source> [--out <path>]
orbit branch status
orbit branch inspect <branch>
orbit branch list
```

其中：

- `orbit bindings init` 默认输出到 stdout，主文档推荐 `--out .harness/vars.yaml`；
- `orbit template save` 读取 runtime repo 中的 `.harness/vars.yaml`。

## 11.2 Harness 命令

Harness 负责运行态工程：

```bash
harness create <path>
harness init [--path <dir>]
harness inspect [--path <dir>]
harness add <orbit-id> [--path <dir>]
harness remove <orbit-id> [--path <dir>]
harness install <orbit-template-source> [--overwrite-existing] [--path <dir>]
harness check [--path <dir>]
harness template save --to <template-branch> [--path <dir>]
harness root [--path <dir>]
```

---

# 12. 验收标准

## 12.1 Runtime 与存储

- repo 中存在 `.harness/runtime.yaml` 时，被识别为 harness runtime repo；
- `harness create` / `harness init` 后允许 `members` 为空，zero-member runtime 仍然有效；
- 同一 repo 不再建模多个 harness；
- runtime 版本化元数据进入 `.harness/`，而不是 `.orbit/`。

## 12.2 安装语义

- `harness install <source>` 可安装本地或远程 orbit template；
- 安装后写入 `.orbit/orbits/<orbit-id>.yaml`；
- 安装后写入 `.harness/installs/<orbit-id>.yaml`；
- 必要时更新 `.harness/vars.yaml`；
- 对同一 `orbit-id` 的再次安装默认失败，只有显式 `--overwrite-existing` 才允许覆盖；
- 首阶段安装 identity 固定使用模板中的 `orbit_id`；
- 安装后不自动 `enter`。

## 12.3 Projection 语义

- `orbit enter / leave / current / status / diff / log / commit / restore` 继续围绕 orbit projection 工作；
- harness 的引入不改变 projection 目标。

## 12.4 Orbit Template

- `orbit template save` 继续可用；
- 保存时读取的是 `.harness/vars.yaml`；
- 生成的 orbit template 不包含 runtime `.harness/` 文件。

## 12.5 Harness Template

- 可以从当前 harness members 成功导出 harness template；
- 导出逻辑基于 member orbit 的模板快照，而不是 runtime repo 全局内容扫描；
- 若 runtime root 存在 `AGENTS.md`，导出时把它视为运行态容器，先规范化为 payload-only 文本，再模板化写入 harness template；
- 导出支持冲突检测与最终模板编辑。

---

# 13. 明确废弃的旧口径

以下口径不应再出现在下一阶段主文档中：

- `default harness`
- “独立 orbit runtime”
- `orbit dependency`
- “runtime branch = orbit runtime branch”
- `orbit template apply` 作为正式安装入口
- `.orbit/vars.yaml` 作为 runtime bindings 主落点
- `.orbit/installs/*.yaml` 作为 runtime install records 主落点
- 一个 repo 内多个 harness 并存

这不是“兼容期内先别强调”，而是下一阶段文档层面直接废弃。
