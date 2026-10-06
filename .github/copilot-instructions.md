# Project Guidelines

Read `PROJECT.md` before project work. It is the source of truth for scope, architecture, demo scenarios, and the implementation sequence. Keep this file concise and update it when approved project decisions change.

## Collaboration

- Keep the user informed before consequential steps and at meaningful milestones.
- Before implementation, summarize the intended scope and wait for explicit approval. Planning and instruction-file edits are not implementation approval.
- Raise unresolved questions, GitHub constraints, and credential dependencies early; do not silently broaden scope.

## Project Rules

- Use Go. Optimize for a reliable five-minute jury demo.
- The final demo must use live Claude tool calling. Test doubles may support automated tests but must never be presented as live model behavior.
- Model VI workflow states with mutually exclusive GitHub Issue labels: `vi:in-progress`, `vi:ready-for-docs`, `vi:needs-clarification`, `vi:docs-review`, and `vi:done`. Run the agent only after the linked implementation PR is merged and `vi:ready-for-docs` is added. Never auto-merge docs.
- Checked-in product code/configuration is authoritative for current behavior, settings, and defaults. The VI is evidence of intent and business context. A contradiction blocks resolution pending human clarification.
- Maintain customer-facing and internal developer docs in this repo. Follow their templates in `PROJECT.md`, keep audiences separate, cite evidence, and do not invent unsupported claims.
- Preserve the inspectable Doc Contract, evidence/tool trace, missing-information and conflict reporting, and deterministic validation.
- Keep Claude credentials in GitHub Actions secrets. Use bounded tools and least-privilege GitHub permissions.
- Prefer standard-library Go and the smallest implementation that demonstrates the end-to-end workflow. Avoid extra frameworks, databases, vector search, multi-agent systems, and custom UI unless approved.

## Validation

Run focused Go tests for changed behavior; use `go test ./...` when the Go project is present. Do not claim live GitHub/Claude validation until the workflow has been exercised with the required secret.
