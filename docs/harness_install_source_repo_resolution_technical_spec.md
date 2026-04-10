# Harness Install Source Repo Resolution Technical Spec

版本：v0.1
状态：proposal
阶段：post-v0.3 enhancement
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/orbit_template_publish_technical_spec.md`
- `docs/testing-strategy.md`

---

## 1. 文档目标

冻结 `harness install <repo-url>` 在面对 orbit template source repo 时的自动解析合同。

当前模型里：

- source branch 由 `.orbit/source.yaml` 标识；
- installable orbit template branch 由 `.orbit/template.yaml` 标识；
- `harness install` 继续只消费 installable template branch；
- 若用户把 source branch 当成 install source，当前实现会 fail-closed 并提示先 `orbit template publish`。

这份文档引入一个更自然的产品行为：

- **允许用户直接把 source repo URL 作为 `harness install` 的输入；**
- **系统自动将 source branch 解析为其已发布的 orbit template branch；**
- **系统不自动 publish，也不改变 install 的正式消费对象仍然是 orbit template branch。**

---

## 2. 要解决的产品问题

当前 source branch / template branch 模型在作者侧是清楚的，但对消费者侧偏重：

1. 用户需要知道 source branch 不可安装；
2. 用户需要知道发布分支命名规则 `orbit-template/<orbit-id>`；
3. 用户需要显式写 `--ref orbit-template/<orbit-id>`；
4. 如果远端默认分支正好就是 source branch，`harness install <repo-url>` 仍然会失败。

对单 orbit 模板仓库来说，这会造成不必要的心智负担。

---

## 3. 非目标

本次增强不做以下事情：

1. 不自动执行 `orbit template publish`；
2. 不让 source branch 变成 installable template branch；
3. 不改变 `orbit template publish` 的 source contract；
4. 不引入 harness template install；
5. 不改变本地 branch install 的输入模型；
6. 不在远端 repo 上做任意 revision 重建或自动切 branch。

---

## 4. 已冻结结论

### 4.1 install 的正式消费对象不变

`harness install` 正式消费的仍然是：

- 本地 orbit template branch
- 远程 orbit template branch

source branch 仍然不是 installable template branch。

### 4.2 source repo URL 可以作为 orbit template branch 的别名输入

当输入是远程 Git URL 且命中 source branch 解析规则时，`harness install` 可以把它当成一个 **source alias**，再解析成对应的已发布 orbit template branch。

因此：

- **用户输入**：source repo URL
- **系统真实安装的对象**：`orbit-template/<orbit-id>`

### 4.3 不自动 publish

若 source branch 存在，但对应的 `orbit-template/<orbit-id>` 不存在：

- install 必须 fail-closed；
- 错误信息必须明确提示先运行 `orbit template publish`。

### 4.4 只支持远程 source alias

第一版 source alias 只对 **远程 Git URL** 生效。

不支持：

- 本地文件系统路径 source repo
- 本地当前 repo 中某个 source branch 名被直接当 alias

原因：

- 当前 `install` 的本地输入模型是“本 repo 里的 revision”
- source alias 的主要价值在消费者安装远程模板仓库时降低心智
- 保持本地分支安装与远端 URL 安装的边界清晰

### 4.5 显式 `--ref` 指向 source branch 时也允许 alias

若用户显式传：

```bash
harness install <repo-url> --ref <source-branch>
```

且该 `ref` 是合法 source branch，则安装流程应：

1. 读取 `.orbit/source.yaml`
2. 取 `publish.orbit_id`
3. 重定向解析到 `orbit-template/<publish.orbit_id>`

仍然不自动 publish。

---

## 5. 远程 source alias 合同

### 5.1 触发条件

以下两种输入触发 source alias 分支：

#### 情况 A：未显式提供 `--ref`

```bash
harness install <repo-url>
```

系统行为：

1. 先检查远端默认分支是否为合法 source branch；
2. 若是，则优先尝试解析 source alias；
3. 若默认分支不是合法 source branch，再按当前远端默认选择逻辑枚举 installable orbit template branches；
4. 若能直接唯一选出 installable orbit template，则沿用现有逻辑；
5. 若不能直接选出，则保持现有 not found / ambiguity 行为。

#### 情况 B：显式 `--ref` 指向 source branch

```bash
harness install <repo-url> --ref main
```

系统行为：

1. 若该 `ref` 不是 installable orbit template branch；
2. 再检查它是否为合法 source branch；
3. 若是，则尝试解析 source alias。

### 5.2 source alias 所需最小条件

要成功把 source branch 映射成 published template branch，必须满足：

1. 远端 `ref` 上存在合法 `.orbit/source.yaml`
2. `.orbit/source.yaml` 中存在 `publish.orbit_id`
3. `publish.orbit_id` 为合法 orbit id
4. 远端存在 `refs/heads/orbit-template/<publish.orbit_id>`
5. 该分支上存在合法 `.orbit/template.yaml`

若任一条件不满足，必须 fail-closed。

### 5.3 第一版不做远端 orbit id 推导

即使 source branch 是单 orbit branch，第一版也不从远端 `.orbit/orbits/*.yaml` 推导目标 orbit id。

第一版只接受：

- `.orbit/source.yaml` 中显式存在 `publish.orbit_id`

原因：

- source alias 的目的是减少消费者心智，而不是扩展远端 source branch 解析器
- 用 `publish.orbit_id` 可避免额外读取和推导逻辑
- 这也和 source branch 当前“可选 orbit_id 一致性锁”设计相契合

---

## 6. 与现有远端选择逻辑的关系

当前远端 orbit template 选择规则在：

- `cmd/orbit/cli/template/remote_source.go`

现有顺序是：

1. 显式 `--ref` -> 只检查该 ref 是否为 installable orbit template branch
2. 未显式 `--ref`：
   - 0 candidates -> not found
   - 1 candidate -> 选中
   - 多 candidate + 唯一 default -> 选中
   - 否则 -> ambiguity

新规则建议改成：

### 6.1 显式 `--ref`

1. 若 `--ref` 是 installable orbit template branch -> 直接选中
2. 若不是 installable orbit template branch：
   - 检查它是否为 source branch
   - 若是，则尝试 source alias
   - 若 alias 失败，返回 source-aware not found
3. 若既不是 template branch，也不是 source branch -> 保持当前 not found

### 6.2 未显式 `--ref`

1. 先解析远端默认分支
2. 若默认分支是 source branch：
   - 优先尝试 source alias
   - 若 alias 成功，则直接选中 alias 目标
   - 若 alias 因 source 合同不完整或 published branch 缺失而失败，则返回 source-aware not found
3. 若默认分支不是 source branch：
   - 再跑现有 orbit template candidate 选择逻辑
4. 若 candidate 逻辑得到唯一 installable template -> 直接选中
5. 若 candidate 逻辑结果是 not found / ambiguity -> 保持现有行为

原因：

- source repo URL 的主要用户心智是“默认分支就是模板作者入口”
- 当默认分支已经明确声明自己是 source branch 时，优先走 source alias 更符合作者/消费者的直觉
- ambiguity 仍然只在“默认分支不是 source branch”时进入现有 candidate 选择逻辑

---

## 7. 输出与错误合同

### 7.1 text/json 需要保留真实安装对象

即使用户输入的是 source repo URL，install 的输出也必须明确区分：

- 用户输入的 source alias
- 实际解析到的 published template branch

建议 JSON 至少新增：

```json
{
  "source": {
    "kind": "remote_git",
    "repo": "https://example.com/acme/template.git",
    "requested_ref": "main",
    "resolved_ref": "orbit-template/issues",
    "resolution_kind": "source_alias"
  }
}
```

text 输出也应增加一行：

```text
resolved_ref: orbit-template/issues
resolution_kind: source_alias
```

### 7.2 关键失败文案

source alias 相关失败至少区分：

1. source branch 存在，但未声明 `publish.orbit_id`
2. source branch 存在，但对应 published template branch 不存在
3. source branch 存在，但目标 published branch 不是合法 orbit template branch

其中第 2 类错误应明确提示：

```text
run `orbit template publish` on the source branch first
```

---

## 8. 测试矩阵

至少补这些场景：

1. `harness install <repo-url>`：
   - 远端默认分支是 source branch
   - `publish.orbit_id` 存在
   - published branch 存在
   - 成功解析并安装
2. `harness install <repo-url> --ref main`：
   - `main` 是 source branch
   - 成功 alias
3. source branch 存在但缺少 `publish.orbit_id`
   - fail-closed
4. source branch 存在但对应 `orbit-template/<id>` 不存在
   - fail-closed，提示 publish first
5. 远端已有多个 installable template branch 且无唯一 default
   - 继续 ambiguity
   - 不自动 source alias
6. 远端 `--ref` 指向普通非 template/non-source branch
   - 保持现有 not found

---

## 9. 推荐拆分

### Issue A：远端 source alias resolution 原语

- 增强 remote selector
- 增加 source alias 输出与错误 contract
- 补 selector tests

### Issue B：`harness install` 接线与 acceptance

- 把 source alias 结果接入 install preview/result
- 冻结 text/json 输出
- 增加 acceptance smoke

---

## 10. 建议结论

source repo 直装是值得尽快做的增强，因为它：

- 不改变 install 的正式消费对象
- 不要求自动 publish
- 不需要引入 harness template install
- 能明显降低单 orbit 模板仓库的消费者心智负担

最稳的第一版是：

- 只支持远程 Git URL source alias
- 只在 source branch 明确声明 `publish.orbit_id` 时生效
- 只把 source alias 映射到已发布的 orbit template branch
- ambiguity 继续 fail-closed
