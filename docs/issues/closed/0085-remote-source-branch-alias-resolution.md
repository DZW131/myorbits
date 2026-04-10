# ISSUE-0085 Remote Source Branch Alias Resolution

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

在远端模板选择器里引入 source alias 解析：当 `harness install` 面对 source branch 时，把它映射到对应的 published orbit template branch，而不是直接报 “publish first”。

## Scope

- 增强远端 source 选择原语：
  - 显式 `--ref` 指向 source branch 时，尝试 alias 到 `orbit-template/<publish.orbit_id>`
  - 未显式 `--ref` 且远端默认分支是 source branch 时，尝试同样 alias
- 仅当 `.orbit/source.yaml` 中存在 `publish.orbit_id` 时允许 alias。
- alias 目标必须存在且是合法 orbit template branch。
- ambiguity 场景保持 fail-closed，不做自动 alias 兜底。
- 更新 selector 层错误类型、reason、必要的 machine-readable contract。

## Done When

- `ResolveRemoteTemplateSource` 或等价远端选择原语支持 source alias。
- 以下场景有稳定测试：
  - source repo 默认分支 alias 成功
  - `--ref <source-branch>` alias 成功
  - 缺少 `publish.orbit_id` fail-closed
  - published branch 缺失 fail-closed，并提示 publish first
  - ambiguity 继续要求显式 `--ref`
- 不改变本地 template branch 解析路径。

## Notes

- 依赖 `ISSUE-0084`。
- 这是 selector / resolution 层 issue，不负责最终 `harness install` 输出面。
- 第一版只支持远程 URL，不扩展本地 source branch alias。
