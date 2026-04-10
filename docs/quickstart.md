# Orbit / Harness Quickstart

这份文档是当前 v0.4 的总入口。

Orbit 的目标不是让所有人都先理解全部控制面，而是让不同用户尽快进入自己的主路径。

---

## 1. 先选你的路径

如果你是执行任务的 worker：

- 看 [docs/worker_guide.md](./worker_guide.md)

如果你是单个 orbit 的作者：

- 看 [docs/orbit_author_guide.md](./orbit_author_guide.md)

如果你是整个 harness 的作者：

- 看 [docs/harness_author_guide.md](./harness_author_guide.md)

如果你要一条从模板源到 runtime、再到 template 导出的完整可执行示例，继续看本文后半部分。

---

## 2. 最小心智模型

- `orbit`
  - 一个局部工作单元
- `harness`
  - 整个工作环境
- 根 `AGENTS.md`
  - 宏观任务编排入口
- orbit brief
  - 当前 orbit 的局部入口

Orbit 只负责外部控制：

- 现在处于哪个 orbit
- 当前看见什么
- 当前允许写什么
- 当前规则、探针、记录目标是什么

Orbit 不负责 worker 的内部执行方法。

---

## 3. 三类用户的最短路径

## 3.1 Worker

```bash
harness inspect
orbit list
orbit show docs
orbit enter docs
orbit current
orbit status
```

做完工作后：

```bash
orbit diff
orbit commit -m "update docs orbit"
orbit leave
```

## 3.2 Orbit 作者

最常见的三个入口：

- 直接维护单个 orbit：`harness init` + `orbit add`
- source 分支开发再发布：`orbit template init-source`
- runtime 优化后写回：`orbit template save <orbit-id>`

## 3.3 Harness 作者

```bash
harness create demo-repo
orbit bindings init <template-source> > /tmp/template.vars.yaml
# merge into .harness/vars.yaml
harness install <template-source> --bindings .harness/vars.yaml --dry-run
harness install <template-source> --bindings .harness/vars.yaml
orbit brief materialize --orbit <orbit-id>
harness inspect
harness check --json
harness template save --to harness-template/workspace
```

完整的“多 orbit 安装 -> 统一变量绑定 -> 编排根 `AGENTS.md`”主路径见 [docs/harness_author_guide.md](./harness_author_guide.md)。

---

## 4. 完整可执行示例

下面这条主路径覆盖当前 v0.4 双二进制的完整骨架：

- 用 `orbit` 定义 orbit、导出单 orbit template
- 用 `orbit template init-source` 验证 `source` revision
- 用 `harness` 创建 runtime、安装模板、检查 runtime
- 用 `orbit` 做 projection 与 runtime writeback
- 用 `harness template save` 导出 harness template

## 4.1 准备二进制

在仓库根目录执行：

```bash
export ORBIT_DEMO_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/orbit-demo.XXXXXX")"
export ORBIT_BIN_DIR="$ORBIT_DEMO_ROOT/bin"
export ORBIT_BIN="$ORBIT_BIN_DIR/orbit"
export HARNESS_BIN="$ORBIT_BIN_DIR/harness"
export TEMPLATE_REPO="$ORBIT_DEMO_ROOT/template-source"
export SOURCE_REPO="$ORBIT_DEMO_ROOT/source-authoring"
export RUNTIME_REPO="$ORBIT_DEMO_ROOT/runtime-repo"
export INSTALL_BINDINGS="$ORBIT_DEMO_ROOT/install-bindings.yaml"

mkdir -p "$ORBIT_BIN_DIR"
sh ./scripts/build_binaries.sh "$ORBIT_BIN_DIR"

printf 'orbit binary: %s\n' "$ORBIT_BIN"
printf 'harness binary: %s\n' "$HARNESS_BIN"
printf 'template repo: %s\n' "$TEMPLATE_REPO"
printf 'source repo: %s\n' "$SOURCE_REPO"
printf 'runtime repo: %s\n' "$RUNTIME_REPO"
```

## 4.2 创建模板源仓库

先准备一个会产出 `orbit-template/docs` 的源仓库：

```bash
mkdir -p "$TEMPLATE_REPO"
cd "$TEMPLATE_REPO"

git init -b main
git config user.name "Orbit Demo"
git config user.email "orbit-demo@example.com"

"$HARNESS_BIN" init
"$ORBIT_BIN" add docs
```

`orbit add docs` 会直接把最小 OrbitSpec skeleton 写到 `.harness/orbits/docs.yaml`。

再补最小变量绑定和 runtime 内容：

```bash
cat > .harness/vars.yaml <<'EOF'
schema_version: 1
variables:
  project_name:
    value: Orbit
    description: Product title
EOF

mkdir -p docs
cat > docs/guide.md <<'EOF'
Orbit guide
EOF
```

提交后导出一个 orbit template branch：

```bash
git add .harness docs
git commit -m "seed docs orbit"

"$ORBIT_BIN" template save docs --to orbit-template/docs
"$ORBIT_BIN" branch inspect orbit-template/docs --json
```

到这里，模板源仓库已经准备好。

## 4.3 补一条 source branch 作者路径

如果你还想验证第四类 revision kind，也就是作者态 `source`，可以单独准备一个 source repo：

```bash
mkdir -p "$SOURCE_REPO"
cd "$SOURCE_REPO"

git init -b main
git config user.name "Orbit Demo"
git config user.email "orbit-demo@example.com"

mkdir -p .orbit/orbits docs
cat > .orbit/config.yaml <<'EOF'
version: 1
shared_scope: []
behavior:
  outside_changes_mode: warn
  block_switch_if_hidden_dirty: true
  commit_append_trailer: true
  sparse_checkout_mode: no-cone
EOF

cat > .orbit/orbits/docs.yaml <<'EOF'
id: docs
description: Docs orbit
include:
  - docs/**
EOF

cat > docs/guide.md <<'EOF'
Source Orbit guide
EOF

git add .orbit docs
git commit -m "seed source authoring repo"

"$ORBIT_BIN" template init-source
"$ORBIT_BIN" branch inspect HEAD --json
```

这里的 `HEAD` 应被识别为 `kind=source`。

## 4.4 创建运行态仓库

现在创建消费这个模板的 runtime repo：

```bash
"$HARNESS_BIN" create "$RUNTIME_REPO"
cd "$RUNTIME_REPO"

git config user.name "Orbit Demo"
git config user.email "orbit-demo@example.com"
```

`harness create` 会初始化新的单控制面 runtime root：

- `.harness/manifest.yaml`
- `.harness/orbits/`

zero-member runtime 在这个阶段是合法状态。

## 4.5 安装模板

把上一步生成的 template source 安装进 runtime repo：

```bash
cat > "$INSTALL_BINDINGS" <<'EOF'
schema_version: 1
variables:
  project_name:
    value: Installed Orbit
EOF

"$HARNESS_BIN" install "$TEMPLATE_REPO" \
  --ref orbit-template/docs \
  --bindings "$INSTALL_BINDINGS"
```

正式安装后会写入：

- `.harness/manifest.yaml`
- `.harness/orbits/docs.yaml`
- `.harness/installs/docs.yaml`
- `.harness/vars.yaml`
- 模板渲染后的 runtime files

兼容文件可能仍存在于 migrated runtime 中，但正式 top-level contract 已经是 `.harness/manifest.yaml` + `.harness/orbits/*.yaml`。

## 4.6 检查运行态并提交首个 runtime commit

先看 runtime 视角：

```bash
"$HARNESS_BIN" inspect
"$HARNESS_BIN" check --json
```

Orbit projection 只对 tracked files 工作，所以安装完成后先把当前 runtime state 落成首个正常 Git commit：

```bash
git add -A
git commit -m "install docs template"
```

## 4.7 进入 orbit 并查看当前 runtime branch

再看 orbit projection 视角：

```bash
"$ORBIT_BIN" enter docs
"$ORBIT_BIN" current
"$ORBIT_BIN" status
"$ORBIT_BIN" diff
"$ORBIT_BIN" leave
```

到这里你已经走通了当前正式主链路：

- `harness create`
- `harness install`
- `harness check`
- `git commit`
- `orbit enter / status / diff`

你也可以顺手验证当前 branch 已经被识别成 runtime：

```bash
"$ORBIT_BIN" branch inspect HEAD --json
```

## 4.8 把 runtime 改动回写到已安装 template

如果当前 runtime member 来自 `harness install`，也就是 `.harness/manifest.yaml` 里该 member 的 `source=install_orbit`，那么 `orbit template save` 可以直接复用 `.harness/installs/<orbit-id>.yaml` 里的 `template.source_ref`：

```bash
cat > docs/guide.md <<'EOF'
Improved Installed Orbit guide
EOF

"$ORBIT_BIN" template save docs
```

如果这是一个 migrated runtime，工作树里还残留兼容 `.orbit/config.yaml`，正式 writeback lane 仍然成立，但这个 legacy config 不会进入 template payload：

```bash
mkdir -p .orbit
cat > .orbit/config.yaml <<'EOF'
version: 1
shared_scope:
  - README.md
behavior:
  outside_changes_mode: warn
  block_switch_if_hidden_dirty: true
  commit_append_trailer: true
  sparse_checkout_mode: no-cone
EOF

cat > docs/guide.md <<'EOF'
Migrated Installed Orbit guide
EOF

"$ORBIT_BIN" template save docs --overwrite
```

## 4.9 导出 harness template

如果当前 runtime members 已经是你想复用的组合，可以直接导出 harness template：

```bash
"$HARNESS_BIN" template save --to harness-template/workspace
```

它会写一个新的 harness template branch，其中包含：

- `.harness/manifest.yaml`
- `.harness/orbits/*.yaml`
- 渲染后的模板文件
- 根目录 `AGENTS.md`（如果 runtime root 存在）
- 当前实现里还可能带 `.harness/template.yaml` 这类 payload metadata bridge，但 branch identity 与 authored truth 已经不依赖它

你也可以顺手把三类 branch 都 inspect 一遍：

```bash
"$ORBIT_BIN" branch inspect orbit-template/docs --json
"$ORBIT_BIN" branch inspect HEAD --json
"$ORBIT_BIN" branch inspect harness-template/workspace --json
```

## 5. 常用补充命令

生成 bindings skeleton：

```bash
(
  cd "$TEMPLATE_REPO"
  "$ORBIT_BIN" bindings init orbit-template/docs
)
```

查看本地 branch taxonomy：

```bash
"$ORBIT_BIN" branch list --json
```

再次安装同一 `orbit-id` 时，默认失败；只有显式加上 `--overwrite-existing` 才允许覆盖：

```bash
"$HARNESS_BIN" install "$TEMPLATE_REPO" \
  --ref orbit-template/docs \
  --bindings "$INSTALL_BINDINGS" \
  --overwrite-existing
```

自动化重放这条文档主路径：

```bash
mise run acceptance:quickstart
```

## 6. 当前推荐心智模型

- `orbit` 负责定义、投影、scope 内操作、单 orbit template
- `harness` 负责 runtime、install、members、check、组合模板导出
- `.harness/manifest.yaml` 表达四类 revision kind 的顶层身份
- `.harness/orbits/*.yaml` 是当前主链路里的 OrbitSpec steady-state host
- source / template / runtime / harness_template 都先看 branch manifest 与 hosted OrbitSpec，再决定各自的 publish、install、inspect 与 export 行为
- `orbit template apply` 仍可作为兼容 wrapper 使用，但不是正式主路径

## 7. Migration Note

当前没有单独发布 `harness migrate-control-plane` 命令。

当前建议是：

- 新仓库直接使用 `harness create` / `harness init`
- 旧仓库若还需要兼容初始化路径，可暂时继续使用 `orbit init`
- 显式迁移工具如果后续确实需要，会作为单独 issue 收口，而不是夹在正式 quickstart 主路径里

## 8. 清理 demo 环境

```bash
rm -rf "$ORBIT_DEMO_ROOT"
```
