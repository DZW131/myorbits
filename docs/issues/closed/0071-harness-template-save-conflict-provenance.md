# ISSUE-0071 Harness Template Save Conflict Provenance

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-26
- Updated: 2026-04-09

## Summary

补强 `harness template save` 的 conflict / ambiguity diagnostics，把 failing candidate merge 的 member provenance 带出来，减少多 member 导出时的人工排障成本。

## Scope

- 为 path conflict、variable conflict、replacement ambiguity 增加 member provenance：
  - 哪些 member 贡献了冲突文件
  - 哪些 member 贡献了冲突变量
  - 哪些 member 命中了 ambiguity literal
- 更新命令错误输出与必要的 machine-readable contract。
- 补单元测试和至少一条集成回归，确保多 member 失败路径可稳定定位。

## Done When

- `harness template save` fail-closed 时能指出冲突涉及的 member 集合。
- merge / ambiguity 相关测试已覆盖 provenance 输出。
- 不改变现有 conflict policy，只增强诊断信息。

## Notes

- 对应 `docs/technical-debt.md` 的 12。
- 这项只增强 diagnostics，不放宽当前 fail-closed merge 合同。

## Resolution

- `harness template save` 现在已经能在 path conflict、variable conflict 与 replacement ambiguity 三条主失败路径上稳定暴露 provenance。
- diagnostics 不再只停留在纯文本错误：dry-run JSON 与 save failure JSON 都已具备 machine-readable contributors / conflict payload，相关命令层与集成测试也已覆盖。
- 该 issue 的目标“增强 diagnostics 而不放宽现有 fail-closed merge 合同”已经满足；更细的 machine-readable 扩展或 mixed-install policy 不再属于这张 issue 的阻塞项。
