# Harness Mixed Install Technical Spec

版本：v0.1
状态：proposal
阶段：post-v0.3 enhancement
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/harness_install_source_repo_resolution_technical_spec.md`
- `docs/orbit_template_publish_technical_spec.md`
- `docs/orbit_member_runtime_technical_spec.md`
- `docs/technical-debt.md`

---

## 1. 文档目标

冻结 `harness install` 后续扩展到 mixed install 时的对象模型、冲突策略和 `AGENTS.md` / 变量处理合同。

这份文档处理三类需求：

1. `harness install` 除 orbit template 外，也能安装 harness template；
2. orbit template 与 harness template 可以安装到同一个 runtime；
3. 第一阶段 mixed install 只允许 disjoint install，第二阶段只允许同 install unit replace。

本文档**不**改变当前已实现的 v0.3 基线；它定义的是下一阶段的增强方案。

---

## 2. 现状与问题

当前代码与文档基线是：

- `harness install` 只支持 orbit template branch；
- `harness template save` 只负责导出 harness template branch；
- 首阶段不做 harness template apply；
- source branch 不是 installable template branch，只能通过 `orbit template publish` 产出 orbit template branch；
- `harness template save` 对根目录 `AGENTS.md` 使用 whole-file lane；
- 变量冲突在 harness template candidate merge 阶段按保守规则 fail-closed。

这意味着：

- 单 orbit 模板安装已经成熟；
- 多 orbit 组合模板只支持导出，不支持重新安装；
- runtime 中若未来同时存在 orbit-template-based install 和 harness-template-based install，当前 ownership / `AGENTS.md` / drift 模型都还没有明确合同。

---

## 3. 产品目标

本增强的目标是：

1. 保持 `harness install` 作为唯一正式安装入口；
2. 让 `harness install` 可以接受：
   - orbit template install unit
   - harness template install unit
   - orbit source repo alias（仍最终解析为 orbit template install unit）
3. 允许一个 runtime 中同时存在多个 install unit；
4. mixed install 第一阶段只支持安全并存；
5. mixed install 第二阶段只支持“同 install unit 覆盖同 install unit”；
6. `AGENTS.md` 不按普通路径冲突处理；
7. 变量冲突继续沿用“兼容即合并，不兼容即失败”的保守规则。

---

## 4. 非目标

本次方案不做以下事情：

1. 不允许 warning 后继续执行任意路径冲突覆盖；
2. 不引入 install-unit 之间的多 owner / ref-count ownership；
3. 不支持 bundle 内单 member 的局部替换；
4. 不引入变量 namespacing、自动重命名或 alias；
5. 不让 harness template 与 orbit template 在 schema 层合并成同一个对象；
6. 不自动 publish source branch。

---

## 5. install unit 模型

### 5.1 不把 harness template 当成 orbit template

`harness template` 不应被建模为“字面意义上的特殊 orbit template”。

更准确的说法是：

- orbit template 和 harness template 都属于 **template install unit**
- 但它们仍然是不同对象类型、不同 manifest、不同 ownership 粒度

这样可以让 `harness install` 走统一管线，而不混淆对象模型。

### 5.2 三种 install input

下一阶段 install 输入可分为：

1. **orbit template install unit**
   - 标记：`.orbit/template.yaml`
2. **harness template install unit**
   - 标记：`.harness/template.yaml`
3. **source alias**
   - 输入仍是 source repo / source branch
   - 但在 install 前先解析到 orbit template install unit

source alias 本身不是 install unit。

---

## 6. Phase 1：mixed disjoint install

### 6.1 已冻结结论

第一阶段 mixed install 只允许 **disjoint install**。

也就是：

- orbit template 与 orbit template 可以并存
- orbit template 与 harness template 可以并存
- harness template 与 harness template 可以并存

前提是冲突集全部通过。

### 6.2 冲突集

第一阶段 mixed install 的 fail-closed 冲突集至少包含：

1. member identity 冲突
2. 普通路径冲突
3. 变量声明冲突
4. `AGENTS.md` lane 冲突

任一冲突都必须 fail-closed。

### 6.3 member identity 冲突

member identity 冲突指：

- 新 install unit 想引入的 orbit id 已存在于 runtime members 中
- 且该 orbit id 不属于“同 install unit replace”的场景

第一阶段对这类冲突全部 fail-closed。

### 6.4 普通路径冲突

普通文件路径冲突仍按现有模板冲突思路处理：

- 同一路径、内容完全相同：允许
- 同一路径、内容不同：fail-closed

但 `AGENTS.md` 不走这条规则。

---

## 7. Phase 2：same install unit replace

### 7.1 已冻结结论

第二阶段只允许：

- orbit install unit 覆盖同一个 orbit install unit
- harness install unit 覆盖同一个 harness install unit

不允许：

- orbit template 覆盖 bundle 内单 member
- harness template 覆盖多个散装 orbit install
- 不同 install unit 之间的危险冲突覆盖

### 7.2 orbit install unit replace

沿用现有 `--overwrite-existing` 心智：

- identity = `orbit_id`
- old owned files 通过 install record replay + cleanup plan 处理

### 7.3 harness install unit replace

建议新增 bundle-level identity：

- identity = `harness_id`

并引入 bundle-level install record。

第一版不支持：

- bundle 内局部 member replace
- bundle 与散装 orbit 的局部互相覆盖

---

## 8. `AGENTS.md` 合同

### 8.1 为什么不能当普通文件

`AGENTS.md` 几乎会出现在绝大多数模板中。

如果 mixed install 里把它按普通 path conflict 处理，则：

- 只要安装两个模板就几乎必然冲突
- mixed install 实际不可用

因此 `AGENTS.md` 必须继续保留专用 lane。

### 8.2 orbit template 的 lane

orbit template 继续沿用当前 lane：

- runtime `AGENTS.md` 是 block 容器
- runtime 根文件只承担容器职责，不是第一性真相源
- 当前 orbit block 的真相源来自 `meta.agents_template` / orchestration 输入，而不是模板 branch 根 `AGENTS.md`
- block identity = `orbit_id`
- 安装时 replace / append 该 orbit block
- reverse sync 若需要发生，只能走显式 orbit 命令 `orbit brief backfill`，不是 install side effect
- orbit template branch 自身不再携带根 `AGENTS.md`

见：

- `cmd/orbit/cli/template/agents_apply.go`
- `docs/orbit_member_runtime_technical_spec.md`

### 8.3 harness template 的 lane

harness template install 不应把根 `AGENTS.md` 当普通文件覆盖。

建议新增 bundle block lane：

- runtime `AGENTS.md` 仍是 block 容器
- runtime 根文件仍不是 harness template 的真相源，只是运行态宿主
- block identity = `harness_id`
- harness template 的根 `AGENTS.md` payload 作为一个完整 block 写入

也就是：

- orbit install -> `orbit_id` block
- harness install -> `harness_id` block

### 8.4 `AGENTS.md` 冲突定义

第一阶段 mixed install 中，`AGENTS.md` lane 的冲突不按普通路径判定，而按 block identity 判定：

- 新 block identity 不存在：允许 append
- 同 install unit replace：允许 replace 同 identity block
- 不同 install unit 试图改写已有 block：fail-closed

### 8.5 对 `harness template save` 的新增要求

一旦 runtime `AGENTS.md` 成为多 install unit block 容器，`harness template save` 就不能继续把运行态整文件原样回收为模板。

已冻结决策：

- **`harness template save` 在导出时必须规范化 / 剥离 runtime block markers。**

也就是说：

- source payload 是运行态容器中按当前顺序展开后的 payload 文本
- 但导出到 harness template 时，必须先去掉 runtime block marker 包装
- 模板中的根 `AGENTS.md` 存储的是 payload，不是运行态 marker 容器本体
- unmarked prose 保持为容器级文本，不自动归因到某个 orbit block

否则：

- runtime block markers 会被卷回模板
- 再次安装时会形成 marker 嵌套或重复 block

### 8.6 仍未冻结的实现细节

以下细节后续还需专门设计，但方向已确定：

1. runtime `AGENTS.md` 中 harness block 的 marker 语法是否复用 orbit marker 语法
2. harness block 与 orbit block 是否共享同一 parser
3. `harness template save` 剥离 markers 时如何稳定保留 block 顺序和 unmarked 文本

---

## 9. 变量冲突合同

### 9.1 已冻结结论

mixed install 与 install 主路径中的变量处理都采用同一条保守合同：

- 兼容即合并
- 不兼容即失败

这里需要明确区分两类问题：

- **声明冲突**
  - 两个 install unit 是否把同名变量当成同一语义对象
- **运行态值复用**
  - runtime 中若已存在该变量值，安装时是否继续沿用它

第一版冻结为：

- 先判断声明是否兼容；
- 声明兼容时，安装默认自动继续；
- runtime 中若已存在该变量值，则默认复用；
- 声明不兼容时，安装 fail-closed；
- 不引入 namespacing / alias / 自动 rename。

### 9.2 当前兼容规则

沿用现有 harness template candidate merge 的规则：

- 描述相同：可合并
- 一边描述为空：用另一边补齐
- `required` 按 OR 合并
- 两边描述都非空且不同：fail-closed

见：

- `cmd/orbit/cli/harness/template_merge.go`

### 9.3 安装期快速路径

为了避免常见同名变量场景阻塞 `harness install`：

- 若 runtime 中不存在该变量：
  - 正常把 install unit 的声明和值写入 `.harness/vars.yaml`
- 若 runtime 中已存在同名变量，且声明兼容：
  - 默认复用 runtime 中已有值
  - 更新后的变量声明按兼容规则合并
  - 命令继续执行，不要求额外 flag
- 若 runtime 中已存在同名变量，但声明不兼容：
  - 在写入任何 install 结果前 fail-closed
  - 诊断必须至少指出变量名和冲突来源

这意味着 install 期的变量冲突处理不应再只看“整个 `.harness/vars.yaml` 文件是否变化”，而应先做 per-variable 的声明兼容判定。

### 9.4 为什么暂不做 namespacing

当前 `.harness/vars.yaml` 仍是全局单命名空间。

若没有 schema-backed namespacing / alias 模型，就不能安全支持：

- 自动 rename
- `<install-unit>-<var>` 命名空间
- warning 后继续执行冲突覆盖

因此 mixed install 第一版里：

- 同名同义变量允许合并
- 同名异义变量必须 fail-closed

### 9.5 deferred 逃生阀

若真实工作流中“同名异义但仍希望快速继续”成为高频需求，后续可以单独设计显式策略面，例如：

- `--var-conflicts=prefer-runtime`

但该能力当前**不属于第一版 mixed install / install 主路径范围**。

在没有 schema-backed alias / persistence model 之前，不允许：

- 默认 warning 后继续
- 交互式 rename
- `<install-unit>-<var>` 自动 namespacing

### 9.6 这与 mixed install 的关系

这意味着 mixed install 第一版是“可组合但保守”的：

- 若两个 install unit 的变量词汇表兼容，则可并存
- 若不兼容，必须由模板作者先解决命名 / 描述层冲突

---

## 10. ownership / provenance 影响

### 10.1 orbit install unit

继续沿用当前模型：

- runtime member source = install
- install record = `.harness/installs/<orbit-id>.yaml`

### 10.2 harness install unit

若要支持 harness template install，必须新增 bundle-level provenance。

建议方向：

- 新增 bundle install record，例如：
  - `.harness/bundles/<harness-id>.yaml`

该 record 至少保存：

- `harness_id`
- source kind / repo / ref / commit
- member ids
- applied_at
- 是否包含 root `AGENTS.md`

若要为 Phase 2 replace 做好基础，建议额外保存：

- owned paths

### 10.3 runtime members

当前 runtime member source 只有：

- `manual`
- `install`

后续 mixed install 需要新增更细粒度来源表达。

推荐方向：

- `install_orbit`
- `install_bundle`

或等价 schema 扩展。

否则：

- `harness check`
- mixed overwrite
- drift provenance

都无法精确归因。

---

## 11. 推荐实现顺序

### Phase A：source repo 直装

先完成：

- source alias resolution
- `harness install` 接线

这一步不引入 harness template install。

### Phase B：harness template install（disjoint only）

1. 增加 harness template remote/local discovery
2. 定义 bundle install record
3. 定义 disjoint conflict set
4. 引入 bundle `AGENTS.md` block lane
5. mixed install 只做 fail-closed disjoint

### Phase C：same install unit replace

1. orbit install unit replace 继续沿用现有 overwrite
2. bundle install unit replace 新增 cleanup / replay 基线
3. 不支持跨 install unit replace

### Phase D：`harness template save` 回路收口

1. runtime `AGENTS.md` marker 规范化
2. 导出时剥离 / 归一化 block markers
3. 增加 mixed runtime -> harness template save 测试

---

## 12. 建议结论

这条增强路线是可行的，但第一版必须明确分层：

- source alias：降低消费者安装 source repo 的心智
- mixed install：先只做 disjoint
- replace：只做 same install unit
- `AGENTS.md`：继续专用 lane，不按普通文件冲突处理
- 变量：继续保守合并

最重要的新冻结结论是：

- **harness template install 不能把根 `AGENTS.md` 当普通文件覆盖**
- **harness template save 必须在导出时规范化 / 剥离 runtime block markers**

这样 mixed install 才既可用，又不会把导出回路做脏。
