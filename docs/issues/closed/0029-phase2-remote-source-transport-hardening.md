# ISSUE-0029 Phase 2 Remote Source Transport Hardening

- Status: closed
- Priority: medium
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Revisit the current remote-source transport so Phase 2 can keep the existing correctness guarantees while reducing unnecessary temp-ref churn and repeated remote fetches.

## Scope

- Review the current remote-source flow:
  - `ls-remote --heads`
  - one shallow temp-ref fetch per candidate for manifest validation
  - another fetch for selected-branch snapshot materialization
- Improve the fast path where it is safe to do so, with priority on:
  - explicit `--ref`
  - avoiding duplicate fetch work for the selected branch
- Keep the storage boundary intact:
  - no `.orbit/` cache
  - any cache or temp artifact must remain repo-local under `.git/orbit/` or temp space
- Preserve fail-closed behavior and temp-ref cleanup guarantees.

## Done When

- The remote apply path keeps the documented selection semantics and cleanup behavior.
- The transport layer avoids at least the obvious duplicate work in the current implementation.
- Regression coverage makes it hard to reintroduce local ref pollution or source-selection drift.

## Notes

- Tracks technical-debt entry `0017`-`0021` fan-out follow-up.
- This is a Phase 2C-2 hardening item, not a blocker for entering the stage.
- 2026-03-21: 已为显式 `--ref` 的 remote apply / bindings init 加上单次 temp-ref fetch 快路径，避免再跑 `ls-remote` 和对同一选中分支重复 fetch；同时保留既有 source selection API 的语义与 temp-ref cleanup 合同。
