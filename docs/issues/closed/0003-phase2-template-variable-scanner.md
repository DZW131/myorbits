# ISSUE-0003 Implement Template Variable Scanner

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

实现模板变量扫描器，遍历模板文件树并识别 `$[A-Za-z_][A-Za-z0-9_]*` 形式的变量引用，输出引用集合、未声明变量和未使用变量，为 `template save` 再扫描和 `template apply` 渲染前校验提供统一原语。

## Scope

- 扫描模板候选文件树中的文本内容。
- 汇总唯一变量集合并保持稳定排序。
- 对比 manifest 声明集合，输出 undeclared / unused。
- 为多文件、重复引用和无变量场景增加测试。
- 明确二进制文件或不可解析文本的处理边界。

## Done When

- 能稳定输出 referenced、undeclared、unused 三类结果。
- 同一变量跨文件重复出现时只汇总一次。
- 未声明变量会被明确标记，不被静默忽略。
- 单测覆盖正常引用、重复引用、空集合和声明不一致场景。

## Notes

- 对应 `docs/context/orbit_phase2_development_plan.md` 的 Phase 2A-0 / 任务 4。
- 主要依据：`docs/context/orbit_phase2_technical_spec.md` 的 8.2、10.2。
- 2026-03-21: 已实现模板变量扫描原语，支持跨文件去重、`referenced` / `undeclared` / `unused` 输出，以及跳过二进制或非法 UTF-8 文件的边界处理，相关单测已补齐。
