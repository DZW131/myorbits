# ISSUE-0004 Implement Runtime-to-Template Replacement Engine

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

实现运行态到模板态的值替换引擎，把 `.orbit/vars.yaml` 中的 concrete values 替换成 `$var_name`，并在存在歧义时 fail-closed，而不是静默生成错误模板。

## Scope

- 支持基于 bindings 的字面量精确替换。
- 按 literal 长度倒序执行替换，避免短值先替换破坏长值匹配。
- 当同一 literal 映射到多个变量时返回 ambiguity，而不是继续替换。
- 产出替换摘要，供 dry-run 和 review 阶段复用。
- 为文本和二进制边界、歧义和顺序规则增加测试。

## Done When

- 文本内容能稳定得到替换后的模板文本。
- 长 literal 优先于短 literal。
- 同 literal 对应多个变量时会明确失败。
- 二进制文件不会被当作文本替换。
- 单测锁定替换顺序、歧义检测和摘要输出。

## Notes

- 对应 `docs/context/orbit_phase2_development_plan.md` 的 Phase 2A-0 / 任务 6。
- 主要依据：`docs/context/orbit_phase2_technical_spec.md` 的 10.1 以及开发计划中的不变量 7。
- 2026-03-21: 已实现 replacement engine，支持文本文件的精确字面量替换、按 literal 长度倒序执行、同 literal 多变量 ambiguity 检测、二进制文件跳过与替换摘要输出，相关单测已补齐。
