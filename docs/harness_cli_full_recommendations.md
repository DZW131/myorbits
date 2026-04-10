# Harness CLI 命令设计（Clean Break 建议稿）

版本：v0.3
状态：建议稿
适用前提：

- `orbit` 负责定义单元、模板单元、projection
- `harness` 负责唯一正式运行态工程
- 一个 repo 只对应一个 harness runtime
- runtime 版本化元数据统一进入 `.harness/`
- 主文档不再承载兼容迁移口径

---

# 1. 文档目的

本文档给出下一阶段的正式 CLI 方向。

目标不是在旧 `orbit template apply`、`default harness`、`migrate` 等过渡概念上继续打补丁，而是直接给出 clean-break 的命令面：

- 用户操作单条 orbit 时，用 `orbit ...`
- 用户操作运行态工程时，用 `harness ...`

---

# 2. 总体设计原则

## 2.1 对象跟命令前缀走

### `orbit ...`

用于操作单条 orbit：

- orbit 定义
- orbit projection
- 单 orbit 模板导出
- orbit 级 branch / template 检查

### `harness ...`

用于操作运行态工程：

- 创建 harness 工程
- 初始化 harness runtime
- 安装 orbit template
- 管理 members
- 检查组合冲突
- 导出 harness template

## 2.2 一仓库一 Harness

本阶段采用一 repo 一 harness 的简化模型。

因此：

- 不需要 `default harness`
- 不需要 `--harness <id>`
- 不需要 `harness list`
- 不需要 `.harness/harnesses/<harness-id>.yaml`

当前 repo 中只有一个正式 harness runtime。

## 2.3 当前目录优先，`--path` 为补充

所有 `harness` 命令都应支持两种寻址方式：

1. 隐式当前目录
2. 显式 `--path <dir>`

寻址规则：

1. 若提供 `--path`，优先使用该目录；
2. 否则使用当前工作目录；
3. 必要时可向上查找最近的 harness root。

## 2.4 主文档不保留兼容别名

以下内容不应进入正式命令推荐：

- `orbit template apply`
- `orbit harness ...`
- `harness migrate`
- `default harness`
- “先兼容，再迁移”的提示文字

如果未来实现层内部需要临时兼容，那是实现问题，不应写进新阶段的主设计。

---

# 3. Harness Root 解析

## 3.1 Harness Root 的定义

一个目录若满足以下条件，则视为 harness root：

- 存在 `.harness/runtime.yaml`
- 结构满足 harness runtime 最小合同

例如：

```text
project-a/
  .harness/
    runtime.yaml
  .orbit/
    orbits/
      docs.yaml
```

## 3.2 命令如何确定目标 repo

统一规则：

1. `--path` 优先
2. 未指定时使用当前目录
3. 如当前目录不是 root，可向上查找最近的 harness root
4. 找不到时：
   - 只读命令报错
   - 创建命令可允许初始化新 harness

---

# 4. 推荐的 Orbit 命令

下面这些命令继续表示“用户在操作单条 orbit”。

## 4.1 Orbit 定义

```bash
orbit add <orbit-id>
orbit validate
orbit list
orbit show <orbit-id>
orbit files <orbit-id>
```

### 作用

- 创建 orbit 定义
- 校验 orbit 定义
- 查看 orbit 定义与文件范围

## 4.2 Orbit Projection

```bash
orbit current
orbit enter <orbit-id>
orbit leave
orbit status
orbit diff
orbit log
orbit commit
orbit restore
```

### 作用

继续围绕当前 harness runtime 中的某条 orbit 做 projection 与 scoped operations。

## 4.3 导出单 Orbit Template

```bash
orbit template save <orbit-id> --to <template-branch>
```

### 作用

导出单条 orbit 的模板态。

### 推荐 flags

- `--dry-run`
- `--edit-template`
- `--overwrite`
- `--default`
- `--json`

### 说明

- runtime bindings 默认来自 `.harness/vars.yaml`
- 输出仍然是 orbit template，不会写入 runtime `.harness/` 文件

## 4.4 生成 Bindings Skeleton

```bash
orbit bindings init <template-source> [--out <path>]
```

### 作用

根据模板源生成 bindings skeleton。

### 推荐默认行为

若未显式传 `--out`，默认目标路径应为：

```text
.harness/vars.yaml
```

原因：

- bindings 现在属于 harness runtime 层
- 不再属于 `.orbit/`

## 4.5 Branch 分类与检查

```bash
orbit branch status
orbit branch inspect <branch>
orbit branch list
```

### 作用

查看 branch 是：

- `template`
- `runtime`
- `plain`

其中：

- `runtime` 表示 harness runtime branch
- `template` 可细分 orbit template / harness template

---

# 5. 推荐的 Harness 命令

这些命令负责当前工作空间或显式目标目录对应的 harness 工程。

## 5.1 `harness create`

```bash
harness create <path>
```

### 作用

创建一个新的 harness 工程目录。

### 推荐行为

- 创建目录（必要时）
- 初始化 `.harness/runtime.yaml`
- 不自动安装 orbit
- 不自动 `enter`

### 推荐 flags

- `--mkdir`
- `--force`
- `--name <harness-id>`
- `--json`

## 5.2 `harness init`

```bash
harness init [--path <dir>]
```

### 作用

把一个已有目录初始化为 harness runtime repo。

### 推荐行为

- 创建 `.harness/runtime.yaml`
- 不清空已有文件
- 若已初始化则稳定失败或明确提示

## 5.3 `harness inspect`

```bash
harness inspect [--path <dir>]
```

### 作用

查看当前 harness runtime 的详细信息。

### 推荐输出

- harness root
- harness id / name
- members
- runtime 初始化状态
- vars / install records 摘要
- 当前 projection（若可获得）

## 5.4 `harness add`

```bash
harness add <orbit-id> [--path <dir>]
```

### 作用

把某条已有 orbit 加入当前 harness 的 members。

### 推荐行为

- 要求 `.orbit/orbits/<orbit-id>.yaml` 已存在
- 检查重复 member
- 检查身份冲突
- 更新 `.harness/runtime.yaml`

## 5.5 `harness remove`

```bash
harness remove <orbit-id> [--path <dir>]
```

### 作用

把某条 orbit 从当前 harness 的 members 中移除。

### 第一版建议

只调整 member 关系，不自动删除：

- 用户文件
- install record
- projection cache

## 5.6 `harness install`

```bash
harness install <orbit-template-source> [--path <dir>]
```

### 作用

把一条 orbit template 安装到当前 harness runtime。

### 这是新的正式安装入口

它是下一阶段对外唯一正式推荐的安装命令。

### 推荐行为

- 解析 orbit template
- 写入运行态文件
- 写入 `.orbit/orbits/<orbit-id>.yaml`
- 写入 `.harness/installs/<orbit-id>.yaml`
- 必要时更新 `.harness/vars.yaml`
- 更新 `.harness/runtime.yaml` 的 members
- 不自动 `enter`

### 推荐 flags

- `--bindings <file>`
- `--interactive`
- `--editor`
- `--overwrite-existing`
- `--dry-run`
- `--json`

## 5.7 `harness check`

```bash
harness check [--path <dir>]
```

### 作用

检查当前 harness 的组合冲突。

### 最少应检查

- duplicate members
- identity collisions
- variable collisions
- content/path collisions

### 推荐行为

默认只读，不写入。

## 5.8 `harness template save`

```bash
harness template save --to <template-branch> [--path <dir>]
```

### 作用

把当前 harness 工程导出为 harness template。

### 推荐行为

- 读取 current harness members
- 为每个 member 生成 orbit template candidate
- 合并多个 candidate
- 检测变量 / 路径冲突
- 可选进入最终模板编辑
- 写出 harness template branch

### 推荐 flags

- `--dry-run`
- `--edit-template`
- `--overwrite`
- `--default`
- `--json`

## 5.9 `harness root`

```bash
harness root [--path <dir>]
```

### 作用

输出目标目录对应的 harness root。

### 价值

这个命令对：

- shell 集成
- editor 集成
- 调试
- 脚本编排

都很有帮助。

---

# 6. 推荐的正式命令总表

## 6.1 Orbit 命令

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

## 6.2 Harness 命令

```bash
harness create <path>
harness init [--path <dir>]
harness inspect [--path <dir>]
harness add <orbit-id> [--path <dir>]
harness remove <orbit-id> [--path <dir>]
harness install <orbit-template-source> [--path <dir>]
harness check [--path <dir>]
harness template save --to <template-branch> [--path <dir>]
harness root [--path <dir>]
```

---

# 7. 不应引入或不应继续推荐的命令

以下命令或口径不应出现在下一阶段主文档中：

```bash
orbit template apply <template-source>
orbit harness create ...
orbit harness add ...
orbit harness remove ...
orbit harness inspect ...
harness migrate ...
harness list ...
```

也不应继续强化以下参数或概念：

- `--harness <id>`
- `default harness`
- “兼容 alias”
- “迁移提示”

原因很简单：

- 它们会重新引入旧世界观；
- 会模糊“一仓库一 harness”的模型；
- 会让 `orbit` / `harness` 边界再次变脏。

---

# 8. 推荐的帮助文案方向

## 8.1 `orbit` 顶层 help

建议突出：

- orbit 定义
- orbit projection
- 单 orbit template 导出
- branch 分类检查

一句话摘要建议：

> `orbit` 用于定义单条 orbit、进入 orbit projection、导出单 orbit template，并查看 orbit 相关 branch 信息。

## 8.2 `harness` 顶层 help

建议突出：

- harness runtime 初始化
- orbit template 安装
- members 管理
- harness template 导出

一句话摘要建议：

> `harness` 用于操作当前工作空间或指定目录下的 harness runtime：创建、初始化、安装 orbit、管理成员、检查冲突与导出 harness template。

---

# 9. 示例路径

## 9.1 创建新工程

```bash
harness create /work/project-a
cd /work/project-a
orbit add docs
orbit add cli
```

## 9.2 安装一个 Orbit Template

```bash
harness install https://example.com/acme/templates.git --path /work/project-a
orbit enter docs
```

## 9.3 导出 Harness Template

```bash
harness template save --to harness-template/project-a --path /work/project-a
```

---

# 10. 最终结论

下一阶段的 CLI 应该直接体现新的对象边界：

- `orbit` = 定义 / projection / 单 orbit template
- `harness` = 运行态工程 / 安装 / 组合 / harness template

不要再在主文档里保留过渡命令或迁移措辞。
从命令面开始，就应该是 break new。
