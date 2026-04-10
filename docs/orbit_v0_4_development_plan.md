# Orbit v0.4 Unified State Model Development Plan

版本：v0.4
状态：主线完成，进入 follow-up 收口
对应 PRD：`docs/orbit_v0_4_prd.md`
对应技术方案：`docs/orbit_v0_4_technical_spec.md`
关联文档：
- `docs/orbit_brief_lane_v0_4_technical_spec.md`
- `docs/testing-strategy.md`
- `docs/orbit_template_authoring_guide.md`
- `docs/context/orbit_v0_4_design_archive.md`
- `docs/harness_centric_runtime_prd.md`（v0.3 历史背景）
- `docs/harness_centric_runtime_technical_spec.md`（v0.3 历史背景）

---

## 0. 当前收口状态

截至 `2026-04-09`：

- v0.4 主线阶段 `Phase 0` 到 `Phase 7` 已按开发计划收口完成。
- 对应主线 issue `0109`、`0110`、`0111`、`0112`、`0113`、`0114`、`0115`、`0116`、`0117`、`0118`、`0119`、`0120`、`0121`、`0122` 已全部归档到 `docs/issues/closed/`。
- 当前仍在 `docs/issues/open/` 的事项，已不再作为 v0.4 主线阻塞项，而是：
  - 旧批次遗留但不阻塞 v0.4 主线完成的 follow-up；
  - post-v0.4 的 harness authoring / mixed-install / policy 增强。

当前 open issue 的角色判断：

- `0087`：install/apply 变量声明兼容策略增强，不阻塞 v0.4 unified state model 主线。
- `0123`、`0124`、`0125`、`0126`：post-v0.4 的 harness author multi-orbit flow 增强，不属于本计划的主线交付范围。
- `0067`、`0071` 已在本次 closeout audit 中归档，不再保留为 open umbrella。

因此，本开发计划后续主要作为：

1. v0.4 已完成边界的冻结基线；
2. 评估后续 follow-up 是否应进入 `v0.5` 或更小增量版本的参考；
3. 技术债与 post-v0.4 增强的分流依据。

---

## 1. 文档目标

本文档把 v0.4 的统一状态模型拆成可执行的开发阶段，目标是明确：

1. 先做什么，后做什么；
2. 哪些兼容路径应尽早删除，而不是继续保留；
3. 每阶段应交付哪些命令、文档、测试与收口结果；
4. 下一步 issue / PR 应如何继续拆分。

---

## 2. 总体开发策略

### 2.1 Clean-break 优先

默认策略不是“先兼容再慢慢迁移”，而是：

1. 先冻结 v0.4 文档基线；
2. 直接让新实现服务新模型；
3. 若确实需要迁移，只提供显式的一跳迁移路径；
4. 迁移完成后删除兼容读写路径。

### 2.2 先收口宿主，再切行为

建议始终按下面顺序推进：

1. 先收口 revision host
2. 再收口 OrbitSpec host
3. 再收口 surface planner
4. 再切命令消费面
5. 再补作者工作流与发布路径

### 2.3 先把作者工作流补对称，再做高级增强

比起继续堆 install / export 细枝末节，更应该先把下面两条路径做对称：

- direct `orbit_template` authoring
- `source -> orbit_template` authoring

因为这两条路径会直接决定：

- brief 如何编辑；
- publish 如何校验；
- export 与 orchestration 如何分层。

### 2.4 文档与入口必须同步切换

这一轮不是“代码先改，文档以后补”。

每切完一个正式边界，都需要同步更新：

- `AGENTS.md`
- 主文档
- help / quickstart / examples
- acceptance fixtures

---

## 3. 阶段总览

建议分为 8 个阶段：

1. Phase 0：文档基线与归档收口
2. Phase 1：Manifest 统一与 revision taxonomy 切换
3. Phase 2：OrbitSpec host 统一与 loader/writer 切换
4. Phase 3：Surface planner 切换与命令消费分层
5. Phase 4：Brief lane foundations
6. Phase 5：Authoring publish 主链路
7. Phase 6：Runtime export / writeback / harness template
8. Phase 7：清理、迁移、测试与发布硬化

---

## 4. Phase 0：文档基线与归档收口

### 4.1 目标

先把 v0.4 的 source of truth 建起来，让后续实现有统一入口。

### 4.2 任务

1. 归档旧 proposal / exploratory 文档到 `docs/context/`
2. 新建 v0.4 PRD / technical spec / development plan
3. 更新 `AGENTS.md` 的 source of truth 顺序
4. 更新 active guide 与 specialized supplement 的引用关系
5. 明确哪些旧文档是历史背景，哪些是正式主线

### 4.3 测试与检查

1. 关键入口文档路径无断链
2. 新主线文档能独立回答状态模型与命令模型
3. 开发入口不再把旧 proposal 当 source of truth

### 4.4 完成标准

- v0.4 主文档已存在
- 旧设计文稿已归档
- 仓库入口口径已切到 v0.4

---

## 5. Phase 1：Manifest 统一与 Revision Taxonomy 切换

### 5.1 目标

让 `.harness/manifest.yaml` 成为唯一 revision identity 控制面。

### 5.2 任务

1. 扩展 manifest schema，完整覆盖：
   - `runtime`
   - `source`
   - `orbit_template`
   - `harness_template`
2. 切 branch classifier / inspect / root resolution
3. 为 source / template / runtime bootstrap 提供统一 manifest writer
4. 删除主链路对以下文件的依赖：
   - `.orbit/source.yaml`
   - `.orbit/template.yaml`
   - `.harness/runtime.yaml`
   - `.harness/template.yaml`
5. 明确 `plain` 只作为 classifier 结果，而不是 manifest kind

### 5.3 测试

1. manifest per-kind round-trip
2. mixed-kind validation
3. branch inspect / classify golden tests
4. create/init/template bootstrap integration tests

### 5.4 完成标准

- revision kind 已可只靠 manifest 识别
- 主链路命令不再要求旧 marker 文件

---

## 6. Phase 2：OrbitSpec Host 统一与 Loader / Writer 切换

### 6.1 目标

让 `.harness/orbits/<orbit-id>.yaml` 成为所有正式 revision kind 的 authored OrbitSpec host。

### 6.2 任务

1. 切 Orbit loader / writer / validator 到 hosted `.harness/orbits/*.yaml`
2. 统一 `orbit add` / `show` / `list` / `validate`
3. 去掉新主链路对 `.orbit/orbits/*.yaml` 与 `.orbit/config.yaml` 的依赖
4. 若需要迁移，提供显式迁移工具或 fixture rewrite 脚本
5. 冻结 `meta.file` 与 host path 一致性规则

### 6.3 测试

1. hosted OrbitSpec parser tests
2. add/show/list/validate integration tests
3. migration fixture tests

### 6.4 完成标准

- 所有正式 revision kind 都以 `.harness/orbits/*.yaml` 为 authored truth
- Orbit 命令对旧 host 的依赖已退出主路径

---

## 7. Phase 3：Surface Planner 切换与命令消费分层

### 7.1 目标

正式把 `projection / orbit_write / export / orchestration` 四个 surface 落为统一 planner。

### 7.2 任务

1. 固化 role -> surface 映射
2. 把 `projection_plan.go` 升级为四面 planner，而不是只产出 projection 路径
3. 更新 `path_classification.go` 与 status 输出，显式展示 surface 信息
4. 切换命令消费面：
   - `enter/leave/files/current` -> `projection`
   - `diff/log/commit/restore` -> `orbit_write`
   - `template save/publish` -> `export`
   - `brief materialize/backfill` -> `orchestration`
5. 更新 ledger 生成逻辑，让 file inventory 与 git state 体现新 surface 语义

### 7.3 测试

1. role -> surface matrix tests
2. tracked / untracked classification tests
3. command-consumer regression tests
4. ledger snapshot tests

### 7.4 完成标准

- 每个正式命令都有唯一明确的 surface 来源
- 不再依赖单个模糊的 “in scope” 布尔值驱动所有行为

---

## 8. Phase 4：Brief Lane Foundations

### 8.1 目标

先把 brief lane 收口为 `runtime / source / orbit_template` 共享的 `orchestration` lane，并把它从 publish / export / writeback 中彻底解耦。

### 8.2 任务

1. 先完成 brief lane 的 revision gating 与 orbit targeting
2. 再实现 `orbit brief materialize` 的 block-preserving container patch
3. 再硬化 `orbit brief backfill` 的 revision matrix 与 local-only write contract
4. 补 brief drift diagnostics 或最小 `--check` 能力
5. 固化 `runtime / source / orbit_template` 三态下统一的命令语义
6. 明确 runtime 下的 brief 更新仍然只停留在本地 structured truth，不自动进入 export/writeback/publish

推荐 issue 顺序：

1. `ISSUE-0119` revision gating 与 target resolution
2. `ISSUE-0120` `orbit brief materialize` 与 block-preserving container patch
3. `ISSUE-0121` `orbit brief backfill` 的 revision matrix 与 local-only write contract
4. `ISSUE-0122` brief lane drift diagnostics
5. `ISSUE-0113` 作为 umbrella issue 收口验收

### 8.3 测试

1. brief materialize / backfill round-trip
2. `runtime/source/orbit_template` 三态 brief lane integration tests
3. revision gating / orbit targeting tests
4. container patch preserve tests
5. local-only backfill tests
6. diagnostics / drift-state tests

### 8.4 完成标准

- `runtime/source/orbit_template` 三态下 brief lane 语义一致
- brief lane 的副作用始终只属于 `orchestration`
- runtime 下的 brief 更新不会被误认为 template publish
- root `AGENTS.md` 的 container patch 与 backfill contract 已冻结

---

## 9. Phase 5：Authoring Publish 主链路

### 9.1 目标

建立在稳定 brief lane 之上，打通 direct template authoring 与 source authoring 的正式 publish 闭环。

### 9.2 任务

1. 支持 `orbit_template` revision 直接 publish
2. 支持 `source -> orbit_template` publish
3. publish 前验证 installable orbit template 不包含根 `AGENTS.md`
4. 若 materialized brief 与 `meta.agents_template` 不一致，给出显式 backfill 路径
5. 更新作者文档与帮助文本

推荐 issue 顺序：

1. `ISSUE-0114` orbit template direct publish
2. `ISSUE-0115` source publish pipeline

### 9.3 测试

1. orbit_template publish integration tests
2. source publish integration tests
3. published payload excludes root `AGENTS.md` tests
4. brief drifted 时的 publish gating tests

### 9.4 完成标准

- 开发者可以直接维护 `orbit_template`
- 复杂 orbit 也可以通过 `source` 开发与发布
- brief 编辑入口与发布校验已对齐

---

## 10. Phase 6：Runtime Export / Writeback / Harness Template

### 10.1 目标

把 runtime 侧优化与正式导出链路收口清楚。

### 10.2 任务

1. 让 `orbit template save` 明确只消费 `export` surface
2. 定义并实现 runtime -> orbit template writeback / publish 主路径
3. 对 runtime provenance 做同步更新
4. 硬化 `harness template save`
5. 让 bundle record 与 runtime member source 一致化

推荐 issue 顺序：

1. `ISSUE-0116` runtime export / writeback
2. `ISSUE-0117` harness template hardening

### 10.3 测试

1. runtime export surface tests
2. runtime writeback integration tests
3. harness template save / install smoke tests
4. provenance consistency tests

### 10.4 完成标准

- runtime 中的 orbit 优化可以通过正式 export/writeback 路径进入 template
- harness template 打包与安装语义稳定

---

## 11. Phase 7：清理、迁移、测试与发布硬化

### 11.1 目标

删除过渡层，让 v0.4 真正成为仓库主线。

### 11.2 任务

1. 删除或彻底隐藏旧兼容入口
2. 移除旧 marker 文件的主链路读写
3. 重写 quickstart / help / examples / acceptance fixtures
4. 明确是否保留一跳迁移工具
5. 完成 release notes 与技术债登记

推荐 issue：

1. `ISSUE-0118` docs/help/acceptance cleanup

### 11.3 测试

1. acceptance smoke 覆盖四类 revision kinds
2. docs-driven quickstart smoke
3. clean repo / migrated repo 两条基本路径验证

### 11.4 完成标准

- 旧兼容层不再出现在正式开发与用户入口
- v0.4 文档、CLI、测试、示例已经一致

---

## 11. 推荐的 Issue / PR 拆分

建议按下列粒度拆 issue：

1. manifest schema 与 classifier
2. hosted OrbitSpec cutover
3. four-surface planner
4. status / ledger / classification cutover
5. brief lane umbrella
6. brief revision gating / target resolution
7. brief materialize container patch
8. brief backfill local-only write
9. brief drift diagnostics
10. orbit_template direct publish
11. source publish pipeline
12. runtime export / writeback
13. harness template hardening
14. docs/help/acceptance cleanup

建议每个主阶段拆成 1 到 3 个 PR，不跨阶段混合。

---

## 12. 开发前置说明

后续开始具体编码前，建议遵守以下顺序：

1. 先把 manifest 与 hosted OrbitSpec 这两个宿主收口
2. 再改 planner 与 command consumers
3. 再把 brief lane 单独收口为 shared orchestration lane
4. 再补 authoring publish 与 runtime export handoff
5. 最后做 cleanup 与 release hardening

如果在实现中发现现有代码与 v0.4 文档冲突，优先更新文档或明确记入技术债，不要继续隐式扩散旧模型。
