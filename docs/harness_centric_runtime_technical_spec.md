# Harness-Centric Runtime Technical Spec（技术方案）

版本：v0.3
状态：实现基线草案
关联文档：
- `docs/harness_centric_runtime_prd.md`
- `docs/harness_centric_runtime_development_plan.md`
- `docs/harness_cli_full_recommendations.md`
- `docs/harness_install_progress_prd.md`
- `docs/harness_install_source_repo_resolution_technical_spec.md`
- `docs/harness_mixed_install_technical_spec.md`
- `docs/context/mvp-product-requirements.md`
- `docs/context/mvp-technical-architecture.md`
- `docs/testing-strategy.md`

---

## 1. 文档定位

本文档定义 Harness-Centric Runtime 的实现方案。

它承接产品 PRD 与 CLI 建议稿，回答以下问题：

1. v0.3 下哪些存储边界正式变化；
2. 当前代码里哪些能力继续复用，哪些能力必须重构；
3. `orbit` 与 `harness` 两条命令线如何分层；
4. 下一阶段应按什么顺序落地。

本文档是 clean-break 技术基线，不是迁移兼容方案。

因此：

- 主文档不继续推荐 `orbit template apply`；
- 不引入 `.orbit/*` 与 `.harness/*` 的双写；
- 不把“兼容旧 runtime host”当作主实现目标；
- 若实现需要短期隐藏兼容入口，只能视为过渡实现细节，而不是产品合同。

---

## 2. 已冻结结论

本阶段直接冻结以下结论：

1. `orbit` 继续负责定义、projection、单 orbit template。
2. `harness` 成为唯一正式运行态对象与正式安装入口。
3. 单 repo 只允许一个 harness runtime。
4. `.orbit/config.yaml` 继续保留为 projection control plane。
5. `.orbit/orbits/*.yaml` 继续保留为 orbit definitions。
6. 运行态版本化元数据从 `.orbit/` 收敛到 `.harness/`。
7. `.git/orbit/state/` 继续只承载 repo-local projection state / cache。
8. `harness install` 默认就是“写入运行态”；只有显式 `--dry-run` 时不落盘。
9. `orbit bindings init` 继续默认写 stdout；只有显式 `--out` 时才写文件。
10. `orbit bindings init --out` 的推荐目标路径改为 `.harness/vars.yaml`。
11. `orbit template save` 读取 runtime bindings 的默认来源改为 `.harness/vars.yaml`。
12. `harness template save` 可以读取运行态根目录 `AGENTS.md` 容器，规范化为 payload-only 文本后做变量替换，再写入 harness template branch。
13. `harness template save` 中的 `AGENTS.md` 不复用 v0.2 orbit-level shared-file lane，而是按 harness-level 模板文件处理。
14. 首阶段不做 harness template apply。
15. 首阶段不把现有共享包整体搬出 `cmd/orbit/cli/*`；先做有边界的增量重构。
16. zero-member harness runtime 合法；`harness create` / `harness init` 后 `members: []` 不视为异常。
17. `harness install` 对同一 `orbit-id` 的再次安装默认 fail-closed；只有显式 `--overwrite-existing` 才允许覆盖既有安装实例。
18. `harness install` 首阶段不支持 `--as`；安装 identity 固定使用模板内 `orbit_id`。
19. 已安装实例的 materialized definition 与运行态文件允许本地编辑；一致性问题由 `harness check` 诊断，不做自动回收或后台重写。

---

## 3. 稳定前提

v0.3 仍然继承以下 MVP 不变量：

1. Git DAG 仍是唯一历史真相源。
2. projection 仍然由 orbit 驱动，而不是由 harness 驱动。
3. `enter` / `leave` 仍然是唯一正式 projection 控制入口。
4. scoped read/write 仍然依赖 pathspec 与 sparse-checkout，不引入 worktree、daemon、服务端状态。
5. `.git/orbit/state/` 仍然不承载版本化内容。

换句话说：

- runtime host 变了；
- projection model 没变。

---

## 4. 对象模型

### 4.1 Orbit

Orbit 继续是最小定义单元。

它负责：

- 文件范围定义；
- projection 目标；
- 单 orbit template 导出源；
- companion control file `.orbit/orbits/<orbit-id>.yaml`。

它不负责：

- 作为正式 runtime object 存在；
- 承载 repo-level vars；
- 承载 install records；
- 表达 runtime members 组合关系。

### 4.2 Orbit Template

Orbit template 继续表示单条 orbit 的模板态快照。

它继续使用：

- `.orbit/template.yaml`
- `.orbit/orbits/<orbit-id>.yaml`
- 模板化后的 owned files
- 既有 remote source resolution 规则

### 4.3 Harness Runtime

Harness runtime 是唯一正式运行态对象。

它表示：

- 当前 repo 是否已经启用运行态；
- 当前 repo 的 harness 身份；
- 当前 repo 当前声明了哪些 orbit members；
- 当前 repo 的 repo-level vars；
- 当前 repo 的安装记录。

### 4.4 Harness Template

Harness template 是多个 orbit members 组合后的模板态快照。

它不是：

- orbit template 的 ref 列表；
- runtime repo 的全量文件盲拷贝；
- 通用 shared-file 系统的第一次落地。

它是：

- 一个独立模板对象；
- 由当前 harness members 组合得到；
- 由 `.harness/template.yaml` 标识。

### 4.5 Projection

Projection 仍然表示“当前工作区进入的 orbit 视图”。

因此：

- current orbit state 仍保存在 `.git/orbit/state/current_orbit.json`；
- `orbit current / enter / leave / status / diff / log / commit / restore` 保持 orbit-centric；
- harness 的存在不改变 projection 的 path resolution 语义。

---

## 5. 存储边界与文件合同

## 5.1 Runtime Repo

运行态仓库允许存在：

```text
.orbit/
  config.yaml
  orbits/
    <orbit-id>.yaml

.harness/
  runtime.yaml
  vars.yaml
  installs/
    <orbit-id>.yaml

.git/orbit/state/
  current_orbit.json
  resolved_scope/
  warnings.json
  last_status.json
  orbit.lock
```

### `.orbit/`

`.orbit/` 在 v0.3 只负责：

- projection control plane
- orbit definitions
- orbit template manifest（仅 orbit template branch）

`.orbit/` 不再负责：

- runtime vars
- install records
- harness runtime identity
- harness members

### `.harness/`

`.harness/` 在 v0.3 负责：

- harness runtime identity
- members
- runtime vars
- install records
- harness template manifest（仅 harness template branch）

### `.git/orbit/state/`

`.git/orbit/state/` 不变，继续只负责：

- current projection state
- projection cache
- warnings / status snapshots
- repo-local lock

它不负责：

- runtime identity
- shared vars
- install history
- template metadata

## 5.2 `.harness/runtime.yaml`

建议合同：

```yaml
schema_version: 1
kind: harness_runtime
harness:
  id: project-a
  name: Project A
  created_at: 2026-03-25T12:00:00Z
  updated_at: 2026-03-25T12:30:00Z
members:
  - orbit_id: cli
    source: manual
    added_at: 2026-03-25T12:05:00Z
  - orbit_id: docs
    source: install
    added_at: 2026-03-25T12:10:00Z
```

规则：

1. `schema_version` 固定为 `1`。
2. `kind` 固定为 `harness_runtime`。
3. `harness.id` 必填，首阶段使用与 `orbit-id` 相同的安全字符规则。
4. `harness.name` 可选。
5. `members` 可以为空数组；空 members 仍然表示一个合法 runtime。
6. `members` 必须显式声明，不做“扫描 repo 自动推断”。
7. `members[].orbit_id` 全局唯一。
8. `members[].source` 首阶段只允许：
   - `manual`
   - `install`
9. 文件写回时对 `members` 采用稳定排序，按 `orbit_id` 排序。

## 5.3 `.harness/vars.yaml`

`.harness/vars.yaml` 复用 v0.2 `.orbit/vars.yaml` 的 schema：

```yaml
schema_version: 1
variables:
  project_name:
    value: HarnessOS
    description: 项目名称
```

变化只有一处：

- host path 从 `.orbit/vars.yaml` 改为 `.harness/vars.yaml`。

因此：

- `bindings` merge / skeleton / codec 可直接复用；
- 固定路径常量必须从 `bindings` 包中抽离。

## 5.4 `.harness/installs/<orbit-id>.yaml`

`.harness/installs/<orbit-id>.yaml` 复用 v0.2 install record schema：

```yaml
schema_version: 1
orbit_id: docs
template:
  source_kind: remote_git
  source_repo: https://example.com/acme/templates.git
  source_ref: orbit-template/docs
  template_commit: abc123
applied_at: 2026-03-25T13:00:00Z
```

变化也只有一处：

- host path 从 `.orbit/installs/` 改为 `.harness/installs/`。

补充规则：

1. 文件名中的 `<orbit-id>` 必须与 `orbit_id` 字段完全一致。
2. 同一 `orbit-id` 最多只允许一个 install record。
3. `--overwrite-existing` 成功覆盖时，install record 在原路径原位更新。

## 5.5 Orbit Template Branch

orbit template branch 合同保持：

```text
.orbit/
  template.yaml
  orbits/
    <orbit-id>.yaml

<owned template files...>
AGENTS.md            # 仅当共享 AGENTS lane 被声明时出现
```

仍然禁止：

- `.harness/runtime.yaml`
- `.harness/vars.yaml`
- `.harness/installs/*.yaml`
- `.git/orbit/state/*`

## 5.6 Harness Template Branch

harness template branch 新增合同：

```text
.harness/
  template.yaml

.orbit/
  orbits/
    <orbit-id>.yaml   # 每个 member 一份

<combined template files...>
AGENTS.md             # 仅当 runtime root 存在该文件时出现
```

harness template branch 中禁止：

- `.harness/runtime.yaml`
- `.harness/vars.yaml`
- `.harness/installs/*.yaml`
- `.orbit/config.yaml`
- `.orbit/template.yaml`
- `.git/orbit/state/*`

## 5.7 `.harness/template.yaml`

建议合同：

```yaml
schema_version: 1
kind: harness_template
template:
  harness_id: project-a
  default_template: false
  created_from_branch: main
  created_from_commit: abc123
  created_at: 2026-03-25T13:30:00Z
  includes_root_agents: true
members:
  - orbit_id: cli
  - orbit_id: docs
variables:
  project_name:
    description: 项目名称
    required: true
```

规则：

1. `kind` 固定为 `harness_template`。
2. `template.harness_id` 必填。
3. `members` 必须与导出时的 runtime members 对齐。
4. `variables` 保存组合后变量元信息，不保存 concrete values。
5. `includes_root_agents=true` 表示模板树中包含根目录 `AGENTS.md`。

---

## 6. 命令面与寻址规则

## 6.1 `orbit ...`

`orbit` 继续负责：

- orbit 定义
- projection
- 单 orbit template 导出
- bindings skeleton 生成
- branch 分类与 inspect

正式保留：

```text
orbit add
orbit validate
orbit list
orbit show
orbit files
orbit current
orbit enter
orbit leave
orbit status
orbit diff
orbit log
orbit commit
orbit restore
orbit template save
orbit bindings init
orbit branch status
orbit branch inspect
orbit branch list
```

不再出现在主文档：

```text
orbit template apply
```

实现策略：

- 可以短期保留 hidden/internal alias；
- 但 help、quickstart、主示例、主测试口径不再把它作为正式入口。

## 6.2 `harness ...`

新增正式命令：

```text
harness create
harness init
harness inspect
harness add
harness remove
harness install
harness check
harness template save
harness root
```

## 6.3 Harness Root 规则

v0.3 首阶段直接冻结：

1. harness root 必须位于 Git repo root。
2. `.harness/runtime.yaml` 必须位于 Git repo root 下。
3. 不支持 repo 内子目录再挂一个独立 harness root。

这样做的原因：

- `orbit` 现有实现已经以 repo root 为绝对基准；
- projection、pathspec、install records、template export 都以 repo root 相对路径工作；
- 先不引入“Git root 与 harness root 分离”的第二套解析语义。

因此：

- `harness root` 实际输出的是“包含合法 `.harness/runtime.yaml` 的 Git repo root”；
- `--path` 只改变从哪个目录开始寻找；
- 找到的目标最终仍然必须是 Git repo root。

## 6.4 `harness create`

语义：

1. 创建目标目录；
2. 若目录不是 Git repo，则执行 `git init`；
3. 初始化 `.orbit/config.yaml` 与 `.orbit/orbits/`；
4. 初始化 `.harness/runtime.yaml`；
5. 初始 `members` 为空数组；
6. 不自动安装 orbit；
7. 不自动 `enter`。

## 6.5 `harness init`

语义：

1. 要求目标目录已位于 Git repo 内；
2. 若尚未初始化，则创建 `.orbit/config.yaml`、`.orbit/orbits/`、`.harness/runtime.yaml`；
3. 若已存在合法 `.harness/runtime.yaml`，默认稳定失败；
4. 初始化后的 runtime 可以暂时没有任何 member；
5. 不清空已有工作树内容。

## 6.6 `orbit bindings init`

冻结口径：

1. 默认输出到 stdout。
2. 仅当显式传 `--out` 时才写文件。
3. 主文档推荐：

```text
--out .harness/vars.yaml
```

而不是自动隐式写这个路径。

原因：

- 保持命令默认非破坏；
- 避免“用户只是看 skeleton，却意外修改 runtime repo”；
- 与现有命令心智保持一致。

## 6.7 `harness install`

这是唯一正式安装入口。

非 `--dry-run` 模式下，它默认直接写入 runtime repo：

1. 渲染并写入运行态文件；
2. 写入 `.orbit/orbits/<orbit-id>.yaml`；
3. 写入 `.harness/installs/<orbit-id>.yaml`；
4. 必要时写入或更新 `.harness/vars.yaml`；
5. 把该 orbit 写入 `.harness/runtime.yaml` 的 members；
6. 不自动 `enter`。

附加规则：

1. 若同一 `orbit-id` 已存在 install-backed 实例，默认 fail-closed。
2. 只有显式 `--overwrite-existing` 才允许覆盖既有 install-backed 实例。
3. 若同一 `orbit-id` 已被 `manual` member 占用，则仍然 fail-closed；首阶段要求用户先显式 `harness remove <orbit-id>` 再安装。
4. 首阶段不支持 `--as`，不允许在安装时重命名 orbit identity。
5. 安装期的同名变量冲突先按 **变量声明兼容性** 判定，而不是只按整个 `.harness/vars.yaml` 文件变化判定。
6. 若 runtime 中已存在同名变量且声明兼容：
   - 默认复用 runtime 中已有值；
   - 更新后的声明按“兼容即合并”规则收口；
   - 安装继续执行。
7. 若 runtime 中已存在同名变量但声明不兼容：
   - 在写入 install 结果前 fail-closed；
   - 诊断至少指出变量名与冲突来源。
8. 第一版不引入变量 namespacing、自动 rename，`--var-conflicts=...` 之类的策略面保持 deferred。

---

## 7. 包结构与重构策略

## 7.1 顶层二进制

新增：

```text
cmd/
  orbit/
    main.go
    cli/
      root.go
      commands/
        ...
  harness/
    main.go
    cli/
      root.go
      commands/
        create.go
        init.go
        inspect.go
        add.go
        remove.go
        install.go
        check.go
        template_save.go
        root.go
```

## 7.2 共享领域包

首阶段不做“大搬家”，继续复用当前共享包：

```text
cmd/orbit/cli/git
cmd/orbit/cli/ids
cmd/orbit/cli/orbit
cmd/orbit/cli/view
cmd/orbit/cli/scoped
cmd/orbit/cli/state
cmd/orbit/cli/template
cmd/orbit/cli/bindings
cmd/orbit/cli/branchinfo
```

新增一个共享领域包：

```text
cmd/orbit/cli/harness
```

负责：

- `.harness/runtime.yaml` schema / load / write
- `.harness/vars.yaml` path host
- `.harness/installs/*.yaml` path host
- harness root resolution
- member mutation
- harness inspect / check shared primitives
- `.harness/template.yaml` schema / load / write

这是一个刻意保守的阶段性选择：

- 先稳定 runtime host 切换；
- 不把阶段目标扩成“顺便重构整个包路径”。

## 7.3 现有包的直接改造点

### `bindings`

`bindings` 保留：

- vars schema codec
- merge engine
- skeleton builder

`bindings` 去掉：

- `.orbit/vars.yaml` 作为固定宿主路径的责任

也就是：

- `VarsFile` 仍在 `bindings`
- `VarsPath` / load / write host 责任迁到 `harness`

### `template`

`template` 保留：

- orbit template manifest
- template source resolver
- render / scan / replacement
- orbit template save

`template` 移除或弱化：

- runtime install record 的宿主路径责任
- “正式安装命令”身份

同时新增：

- harness template manifest
- harness template save builder

### `branchinfo`

`branchinfo` 需要升级为：

- 识别 orbit template
- 识别 harness template
- 识别 harness runtime
- `kind` 仍保持 `template|runtime|plain`
- 新增 subtype：
  - `template_kind=orbit|harness`

---

## 8. 关键流程

## 8.1 `orbit template save`

新流程只改一处：

- runtime bindings 来源从 `.orbit/vars.yaml` 改为 `.harness/vars.yaml`。

其他仍保持：

- 以 orbit scope 为模板内容来源；
- 模板变量识别 / runtime-to-template replacement 仅对 Markdown 文件（`.md`）生效；
- temp dir 编辑；
- 写入 `.orbit/template.yaml`；
- 不污染 runtime worktree；
- 不写 runtime `.harness/*`。

## 8.2 `harness install`

建议直接复用现有 apply 内核，但改为 harness-hosted runtime write：

1. 解析本地或远程 orbit template source；
2. 从模板读取唯一的 `orbit_id`，并将它作为唯一安装 identity；
3. 解析 bindings：
   - `--bindings`
   - `.harness/vars.yaml`
   - `--interactive` / `--editor`
   - 模板变量渲染仅对 Markdown 文件（`.md`）生效，其他文件原样保留
4. 检查当前 runtime 中是否已存在同一 `orbit-id`：
   - 若是 install-backed 实例且未传 `--overwrite-existing`，fail-closed；
   - 若是 install-backed 实例且传了 `--overwrite-existing`，进入覆盖更新路径；
   - 若是 `manual` member 或只有本地 definition 但没有 install record，fail-closed；
5. 渲染运行态文件；
6. 做 fail-closed conflict analysis；
7. 若为覆盖更新路径，先根据现有 install record 重新解析旧模板快照，重建旧 owned file set；若无法安全重建，则 fail-closed；
8. 非 `--dry-run` 时写入：
   - owned runtime files
   - `.orbit/orbits/<orbit-id>.yaml`
   - `.harness/installs/<orbit-id>.yaml`
   - `.harness/vars.yaml`（如需要）
   - `.harness/runtime.yaml` member set
9. 覆盖更新成功后：
   - install record 原位更新；
   - `members` 中该项继续保留 `source=install`；
   - 旧 owned file set 中已不再被新模板拥有的路径，只有在安全确认后才删除。

这里推荐实现方式：

- 把现有 `template apply` 改造成共享 service；
- `harness install` 成为正式 CLI 封装；
- `orbit template apply` 若暂时保留，仅作 hidden wrapper。

## 8.3 `harness add`

流程：

1. 读取 `.harness/runtime.yaml`
2. 校验 `.orbit/orbits/<orbit-id>.yaml` 存在且合法
3. 若 member 已存在则 fail-closed
4. 以 `source=manual` 追加 member
5. 更新 `updated_at`

## 8.4 `harness remove`

流程：

1. 读取 `.harness/runtime.yaml`
2. 若 member 不存在则 fail-closed
3. 从 member set 删除
4. 更新 `updated_at`
5. 不自动删除：
   - user files
   - `.harness/installs/<orbit-id>.yaml`
   - projection cache
6. 若对应 install record 仍存在，则后续同一 `orbit-id` 的 reinstall 继续按 overwrite 路径处理；首阶段不把 `remove` 视为“释放 install slot”。

## 8.5 `harness inspect`

输出建议至少包含：

- harness root
- harness id / name
- member count / members
- vars count
- install count
- current projection（仅从 `.git/orbit/state/current_orbit.json` best-effort 读取）

zero-member runtime 下：

- `member_count=0`
- `members=[]`
- 仍然视为有效 inspect 结果，而不是异常状态

## 8.6 `harness check`

首阶段最少检查：

1. `runtime.yaml` schema 合法性
2. duplicate member
3. member 指向的 orbit definition 缺失
4. install record 与 members 不一致
5. install record orbit_id 与路径不一致
6. 如果尝试导出 harness template，组合后会发生：
   - variable collisions
   - content/path collisions

推荐实现：

- `harness check` 直接复用 `harness template save --dry-run` 的部分组合分析原语；
- 但默认不写 branch。
- zero-member runtime 只要 schema 合法，就应返回成功检查结果，不额外制造 warning。

drift 合同：

1. install-backed 实例允许在 runtime repo 中发生本地编辑。
2. `harness check` 负责把这种编辑显式分类为 drift，而不是自动回滚。
3. 对每个 install-backed member，`harness check` 应尽量根据 install record 中的 source pin 与当前 `.harness/vars.yaml` 重建期望输出。
4. 首阶段至少报告三类 drift：
   - `definition_drift`：当前 `.orbit/orbits/<orbit-id>.yaml` 与期望 definition 不一致；
   - `runtime_file_drift`：当前 materialized runtime files 与期望渲染结果不一致；
   - `provenance_unresolvable`：无法根据 install record 安全解析或重建期望输出。
5. drift 是诊断结果，不自动阻断 projection；推荐修复路径是显式重新执行 `harness install --overwrite-existing` 或人工整理本地内容。

## 8.7 `harness template save`

流程：

1. 读取 `.harness/runtime.yaml`
2. 取出 members
3. 对每个 member 构建 orbit template candidate
4. 合并 candidate files
5. 合并 variable declarations
6. 额外处理根目录 `AGENTS.md`
7. 进入可选最终模板编辑
8. 写出 harness template branch

### 成员模板候选构建规则

对每个 member orbit：

1. 读取其 orbit definition；
2. 用现有 orbit scope resolver 得到 user scope；
3. 读取 runtime file content；
4. 仅对 Markdown 文件用 `.harness/vars.yaml` 做 replacement；其它文件原样保留；
5. 自动注入 `.orbit/orbits/<orbit-id>.yaml`；
6. 不生成 `.orbit/template.yaml`；
7. 不使用 orbit-level AGENTS shared lane。

### `AGENTS.md` 规则

`harness template save` 对根目录 `AGENTS.md` 采用 whole-file 语义：

1. 若 runtime root 存在 `AGENTS.md`，读取该运行态容器文件；
2. 若其中包含 runtime block marker，先规范化为 payload-only 文本，再继续导出；
3. 对规范化后的内容执行与普通 Markdown 模板文件一致的变量替换；
4. 将替换后的结果直接写入 harness template branch 根目录 `AGENTS.md`；
5. 在 `.harness/template.yaml` 中将 `includes_root_agents=true`；
6. 它不做 orbit block 归属推断，也不把 unmarked prose 归因给某个 orbit；
7. 不依赖 `.orbit/template.yaml` 的 `shared_files` 声明；
8. 不把它当作本阶段的通用 shared-file 平台起点；
9. runtime root `AGENTS.md` 在这里被视为容器 / 导出输入，不是第一性真相源。

这个规则与 orbit template save 明确不同：

- orbit template save 仍沿用 v0.2 的 orbit-level AGENTS lane；
- harness template save 把 runtime root `AGENTS.md` 当作 harness-level 模板文件。

### 合并冲突规则

#### 路径冲突

若两个 member candidate 想写同一路径：

- 内容完全相同：允许合并
- 内容不同：fail-closed

#### 变量冲突

若多个 member 引用同名变量：

- `description` 相同：允许合并
- 一个为空、一个非空：取非空值
- 两个都非空且不同：fail-closed

`required` 采用 OR 规则：

- 任一 member 需要该变量，则组合后 `required=true`

---

## 9. Branch Classification

## 9.1 分类规则

分类优先级建议如下：

1. valid `.harness/template.yaml` -> `kind=template`, `template_kind=harness`
2. valid `.orbit/template.yaml` -> `kind=template`, `template_kind=orbit`
3. valid `.harness/runtime.yaml` + valid `.orbit/config.yaml` -> `kind=runtime`
4. 其它 -> `kind=plain`

额外约束：

1. 若同一 revision 同时存在合法 `.harness/template.yaml` 与 `.orbit/template.yaml`，按 invalid conflict 处理，不静默选一个。
2. runtime branch 不再依赖 `.orbit/installs/*` 判断。
3. runtime branch 即使当前 `member_count=0`、`definition_count=0`，也可以被识别为 harness runtime branch。

## 9.2 Inspect 输出

`orbit branch inspect` 在 v0.3 应扩展为：

- `kind`
- `template_kind`（当 kind=template）
- `harness_id`
- `member_count`
- `member_ids`
- `definition_count`
- `definition_ids`
- `install_count`
- `install_ids`
- `includes_root_agents`（当 template_kind=harness）

补充约束：

1. 当 `kind=runtime` 且当前 branch 是 zero-member runtime 时：
   - `member_count=0`
   - `definition_count=0`
   - `install_count=0`
   - `member_ids` / `definition_ids` / `install_ids` 为空数组
2. inspect 输出保持“branch 上可证明的信息”原则，不从 repo-local `.git/orbit/state/*` 推断 branch 元数据。

---

## 10. 测试要求

## 10.1 Unit Tests

新增必须覆盖：

1. `.harness/runtime.yaml` schema / load / write
2. harness root resolution
3. member add/remove 规则
4. `.harness/vars.yaml` host path
5. `.harness/installs/*.yaml` host path
6. orbit template save 改从 `.harness/vars.yaml` 读 vars
7. `harness install` 结果写入 `.harness/*`
8. `harness install` 默认重复安装失败，`--overwrite-existing` 覆盖成功路径
9. zero-member runtime 的 schema / inspect / classifier 行为
10. `harness check` 的 drift 分类
11. `harness template save` 的：
   - candidate merge
   - variable collision
   - path collision
   - root `AGENTS.md` replacement
   - root `AGENTS.md` runtime marker stripping / payload normalization
12. branch classifier 的 `template_kind=harness|orbit`

## 10.2 Temp Repo Integration Tests

新增必须覆盖：

1. `harness create`
2. `harness init`
3. `harness root`
4. `harness inspect`
5. `harness add`
6. `harness remove`
7. `harness install <local-branch>`
8. `harness install <git-url>`
9. `harness install` 同 ID 默认失败
10. `harness install --overwrite-existing`
11. `harness check`
12. `harness template save`
13. `orbit template save` 读取 `.harness/vars.yaml`
14. `orbit bindings init` 默认 stdout
15. `orbit bindings init --out .harness/vars.yaml`
16. `orbit branch status/list/inspect` 的 harness-aware 输出

## 10.3 明确回归点

以下旧能力必须保持：

1. `orbit enter / leave / current / status / diff / log / commit / restore`
2. `.git/orbit/state/*` 读写合同
3. sparse-checkout hidden-dirty gate
4. orbit template save 的 temp-dir final edit
5. remote template source resolution

---

## 11. 建议实现顺序

建议按下面顺序推进：

### Phase 3A：Runtime Host 冻结

1. 新增本文档
2. 新增 `.harness/runtime.yaml` / `.harness/template.yaml` 合同
3. 新增共享 `harness` 包
4. 把 vars/install host 从 `.orbit/*` 切到 `.harness/*`

### Phase 3B：Harness CLI 最小闭环

1. `harness create`
2. `harness init`
3. `harness root`
4. `harness inspect`
5. `harness add`
6. `harness remove`

### Phase 3C：正式安装入口切换

1. 共享 apply service 改造成 harness-hosted runtime write
2. 新增 `harness install`
3. `orbit template apply` 退为 hidden/internal wrapper
4. quickstart/help 切到 `harness install`

### Phase 3D：Branch 与检查升级

1. harness-aware branch classifier
2. `harness check`
3. `orbit branch status/list/inspect` 输出升级

### Phase 3E：Harness Template 导出

1. member candidate builder
2. merge engine
3. root `AGENTS.md` replacement lane
4. `harness template save`

---

## 12. 当前代码的直接落点

首轮实现会直接触达以下现有代码：

1. `cmd/orbit/cli/bindings/vars.go`
2. `cmd/orbit/cli/template/install.go`
3. `cmd/orbit/cli/template/save.go`
4. `cmd/orbit/cli/template/apply.go`
5. `cmd/orbit/cli/template/content_builder.go`
6. `cmd/orbit/cli/branchinfo/classify.go`
7. `cmd/orbit/cli/branchinfo/inspect.go`
8. `cmd/orbit/cli/root.go`
9. `docs/quickstart.md`
10. 帮助与集成测试

这意味着 v0.3 不是“加几个新命令”。

它本质上是：

- 一次 runtime host 切换
- 一次正式安装入口切换
- 一次 template branch taxonomy 扩展

但它不是：

- projection model 重写
- Git adapter 重写
- `.git/orbit/state/` 重构

---

## 13. 非目标

本阶段明确不做：

1. 多 harness 共存
2. harness template apply
3. dual-write `.orbit/*` 与 `.harness/*`
4. 自动迁移旧 repo
5. 把现有所有共享包统一搬到新目录
6. 泛化 harness-level shared-file framework
7. 重新设计 orbit projection 语义

---

## 14. 一句话总结

v0.3 的正确技术方向不是“继续扩展 orbit runtime”，而是：

**保留 orbit projection 内核，把 runtime host 明确切到 `.harness/`，再以 `harness install` 和 `harness template save` 建立新的正式运行态闭环。**
