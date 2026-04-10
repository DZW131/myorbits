# ISSUE-0089 Harness Install Progress Mode And Emitter

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

建立 `harness install` 的进度模式与统一 emitter 原语，为后续 install 阶段接线提供稳定基础。

## Scope

- 新增 `--progress <mode>`：
  - `auto`
  - `plain`
  - `quiet`
- 在命令层或紧邻命令层的位置建立 install progress emitter：
  - 输出目标固定为 `stderr`
  - 输出格式固定为稳定阶段行
  - 默认不污染 `stdout`
- `auto` 的第一版规则：
  - TTY 时输出
  - 非 TTY 时静默
- `plain` 总是输出
- `quiet` 总是静默
- 明确 `--json` 不改变 progress 行的 `stderr` 策略

## Done When

- help / flag surface 可见 `--progress`
- 进度 emitter 有稳定单测或命令级测试覆盖：
  - `plain` 输出阶段行到 `stderr`
  - `quiet` 不输出
  - `stdout` 最终结果不受影响
  - `--json` 的 stdout JSON 不受影响

## Notes

- 依赖 `ISSUE-0088`
- 第一版不做 spinner、动态重绘、heartbeat
- 若需要 TTY 判定，优先采用可测试的轻量封装，而不是把命令层塞满终端细节
