# Orbit Template Publish PRD

版本：v0.1
状态：proposal
阶段：post-v0.3 enhancement
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/harness_centric_runtime_development_plan.md`
- `docs/technical-debt.md`

---

## 1. 文档目标

定义一个新的作者侧命令：

```bash
orbit template publish
```

它的目标不是改变 `harness install` 的消费边界，而是把“作者 source branch 导出最新 orbit template branch”的动作收口成一个明确、低心智负担、可自动化的发布入口。

本文档只定义产品行为与命令边界，不覆盖具体代码拆分与测试矩阵。

---

## 2. 背景

当前 v0.3 的推荐模型是：

- source branch 作为模板作者分支
- `orbit-template/<orbit-id>` 作为正式模板发布分支
- 用户侧始终通过已发布的 template branch 执行 `harness install`
- `orbit template save` 继续保留为底层显式原语，也可用于从 runtime-like 分支提取 orbit template

这个边界是正确的，但当前作者工作流仍然偏手工：

```bash
orbit template save issues --to orbit-template/issues --default
git push origin orbit-template/issues
```

对于以“维护一个 orbit 模板仓库”为主的作者来说，这一步重复、机械，而且容易忘记执行，导致：

- source branch 已更新，但 template branch 仍旧；
- 模板消费者安装到的不是最新模板态；
- source branch / release branch 的关系需要人工记忆。

因此需要一个显式的作者侧发布命令，来把“发布最新模板态”收口成一个更稳定的入口。

---

## 3. 设计原则

1. `harness install` 仍然只消费已经发布好的 template branch，不隐式触发发布。
2. 发布动作属于模板作者侧，不属于模板消费者侧。
3. source branch 继续是 source of truth；template branch 继续是发布产物。
4. 不把 source branch 与 release branch 重新混为一个概念。
5. source branch 必须有显式 marker，不靠启发式猜测。
6. 默认行为应足够省脑，但远端写入必须显式。
7. source branch 是单 orbit 的专业作者分支，不承担多 orbit authoring。

---

## 4. 目标

新增一个显式作者命令：

```bash
orbit template publish
```

它应支持两层能力：

### 4.1 基础版

在当前本地作者仓库内：

- 识别当前 repo 是一个合法的 orbit 模板源仓库；
- 解析要发布的 orbit id；
- 执行一次等价于 `orbit template save` 的发布动作；
- 更新本地 `orbit-template/<orbit-id>` 分支；
- 不自动 push。

### 4.2 扩展版

支持显式控制：

```bash
orbit template publish --orbit <orbit-id>
orbit template publish --push
orbit template publish --remote <remote>
orbit template publish --default
```

`--orbit` 在 source branch 模型里不是主要选择器；它只作为显式一致性校验入口保留。

---

## 5. 非目标

本次不做：

- `harness install` 中的隐式发布
- 从用户 runtime repo 反向触发模板作者仓库发布
- 后台自动发布 daemon
- 远端模板仓库的自动重建服务
- 多 orbit 批量 publish
- harness template publish
- 让 source branch 支持多 orbit authoring

---

## 6. 用户模型

### 6.1 Source Branch

source branch 负责：

- 手工编辑单 orbit 的模板源内容
- 维护作者说明、README、测试脚本、开发辅助文件
- 通过 `orbit template publish` 产出正式 template branch

第一版通过：

```text
.orbit/source.yaml
```

显式声明 source branch 身份、合法的 `source_branch`，以及可选的 `publish.orbit_id`。

source branch 的产品合同进一步收紧为：

- 只允许承载一个 orbit definition
- 不应携带 `.harness/*`
- 直接展示模板原始内容，而不是 runtime bindings 展开后的内容
- 可以存在作者说明、测试脚本、开发辅助文件，但这些文件通常不进入已发布 template branch

### 6.2 发布分支

发布分支负责：

- 保存可被 `harness install` 消费的模板态内容
- 作为 source artifact 被安装

发布分支不应被手工编辑。

### 6.3 消费者仓库

消费者仓库只做：

```bash
harness install <repo-url> --ref orbit-template/<orbit-id>
```

消费者不应承担发布模板态的责任。

---

## 7. 命令面

### 7.1 基础命令

```bash
orbit template publish
```

含义：

- 在当前 repo 内发布一个 orbit template branch
- 若只存在一个 orbit definition，可自动推断该 orbit
- 发布 ref 固定为：

```text
orbit-template/<orbit-id>
```

### 7.2 显式 orbit

```bash
orbit template publish --orbit issues
```

用于：

- 对 source branch 的唯一 orbit id 做显式一致性校验
- 迁移阶段避免作者完全依赖自动推断

### 7.3 发布并推送

```bash
orbit template publish --push
```

含义：

- 先更新本地 template branch
- 再把该 branch push 到默认 remote

默认 remote 为：

```text
origin
```

### 7.4 显式 remote

```bash
orbit template publish --push --remote origin
```

用于：

- 指定发布目标 remote
- 避免在多 remote 环境里依赖隐式推断

### 7.5 default template 标记

```bash
orbit template publish --default
```

它等价于把 `--default` 透传给当前 `orbit template save`。

这项保持显式，不自动开启。

---

## 8. 仓库识别与默认解析

### 8.1 合法 source branch

`orbit template publish` 只应在本地 source branch 上运行。

最小判定条件：

- 当前目录位于 Git repo 内
- 当前 revision 上存在合法 `.orbit/source.yaml`
- 存在合法 `.orbit/config.yaml`
- 恰好存在一个合法 `.orbit/orbits/*.yaml`
- 当前 revision 上不存在 `.orbit/template.yaml`
- 当前 revision 上不存在 `.harness/runtime.yaml`
- 当前 revision 上不存在 `.harness/vars.yaml`
- 当前 revision 上不存在 `.harness/installs/*`
- 当前 revision 上不存在 `.harness/template.yaml`

若不满足，命令应 fail-closed。

`harness install` 不消费 `.orbit/source.yaml`；它继续只消费带有效 `.orbit/template.yaml` 的已发布 template branch。

### 8.2 orbit 自动解析

默认解析规则：

- source branch 产品合同要求恰好一个 orbit definition
- `publish` 默认发布这个唯一 orbit
- 若显式给出 `--orbit`，其值必须与唯一 orbit id 一致
- `.orbit/source.yaml` 中若配置了 `publish.orbit_id`，其值必须与唯一 orbit id 一致

### 8.3 publish ref 解析

publish ref 不支持命令级自定义命名。

固定规则为：

```text
orbit-template/<orbit-id>
```

这样做的目标是避免引入 template branch 命名漂移。

---

## 9. 执行模型

`orbit template publish` 的本质是一个更高层的作者命令封装。

基础版行为：

1. 校验当前 repo 是合法模板源仓库
2. 解析 orbit id
3. 解析固定 publish ref
4. 调用等价于 `orbit template save <orbit-id> --to orbit-template/<orbit-id>` 的保存路径
5. 输出发布结果摘要

扩展版行为：

1. 完成上述本地发布
2. 若显式给出 `--push`，再执行 push

### 9.1 输出重点

text / json 输出应至少覆盖：

- `orbit_id`
- `publish_ref`
- `default_template`
- `pushed`
- `remote`（若 push）

---

## 10. 与现有命令的关系

### 10.1 与 `orbit template save`

关系：

- `orbit template save` 继续保留，作为底层保存原语
- `orbit template publish` 是作者侧更高层、更省脑的发布封装
- 从 runtime-like 分支反推出模板的路径仍然存在，但它不应与 source branch 专业 authoring 模型混在同一 branch 身份里

也就是说：

- `save` 偏底层、显式
- `publish` 偏作者工作流、带 source branch 合同与默认解析

### 10.2 与 `orbit template init-source`

推荐补一个配套的作者入口：

```bash
orbit template init-source
```

它的职责是：

- 读取当前分支名，写入 `.orbit/source.yaml`
- 默认把 `source_branch` 设为当前分支
- 自动写入或校验唯一 orbit 的 `publish.orbit_id`

这不是 `publish` 的一部分，但它会显著降低 source branch 的初始化成本。

### 10.3 与 `harness install`

关系：

- `harness install` 仍只消费已经存在的 template branch
- `install` 不隐式触发 publish
- source repo 的发布责任不迁移给消费者 runtime

---

## 11. 推荐工作流

### 11.1 模板作者

```bash
git switch <source-branch>
# 编辑模板源内容
git commit -m "..."
orbit template publish --push --default
```

### 11.2 模板消费者

```bash
harness install https://github.com/<owner>/<repo>.git --ref orbit-template/issues
```

这样可以保持：

- source branch 与 release branch 的清晰边界
- 用户安装链路的稳定性
- 作者发布动作的低心智负担

---

## 12. 一句话总结

`orbit template publish` 的职责是：

**把模板作者仓库中的单 orbit source branch，显式、稳定地发布成最新 orbit template branch；source branch 由 `.orbit/source.yaml` 标记并声明 `source_branch` 与可选 `publish.orbit_id`，而消费者侧 install 继续只消费已发布的 template branch。**
