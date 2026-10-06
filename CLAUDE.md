# Project Guidance

Read `PROJECT.md` before making project decisions. It is the source of truth for the demo scope, architecture, scenarios, and implementation sequence; update it when an approved decision changes.

## Working With the User

- Keep the user informed before consequential steps and at meaningful milestones.
- Before implementation, summarize the intended scope and wait for explicit approval. Planning or instruction-file edits do not authorize implementation. Do not start or silently expand implementation work.
- Surface ambiguity, GitHub limitations, credentials, and other demo risks early. Make small, reversible decisions where the brief leaves details open.
- Use `gh` for routine repository inspection and for issue/PR setup or updates that advance requested work; do not ask for separate permission to use the CLI.
- Once implementation is authorized and focused checks pass, you may commit and push scoped changes to a suitable remote feature/demo branch without another confirmation. Do not force-push, rewrite shared history, push directly to `main`, or merge PRs unless explicitly asked. Never auto-merge documentation PRs.

## Project Constraints

- Implement the prototype in Go and optimize for a reliable five-minute jury demo.
- Demonstrate one live Claude tool-using ResolutionAgent in the final demo. Test doubles are for automated tests only; never present simulated model behavior as live Claude behavior.
- Use GitHub Issues and mutually exclusive `vi:*` labels for the VI workflow. `vi:ready-for-docs` may trigger work only after the linked implementation PR is merged. Move blocked VIs to `vi:needs-clarification`; use `vi:docs-review` while a docs PR awaits human review; never auto-merge docs.
- Treat checked-in product code/configuration as authoritative for current product behavior, settings, and defaults. The VI supplies intent and business context. If the VI contradicts product behavior, surface evidence and block resolution pending human clarification.
- Maintain both customer-facing and internal developer documentation in this repository. Follow the audience-specific templates in `PROJECT.md`; do not leak internal-only details into customer docs or invent unsupported claims.
- Keep the Doc Contract, evidence references, tool trace, missing information, conflicts, and deterministic validation visible in the workflow output.
- Keep Claude credentials in GitHub Actions secrets. Never commit secrets. Use bounded tools and least-privilege GitHub permissions.
- Prefer standard-library Go and focused code over databases, vector search, elaborate multi-agent frameworks, or a custom UI unless the user approves expanding scope.

## Validation

Run focused Go tests for changed behavior; use `go test ./...` when the Go project is present. Do not claim the live demo path is verified until it has been exercised with the configured Claude secret and GitHub workflow.
