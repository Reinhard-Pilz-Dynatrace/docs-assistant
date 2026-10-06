# Project Progress

Last checked: 2026-10-06. For scope, architecture, demo scenarios, and the complete implementation plan, see [PROJECT.md](PROJECT.md).

## Implemented

- Go product fixtures for three independent fictional features (`process_monitoring`, `log_collection`, `network_monitoring`) plus an intentionally unmapped helper package (`product/text_util`).
- Audience-aware docs mapping (`mappings/docs-map.yaml`) with per-feature schema concepts, source files, and customer/internal docs; templates with required headings. The executor derives feature, source evidence, and diff context from the mapping.
- Evidence-backed Doc Contract types and deterministic validation (contract shape, evidence IDs, allowed mapped targets, template headings).
- Live Claude Messages API tool loop (`inspect_vi` → `read_documentation_context` → `submit_resolution`) with a bounded turn count. The run ends when a submission succeeds; GitHub write failures abort instead of being retried; the tool trace is saved even on failure.
- Claude access through either `ANTHROPIC_API_KEY` (direct API) or `ANTHROPIC_AUTH_TOKEN` (Bearer) with an optional `ANTHROPIC_BASE_URL` gateway. The Action summary names the model and gateway host.
- GitHub integration: mutually exclusive `vi:*` labels, issue comments, docs branch/PR creation, premature-closure reopening.
- GitHub Actions: `documentation-agent.yml` (live agent) and `ci.yml` (gofmt, vet, test on PRs; pending merge of PR #17 if not yet on `main`).
- Compact, GitHub-rendered report format for docs PRs, clarification comments, and job summaries (tables, collapsible evidence, short evidence labels, measured context reduction), seen on real live-run output.

## Verified live (real Claude through the gateway, real GitHub Actions)

| Scenario | VI | Result |
| --- | --- | --- |
| A: feature docs | #1 | Docs PR #11 (human-merged), VI `vi:done` |
| B: documentation regression (30 → 60 s) | #2 | Docs PR #9 (human-merged), VI `vi:done` |
| Blocked VI (VI says 50 MB, code ships 200 MB) | #13 | `vi:needs-clarification` with evidence and a question, no docs PR |
| C: no documentation impact | #15 | `vi:done`, no docs PR |
| Premature closure | #13 | Closed early, automatically reopened with an explanation |

The first four rows used the earlier, more cluttered report format; Set A runs use the current one. Docs PRs were reviewed and merged by a human; the agent never merges documentation.

## Demo sets

Two independent sets of four VIs (live, regression, blocked, no impact) exist on separate fictional features, labeled `demo:set-a` (team) and `demo:set-b` (jury). Set A's regression (#26), blocked (#28) and no-impact (#30) tickets were pre-run live; its live ticket (#24) and all of Set B (#33, #35, #37, #39) are unrun. Tickets are single-use. The presenter runbook and the brief for building slides are in [DEMO.md](DEMO.md).

## Setup

- Actions secret `ANTHROPIC_AUTH_TOKEN`; repository variables `ANTHROPIC_BASE_URL` and `ANTHROPIC_MODEL` (currently `claude-sonnet-5-5`).
- Repository setting "Allow GitHub Actions to create and approve pull requests" must be enabled for the agent to open docs PRs.

## Remaining

- Docs PR #32 (Set A regression) awaits human review.
- Take the screenshots listed in `DEMO.md` and build the slides.
- A run that ends without a submission still saves no trace artifact.
- VI #13 is deliberately left in `vi:needs-clarification` as the blocked example.
- Keep `PROJECT.md` as the source of truth if scope or architecture decisions change.
