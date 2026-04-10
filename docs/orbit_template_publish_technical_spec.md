# Orbit Template Publish Technical Spec

版本：v0.1
状态：proposal
阶段：post-v0.3 enhancement
对应 PRD：`docs/orbit_template_publish_prd.md`
关联文档：
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/harness_centric_runtime_development_plan.md`
- `docs/technical-debt.md`

---

## 1. 文档目标

冻结 `orbit template publish` 的实现合同，解决它与现有 `orbit template save`、模板分支写入逻辑、以及 Git push 行为之间的边界问题。

这份文档的重点不是扩展新能力，而是把当前最容易产生分歧的实现点先讲清楚，避免直接编码时把作者工作流做成隐式、脆弱或高副作用的命令。

---

## 2. 为什么需要单独 Technical Spec

从代码现状看，`orbit template publish` 不是一个“纯 help 层别名”。

它和现有实现之间至少有四个关键岔口：

1. 当前 `orbit template save` 明确要求显式 `--to`，而 `publish` 要固定到 `orbit-template/<orbit-id>`。
2. 当前 `orbit template save` 默认在目标 branch 已存在时 fail-closed；而 `publish` 的使用场景天然要求重复发布。
3. 当前模板保存逻辑是“基于当前 worktree / HEAD 回退的本地保存”；如果要实现“自动对主发布分支发布”，必须先明确是：
   - 只允许在配置指定的 source branch 上运行；
   - 还是允许在任意 branch 上读取目标 source branch 的内容发布。
4. 当前代码里还没有正式的 Git push 原语与输出合同；`--push` 不是零成本附加。

因此，这项增强的代码体量未必非常大，但**行为合同的风险不小**，值得先补一份实现规格。

---

## 3. 现有代码基线

### 3.1 可直接复用的部分

- `cmd/orbit/cli/commands/template_save.go`
- `cmd/orbit/cli/template/save.go`
- `cmd/orbit/cli/template/content_builder.go`
- `cmd/orbit/cli/git/template_branch.go`

当前已经具备：

- orbit template tree 构建
- manifest 生成
- 本地 template branch 写入
- branch 已存在时的 overwrite 语义
- `--default` manifest 标记
- text / json 输出与测试基线

### 3.2 当前最重要的既有合同

1. `orbit template save` 是显式底层原语；
2. `WriteTemplateBranch(..., Overwrite=true)` 允许重复写入同一 template branch；
3. `WriteTemplateBranch` 的 overwrite 不是 destructive reset，而是向既有 target branch 追加一个新 commit；
4. 模板保存读取的是当前 repo 的本地状态，不是一个任意 revision 读取器；
5. 当前命令树里没有正式 publish 命令，也没有 Git push 的命令级封装。
6. 当前 `template save` 没有 no-op publish 检测，重复写入相同模板内容也可能产生新 commit。

---

## 4. 最小实现原则

为了控制风险，`orbit template publish` 的第一版实现遵循以下原则：

1. 它是作者侧显式命令，不改变 `harness install`。
2. 它是 `orbit template save` 的高层封装，不重写 template save 内核。
3. 它固定发布命名，不支持 `--to`。
4. 它只支持“在当前本地作者仓库内发布”，不做远程仓库的自动重建。
5. 它不做“自动切到 source branch 再发布”，而是要求调用者已经位于配置指定的 source branch。
6. source branch 是单 orbit 的专业作者分支；runtime-like 提取模板仍由 `orbit template save` 负责。

---

## 5. 命令合同

### 5.1 命令面

```bash
orbit template publish
orbit template publish --orbit <orbit-id>
orbit template publish --push
orbit template publish --remote <remote>
orbit template publish --default
```

不支持：

```bash
orbit template publish --to ...
```

说明：

- `--orbit` 在 source branch 模型里不是主要选择器
- source branch 强合同要求恰好一个 orbit definition
- `--orbit` 仅作为显式一致性校验入口保留

### 5.2 固定发布命名

发布目标 ref 固定为：

```text
orbit-template/<orbit-id>
```

命令不提供自定义 branch 命名入口。

原因：

- 避免命名漂移
- 避免 source repo README / CI / 安装命令长期写死多个 ref 别名
- 让作者与消费者都只记一个稳定目标

---

## 6. Source Branch 识别

`orbit template publish` 的第一版引入一个显式 source branch marker：

```text
.orbit/source.yaml
```

它的语义不是“整个 Git 仓库都是模板仓库”，而是：

- 当前 revision / branch 是一个 orbit template source branch
- 它是作者维护和发布模板的输入分支
- 它不是 `harness install` 可直接消费的 installable template branch

### 6.1 最小合同

当前 branch 要成为合法 source branch，必须同时满足：

- 当前目录位于 Git repo 内
- 当前 revision 上存在合法 `.orbit/source.yaml`
- 当前 revision 上存在合法 `.orbit/config.yaml`
- 恰好存在一个合法 `.orbit/orbits/*.yaml`
- 当前 revision 上不存在 `.orbit/template.yaml`
- 当前 revision 上不存在 `.harness/runtime.yaml`
- 当前 revision 上不存在 `.harness/vars.yaml`
- 当前 revision 上不存在 `.harness/installs/*`
- 当前 revision 上不存在 `.harness/template.yaml`

若不满足，`orbit template publish` 必须 fail-closed。

### 6.2 `.orbit/source.yaml` 合同

第一版建议采用最小 schema：

```yaml
schema_version: 1
kind: orbit_template_source
source_branch: main
publish:
  orbit_id: issues
```

合同：

- `schema_version` 固定为 `1`
- `kind` 固定为 `orbit_template_source`
- `source_branch` 必须是合法本地 branch 名
- `source_branch` 不做隐式默认值，必须显式声明
- `publish.orbit_id` 可选
- 若声明 `publish.orbit_id`，它必须与唯一合法 orbit definition 的 id 一致

### 6.3 与 installable template branch 的区分

source branch 与 installable template branch 必须严格区分：

- 有效 `.orbit/source.yaml`：source branch
- 有效 `.orbit/template.yaml`：orbit template branch
- 有效 `.harness/template.yaml`：harness template branch

这意味着：

- source branch 只服务作者工作流
- source branch 展示模板原始内容，而不是 runtime bindings 展开后的内容
- source branch 可以保留作者说明、README、测试脚本、开发辅助文件
- install 继续只消费带 `.orbit/template.yaml` 的已发布 template branch
- `.orbit/source.yaml` 不参与 remote template discovery
- `.orbit/source.yaml` 不进入 orbit template content
- `.orbit/source.yaml` 不进入 harness template content
- `.harness/*` 不属于 source branch 合法内容

若用户显式把 source branch 当作 `harness install --ref ...` 的输入，命令应继续像今天一样失败，因为它不是有效 template branch。

### 6.4 冲突合同

同一个 revision 上若同时存在：

- 有效 `.orbit/source.yaml`
- 有效 `.orbit/template.yaml`

第一版按冲突状态 fail-closed 处理。

原因：

- source branch 是作者输入态
- template branch 是发布产物态
- 同一 revision 同时声明两者会让 `publish`、`install` 和 branch inspect 的语义都变得模糊

### 6.5 单 orbit 强合同

第一版将 source branch 收紧为单 orbit 作者分支：

- source branch 必须恰好包含一个合法 orbit definition
- `publish` 默认发布该唯一 orbit
- 若显式传 `--orbit`，其值必须与唯一 orbit id 一致
- 若 `.orbit/source.yaml` 中存在 `publish.orbit_id`，其值也必须与唯一 orbit id 一致

多 orbit authoring 不是 source branch 模型的一部分。

若作者需要从 runtime-like 分支或多 orbit 状态中提取模板，继续使用：

```bash
orbit template save <orbit-id> --to <template-branch>
```

---

## 7. Source Branch 合同

这是这项增强里最重要的收口点。

### 7.1 不采用的方案

第一版不做：

- 在任意 branch 上隐式读取配置指定的 source branch
- 自动切换到 source branch 再执行 publish
- 用临时 detached checkout / temp worktree 重建 source branch 内容

原因：

- 现有 `template save` 内核就是基于当前 repo 本地状态工作的；
- 如果把 `publish` 做成“无论你在哪个分支都自动发布 source branch”，就会把保存路径从“当前 worktree 发布器”升级为“任意 revision 发布器”，改动面明显扩大；
- 自动切换 branch 会增加本地状态副作用和失败恢复复杂度。

### 7.2 第一版采用的合同

第一版明确要求：

- 当前 branch 必须等于 `.orbit/source.yaml` 中声明的 `source_branch`
- 当前 branch 上必须可见合法 `.orbit/source.yaml`

若不在该 branch，命令应 fail-closed。

错误语义应直接告诉用户：

- 当前 publish 只支持在 source branch 上运行
- source branch 由 `.orbit/source.yaml` 的 `source_branch` 定义
- 若要发布，请先切回该分支

### 7.3 Worktree 状态合同

第一版还要求：

- worktree 必须是 clean

原因：

- 当前模板保存会读取当前本地状态；
- 若允许 dirty worktree，`publish` 就会把“尚未提交的作者草稿”发布成正式模板态；
- 这和“从 source branch 发布稳定模板态”的目标冲突。

因此：

- 非 clean worktree 时 fail-closed

这里的 clean 检查应以 tracked changes 为主，与现有 Orbit/Git 语义保持一致。

---

## 8. 重复发布合同

当前 `orbit template save` 对已存在目标 branch 默认失败。

这对 `save` 是对的，但对 `publish` 不合适，因为：

- `publish` 天然是重复动作
- 作者需要稳定地把 source branch 的最新状态发布到同一个固定 ref

因此第一版合同是：

- `orbit template publish` 内部总是以 `overwrite=true` 调用底层 template branch writer

但这里的 overwrite 语义不是 destructive reset：

- 继续沿用当前 `WriteTemplateBranch` 的语义
- 若 branch 已存在，则基于现有 target branch 追加新的 commit
- 不做 `git reset --hard`
- 不做 force history rewrite

这意味着：

- 本地重复发布是正常行为
- 后续 `git push` 默认应是 fast-forward 友好的

### 8.1 no-op publish 检测

第一版必须增加 no-op 检测。

合同：

- 若新生成的模板 tree 与当前目标 template branch HEAD 的 tree 完全一致，则不创建新 commit
- 命令输出应明确标记：
  - `local_publish.success=true`
  - `local_publish.changed=false`

这样做的原因：

- 避免重复 publish 在“内容未变化”时制造无意义 commit
- 避免 `--push` 因远端 freshness check 被阻止后，作者重复执行 publish 时累积本地噪音 commit

若目标 branch 不存在，则正常创建第一次发布 commit。

---

## 9. Push 合同

### 9.1 `--push` 是显式行为

默认：

- 只更新本地 `orbit-template/<orbit-id>` 分支

只有显式给出：

```bash
--push
```

才会继续 push 到远端。

### 9.2 remote 解析

默认 remote：

```text
origin
```

显式 remote：

```bash
orbit template publish --push --remote origin
```

若给出 `--remote` 但未给 `--push`，命令应 fail-closed，因为此时 remote 参数没有意义。

### 9.3 Source Branch Freshness 检查

只有显式给出 `--push` 时，才执行 source branch freshness 检查。

检查对象：

- 本地 source branch `source_branch`
- 远端 `<remote>/<source_branch>`

合同定义：

- 若本地 source branch 与远端 source branch 相等：允许 push
- 若本地 source branch ahead 于远端 source branch：允许 push
- 若本地 source branch behind 于远端 source branch：阻止 push
- 若本地 source branch 与远端 source branch diverged：阻止 push

如果检查结果是 behind 或 diverged：

- 不执行 template branch 远端 push
- 本地 publish 结果仍然保留
- 命令返回非零
- 输出必须明确说明：
  - 本地 publish 成功
  - 远端 push 未执行
  - 原因是 source branch not up to date

如果远端 source branch 不存在，第一版按 fail-closed 处理。

### 9.4 push 行为边界

第一版：

- 只做普通 push
- 不支持 `--force`
- 不自动创建复杂 upstream 配置面

如果 push 失败，命令直接报错，但本地 template branch 发布结果仍保留。

---

## 10. 输出合同

第一版建议提供 text / json 两种输出。

### 10.1 结构原则

输出必须明确区分两段结果：

- 本地 publish 是否成功
- 远端 push 是否成功

即使 `--push` 失败，也不能把“本地 publish 成功”的事实吞掉。

### 10.2 text 输出最低字段

建议 text 输出至少包含：

- `orbit_id`
- `publish_ref`
- `source_branch`
- `local_publish.success`
- `local_publish.changed`
- `local_publish.commit`（changed=true 时）
- `remote_push.attempted`
- `remote_push.success`
- `remote_push.remote`（attempted=true 时）
- `remote_push.reason`（失败或未执行时）

### 10.3 json 输出最低字段

```json
{
  "orbit_id": "issues",
  "publish_ref": "refs/heads/orbit-template/issues",
  "branch": "orbit-template/issues",
  "source_branch": "main",
  "default_template": true,
  "local_publish": {
    "success": true,
    "changed": true,
    "commit": "<sha>"
  },
  "remote_push": {
    "attempted": false,
    "success": false
  }
}
```

若 `--push` 且成功，则：

```json
"remote_push": {
  "attempted": true,
  "success": true,
  "remote": "origin"
}
```

若 `--push` 但因为 source branch behind / diverged 被阻止，则：

```json
"remote_push": {
  "attempted": false,
  "success": false,
  "remote": "origin",
  "reason": "source_branch_not_up_to_date"
}
```

---

## 11. 命令层与包边界

### 11.1 命令层

新增：

- `cmd/orbit/cli/commands/template_publish.go`

挂载到：

- `cmd/orbit/cli/commands/template.go`

### 11.2 可复用逻辑

尽量复用：

- `orbittemplate.BuildTemplateSavePreview`
- `orbittemplate.SaveTemplateBranch`
- `gitpkg.WriteTemplateBranch`

### 11.3 建议新增的小原语

建议新增一个很薄的高层封装，例如：

- `cmd/orbit/cli/template/publish.go`

职责：

- 解析 orbit id
- 读取 `.orbit/source.yaml`
- 断言当前 branch == source branch
- 断言 clean worktree
- 计算固定 publish branch
- 复用现有 save path
- 执行 no-op 检测
- 可选执行 source branch freshness 检查
- 可选 push

这样命令层仍保持薄。

### 11.4 配套作者入口

建议新增一个配套命令：

```bash
orbit template init-source
```

它不属于 `publish` 的最小功能闭环，但应作为紧随其后的作者体验补强。

最小职责：

- 读取当前 branch 名并写入 `.orbit/source.yaml`
- 将 `source_branch` 设为当前 branch
- 在恰好一个 orbit definition 时写入 `publish.orbit_id`
- 在 detached HEAD、多 orbit、或存在 `.harness/*` / `.orbit/template.yaml` 时 fail-closed

---

## 12. 测试矩阵

至少补下面这些测试：

### 12.1 命令与解析

- 单 orbit source branch 下 `orbit template publish` 自动解析 orbit id
- `.orbit/source.yaml` 中声明 `publish.orbit_id` 时会做一致性校验
- source branch 含多个 orbit definitions 时失败
- source branch 上显式 `--orbit` 与唯一 orbit id 不一致时失败

### 12.2 branch / worktree 前置条件

- 当前 branch 没有合法 `.orbit/source.yaml` 时失败
- 当前不在 `.orbit/source.yaml` 指定的 source branch 时失败
- 当前 worktree dirty 时失败
- 当前 revision 同时出现 `.orbit/source.yaml` 与 `.orbit/template.yaml` 时失败
- 当前 revision 出现 `.harness/*` 时失败

### 12.3 发布行为

- 首次 publish 成功创建 `orbit-template/<orbit-id>`
- 重复 publish 成功更新同一 branch
- 无变化 publish 返回 `changed=false` 且不创建新 commit
- `--default` 正确进入 manifest
- 发布结果不包含 `.orbit/source.yaml`

### 12.4 push 行为

- `--push` 前会检查 source branch 新鲜度
- `--push` 时向默认 remote push 成功
- `--push --remote <name>` 时向指定 remote push 成功
- `--remote` 不配 `--push` 时失败
- 本地 source branch behind 远端时阻止 push
- 本地 source branch diverged 于远端时阻止 push
- push 失败时本地 publish 结果保留，命令返回非零

### 12.5 输出合同

- text 输出字段稳定
- `--json` 输出字段稳定

---

## 13. 风险判断

### 13.1 当前最主要的风险

如果不加这份技术方案，最可能出现的问题是：

1. 直接把 `publish` 做成 `template save` 的轻包装，但遗漏“重复发布需要 overwrite”；
2. 把“自动对 source branch 发布”误做成“在任意 branch 上偷偷读 source branch 或偷偷切 branch”；
3. 在没有明确 push 合同的情况下，引入半成品远端写入逻辑；
4. 让 `--to` 重新进入命令面，导致你已经明确拒绝的命名漂移回归。

### 13.2 结论

这项改动**代码规模中等，但行为合同风险偏高**。

因此建议：

- 先按这份 technical spec 收口
- 再进入实现

而不是只凭 PRD 直接写代码。

---

## 14. 一句话总结

`orbit template publish` 的第一版应被实现成：

**一个运行在 clean、single-orbit source branch 上的作者侧显式发布命令；source branch 由 `.orbit/source.yaml` 明确标记并声明 `source_branch`，可选声明 `publish.orbit_id` 作为一致性锁，发布固定落到 `orbit-template/<orbit-id>`，本地重复发布默认可用，内建 no-op 检测，远端 push 必须显式开启且先通过 source branch freshness 检查。**
