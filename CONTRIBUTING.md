# Contributing

当前仓库采用一套轻量但明确的协作约束：分支开发、提交前本地校验、PR 合并到 `main`。

## Git Workflow

对于较大功能、架构调整或行为变更，先开 issue 或先完成方案对齐，再开始实现。

1. 从 `main` 拉出分支开发。
2. 分支命名建议：
   - `feature/<topic>`
   - `fix/<topic>`
   - `docs/<topic>`
   - `chore/<topic>`
3. 提交信息保持清晰、描述式、祈使句。
4. 通过 Pull Request 合并回 `main`，不要直接推送功能改动到 `main`。

示例：

```bash
git checkout main
git pull --ff-only
git checkout -b feature/spec-scan-classifier
```

## Before Every Commit

提交前执行：

```bash
mise run fmt
mise run lint
mise run test:ci
```

当前仓库尚未落地 Go 模块时，这些任务会自动跳过；一旦加入 Go 代码，就会转为强校验。

## Before Merge Or RC

在准备合并 MVP 行为改动、收口阶段 PR，或做演示前，再额外执行一次：

```bash
mise run acceptance:mvp
```

这条脚本会在独立 temp repo 中完成一轮端到端回归：

- `init`
- `add`
- `validate`
- `enter`
- `status`
- `diff`
- `log`
- `commit`
- `restore`
- `leave`

它的目标不是替代 `go test`，而是确认 Orbit 当前作为一个完整 CLI 工具时的主路径仍然可用。

## Code Style

当前仓库采用以下默认约束：

- Go 代码必须通过 `gofmt -s`。
- Go 代码必须通过 `golangci-lint`。
- 错误处理必须显式，不允许静默忽略。
- 命名保持可读、具体，避免缩写泛滥。
- 新增逻辑默认需要配套测试。

## Pull Request Expectations

PR 描述至少包含：

- 变更目的
- 主要修改点
- 本地验证方式
- 是否影响现有命令、文档结构或状态文件
