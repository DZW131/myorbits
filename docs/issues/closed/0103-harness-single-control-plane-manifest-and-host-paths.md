# ISSUE-0103 Harness Single Control Plane Manifest And Host Paths

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-05
- Updated: 2026-04-06

## Summary

为单目录控制面方案建立新的 `.harness/manifest.yaml` schema、validator、codec 与 `.harness/orbits/*` host path helpers，作为后续命令切换的共享基础。

## Scope

- 定义 `manifest.yaml` 顶层 schema：
  - `kind=runtime`
  - `kind=orbit_template`
  - `kind=harness_template`
- 新增 per-kind validator 与 stable codec。
- 新增 `.harness/manifest.yaml` 路径 helper。
- 新增 `.harness/orbits/<orbit-id>.yaml` 路径 helper。
- 不切命令主链路。

## Done When

- 代码内能稳定读写三种 manifest。
- 非法 kind、混合字段、缺失必填字段都会 fail-closed。
- path helper / schema 单测通过。

## Notes

- 对应 `docs/context/harness_single_control_plane_technical_spec.md` 的第 4、5 节。
- 是后续 OrbitSpec 宿主切换、branch classification 与 bootstrap 的共同前置项。
- Completed:
  - 新增 `.harness/manifest.yaml` 的 `ManifestFile` schema、raw parse、per-kind validator 与 stable codec。
  - 新增 `.harness/manifest.yaml` 与 `.harness/orbits/<orbit-id>.yaml` path helpers。
  - 覆盖 `runtime`、`orbit_template`、`harness_template` 三种 `kind` 的 round-trip、mixed-field、missing-field 与 invalid-member 测试。
