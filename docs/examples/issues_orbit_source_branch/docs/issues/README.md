# Issues Orbit Rules

This orbit owns the local issue-management loop.

## Goal

Turn issue work into a small observable loop with explicit traces and simple outcomes.

## State Model

Each issue must be in exactly one of these states:

- `open`
- `triaged`
- `in_progress`
- `blocked`
- `closed`

Each issue should also expose one result signal:

- `pending`
- `success`
- `failed`

## Required Trace

Every issue file should contain:

- a `Status` line
- an `Outcome` line
- an `Evidence` section

Before moving an issue to `closed`, the issue file must also contain:

- a `Closure Note` section
- an `Evidence` section with at least one concrete artifact or observation

## Lightweight Probe

Use `tools/check-issues.sh` only as a cheap ready / not_ready probe for the
basic trace structure of issue files.

It should stay fast, local, and structural. Deeper QA belongs to higher-level
harness flows, not to the orbit's basic contract.
