# Local Issue System

This repository keeps a local, file-based issue tracker under `docs/issues/`.

## Layout

```text
docs/issues/
  README.md
  open/
    NNNN-slug.md
  closed/
    NNNN-slug.md
```

## Rules

- One Markdown file equals one issue.
- Open issues live in `docs/issues/open/`.
- Closed issues move to `docs/issues/closed/`.
- Prefer zero-padded numeric filenames plus a short slug, for example `0001-template-save-schema.md`.
- Keep the full history for an issue in the same file instead of splitting notes across multiple docs.

## Suggested Issue Template

```md
# ISSUE-0001 Title

- Status: open
- Priority: medium
- Owner:
- Created: 2026-03-21
- Updated: 2026-03-21

## Summary

One paragraph describing the problem or desired outcome.

## Scope

- In scope item
- In scope item

## Done When

- Clear acceptance condition
- Clear acceptance condition

## Notes

- Optional working notes
```

## Workflow

1. Create a new file in `docs/issues/open/`.
2. Update the same file as the issue evolves.
3. When the issue is complete or intentionally stopped, update its status and move it to `docs/issues/closed/`.
