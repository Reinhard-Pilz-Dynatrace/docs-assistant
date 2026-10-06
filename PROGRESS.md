# Project Progress

Last checked: 2026-10-06. For scope, architecture, demo scenarios, and the complete implementation plan, see [PROJECT.md](PROJECT.md).

## Implemented

- Go product fixture for process monitoring, YAML configuration parsing, and behavior tests.
- Audience-aware docs mapping, customer and internal developer pages, and required-section templates.
- Evidence-backed Doc Contract types and deterministic validation for evidence IDs, allowed mapped targets, and template headings.
- Live Claude Messages API tool loop with a bounded tool-turn count, tool-result trace, and a hard requirement for `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL`.
- Bounded agent tools to inspect a VI and merged implementation PR, read mapped product/docs/template context, and submit a proposal, no-impact decision, or clarification block.
- GitHub integration for mutually exclusive `vi:*` labels, issue comments, docs branch/PR creation, and reopening a prematurely closed VI.
- GitHub Actions workflow, VI issue form, PR template, persistent Claude/Copilot guidance, and demo instructions in [README.md](README.md).

## Current GitHub Demo

Repository: [docs-assistant](https://github.com/Reinhard-Pilz-Dynatrace/docs-assistant), private.

- VI [#1: Container process census](https://github.com/Reinhard-Pilz-Dynatrace/docs-assistant/issues/1) is open with `vi:in-progress`.
- Implementation PR [#3](https://github.com/Reinhard-Pilz-Dynatrace/docs-assistant/pull/3) adds container process monitoring configuration and a test.
- VI [#2: 60-second scans](https://github.com/Reinhard-Pilz-Dynatrace/docs-assistant/issues/2) is open with `vi:in-progress`.
- Implementation PR [#4](https://github.com/Reinhard-Pilz-Dynatrace/docs-assistant/pull/4) changes the scan interval from 30 to 60 seconds and adds a config-backed test.
- Both PRs were last checked as open and mergeable against `main`; neither had CI checks configured. Their test-file hunks overlap, so check/resolve a possible merge conflict after merging either one.
- Neither VI has `vi:ready-for-docs`; no live documentation-resolution workflow has been triggered.

Current local branch: `demo/vi-2-scan-interval` at `7ccbe4e`. Local process-monitoring config is set to 60 seconds. The worktree was clean at the last check.

## Validation And Remaining Gates

- `go test -count=1 ./...` passes.
- `go vet ./...` passes.
- YAML workflow, issue form, and docs mapping are covered by tests.
- A separate `claude -p` CLI smoke command exited successfully, but that does not verify this app's Messages API integration or the GitHub Actions path.
- The app's live path still requires `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL`; the local API-key environment variable was absent at the last check. Configure the Actions secret and model variable before triggering `vi:ready-for-docs`.
- No live Claude tool-call run, generated docs PR, or end-to-end GitHub Action has been verified yet. Do not present tests or the CLI smoke check as that verification.
- GitHub Actions currently reports no checks on implementation PRs. Documentation PRs must remain human-reviewed and must never be auto-merged.

## Resume

1. Review and merge the implementation PRs, resolving any overlap in `process_monitor_test.go`.
2. Confirm `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL` are available to Actions.
3. Add `vi:ready-for-docs` to one VI at a time and inspect the Action summary, Doc Contract, tool trace, and docs PR or clarification result.
4. Keep `PROJECT.md` as the source of truth if scope or architecture decisions change.
