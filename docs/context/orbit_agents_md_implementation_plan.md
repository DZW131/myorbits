# Orbit AGENTS.md 实施规划（V0.2 历史规划）

版本：V0.2  
状态：historical implementation note；post-v0.3 不再作为当前实施入口  
关联文档：`docs/context/orbit_agents_md_development.md`

---

## 1. 目标

本文档把 `docs/context/orbit_agents_md_development.md` 中已经冻结的设计，进一步整理成一份可执行的开发工作参考。

说明：

- 本文档对应的是已经完成的 V0.2 AGENTS shared-file lane；
- post-v0.3 若继续推进 `AGENTS.md`，应改读 `docs/orbit_member_runtime_technical_spec.md` 与 `docs/orbit_member_runtime_development_plan.md`；
- 本文档保留的意义主要是解释旧实现为何存在，而不是指导新的 AGENTS cutover。

它不重新定义产品语义，而是回答 3 个问题：

1. 当前代码需要在哪些边界上改动；
2. 应该按什么顺序推进，才能减少返工；
3. 每一段改动应该用什么类型的测试先约束行为。

---

## 2. 已冻结前提

以下前提已在设计文档中确认，本实施规划直接继承：

1. `AGENTS.md` 不是普通 owned file，而是专用 shared-file lane。
2. V0.2 first cut 只支持 `replace-block + create-if-absent`。
3. 模板态根目录 `AGENTS.md` 是 plain payload，不带 marker。
4. 运行态 `AGENTS.md` 使用 HTML comment + XML 风格属性 marker。
5. `.orbit/template.yaml` 通过 `shared_files` 显式声明 AGENTS shared entry。
6. 异常 marker 场景全部 fail-closed。
7. `orbit validate` 后续需要增加运行态 `AGENTS.md` 结构校验。
8. `include_unmarked_content: true` 的默认 save 语义是已接受风险。

---

## 3. 当前代码边界

### 3.1 直接受影响的模块

- `cmd/orbit/cli/template/manifest.go`
  - 目前还没有 `shared_files`
- `cmd/orbit/cli/template/content_builder.go`
  - 目前会把根目录 `AGENTS.md` 当普通模板文件
- `cmd/orbit/cli/template/save.go`
  - 目前没有 runtime `AGENTS.md` 抽取逻辑
- `cmd/orbit/cli/template/apply.go`
  - 目前是文件级 render / write 回路
- `cmd/orbit/cli/commands/validate.go`
  - 目前没有 `AGENTS.md` runtime marker 校验

### 3.2 最容易误伤主线的点

1. `AGENTS.md` 同时走普通文件和 shared-file 两条路径，导致重复写入或行为冲突。
2. 在 apply 路径里直接复用普通文件覆盖逻辑，错误覆盖运行态 `AGENTS.md`。
3. 在 validate 里对普通 repo 的自定义 `AGENTS.md` 误报。

### 3.3 本轮不做的事

- 泛化多 shared-file 平台
- 通用 block merge engine
- 自动修复 malformed marker
- 把 `AGENTS.md` 重新纳入普通 owned-file 模型

---

## 4. 建议的开发顺序

### 4.1 ISSUE-0030：Manifest And Source Routing

先扩展 manifest 合同，并打通 source-routing 分流。

建议先做这个阶段的原因：

- 它是 save / apply / validate 三条链路的总开关；
- 不先分流，后续所有行为都会继续把 `AGENTS.md` 当普通文件处理；
- 改动集中在 schema、validation、source loading，风险最可控。

测试优先级：

- manifest parse / validate 单测
- local / remote source loader 的回归测试

### 4.2 ISSUE-0031：Runtime Parser And Validator

建立 `AGENTS.md` runtime parser / validator 原语。

建议保持这一层为纯逻辑模块，不依赖 git、state、command。  
save / apply / validate 都复用它，不在各自路径里复制语义。

测试优先级：

- parser 正常结构单测
- malformed / nested / duplicate / mismatch 全覆盖单测

### 4.3 ISSUE-0032：Template Save Extraction

在 `template save` 中接入 AGENTS 专用抽取。

关键行为：

- 保留 unmarked span
- 保留当前 orbit block 内部内容
- 删除其他 orbit block
- 缺少当前 orbit block 时 warning 后继续
- 空 payload 时不生成 `shared_files` entry，也不写模板态 `AGENTS.md`

测试优先级：

- service 层 preview / save 单测
- `template save` CLI 集成测试
- `--edit-template` 路径回归测试

### 4.4 ISSUE-0033：Template Apply Merge

在 `template apply` 中接入 AGENTS 专用 merge writer。

关键行为：

- 文件不存在则创建
- 同 orbit block 存在则原地替换并 warning
- 同 orbit block 不存在则 EOF 追加
- malformed runtime `AGENTS.md` 阻断 apply

测试优先级：

- service 层 apply preview / apply 单测
- local apply 集成测试
- remote apply 集成测试

### 4.5 ISSUE-0034：Validate Runtime Markers

把 runtime marker 校验接到 `orbit validate`。

建议：

- 触发条件保持保守
- 只有在仓库显式使用 AGENTS shared lane 时才检查
- 输出继续沿用现有 validate 风格

测试优先级：

- `orbit validate` 集成测试
- 错误摘要和退出语义测试

### 4.6 ISSUE-0035：Integration And Doc Sync

最后收口：

- 跨 save / apply / validate 的端到端测试
- help / quickstart / debt 同步
- 文档改成“已实现合同”

---

## 5. 推荐的代码切分方式

建议保持“新能力集中、命令层保持薄”的 repo 风格。

推荐切分：

- `cmd/orbit/cli/template/manifest.go`
  - 负责 `shared_files` schema 和 validation
- `cmd/orbit/cli/template/agents.go` 或 `agents_runtime.go`
  - 负责 runtime parser / validator / marker render helper
- `cmd/orbit/cli/template/save.go`
  - 负责 AGENTS save extraction 编排
- `cmd/orbit/cli/template/apply.go`
  - 负责 AGENTS apply merge 编排
- `cmd/orbit/cli/commands/validate.go`
  - 只做 validate 接线，不承载解析细节

不建议：

- 把 runtime marker 解析塞进 command 层
- 在 `apply.go` 中直接用字符串拼接完成所有逻辑而没有独立 parser
- 为了这次能力引入通用 shared-file 抽象层

---

## 6. TDD 建议

### 6.1 先红再绿的最小粒度

推荐每个 issue 都按下面顺序推进：

1. 先补最相关单元测试或集成测试
2. 明确初始失败点与目标行为一致
3. 写最小实现让红测转绿
4. 在测试保护下再做必要小重构

### 6.2 建议的测试层次

- schema / parser：单元测试优先
- save / apply：service 层单测 + CLI 集成测试
- validate：CLI 集成测试优先
- local / remote：只在行为确实不同的地方分别覆盖

### 6.3 关键回归用例

至少要覆盖：

- shared `AGENTS.md` 不再作为普通 owned file 重复处理
- save 保留 unmarked 内容
- save 遇到缺失当前 orbit marker 时 warning 后继续
- apply 在 replace / append / create 三条路径上行为稳定
- malformed runtime `AGENTS.md` 会阻断 apply 和 validate
- local 与 remote apply 在 AGENTS lane 上语义一致

---

## 7. 风险与缓解

### 7.1 已接受风险

`include_unmarked_content: true` 可能把 repo-level prose 带入多个模板，后续 repeated apply 需要人工整理。

处理策略：

- 通过测试固定当前行为
- 继续保留在 `docs/technical-debt.md`

### 7.2 高风险实现点

1. 普通文件路径与 AGENTS shared lane 未完全分离
2. apply 仍落到普通 overwrite 语义
3. validate 触发范围过宽

缓解方式：

- 先做 0030 与 0031
- 在 0032 / 0033 中明确测试“不走普通路径”
- validate 只在显式使用 shared lane 时触发

---

## 8. 建议的交付方式

推荐按 issue 顺序逐个提交，不建议把全部功能挤在一个超大提交里。

推荐节奏：

1. `0030 + 0031`
2. `0032`
3. `0033`
4. `0034`
5. `0035`

若中途发现 `validate` 触发条件仍有明显歧义，可以把 `0034` 稍后，但不建议跳过 `0030 / 0031` 直接做 save / apply。

---

## 9. 本文档用途

本文档用于帮助后续开发时快速回答：

- 下一步应该先动哪一层
- 哪些文件值得改，哪些地方不要先动
- 每一段实现应该先补什么测试

它不替代 `docs/context/orbit_agents_md_development.md` 的产品合同；若两者冲突，应以产品合同为准，并先更新本文档。
