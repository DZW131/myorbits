# Orbit v0.4 Design Archive

状态：archive index
当前主线：
- `docs/orbit_v0_4_prd.md`
- `docs/orbit_v0_4_technical_spec.md`
- `docs/orbit_v0_4_development_plan.md`

---

## 1. 文档目的

本文档用于说明：

1. 哪些旧设计文稿已经归档；
2. 它们与 v0.4 正式主文档的关系是什么；
3. 后续开发时哪些文档应当优先读，哪些只作为背景参考。

---

## 2. 已归档的设计文稿

以下文档已从活跃设计主线退到 `docs/context/`，作为背景材料保留：

- `docs/context/orbit_state_and_workflow_unification.md`
- `docs/context/orbit_content_and_state_optimization.md`
- `docs/context/orbit_member_filesystem_behavior.md`
- `docs/context/harness_single_control_plane_proposal.md`
- `docs/context/harness_single_control_plane_technical_spec.md`
- `docs/context/harness_single_control_plane_development_plan.md`

这些文档仍有价值，但它们的职责已经变成：

- 保留设计讨论过程；
- 解释 v0.4 若干结论的来源；
- 为后续实现提供历史背景。

它们不再与 v0.4 主文档并列作为正式 source of truth。

---

## 3. 仍可参考的历史主线

以下文档保留在 `docs/`，但在 v0.4 之后主要作为历史背景或专项补充：

- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/harness_centric_runtime_development_plan.md`
- `docs/orbit_member_runtime_technical_spec.md`
- `docs/orbit_member_runtime_development_plan.md`

建议理解如下：

- `harness_centric_runtime_*`：v0.3 主线背景
- `orbit_member_runtime_*`：围绕 orchestration / ledger / member model 的专项补充

若它们与 v0.4 主文档冲突，以 v0.4 主文档为准。

---

## 4. 后续阅读顺序

后续开发建议按下面顺序阅读：

1. `docs/orbit_v0_4_prd.md`
2. `docs/orbit_v0_4_technical_spec.md`
3. `docs/orbit_v0_4_development_plan.md`
4. `docs/testing-strategy.md`
5. `docs/orbit_template_authoring_guide.md`（涉及模板作者工作流时）
6. `docs/orbit_member_runtime_technical_spec.md` / `docs/orbit_member_runtime_development_plan.md`（涉及 orchestration / ledger 时）
7. 归档设计文稿

---

## 5. 归档规则

从 v0.4 开始，新增高频设计文档仍应先放在 `docs/`。

当某份文档不再是活跃 source of truth，而只是背景材料时，再移动到 `docs/context/`。
