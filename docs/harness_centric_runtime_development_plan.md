# Harness-Centric Runtime 开发计划

版本：v0.3
状态：执行基线
对应 PRD：`docs/harness_centric_runtime_prd.md`
对应技术方案：`docs/harness_centric_runtime_technical_spec.md`
关联文档：
- `docs/harness_cli_full_recommendations.md`
- `docs/context/mvp-product-requirements.md`
- `docs/context/mvp-technical-architecture.md`
- `docs/testing-strategy.md`

---

## 1. 文档目标

本文档基于当前 v0.3 PRD 与技术方案，给出下一阶段的实际开发推进计划。

目标是明确：

1. 开发阶段如何切分；
2. 每阶段先做什么、后做什么；
3. 哪些代码边界应复用，哪些需要重构；
4. 每阶段的测试、收口与完成标准；
5. issue / 分支 / PR 应如何拆分。

本文档只覆盖 Harness-Centric Runtime 主线，不扩展到：

- harness template apply
- 多 harness 共存
- runtime host 的 dual-write 兼容
- 大规模包路径重命名
- 通用 harness-level shared-file 平台

---

## 2. Source Of Truth

开发顺序以下列文档为准：

1. `docs/context/mvp-product-requirements.md`
2. `docs/context/mvp-technical-architecture.md`
3. `docs/testing-strategy.md`
4. `docs/harness_centric_runtime_prd.md`
5. `docs/harness_centric_runtime_technical_spec.md`
6. `CONTRIBUTING.md`
7. `AGENTS.md`

开发计划必须服从以下已冻结合同：

- projection 继续由 orbit 驱动；
- runtime 正式对象切换为 harness；
- `.orbit/config.yaml` 继续作为 projection control plane；
- runtime 版本化元数据迁入 `.harness/`；
- `harness install` 成为唯一正式安装入口；
- zero-member runtime 合法；
- 同一 `orbit-id` 的重复安装默认失败，只有显式 `--overwrite-existing` 才允许覆盖；
- `orbit bindings init` 默认 stdout，推荐 `--out .harness/vars.yaml`；
- `harness template save` 对根目录 `AGENTS.md` 采用 whole-file 语义。

---

## 3. 总体开发策略

### 3.1 核心原则

整个实现过程中应持续满足以下原则：

- 保留 orbit projection kernel，不重写 `enter / leave / status / diff / log / commit / restore` 主链路。
- 用“上层重构、下层复用”的方式推进；先完成 host 切换与 bootstrap，再切安装入口与模板导出。
- `orbit` 与 `harness` 的职责边界必须在命令层、存储层、文档层同时收口。
- 一切 runtime 级写入都以 `.harness/*` 为正式宿主，不做 `.orbit/*` / `.harness/*` 双写。
- 所有 install、overwrite、template merge 都应 fail-closed，不静默覆盖。
- drift 由 `harness check` 显式诊断，不做后台修复或隐式重渲染。
- 一旦命令进入正式命令面，其 text / json 输出契约就应尽早用测试冻结，不把稳定化留到最后。
- 双二进制带来的 build / CI / help / completion 入口要和命令实现同步推进，不在最后补。
- 文档、help、quickstart 的切换不能落后于正式命令面。

### 3.2 开发排序原则

本阶段推荐严格按下面顺序开发：

1. 先做 preflight foundations，把路径宿主、harness domain skeleton、双二进制入口先搭好。
2. 先打通 harness bootstrap，再补最小 branch 可观测性。
3. 先交付 install 基础路径，再单独实现 overwrite / reinstall / drift foundations。
4. 先让 runtime / template branch taxonomy 完整可见，再补 `harness check`。
5. 先完成 harness template save 主链路，再做文档、发布与硬化收尾。
6. 先在现有共享包上做增量改造，不同时发起大规模目录迁移。

这样排序的原因：

- 路径宿主与 domain skeleton 不先收口，后续会在 `bindings`、`template`、`branchinfo` 中到处散落 `.orbit/*` / `.harness/*` 逻辑；
- `harness create/init/root/inspect` 是后续一切命令的前置基础设施；
- install 的基础路径与 overwrite 路径不是同一风险等级，应分开落地；
- branch info 与 check 需要建立在 install / members / template taxonomy 已落地之后；
- `harness template save` 是组合型能力，应建立在 install、vars、member、drift 分类都已稳定之后。

### 3.3 命令批次视图

建议按下面六批交付：

1. 第一批：preflight foundations、`harness` 二进制入口、path host helper
2. 第二批：`harness create`、`harness init`、`harness root`、`harness inspect`、最小 runtime branch 识别
3. 第三批：`harness add`、`harness remove`、`orbit bindings init` host 切换、`orbit template save` vars host 切换
4. 第四批：`harness install` 基础路径、hidden `orbit template apply` wrapper
5. 第五批：`harness install --overwrite-existing`、owned file reconstruction、drift foundations、完整 branch/check
6. 第六批：`harness template save`、quickstart/help/release/docs 收口

### 3.4 分支与 PR 策略

遵循 `CONTRIBUTING.md`：

- 从 `main` 拉分支开发，通过 PR 合并回 `main`；
- 每个阶段一个主分支、一个主 PR；
- 若阶段改动过大，可再拆 2 到 4 个子 PR，但不要跨阶段混合。

建议分支名：

- `feature/v0.3-preflight-foundations`
- `feature/v0.3-harness-bootstrap`
- `feature/v0.3-runtime-host-switch`
- `feature/v0.3-harness-install-basic`
- `feature/v0.3-harness-install-overwrite`
- `feature/v0.3-branch-check`
- `feature/v0.3-harness-template-save`
- `feature/v0.3-release-hardening`

---

## 4. 代码边界与责任划分

### 4.1 直接复用的稳定内核

原则上尽量不大动：

- `cmd/orbit/cli/git`
- `cmd/orbit/cli/state`
- `cmd/orbit/cli/ids`
- `cmd/orbit/cli/view`
- `cmd/orbit/cli/scoped`

这些包继续负责：

- repo / git 适配
- repo-local runtime state
- orbit id 与路径校验
- projection 进入与离开
- scoped diff/log/commit/restore

### 4.2 本阶段的主要改造层

重点改造：

- CLI 入口层
- runtime model / root discovery
- bindings / vars host
- install record host
- branch classification / inspect
- template install / harness template export

直接会触达的代码区域：

- `cmd/orbit/cli/root.go`
- `cmd/orbit/cli/bindings/*`
- `cmd/orbit/cli/template/*`
- `cmd/orbit/cli/branchinfo/*`
- 新增 `cmd/harness/*`
- 新增共享 `cmd/orbit/cli/harness/*`

### 4.3 新增共享领域包

首阶段新增：

- `cmd/orbit/cli/harness`

负责：

- `.harness/runtime.yaml` schema / load / write
- `.harness/vars.yaml` host path
- `.harness/installs/*.yaml` host path
- harness root resolution
- member mutation
- inspect / check 原语
- `.harness/template.yaml` schema / load / write

这一步是“加新域包”，不是“整体搬家”。

### 4.4 Preflight 重构边界

在正式进入命令开发前，先收口以下共享前置项：

- 把 `.orbit/vars.yaml`、`.orbit/installs/*` 相关硬编码路径抽成 host helper；
- 让 `bindings` 只负责 schema / merge / skeleton，不再拥有固定宿主路径；
- 让 install record 的 host 路径责任从 `template` 侧剥离到 `harness` 侧；
- 为后续 dual-binary build / test 留出独立入口。

---

## 5. 阶段划分总览

建议分为 7 个阶段：

- `Phase 3A-0`：Preflight Foundations
- `Phase 3A-1`：Harness Bootstrap 与最小可观测性
- `Phase 3B-1`：Runtime Host 切换与 Install Basic Path
- `Phase 3B-2`：Overwrite / Reinstall / Drift Foundations
- `Phase 3C`：完整 Branch / Check 升级
- `Phase 3D`：Harness Template Save
- `Phase 3E`：文档、发布与硬化收尾

下面按阶段展开。

---

## 6. Phase 3A-0：Preflight Foundations

这一阶段优先把路径宿主、shared primitives 和新的二进制入口搭起来，不追求完整功能闭环。

状态：已完成（2026-03-26）

### 6.1 目标

冻结 host 前置基础设施，确保后续命令都建立在统一路径责任与共享原语上。

### 6.2 交付物

- `cmd/harness/main.go`
- `cmd/harness/cli/root.go`
- `cmd/orbit/cli/harness/*`
- path host helper
- `.harness/runtime.yaml` schema / load / write
- `.harness/template.yaml` schema / load / write
- `.harness/vars.yaml` / `.harness/installs/*.yaml` host path helper
- dual-binary build / test 最小入口

### 6.3 开发内容

任务 1：抽离路径宿主与共享 helper

- 把 `.orbit/vars.yaml` 与 `.orbit/installs/*` 的硬编码路径从现有调用点中移除；
- 建立 `.harness/vars.yaml` 与 `.harness/installs/*` 的统一 host helper；
- 收口 install record 与 vars 的读写入口，避免后续命令直接拼路径。

任务 2：实现 harness runtime / template schema

- 定义 `.harness/runtime.yaml` 结构与校验；
- 定义 zero-member runtime 的合法写法；
- 定义 `.harness/template.yaml` 结构与校验；
- 补齐稳定写回排序与时间字段更新策略。

任务 3：建立双二进制最小入口

- 新增 `harness` 顶层二进制；
- 保持 `orbit` 入口不变；
- 让两者共享现有底层包与新增 `harness` 领域包。
- 把 `harness` 纳入最小 build / test 入口，确保后续命令不是“代码存在但无法发布”。

### 6.4 测试要求

必须新增：

- host helper 单测
- `.harness/runtime.yaml` schema 单测
- `.harness/template.yaml` schema 单测
- zero-member runtime 单测

### 6.5 完成标准

- 现有业务代码不再依赖硬编码 `.orbit/vars.yaml` / `.orbit/installs/*` 路径
- 可以从代码中稳定加载 / 写回 `.harness/runtime.yaml`
- `harness` 二进制可启动并提供稳定 root command 入口

本阶段已交付：

- `cmd/orbit/cli/harness/*` 共享领域包与 `.harness/*` path host helper
- `.harness/runtime.yaml` / `.harness/template.yaml` schema、validation 与单测
- `cmd/harness/main.go` 与 `cmd/harness/cli/root.go` 独立入口

---

## 7. Phase 3A-1：Harness Bootstrap 与最小可观测性

这一阶段完成 runtime bootstrap 命令、root 解析和最小 branch 可观测性。

### 7.1 目标

建立新的 runtime repo 初始化闭环，并让 harness runtime 在 branch 维度可被最小识别。

### 7.2 交付物

- `harness create`
- `harness init`
- `harness root`
- `harness inspect`
- 最小 harness runtime branch classifier

### 7.3 开发内容

任务 1：实现 bootstrap 命令

- `harness create`
  - 创建目录
  - 必要时执行 `git init`
  - 初始化 `.orbit/config.yaml`
  - 初始化 `.orbit/orbits/`
  - 初始化 `.harness/runtime.yaml`
- `harness init`
  - 要求目录已位于 Git repo 内
  - 已存在合法 runtime 时默认失败
- `harness root`
  - 输出包含合法 runtime 的 Git repo root
- `harness inspect`
  - 输出 harness root、harness id、member count、vars count、install count、current projection
  - 首批就冻结稳定 text / json 输出契约

任务 2：实现 harness root resolution

- 从任意目录回溯 Git repo root；
- 校验 `.harness/runtime.yaml` 必须位于 Git root；
- 冻结“不支持 repo 内子目录独立 harness root”的规则。

任务 3：补齐最小 branch 可观测性

- 让 classifier 至少能识别 harness runtime；
- 正确识别 zero-member runtime；
- 让 bootstrap 后的 repo 能被 branch inspect / status 基本理解，方便后续调试。

### 7.4 测试要求

必须新增：

- `harness create` temp repo 集成测试
- `harness init` temp repo 集成测试
- `harness root` temp repo 集成测试
- `harness inspect` temp repo 集成测试
- harness runtime 最小 classifier 测试
- `harness inspect` text / json golden tests

### 7.5 完成标准

- 用户可以从空目录或现有 Git repo 初始化 harness runtime
- runtime 初始化后允许 `members: []`
- bootstrap 后的 branch 已能被最小正确识别
- `harness inspect` 输出契约已冻结

---

## 8. Phase 3B-1：Runtime Host 切换与 Install Basic Path

这一阶段完成 vars/install host 切换，并交付不含 overwrite 的 install 主路径。

### 8.1 目标

打通 orbit template -> harness runtime 的基本安装闭环。

### 8.2 交付物

- `harness add`
- `harness remove`
- `orbit bindings init` 默认 stdout、推荐 `--out .harness/vars.yaml`
- `orbit template save` 读取 `.harness/vars.yaml`
- `harness install`
- 共享 apply service
- hidden/internal `orbit template apply` wrapper
- `.harness/installs/<orbit-id>.yaml` 正式写入

### 8.3 开发内容

任务 1：完成 runtime host 切换

- `bindings` 包移除固定 `.orbit/vars.yaml` 宿主路径责任；
- `orbit bindings init` 保持 stdout 默认；
- `orbit bindings init --out` 主文档推荐 `.harness/vars.yaml`；
- `orbit template save` 改从 `.harness/vars.yaml` 读取默认 bindings；
- install record 正式落到 `.harness/installs/*`。

任务 2：实现 member 基础管理

- `harness add`
  - 校验 definition 存在
  - 以 `source=manual` 写入 members
- `harness remove`
  - 只更新 `members`
  - 不自动删除 definition / install record / runtime files
  - 后续同 ID reinstall 仍由 install record / overwrite 规则约束

任务 3：抽共享 apply service

- 把现有 `template apply` 内核收敛成共享 service；
- 把 runtime file write、definition write、install record write、vars write、member update 统一纳入新 service；
- 确保非 `--dry-run` 时默认落盘。

任务 4：实现 `harness install` basic path

- 解析本地 / 远程 orbit template source；
- 从模板读取唯一 `orbit_id`；
- 读取 `.harness/vars.yaml` 与显式 bindings；
- 进行渲染与 fail-closed conflict analysis；
- 写入：
  - runtime files
  - `.orbit/orbits/<orbit-id>.yaml`
  - `.harness/installs/<orbit-id>.yaml`
  - `.harness/vars.yaml`
  - `.harness/runtime.yaml`
- 先只交付“首次安装成功 / 重复安装失败”的基础路径。

### 8.4 测试要求

必须新增：

- `harness add/remove` 单测与集成测试
- `orbit bindings init` 默认 stdout 测试
- `orbit bindings init --out .harness/vars.yaml` 测试
- `orbit template save` 读取 `.harness/vars.yaml` 测试
- `harness install <local-branch>` 集成测试
- `harness install <git-url>` 集成测试
- 同 ID 默认失败测试
- `orbit template apply` hidden wrapper 回归测试

### 8.5 完成标准

- vars/install host 已全面切到 `.harness/*`
- `harness install` 基础路径稳定可用
- install record 全面写入 `.harness/installs/*`

---

## 9. Phase 3B-2：Overwrite / Reinstall / Drift Foundations

这一阶段单独处理 install 最难的覆盖更新路径，并为后续 `harness check` 准备共享原语。

### 9.1 目标

稳定落地 overwrite、reinstall 和 owned file reconstruction 的边界。

### 9.2 交付物

- `harness install --overwrite-existing`
- old owned file reconstruction primitive
- reinstall / overwrite contract
- drift replay primitives

### 9.3 开发内容

任务 1：实现 overwrite 合同

- 同一 `orbit-id` 默认失败；
- 只有显式 `--overwrite-existing` 才进入覆盖路径；
- 首阶段不支持 `--as`；
- 若 install record 仍存在，则同 ID reinstall 继续按 overwrite 路径处理，而不是当作全新安装。

任务 2：实现 owned file reconstruction

- 根据 install record 重新解析旧模板快照；
- 重建旧 install-owned file set；
- 对旧来源不可安全重建的情况 fail-closed；
- 只在安全确认后删除旧模板已不再拥有的路径。

任务 3：准备 drift replay primitives

- 复用 overwrite 路径中的 source replay 与期望输出重建能力；
- 为后续 `definition_drift` / `runtime_file_drift` / `provenance_unresolvable` 诊断提供共享原语。

### 9.4 测试要求

必须新增：

- `--overwrite-existing` 成功覆盖测试
- remove 后同 ID reinstall 进入 overwrite 路径测试
- 旧 owned file set 无法安全重建时 fail-closed 测试
- drift replay primitives 单测

### 9.5 完成标准

- overwrite 路径稳定可用
- reinstall 语义明确且有测试覆盖
- drift replay 能支撑后续 `harness check`

---

## 10. Phase 3C：完整 Branch / Check 升级

这一阶段补齐完整可观察性和诊断能力，让新模型在 branch 层和检查层稳定可见。

### 10.1 目标

让 branch classification、inspect、runtime check 全部理解 harness runtime / harness template。

### 10.2 交付物

- harness-aware `orbit branch status`
- harness-aware `orbit branch inspect`
- harness-aware `orbit branch list`
- `harness check`

### 10.3 开发内容

任务 1：升级完整 branch classifier

- 识别 orbit template
- 识别 harness template
- 识别 harness runtime
- 处理 `.harness/template.yaml` 与 `.orbit/template.yaml` 同时出现的 invalid conflict
- 正确识别 zero-member runtime

任务 2：升级 inspect 契约

- 输出：
  - `kind`
  - `template_kind`
  - `harness_id`
  - `member_count`
  - `member_ids`
  - `definition_count`
  - `definition_ids`
  - `install_count`
  - `install_ids`
  - `includes_root_agents`
- inspect 只基于 branch 上可证明的信息，不读 `.git/orbit/state/*`
- 用 text / json golden tests 冻结输出。

任务 3：实现 `harness check`

- 校验 runtime schema
- 校验 duplicate member
- 校验 definition 缺失
- 校验 install record 与 members 不一致
- 校验 install record path / orbit_id 不一致
- 诊断 drift：
  - `definition_drift`
  - `runtime_file_drift`
  - `provenance_unresolvable`

### 10.4 测试要求

必须新增：

- branch classifier 单测
- zero-member runtime inspect 测试
- harness template / orbit template 分类测试
- `harness check` drift 分类测试
- `harness check` 在 zero-member runtime 下成功返回测试
- `orbit branch inspect` text / json golden tests

### 10.5 完成标准

- 新 branch taxonomy 在 text / json 输出中稳定
- drift 分类可复现、可测试
- `harness check` 能给出可操作的诊断结果

---

## 11. Phase 3D：Harness Template Save

这一阶段实现多 member runtime -> harness template 的正式导出闭环。

### 11.1 目标

让当前 harness runtime 可以导出为 harness template branch，并稳定处理变量、路径和根目录 `AGENTS.md`。

### 11.2 交付物

- `harness template save`
- member candidate builder
- candidate merge engine
- `.harness/template.yaml` writer

### 11.3 开发内容

任务 1：实现 member candidate builder

- 针对每个 member 读取 `.orbit/orbits/<orbit-id>.yaml`
- 用现有 scope resolver 得到 user scope
- 从 runtime repo 读取当前文件内容
- 用 `.harness/vars.yaml` 做 replacement
- 自动注入 `.orbit/orbits/<orbit-id>.yaml`

任务 2：实现组合与冲突分析

- 合并 candidate files
- 合并变量声明
- 路径冲突：
  - 相同内容允许
  - 不同内容 fail-closed
- 变量冲突：
  - 描述一致允许
  - 一个空一个非空取非空
  - 两个非空且不同 fail-closed
  - `required` 按 OR 合并

任务 3：实现根目录 `AGENTS.md` whole-file lane

- runtime root 存在 `AGENTS.md` 时按运行态容器读取；
- 先剥离 runtime markers，规范化为 payload-only 文本；
- 按普通模板文件做变量替换；
- 写入 harness template branch 根目录；
- 在 `.harness/template.yaml` 中写 `includes_root_agents=true`。

### 11.4 测试要求

必须新增：

- candidate merge 单测
- variable collision 单测
- path collision 单测
- root `AGENTS.md` replacement 测试
- root `AGENTS.md` marker/comment preservation 测试
- `harness template save` 集成测试

### 11.5 完成标准

- 能从多 member runtime 稳定导出 harness template
- `AGENTS.md` whole-file lane 行为稳定
- 不生成 `.orbit/template.yaml`

---

## 12. Phase 3E：文档、发布与硬化收尾

这一阶段统一对外口径，补完帮助、quickstart、测试矩阵与遗留入口收口。

### 12.1 目标

让代码实现、CLI help、发布入口、文档与测试矩阵完全一致。

### 12.2 交付物

- `docs/quickstart.md` 更新
- CLI help 文案更新
- dual-binary build / release / completion 收口
- JSON / text 输出收口
- 测试矩阵补齐

### 12.3 开发内容

任务 1：文档与帮助切换

- quickstart 主路径切到 `harness install`
- 主文档不再推荐 `orbit template apply`
- help 文案与示例切到 `harness` / `orbit` 双前缀模型

任务 2：发布面收口

- 把 `harness` 纳入 build / release / CI 正式产物；
- 补齐 shell completion、安装文档或等价的命令发现入口；
- 确保双二进制不会在发布链路中丢失其中一个入口。

任务 3：测试与收口

- 对照 `docs/testing-strategy.md` 补齐最低覆盖矩阵；
- 复查 `mise run fmt`、`mise run lint`、`mise run test:ci`；
- 补齐 branch / inspect / install / template save 的 JSON 契约测试。

任务 4：遗留入口处理

- `orbit template apply` 若保留，只作为 hidden/internal wrapper；
- 不让旧 runtime host 概念继续出现在正式输出与帮助中。

### 12.4 完成标准

- README / quickstart / help / PRD / spec / development plan 口径一致
- 双二进制 build / release 入口稳定
- 正式命令面稳定
- 回归测试通过

---

## 13. Issue 拆分建议

建议按下面粒度拆本地 issue：

1. path host helper 与 `.harness/*` shared primitives
2. harness 顶层二进制与最小 build / test 入口
3. `harness create/init/root/inspect`
4. 最小 harness runtime classifier
5. `harness add/remove`
6. vars host 切换
7. shared apply service 重构
8. `harness install` basic path
9. `harness install --overwrite-existing`
10. old owned file reconstruction
11. branch classifier / inspect / list 完整升级
12. `harness check`
13. `harness template save`
14. root `AGENTS.md` whole-file lane
15. build / release / help / quickstart 收口
16. 测试矩阵与 hardening

拆分原则：

- 每个 issue 都应形成独立可 review 的边界；
- install basic path、overwrite、owned file reconstruction 不要塞进同一个 PR；
- install overwrite 和 drift check 不要与 bootstrap 命令混在同一个 PR；
- `AGENTS.md` whole-file lane 单独拆 issue，避免污染普通 merge 逻辑。

---

## 14. 阶段验收与验证

每个阶段合并前，至少执行：

```bash
mise run fmt
mise run lint
mise run test:ci
```

阶段总体验收要求：

- `orbit` projection 主链路无行为回退；
- `.harness/*` 成为唯一正式 runtime metadata host；
- `harness install` 成为唯一正式安装入口；
- zero-member runtime、overwrite-existing、drift check、root `AGENTS.md` 都有测试覆盖；
- branch inspect / status / list 能正确表达 harness runtime / harness template。

---

## 15. 一句话总结

v0.3 的开发顺序应是：

**先做 preflight foundations 与双二进制入口，再完成 bootstrap 和最小可观测性，再交付 install basic path，随后单独收口 overwrite / drift，最后补完整 branch/check、template save 与发布硬化。**
