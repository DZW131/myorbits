# Harness Single Control Plane Development Plan

版本：v0.1
状态：proposal
阶段：post-v0.3 exploration
对应提案：`docs/context/harness_single_control_plane_proposal.md`
对应技术方案：`docs/context/harness_single_control_plane_technical_spec.md`
关联文档：
- `docs/orbit_member_runtime_technical_spec.md`
- `docs/testing-strategy.md`
- `docs/harness_centric_runtime_technical_spec.md`

---

## 1. 文档目标

本文档把单目录控制面方案拆成可执行阶段，回答：

1. 先做哪些底层切换；
2. 哪些代码面需要一起切；
3. 哪些兼容策略明确不做；
4. 每阶段的测试和完成标准是什么；
5. 对应 issue 应如何拆分。

---

## 2. 开发原则

整个推进过程遵守以下原则：

1. 先冻结文档，再切路径宿主与 loader。
2. 先收口 `.harness/manifest.yaml`，再切命令行为。
3. 不做 `.orbit/*` 与 `.harness/*` 双写。
4. 不把 repo-local `.git/orbit/state/*` 并入版本化控制面。
5. 先切 authoring / branch classification，再切 projection kernel。
6. 文档、help、测试与命令面同步切换，不留长期“旧路径示例”。

---

## 3. 总体策略

建议分为五个阶段：

1. 文档冻结与 schema primitives
2. OrbitSpec 宿主切换与 authoring 命令切换
3. manifest-based branch classification 与 runtime bootstrap
4. projection / template / install 主链路切换
5. 测试、帮助、文档与可选迁移收尾

其中默认策略如下：

1. steady-state 不读 `.orbit/config.yaml`
2. steady-state 不读 `.orbit/orbits/*.yaml`
3. 如需迁移旧仓库，放到显式后置阶段处理

---

## 4. 阶段总览

### Phase 1：文档冻结与 Schema Primitives

目标：

- 冻结 proposal、technical spec、development plan
- 引入 `.harness/manifest.yaml` 与 `.harness/orbits/*` 的 schema / path host helper

任务：

1. 新增 manifest schema、validator、codec。
2. 新增 `.harness/orbits/` path helpers。
3. 定义 runtime / orbit_template / harness_template 三种 `kind`。
4. 定义新的 branch classifier primitives。

测试：

1. manifest codec 单测
2. per-kind validator 单测
3. path helper 单测

完成标准：

- 代码中已有新的 manifest 领域模型
- 可以独立读写 `.harness/manifest.yaml`
- 不触碰现有命令主链路

### Phase 2：OrbitSpec 宿主切换与 Authoring Commands

目标：

- 把 OrbitSpec 的 steady-state 宿主切到 `.harness/orbits/*.yaml`
- 让 authoring 命令不再依赖 `.orbit/config.yaml`

任务：

1. 更新 OrbitSpec path host 与 `meta.file` 校验。
2. 切换 control loader 到 `.harness/orbits/*.yaml`。
3. 切换：
   - `orbit add`
   - `orbit show`
   - `orbit list`
   - `orbit validate`
4. 删除或隔离旧 global config loader 在 authoring 命令上的依赖。

测试：

1. add/show/list/validate 集成测试
2. OrbitSpec path / meta.file 校验测试
3. 非法或缺失 manifest/orbit definitions 测试

完成标准：

- authoring 命令只读写 `.harness/orbits/*`
- steady-state 不再要求 `.orbit/config.yaml`

### Phase 3：Manifest-Based Classification 与 Runtime Bootstrap

目标：

- 用 `.harness/manifest.yaml` 统一表达 branch 顶层身份
- 让 bootstrap 命令落到新合同

任务：

1. `harness init` / `harness create` 改写 `manifest.kind=runtime`。
2. branch classifier 改为只看 manifest。
3. `harness root` / `inspect` 改为基于 manifest。
4. 明确 `orbit init` 的兼容策略：
   - hidden wrapper
   - 或稳定失败并给迁移提示

测试：

1. runtime / orbit template / harness template 分类测试
2. zero-member runtime 测试
3. init/create/root/inspect 集成测试

完成标准：

- runtime / template 顶层身份只由 manifest 决定
- bootstrap 不再写 `.orbit/*`

### Phase 4：Projection / Template / Install 主链路切换

目标：

- 完成运行态命令、模板导出、模板安装对新控制面的切换

任务：

1. `enter/files/status/current/diff/log/commit/restore`
   - 只在 `kind=runtime` 下运行
   - 只基于 OrbitSpec + role-aware scope
2. 移除 `shared_scope` / `projection_visible` / `behavior` 的 steady-state 依赖。
3. `orbit template save`
   - 输出 `kind=orbit_template`
4. `harness template save`
   - 输出 `kind=harness_template`
5. `harness install`
   - 从新模板合同读取
   - 写回 runtime manifest / vars / installs

测试：

1. projection commands runtime-only gating
2. role-aware scope 行为回归
3. orbit template save 集成测试
4. harness template save 集成测试
5. harness install 集成测试

完成标准：

- 主要命令链路已不再依赖旧 `.orbit/*` 控制文件
- 模板读写都走新 manifest 合同

### Phase 5：收尾与可选迁移

目标：

- 清理帮助文档、测试矩阵与遗留引用
- 评估是否需要显式迁移工具

任务：

1. 更新 quickstart / help / release / testing docs。
2. 删除或标记旧 `.orbit/*` 示例。
3. 增补 acceptance matrix：
   - runtime
   - orbit template
   - harness template
4. 可选：
   - 评估 `harness migrate-control-plane` 一次性迁移命令

测试：

1. help / docs 引用回归
2. acceptance smoke
3. 若实现迁移器，再补迁移前后 fixture 测试

完成标准：

- 文档与实现口径一致
- 新控制面有完整 acceptance 回归网

---

## 5. 代码边界建议

### 5.1 主要新增 / 改造区域

重点触达：

- `cmd/orbit/cli/harness/*`
- `cmd/orbit/cli/orbit/*`
- `cmd/orbit/cli/branchinfo/*`
- `cmd/orbit/cli/commands/*`

### 5.2 优先复用的稳定内核

原则上尽量复用：

- `cmd/orbit/cli/git`
- `cmd/orbit/cli/state`
- `cmd/orbit/cli/ids`
- `cmd/orbit/cli/scoped`
- `cmd/orbit/cli/view`

因为本方案主要改的是：

- 版本化控制面宿主
- loader
- classification
- command wiring

而不是 Git-native projection 基础设施。

---

## 6. 风险与收口点

### 6.1 最大风险

1. branch classifier 与 loader 同时切换，容易让模板态 / 运行态误判。
2. 去掉 `.orbit/config.yaml` 后，旧 path overlay 语义必须由 role-aware model 完整接住。
3. `orbit template save` 与 `harness install` 的 manifest 合同若不同步，模板闭环会断。

### 6.2 收口策略

1. 先冻结 manifest schema，再切 classifier。
2. 先切 authoring commands，再切 runtime commands。
3. 先让 orbit template / harness template 都能稳定写出新 manifest，再切 install 读取。
4. 用真实 temp repo integration tests 保护三类 branch 形态。

---

## 7. Issue 拆分建议

建议至少拆成以下五个 issue：

1. manifest schema 与 host path helpers
2. OrbitSpec 宿主切换与 authoring 命令
3. manifest-based runtime bootstrap 与 branch classification
4. runtime kernel / template / install 主链路切换
5. docs / help / acceptance hardening 与可选迁移评估

---

## 8. 一句话总结

单目录控制面实现的正确推进顺序应是：

**先立 manifest 与 OrbitSpec 新宿主，再切 authoring 与 branch classification，最后切 runtime/template/install 主链路，并用 docs/help/tests 一起收尾。**
