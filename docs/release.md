# Orbit / Harness Release

这份文档定义当前仓库的最小发布链路。

当前发布对象是双二进制 CLI：

- `orbit`
- `harness`

GoReleaser 会在一次 release 中同时构建、打包并上传这两个二进制。

## 1. 前提

发布前先保证工作树干净，并完成仓库标准校验：

```bash
mise run fmt
mise run lint
mise run test:ci
```

GoReleaser 使用默认产物目录 `dist/`，仓库通过 `.gitignore` 忽略该目录。

## 2. 本地检查配置

安装 GoReleaser 后，先做配置校验：

```bash
goreleaser check
```

再跑一次本地 snapshot 构建，确认双二进制都能被正确打包：

```bash
goreleaser release --snapshot --clean
```

## 3. 正式发布

正式发布使用语义化标签：

```bash
git tag -a v0.4.0 -m "v0.4.0"
git push origin v0.4.0
```

GitHub Actions 会在 tag push 后执行 GoReleaser，并创建一个 draft release。

## 4. 当前配置边界

当前仓库已包含：

- `.goreleaser.yaml`
- `.github/workflows/release.yml`

当前配置会：

- 构建 `linux` / `darwin` / `windows`
- 构建 `amd64` / `arm64`
- 将 `orbit` 与 `harness` 打进同一个平台归档
- 生成 `checksums.txt`
- 先创建 draft release，避免直接发布不可回滚产物

当前配置不包含：

- Homebrew tap 发布
- 签名 / notarization
- SBOM / provenance

这些能力后续可以在 release 面稳定后再加。
