# Orbit AGENTS.md 扩展方案（V0.2 历史合同）

版本：V0.2
状态：historical compatibility contract；0030-0034 已落地；post-v0.3 已由 `docs/orbit_member_runtime_technical_spec.md` / `docs/orbit_member_runtime_development_plan.md` 接管
关联文档：`docs/context/orbit_phase2_prd.md`、`docs/context/orbit_phase2_technical_spec.md`、`docs/context/orbit_storage_boundary.md`

---

## 1. 文档定位

本文档定义 `AGENTS.md` 在 Orbit V0.2 收尾扩展中的最小可实施模型。

说明：

- 本文档只记录已经实现过的 V0.2 compatibility lane；
- 它不再是 post-v0.3 `AGENTS.md` 行为的主合同；
- 若与 `docs/orbit_member_runtime_technical_spec.md` 或 `docs/orbit_member_runtime_development_plan.md` 冲突，以后两者为准。

它解决的问题不是“是否需要单独处理 AGENTS”，而是：

1. 在不破坏 Phase 2 主线模板 contract 的前提下，如何支持 `AGENTS.md`；
2. 当前版本到底支持哪些语义，不支持哪些语义；
3. 哪些行为应直接冻结为实现合同，避免边做边猜。

本文档是 `AGENTS.md` 扩展的实现基线。当前仓库已按本合同完成 `0030` 到 `0034` 的主实现。若后续代码与本文档冲突，应先更新本文档或停下确认，而不是直接按猜测实现。

---

## 2. V0.2 冻结结论

先固定 8 条结论：

1. `AGENTS.md` 不是普通 owned file，而是一个单独建模的 shared file。
2. Phase 2 主线闭环不依赖 `AGENTS.md` 的特殊处理能力。
3. V0.2 的 first cut 只支持 `replace-block + create-if-absent`。
4. V0.2 不支持泛化 `append` merge engine，也不支持多 shared-file 的统一抽象落地。
5. 模板态中的 AGENTS payload 不带 marker；marker 只存在于运行态 `AGENTS.md`。
6. `.orbit/template.yaml` 必须扩展 shared-file entry，用来声明该模板包含 AGENTS fragment。
7. 运行态 `AGENTS.md` 中的异常 marker 场景全部 fail-closed。
8. `orbit validate` 已增加对运行态 `AGENTS.md` marker 合法性的专项校验。

一句话：

> V0.2 的 AGENTS 扩展不是“做一个通用 shared-file 系统”，而是“为 AGENTS.md 增加一个受控、可验证、可回退的专用 lane”。

### 2.1 与 post-v0.3 收敛方向的关系

本文件记录的是 **当前已实现的 V0.2 合同**。

后续若切到 member runtime / mixed install 收敛模型，应额外冻结以下方向：

1. 运行态根 `AGENTS.md` 永远只是 block 容器，不是第一性真相源。
2. orbit template lane 的长期目标不再默认吸收 unmarked prose；目标行为应只提取“当前 orbit block 内容”。
3. 若开发者希望把运行态当前 orbit block 回填为真相源，应通过显式命令完成，而不是让 save / apply / validate 自动反向同步。
4. future member-generated block 的手工编辑默认只影响运行态容器；只有显式回填后，才会改动结构化真相源。

也就是说：

- 本文档第 6 节与第 9 节描述的是 **当前 V0.2 已实现行为**；
- 它们不是 post-v0.3 的最终目标行为；
- 未来收敛以 `docs/orbit_member_runtime_technical_spec.md` 为准。

---

## 3. 当前主线实现带来的直接约束

当前 Phase 2 主线代码已经明显固化为“整文件模板”模型：

- `template content builder` 产出的是按路径组织的完整文件内容；
- `render` 以整文件为单位做变量替换；
- `template apply` 当前默认按完整文件写回；
- conflict analyzer 当前是 path 级冲突，不是 block 级冲突。

因此，`AGENTS.md` 扩展不能假装“只是多一个普通文件”，也不应在 V0.2 直接反向抽象成一个泛化 shared-file 框架。

V0.2 应采取的做法是：

1. 保持主线普通文件 save/apply 语义不变；
2. 为 `AGENTS.md` 增加一条显式声明、显式解析、显式写入的专用路径；
3. 在必要处对 source loader、apply writer、validate 增加专门分支；
4. 不要求本轮把所有 file/block merge 问题抽象到通用层。

这是一种刻意保守的设计，目的是减少返工面，而不是追求一次性抽象到位。

---

## 4. 模板 branch 合同

### 4.1 模板 branch 中允许的 AGENTS 扩展落点

当模板包含 AGENTS fragment 时，模板 branch 允许额外出现：

```text
.orbit/
  template.yaml
  orbits/
    <orbit-id>.yaml

AGENTS.md
<owned user files...>
```

说明：

- 根目录 `AGENTS.md` 是模板态 payload，不是运行态完整 `AGENTS.md`；
- 该文件只保存 fragment payload，不保存 marker；
- 它不替代 `.orbit/template.yaml` 中的 shared-file 声明；
- 它在模板 branch 中与运行态目标路径同名，但语义由 manifest 决定；
- 它不允许写入 `.git/orbit/state/*`；
- 仍然禁止回退到历史 `.orbit-template/` 目录方案。

### 4.2 `.orbit/template.yaml` 扩展合同

当模板包含 AGENTS fragment 时，`template.yaml` 必须声明 shared-file entry：

```yaml
schema_version: 1
kind: template
template:
  orbit_id: docs
  default_template: false
  created_from_branch: main
  created_from_commit: abc123
  created_at: 2026-03-21T10:00:00Z
variables:
  project_name:
    description: 项目名称
    required: true
shared_files:
  - path: AGENTS.md
    kind: agents_fragment
    merge_mode: replace-block
    include_unmarked_content: true
```

V0.2 约束：

- `path` 在本轮只允许 `AGENTS.md`；
- `kind` 在本轮只允许 `agents_fragment`；
- `merge_mode` 在本轮只允许 `replace-block`；
- `include_unmarked_content` 明确记录本轮约定的 save 语义；
- 若模板不包含 AGENTS fragment，则 `shared_files` 可省略。

这组字段的作用不是做未来通用 shared-file 平台，而是让当前 source resolver / apply 路径能够明确识别：

- 这是 AGENTS shared payload；
- 模板 branch 中的根目录 `AGENTS.md` 不应被当作普通 owned file；
- 它不应该被当作普通 owned file 处理。

---

## 5. 运行态 marker 合同

### 5.1 语法

V0.2 冻结为 HTML comment + XML 风格属性：

```md
<!-- orbit:begin orbit_id="docs" -->
## Orbit: docs

- 优先阅读 docs/README.md
<!-- orbit:end orbit_id="docs" -->
```

说明：

- `orbit_id` 是必填属性；
- begin / end 的 `orbit_id` 必须一致；
- marker 本身只存在于运行态 `AGENTS.md`；
- 模板态根目录 `AGENTS.md` 不包含 begin / end marker。

### 5.2 合法性规则

以下场景全部 fail-closed：

1. begin / end 不配对；
2. begin / end 的 `orbit_id` 不一致；
3. nested block；
4. 同一个 `orbit_id` 出现多个 block；
5. 无法解析属性；
6. marker 文本被手工改坏，导致无法稳定识别块边界。

V0.2 不做自动修复，也不做“尽量猜测”。

---

## 6. `template save` 语义

### 6.1 输入来源

若当前运行态 repo 存在 `AGENTS.md`，则 `template save` 在普通模板文件流程之外，额外执行 AGENTS 提取逻辑。

它必须先把运行态 `AGENTS.md` 解析为有序片段序列：

1. unmarked span
2. marker block

并保留原文顺序。

### 6.2 提取规则

设当前保存的 orbit id 为 `docs`，则模板态根目录 `AGENTS.md` payload 的提取规则为：

1. 保留所有 unmarked span，内容原样写入；
2. 若存在 `orbit_id="docs"` 的合法 block，则保留该 block 的内部内容，但不保留 marker；
3. 删除其他 orbit 的 marker block；
4. 保持剩余内容的原始顺序。

这意味着模板 payload 是“当前 orbit block 内容 + 所有未归属 marker 的内容”的组合，而不是运行态完整 `AGENTS.md` 的字节拷贝。

### 6.3 缺少当前 orbit marker 的行为

若运行态 `AGENTS.md` 存在，但没有当前 orbit 对应的 marker block：

1. `template save` 必须发出 warning；
2. 仍按 6.2 的规则继续执行；
3. 此时 payload 只包含 unmarked span；
4. 不允许静默猜测“哪些段落属于当前 orbit”。

### 6.4 空 payload 的默认处理

若按上述规则提取后，payload 为空：

- 默认不生成 `shared_files` entry；
- 默认不写模板态根目录 `AGENTS.md`；
- 普通模板保存流程继续。

这是基于当前决策推导出的默认值，目的是避免制造一个空 block 的伪 shared-file 条目。

### 6.5 与变量替换和编辑的关系

对模板态根目录 `AGENTS.md` payload：

1. 变量替换规则与普通模板文件一致；
2. `--edit-template` 允许在 temp dir 中编辑 payload；
3. 编辑的是 payload 本体，不是运行态完整 `AGENTS.md`；
4. save 过程中绝不回写 runtime repo 的 `AGENTS.md`。

---

## 7. `template apply` 语义

### 7.1 输入形态

模板态中的 AGENTS payload 来自：

```text
AGENTS.md
```

其内容不带 marker，仅包含 plain fragment payload。

`template apply` 对该 payload：

1. 先按普通模板变量规则完成渲染；
2. 再根据当前 orbit id 为其包裹运行态 marker；
3. 最终写入或更新运行态 `AGENTS.md`。

### 7.2 目标 repo 不存在 `AGENTS.md`

若目标 repo 不存在 `AGENTS.md`：

- 创建 `AGENTS.md`；
- 文件内容为：

```md
<!-- orbit:begin orbit_id="<orbit-id>" -->
<rendered payload>
<!-- orbit:end orbit_id="<orbit-id>" -->
```

### 7.3 目标 repo 已存在同 orbit block

若目标 repo 已存在相同 `orbit_id` 的合法 marker block：

- 原地替换该 block 的内部内容；
- 保留 block 外的其他内容；
- 发出 warning，明确告知“已替换现有 orbit block”；
- 不允许静默替换而不提示。

### 7.4 目标 repo 已存在 `AGENTS.md` 但不存在同 orbit block

若目标 repo 已存在 `AGENTS.md`，且 marker 合法，但不存在同 orbit block：

- 在文件末尾追加一个空行；
- 再追加新的 marker block；
- 不修改已有 unmarked 内容或其他 orbit block。

这里的“追加到末尾”是一个明确的 V0.2 行为，不做更复杂的插入位置推断。

### 7.5 失败条件

以下场景全部 fail-closed：

1. 目标 `AGENTS.md` marker 非法；
2. 同 orbit block 重复；
3. payload 渲染失败；
4. marker 生成失败；
5. 无法稳定判定替换位置。

---

## 8. `orbit validate` 扩展点

V0.2 当前实现中，`orbit validate` 已增加运行态校验项：

1. `AGENTS.md` begin / end 是否配对；
2. `orbit_id` 属性是否存在且匹配；
3. 是否存在 nested block；
4. 是否存在重复 `orbit_id` block；
5. marker 文本是否满足冻结语法；
6. 若 repo 中存在 `AGENTS.md` marker，是否能稳定解析为有序 segment 序列。

这部分校验面向运行态 repo，不是模板态 payload 校验的替代。

---

## 9. 已接受风险

### 9.1 默认纳入 unmarked 内容

本轮明确接受一个有意识的产品风险：

- `template save` 默认会把运行态 `AGENTS.md` 中所有 unmarked 内容一并纳入模板 payload。

这会带来几个真实风险：

1. repo-level 的通用说明可能被多个 orbit 模板重复打包；
2. 多模板重复 apply 后，运行态 `AGENTS.md` 可能出现重复或风格漂移；
3. unmarked 内容的“所有权”不再完全清晰；
4. 开发者需要在模板和运行态手工整理一部分共享说明。

V0.2 仍接受这个风险，原因是：

1. 你当前明确偏好“先让开发者自己编辑解决”，而不是过早引入复杂自动分段；
2. 这能保持 save 语义简单、直接；
3. 风险主要落在内容治理，不会破坏 Git / storage boundary / template branch contract。

这条风险应在真正实现时继续保留为显式文档，而不是默默藏在实现里。

注意：

- 这里记录的是 **当前实现仍然存在的兼容风险**；
- 它不再代表后续 member runtime 收敛后的目标行为；
- 长期目标应改为“只提取当前 orbit block 内容，不默认纳入 unmarked prose”。

---

## 10. 明确不做的内容

以下内容不纳入本轮 `AGENTS.md` first cut：

1. 泛化 `append` merge engine；
2. 多 shared-file 的统一抽象框架；
3. 无 marker 旧仓库的自动迁移；
4. 自动识别 unmarked 内容属于哪个 orbit；
5. 复杂插入位置策略；
6. 对 malformed marker 的自动修复；
7. 模板态 payload 中保存 marker；
8. 把 AGENTS shared metadata 写入 `.git/orbit/state/*`。

---

## 11. 最小验收对照

当前实现已覆盖的最小验收项：

1. `template save` 只保留 unmarked 内容和当前 orbit block 内容；
2. `template save` 会丢弃其他 orbit block；
3. 无当前 orbit marker 时会 warning，但不静默猜测；
4. `template save --edit-template` 不污染 runtime `AGENTS.md`；
5. `template apply` 在无 `AGENTS.md` 时可创建带 marker 的文件；
6. `template apply` 在已有同 orbit block 时原地替换并告警；
7. `template apply` 在无同 orbit block 时追加到文件末尾；
8. malformed / duplicate / nested marker 全部 fail-closed；
9. `orbit validate` 能发现非法运行态 marker；
10. shared-file 元数据与 payload 不写入 `.git/orbit/state/*`。

---

## 12. 实施记录

本轮实际采用的落地顺序是：

1. 先扩展 `.orbit/template.yaml` schema 和 source loader；
2. 再实现 runtime `AGENTS.md` parser / validator；
3. 然后实现 `template save` 的提取路径；
4. 然后让 source loader / apply path 按 manifest 把根目录 `AGENTS.md` 识别为 shared payload，而不是普通 owned file；
5. 最后接 `template apply` 的 replace-block / create-if-absent 写入路径。

原因很简单：

- 先冻结 manifest 和 parser，才能让 save/apply 的行为有稳定合同；
- 先做 parser / validate，才能把 fail-closed 边界守住；
- 如果反过来从 apply 写入开始做，最容易在异常 marker 场景里埋隐患。
