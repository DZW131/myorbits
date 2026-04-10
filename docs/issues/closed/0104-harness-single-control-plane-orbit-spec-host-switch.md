# ISSUE-0104 Harness Single Control Plane Orbit Spec Host Switch

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-06

## Summary

把 OrbitSpec 的 steady-state 宿主从 `.orbit/orbits/*.yaml` 切到 `.harness/orbits/*.yaml`，并让 authoring 命令同步切换到新路径。

## Scope

- 切换 OrbitSpec path host。
- 更新 `meta.file` 校验到 `.harness/orbits/<orbit-id>.yaml`。
- control loader 改读 `.harness/orbits/*.yaml`。
- 切换：
  - `orbit add`
  - `orbit show`
  - `orbit list`
  - `orbit validate`
- steady-state 不再依赖 `.orbit/config.yaml`。

## Done When

- authoring 命令只读写 `.harness/orbits/*`。
- OrbitSpec 新路径与 `meta.file` 校验稳定。
- 相关集成测试通过。

## Notes

- 对应 `docs/context/harness_single_control_plane_technical_spec.md` 的第 6、8.2 节。
- 依赖 ISSUE-0103。
- Completed:
  - OrbitSpec steady-state host 已切到 `.harness/orbits/*.yaml`，`meta.file` 校验与 hosted loader 已同步更新。
  - `orbit add` / `show` / `list` / `validate` 已切到 hosted definitions，不再要求先有 `.orbit/config.yaml`。
  - 相关 path / loader / authoring integration tests 已通过。
