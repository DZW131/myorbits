# ISSUE-0088 Harness Install Progress

- Status: closed
- Priority: high
- Owner:
- Created: 2026-04-01
- Updated: 2026-04-01

## Summary

为 `harness install` 增加稳定的阶段性进度输出，解决远端模板安装、source alias 解析、bindings 处理和写盘阶段长期静默的问题，同时保持现有 `stdout` / `--json` 契约不被污染。

## Scope

- 协调本批次的执行子项：
  - progress mode 与 emitter 原语
  - `harness install` 阶段进度接线
  - text/json / dry-run / source alias / harness template 路径的回归测试
- 冻结以下产品边界：
  - 进度默认输出到 `stderr`
  - `stdout` 继续只输出最终结果
  - `--json` 继续只输出最终 JSON
  - 第一版只做稳定阶段行，不做 spinner / 原生 Git progress 透传

## Done When

- `harness install` 支持 `--progress auto|plain|quiet`
- 交互式用户在长耗时 install 中能看到稳定阶段反馈
- `--json` 输出不被过程进度污染
- 相关测试、help、输出契约已冻结

## Notes

- 对应 `docs/harness_install_progress_prd.md`
- 不包含 `orbit template apply` 兼容 wrapper 的进度设计
- 不包含 raw Git progress 透传
