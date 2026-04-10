# ISSUE-0106 Harness Single Control Plane Runtime Template Install Cutover

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-06

## Summary

把运行态命令、模板导出与模板安装主链路切到单目录控制面，去掉对 `.orbit/config.yaml`、`shared_scope`、`projection_visible`、`behavior` 的 steady-state 依赖。

## Scope

- 运行态命令只允许在 `kind=runtime` 下运行：
  - `enter`
  - `leave`
  - `current`
  - `files`
  - `status`
  - `diff`
  - `log`
  - `commit`
  - `restore`
- scope 解析改为只依赖 OrbitSpec + role-aware plan。
- `orbit template save` 写 `kind=orbit_template` manifest。
- `harness template save` 写 `kind=harness_template` manifest。
- `harness install` 从新模板合同读取并写回 runtime manifest / vars / installs。

## Done When

- 主链路命令已不再依赖旧 `.orbit/*` 控制文件。
- orbit / harness 两类模板都能稳定写出并被安装链路消费。
- 相关 integration tests 通过。

## Notes

- 对应 `docs/context/harness_single_control_plane_technical_spec.md` 的第 8 节。
- 依赖 ISSUE-0104 与 ISSUE-0105。
- Completed:
  - runtime 命令主链路已能在单控制面 runtime 上工作，不再要求旧 `.orbit/config.yaml` 才能 `enter/current/status/diff/leave`。
  - `orbit template save` 会写 `kind=orbit_template` branch manifest，`harness template save` 会写 `kind=harness_template` branch manifest。
  - `harness install` / template source loading / branch inspect 已切到单控制面合同；相关 integration tests 与 quickstart acceptance 已通过。
