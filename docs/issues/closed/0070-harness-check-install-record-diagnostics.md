# ISSUE-0070 Harness Check Install-Record Diagnostics

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-26
- Updated: 2026-04-06

## Summary

细化 `harness check` 对 install-record 结构故障的诊断种类，让 path mismatch 与 record payload invalid 不再混在同一个 finding kind 里。

## Scope

- 为 install-record structural failures 引入更精确的 finding kind：
  - `install_path_mismatch`
  - `install_record_invalid`
- 调整 `harness check` 的 text / json 输出与测试断言。
- 保持当前 fail-closed 语义；不在这一轮引入 best-effort drift probe。

## Done When

- `harness check` 能稳定区分“路径合同错误”和“record schema / payload 错误”。
- 相关 JSON / text 输出契约已更新并有回归测试。
- 现有成员 / drift 诊断不会因这次细化而回退。

## Notes

- 对应此前 `docs/technical-debt.md` 中关于 coarse install-record finding kind 的残余项；该残余项已随本 issue 完成一起移除。
- 这项优先做 finding kind 拆分，不把范围扩大到新的 drift reconstruction 策略。
- Completed:
  - `harness check` 现在把 malformed install-record payload 归类为 `install_record_invalid`。
  - 文件名 / `orbit_id` 不匹配仍保持 `install_path_mismatch`，原有 path-contract 语义不变。
  - JSON 与 text 输出都已通过回归测试，现有 drift / member 诊断没有回退。
