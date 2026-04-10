# ISSUE-0019 Phase 2 Remote Temp Ref Fetch And Read

- Status: closed
- Priority: high
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

Implement temp-ref fetch and remote revision reads so Phase 2C-1 can inspect one chosen remote template branch without polluting normal refs or depending on a checked-out worktree.

## Scope

- Add Git helpers to fetch a remote revision into a temporary local ref.
- Read `.orbit/template.yaml` and template tree contents from the fetched revision.
- Clean up the temporary ref after successful reads and on failure paths.
- Keep the Git layer limited to transport and revision access; do not move apply orchestration into `git`.
- Avoid treating temporary refs as required long-term state.

## Done When

- The code can fetch one remote revision into a temporary ref and read files from it.
- Temporary refs are cleaned up after the read lifecycle completes.
- Remote read failures do not leave normal local refs polluted.
- The implementation does not require switching branches or using Git worktrees.

## Notes

- Corresponds to `docs/context/orbit_phase2_development_plan.md` Phase 2C-1 / task 3.
- Primary spec references: `docs/context/orbit_phase2_technical_spec.md` sections 7.4, 15, and 17.3.
- 2026-03-21: 已实现 temp-ref fetch 与远程 revision 读取的第一轮落地，新增 `WithFetchedRemoteRef` 保证临时 ref 生命周期受控，并在 template 层补齐基于选中 remote candidate 的完整模板源快照读取；测试覆盖成功路径、回调失败清理、远程模板树加载与 forbidden control files 场景。
