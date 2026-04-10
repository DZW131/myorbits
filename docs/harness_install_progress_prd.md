# Harness Install Progress PRD（产品需求文档）

版本：v0.1
状态：可进入技术设计
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_technical_spec.md`
- `docs/harness_install_source_repo_resolution_technical_spec.md`
- `docs/harness_mixed_install_technical_spec.md`
- `docs/testing-strategy.md`

---

# 1. 文档目的

本文档定义 `harness install` 的执行过程可观测性需求。

当前 `harness install` 在远端 source 解析、template 选择、fetch、bindings 解析、冲突检查、写盘等阶段之间几乎没有中间反馈。对用户而言，命令经常表现为：

- 输入命令后长时间无输出；
- 不清楚当前是否仍在执行；
- 不知道卡在网络、模板解析、bindings、冲突检查还是写盘阶段。

本文档的目标是把 `harness install` 收成：

- 有稳定阶段反馈；
- 不破坏现有 `stdout` / `--json` 契约；
- 允许后续再扩展更细粒度的 Git transport 进度。

---

# 2. 核心结论

## 2.1 默认增加安装阶段进度

`harness install` 应输出稳定的阶段性进度信息，避免用户在长耗时安装中“无感等待”。

## 2.2 进度默认写到 `stderr`

进度输出默认写到 `stderr`，而不是 `stdout`。

这样可以保持：

- `stdout` 继续保留最终结果输出；
- `--json` 继续只输出最终 JSON；
- 过程性信息不污染脚本消费面。

## 2.3 第一版使用稳定阶段行，不做动画

第一版不做 spinner、进度条、动态重绘，而是只输出稳定的阶段行。

这可以保证：

- TTY 与非 TTY 行为更稳定；
- 测试更容易冻结；
- 日志回放更清楚。

## 2.4 原生 Git 网络进度不作为第一版默认能力

当前 `harness install` 底层不是 `git clone`，而是远端 `ls-remote` / 临时 `fetch --depth=1`。

因此第一版的默认可见性目标不是“直接透传 Git 原生进度”，而是“让用户知道 install 正在进行到哪一步”。

raw Git progress 可以作为后续可选增强，但不应阻塞第一版。

---

# 3. 用户问题与目标

## 3.1 当前问题

### 问题 A：长时间无反馈

远端安装尤其在以下阶段容易让用户误以为命令卡死：

- 枚举远端模板分支；
- source branch alias 解析；
- fetch 选中的模板分支；
- bindings 读取与补全；
- conflict / overwrite 检查。

### 问题 B：失败时缺少阶段上下文

即使最终报错，用户也往往不知道失败发生在：

- 远端选择；
- template snapshot 读取；
- bindings；
- conflict 检查；
- 写盘。

### 问题 C：不能破坏现有机器消费面

`harness install --json` 已有稳定结果契约。  
进度输出不能破坏：

- JSON 输出结构；
- 现有 stdout 文本结果；
- 测试与脚本消费方式。

## 3.2 产品目标

本次改造要同时满足：

1. 用户能看到 install 正在推进；
2. 用户能知道当前处于哪个阶段；
3. 失败时能知道大致停在哪一步；
4. `stdout` 最终结果契约不变；
5. `--json` 最终输出契约不变；
6. 不引入复杂动态 UI。

---

# 4. 适用范围

本 PRD 只覆盖：

- `harness install`

包括：

- 本地 orbit template install
- 远程 orbit template install
- source repo alias 解析后的 install
- harness template install
- mixed install 未来扩展下的统一进度模型

本 PRD 不覆盖：

- `orbit template apply` 兼容 wrapper 的进度设计
- `orbit template publish`
- `harness template save`
- Git push/publish 的远端进度体验

---

# 5. 正式产品决策

## 5.1 进度输出通道

### Text 模式

- 最终结果：继续输出到 `stdout`
- 过程进度：输出到 `stderr`

### JSON 模式

- 最终 JSON：继续输出到 `stdout`
- 过程进度：如启用，仍输出到 `stderr`

也就是说：

- `stdout` 永远只承载最终结果
- `stderr` 承载过程性信息

## 5.2 第一版进度样式

第一版只支持**稳定阶段行**。

不做：

- spinner
- 百分比进度条
- 动态重绘
- Git 原生 fetch 文本的默认透传

## 5.3 第一版阶段模型

`harness install` 第一版至少应有以下阶段：

1. `resolving install source`
2. `resolving remote template candidates`
   - 仅在远程 URL 且需要远端选择时出现
3. `source branch detected; resolving published template`
   - 仅在 source repo alias 解析时出现
4. `fetching selected template`
5. `resolving bindings`
6. `checking conflicts`
7. `writing files`
8. `updating runtime metadata`
9. `install complete`

若安装对象是 harness template，可增加更具体的中间阶段：

- `loading harness template manifest`
- `validating harness template members`

## 5.4 进度模式

第一版引入一个显式进度模式开关：

```bash
harness install ... --progress <mode>
```

支持：

- `auto`
  - 默认值
  - TTY 时显示阶段进度
  - 非 TTY 时静默
- `plain`
  - 总是输出稳定阶段行
- `quiet`
  - 不输出过程进度

## 5.5 失败时的阶段上下文

当 install 失败时，用户至少应能从已输出的阶段行判断：

- 命令已进入哪一个阶段；
- 最后停留在哪一个阶段。

第一版不要求把阶段信息额外编码到最终错误对象里，但要保证：

- 进度行顺序稳定；
- 错误前最后一条阶段行对排查有意义。

---

# 6. Raw Git Progress 的定位

## 6.1 第一版不默认透传

尽管用户会把远端拉取直觉理解为“git clone / git fetch”，但第一版不默认透传 Git 原生进度。

原因：

- 当前底层 transport 不是 clone；
- 原生 Git 进度只覆盖网络阶段，不覆盖模板选择、bindings、冲突检查、写盘；
- 原生进度文本较吵，不适合作为默认 CLI 产品输出。

## 6.2 后续可选扩展

后续若确实需要更细网络可见性，可以单独设计：

```bash
harness install ... --git-progress
```

或：

```bash
harness install ... --progress verbose
```

但这不属于第一版需求。

---

# 7. 用户可见行为

## 7.1 Text 模式示例

```text
$ harness install https://example.com/acme/issues.git
progress: resolving install source
progress: resolving remote template candidates
progress: source branch detected; resolving published template
progress: fetching selected template
progress: resolving bindings
progress: checking conflicts
progress: writing files
progress: updating runtime metadata
progress: install complete
installed orbit issues into harness /repo/path
source_ref: orbit-template/issues
member_count: 1
files: 6
warnings: none
```

## 7.2 JSON 模式示例

```text
$ harness install https://example.com/acme/issues.git --json
stderr:
progress: resolving install source
progress: resolving remote template candidates
progress: fetching selected template
...

stdout:
{
  "dry_run": false,
  "harness_root": "/repo/path",
  "source": {
    "kind": "remote_git",
    "repo": "https://example.com/acme/issues.git",
    "ref": "orbit-template/issues",
    "requested_ref": "main",
    "resolved_ref": "orbit-template/issues",
    "resolution_kind": "source_alias",
    "commit": "abc123"
  },
  "orbit_id": "issues",
  "written_paths": [
    "AGENTS.md",
    ".orbit/orbits/issues.yaml"
  ],
  "member_count": 1
}
```

---

# 8. 非目标

本次不做：

1. 真实字节级网络进度条；
2. clone/fetch 原生 stderr 的默认透传；
3. 安装耗时统计或 ETA；
4. 多线程并行安装进度聚合；
5. 结构化实时 JSON event stream。

---

# 9. 验收标准

## 9.1 可见性

- `harness install` 在长耗时场景下不再“完全静默”
- 用户能从 `stderr` 看见当前阶段

## 9.2 契约稳定性

- `stdout` 最终 text 输出仍保持稳定
- `--json` 的 stdout JSON 结构不被过程输出污染

## 9.3 模式控制

- `--progress auto|plain|quiet` 行为稳定
- 默认模式适合交互式用户

## 9.4 覆盖场景

以下场景都应具备稳定进度体验：

- 本地 orbit template install
- 远程 orbit template install（显式 `--ref`）
- 远程 orbit template install（自动选择）
- source repo alias install
- harness template install
- dry-run preview

---

# 10. 后续扩展

后续可在本 PRD 之上继续扩展：

1. raw Git progress 可选模式
2. 长耗时 heartbeat
3. 更细粒度阶段与来源诊断
4. install / check / publish 的统一 progress 框架
