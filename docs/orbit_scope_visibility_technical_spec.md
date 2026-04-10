# Orbit Scope Visibility Technical Spec

版本：v0.1  
状态：proposal  
阶段：post-v0.3 enhancement  
关联文档：
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/orbit_template_publish_technical_spec.md`
- `docs/testing-strategy.md`
- `docs/context/mvp-product-requirements.md`
- `docs/context/mvp-technical-architecture.md`

---

## 1. 文档目标

冻结 Orbit 当前 `shared_scope` / `projection_visible` 语义与 scope 分层方案，解决两个已经暴露出来的问题：

1. 历史 `always_visible` 命名已经不足以表达真实行为，且现已被 `shared_scope` 取代；
2. 当前实现里“投影可见”与“模板拥有”基本绑定，缺少 projection-only 文件层。

这份文档定义两项变化：

- `always_visible` 更名为 `shared_scope`
- 新增 `projection_visible`

同时明确它们对 projection、status、scoped operations、template save/publish 的影响边界。

---

## 2. 当前问题

当前代码基线里，repo 级共享 scope 不只是“可见”。

它目前同时参与：

- orbit tracked scope 解析
- untracked path 的 orbit 归属判断
- sparse projection
- status in-scope / out-of-scope 分类
- `orbit template save`
- `orbit template publish`
- `harness template save` 中 member candidate 的文件输入

当前链路见：

- `cmd/orbit/cli/orbit/resolve.go`
- `cmd/orbit/cli/orbit/match.go`
- `cmd/orbit/cli/template/save.go`
- `cmd/orbit/cli/template/content_builder.go`
- `cmd/orbit/cli/harness/template_candidate.go`

这意味着：

- repo 级“跨 orbit 可见”文件会自动进入 orbit template
- 当前没有正式机制表达“只在投影里可见，但不属于模板态”

这已经和作者仓库 / source branch 的使用心智发生冲突。

---

## 3. 设计目标

本次方案的目标是：

1. 保留 repo 级共享 scope 的能力；
2. 让 repo 级共享 scope 可以被 orbit `exclude` 排除；
3. 引入 projection-only 文件层；
4. 明确哪些文件属于 orbit template owned content；
5. 不把 source branch 的作者说明文件、辅助文件自动带进 template branch；
6. 在不改 Git-native projection 内核前提下，尽量局部收口。

非目标：

1. 不引入工作区外服务或索引层；
2. 不改变 `.harness/*` runtime host 边界；
3. 不改变 `harness install` 的 template 识别方式；
4. 不把 `.orbit/source.yaml` 变成可投影业务文件。

---

## 4. 已冻结结论

### 4.1 配置字段

`GlobalConfig` 从：

```yaml
always_visible:
```

改为：

```yaml
shared_scope:
projection_visible:
```

建议合同：

```yaml
version: 1
shared_scope:
  - LICENSE
projection_visible:
  - README.md
behavior:
  outside_changes_mode: warn
  block_switch_if_hidden_dirty: true
  commit_append_trailer: true
  sparse_checkout_mode: no-cone
```

### 4.2 语义分层

- `include`
  - orbit-owned
  - template-owned
  - scoped-ops-owned
- `shared_scope`
  - repo 级共享 owned scope
  - template-owned
  - scoped-ops-owned
  - 可以被 orbit `exclude` 排除
- `projection_visible`
  - 只影响 orbit projection 和 status
  - 不进入 orbit template save / publish
  - 不进入 harness template member candidate
  - 不进入 orbit scoped operations

### 4.3 优先级

产品口径可以表达为：

```text
exclude > include > shared_scope
exclude > projection_visible
```

实现上更准确的集合公式是：

```text
owned = (include ∪ shared_scope) - exclude
projection_only = (projection_visible - exclude) - owned
```

也就是说：

- `shared_scope` 不再能绕过 `exclude`
- `projection_visible` 只补充 projection 层，不补充 owned 层

### 4.4 `projection_visible` 的正式边界

`projection_visible` 影响：

- sparse projection
- `orbit enter`
- `orbit status`
- `orbit files` 的默认投影结果
- untracked path 的 in-scope 判定
- hidden-dirty gate 中“哪些 tracked paths 不会被隐藏”的判定

`projection_visible` 不影响：

- `orbit template save`
- `orbit template publish`
- `harness template save`
- `orbit diff`
- `orbit log`
- `orbit commit`
- `orbit restore`

这是本次方案里最重要的行为收口。

---

## 5. 新的 scope 模型

### 5.1 ScopeSet 合同

建议把当前 `ScopeSet` 收成四层：

```go
type ScopeSet struct {
    ControlReadPaths      []string
    OwnedPaths            []string
    ProjectionOnlyPaths   []string
    CompanionPaths        []string
    ScopedOperationPaths  []string
    ProjectionPaths       []string
}
```

含义：

- `OwnedPaths`
  - orbit definition 真正拥有的用户文件
  - 由 `include + shared_scope - exclude` 得到
  - template save/publish 的普通文件输入
- `ProjectionOnlyPaths`
  - 只投影、不拥有的文件
  - 由 `projection_visible - exclude - owned` 得到
- `ScopedOperationPaths`
  - orbit scoped ops 的 pathspec scope
  - `OwnedPaths + CompanionPaths`
- `ProjectionPaths`
  - 实际 sparse projection scope
  - `OwnedPaths + ProjectionOnlyPaths + CompanionPaths`

### 5.2 集合公式

对一个 orbit 和 tracked files 集合，定义：

```text
owned_tracked =
  tracked_non_control_plane
  ∩ (include ∪ shared_scope)
  - exclude

projection_only_tracked =
  tracked_non_control_plane
  ∩ projection_visible
  - exclude
  - owned_tracked

scoped_operation_paths =
  owned_tracked ∪ companion_paths

projection_paths =
  owned_tracked ∪ projection_only_tracked ∪ companion_paths
```

### 5.3 untracked path 判定

`PathMatchesOrbit(...)` 也要同步改成：

```text
in_scope_untracked =
  ((include ∪ shared_scope ∪ projection_visible) - exclude)
```

它服务：

- `orbit status`
- `orbit enter` 的 outside-untracked warning

这里的 `in scope` 是“当前 orbit 投影视图里会被视为相关”，不是“template-owned”。

---

## 6. 命令行为变化

### 6.1 `orbit enter`

保持：

- sparse-checkout 仍按 `ProjectionPaths`
- hidden-dirty gate 仍按 `ProjectionPaths`

变化：

- `projection_visible` 命中的 tracked 文件会被投影出来
- 这些文件不会因为“只属于 projection-only”而被隐藏

### 6.2 `orbit status`

保持：

- tracked path 的 in-scope / out-of-scope 继续按当前投影视图判断
- untracked path 继续走 pattern match 逻辑

变化：

- `projection_visible` 命中的文件会被视为 in-scope
- 这只影响 status，不代表这些文件属于 template-owned 或 scoped commit 范围

### 6.3 `orbit files`

第一版建议：

- 继续默认输出 `ProjectionPaths`
- 不在第一轮改 text 输出语义

原因：

- 用户运行 `orbit files` 时，更关心“当前 orbit 会看到什么”
- 这与 sparse projection 一致

若需要增强可观测性，后续再考虑：

- `orbit files --json` 扩展为同时输出 `projection_files` / `scoped_files`
- 或新增 `orbit files --kind projection|scoped`

这不是本轮必须项。

### 6.4 `orbit diff/log/commit/restore`

这是本次方案和现有实现最大的分叉点。

当前代码里，这些 scoped operations 复用的是 `ProjectionPaths`。

本次方案要求改为：

- `diff/log/commit/restore` 走 `ScopedOperationPaths`
- 即：`OwnedPaths + CompanionPaths`

因此：

- `projection_visible` 文件能被看见
- 但不会进入 scoped commit / restore / diff / log

这正是“投影可见，但不算进模板态/owned scope”的核心价值。

### 6.5 `orbit template save` / `orbit template publish`

改为只消费：

- `OwnedPaths`

不再消费：

- `ProjectionOnlyPaths`

这意味着：

- `shared_scope` 仍会进入 orbit template
- `projection_visible` 不会进入 orbit template

### 6.6 `harness template save`

`cmd/orbit/cli/harness/template_candidate.go` 当前也复用 `UserDataPaths`。

本次方案下应改为：

- member candidate 只消费 `OwnedPaths`

因此：

- `projection_visible` 不会进入 harness template member 文件集
- 这与 orbit template save/publish 保持一致

---

## 7. 配置校验规则

### 7.1 `shared_scope`

`shared_scope` 使用业务型 pattern 校验，语义为：

- 只能命中业务文件
- 不允许命中 control-plane / runtime-plane 路径

至少要拒绝命中：

- `.orbit/config.yaml`
- `.orbit/orbits/*.yaml`
- `.orbit/template.yaml`
- `.orbit/source.yaml`
- `.harness/**`
- `.git/orbit/state/**`

### 7.2 `projection_visible`

`projection_visible` 也应使用与 `shared_scope` 相同的路径安全校验：

- 允许业务文件
- 不允许命中 control-plane / runtime-plane 路径

原因：

- projection-visible 文件仍会进入 sparse view 和 status
- 若允许它命中 `.orbit/*` 或 `.harness/*`，会重新污染控制平面边界

### 7.3 废弃字段处理

产品尚未发布，不保留历史兼容。

- `always_visible` 视为废弃字段
- 解析时遇到 `always_visible` 必须 fail-closed
- 配置必须改用 `shared_scope`
- 所有默认输出与文档示例只使用 `shared_scope`

---

## 8. 默认配置建议

历史默认配置曾把 `README.md` 放进 repo 级共享 scope。

在新模型下，这个默认值不再合适。

建议改为：

```yaml
version: 1
shared_scope: []
projection_visible: []
behavior:
  outside_changes_mode: warn
  block_switch_if_hidden_dirty: true
  commit_append_trailer: true
  sparse_checkout_mode: no-cone
```

原因：

- 默认把 `README.md` 共享到所有 orbit，会继续放大“作者说明文件误入 template”问题
- 让 repo 显式声明共享 scope，比隐式默认更清晰

---

## 9. 特殊文件约束

### 9.1 `AGENTS.md`

当前 `AGENTS.md` 已经有特殊 lane，不走普通模板文件通道。

本次方案补一条明确边界：

- 若 `AGENTS.md` 只命中 `projection_visible`
  - 它可以在 orbit projection 中出现
  - 但不应进入 orbit template save/publish 的 special lane

因此：

- `AGENTS.md` 是否进入 template，仍然取决于它是否属于 owned scope
- 不能因为“投影可见”而自动进入 template

### 9.2 `.orbit/source.yaml`

`.orbit/source.yaml` 继续：

- 不进入 projection-owned 业务文件层
- 不进入 template content
- 不进入 harness template content

### 9.3 `.harness/*`

`.harness/*` 继续：

- 不允许进入 `shared_scope`
- 不允许进入 `projection_visible`
- 不允许进入 orbit template content

---

## 10. 对当前代码的影响评估

## 10.1 结论

这项改动不是“大重写”，但也不是“小修补”。

**影响等级：中等偏大。**

原因不是 schema rename 本身，而是：

- 当前代码把“投影视图”和“scoped operation scope”几乎等同处理；
- 你现在要求 `projection_visible` 只影响 projection/status，不影响 template/scoped ops；
- 这会迫使内核把一个 scope 拆成两个正式层次。

### 10.2 受影响模块

直接受影响：

- `cmd/orbit/cli/orbit/config.go`
- `cmd/orbit/cli/orbit/validate.go`
- `cmd/orbit/cli/orbit/resolve.go`
- `cmd/orbit/cli/orbit/match.go`
- `cmd/orbit/cli/view/enter.go`
- `cmd/orbit/cli/view/status.go`
- `cmd/orbit/cli/scoped/scope.go`
- `cmd/orbit/cli/scoped/diff.go`
- `cmd/orbit/cli/scoped/log.go`
- `cmd/orbit/cli/scoped/commit.go`
- `cmd/orbit/cli/scoped/restore.go`
- `cmd/orbit/cli/template/save.go`
- `cmd/orbit/cli/template/content_builder.go`
- `cmd/orbit/cli/harness/template_candidate.go`
- `cmd/orbit/cli/commands/files.go`
- 相关 unit / integration tests

文档受影响：

- `README.md`
- `docs/testing-strategy.md`
- `docs/context/mvp-product-requirements.md`
- `docs/context/mvp-technical-architecture.md`
- `docs/context/orbit_storage_boundary.md`
- `docs/context/two-scope-refactor.md`

### 10.3 为什么不是“小改动”

如果只是：

- `always_visible` 改名为 `shared_scope`
- 且语义仍保持“会进 projection 也会进 template”

那改动量是中等偏小。

但你现在明确要求：

- `shared_scope` 要受 `exclude` 约束
- 新增 `projection_visible`
- `projection_visible` 不能进入 template save/publish
- `projection_visible` 也不能进入 scoped ops

这就会改动：

- scope resolution 合同
- scoped operations scope 来源
- status / untracked 分类逻辑
- template build 输入边界
- 默认配置
- 测试金字塔中的一整片基线

所以更准确的判断是：

**内核层中等偏大改动，但边界相对集中，不会波及 harness runtime host。**

---

## 11. 推荐实现顺序

### Phase A：合同与解析层

1. 新增 `shared_scope` / `projection_visible` schema
2. 移除旧 `always_visible` 兼容读取
3. 重写 `ResolveScopeSet(...)`
4. 重写 `PathMatchesOrbit(...)`
5. 补 unit tests

### Phase B：命令内核层

1. `enter` / `status` / `files` 改用新 `ScopeSet`
2. `scoped/*` 改走 `ScopedOperationPaths`
3. projection cache 继续缓存 `ProjectionPaths`
4. 补 integration tests

### Phase C：template 层

1. `orbit template save/publish` 改用 `OwnedPaths`
2. `harness template candidate` 改用 `OwnedPaths`
3. 补 template/save/publish/harness template tests

### Phase D：默认配置与文档

1. 默认配置改成空 `shared_scope` / 空 `projection_visible`
2. 更新 README / testing docs / context docs
3. 补一致性说明

---

## 12. 建议结论

这个方案在产品语义上是成立的，而且比当前模型更清晰：

- `shared_scope` = 共享 owned scope
- `projection_visible` = 仅投影可见层

但它的真正代价在于：

- 你不再能把 `ProjectionPaths` 同时当成“看见什么”和“scoped ops 作用到什么”

因此：

- 如果你坚持 `projection_visible` 只影响 projection/status，这项改动应按一次正式 enhancement 做，不建议夹带进别的小需求。
- 如果你只想快速止血 `README.md` 这类文件误入 template，短期更小的替代方案只是：
  - 把 `always_visible` 改名成 `shared_scope`
  - 同时把它改成可被 `exclude` 排除
  - 暂时不引入 `projection_visible`

但既然你已经明确要两者都要，当前更合理的做法就是按本 spec 走完整分层。
