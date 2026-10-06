# Demo runbook: a documentation agent that shows its evidence

This file is the single source for the four-minute demo and for the slides built from it.

## Brief for Claude Desktop (use this to generate the slides)

> **Purpose.** Create a short slide deck (about 6 slides) from the scenes below for a four-minute live demo. Each scene has `Slide:`, `Show:` and `Say:` lines. Use `Slide:` for the title and the one idea, and `Say:` for speaker notes.
>
> **Goal of the talk.** Show that documentation can be kept in sync with a product by a system that works from evidence, and that stops and asks a human when the sources disagree. The message is "this is not an AI that writes documentation; it is a system that keeps product knowledge and documentation synchronized".
>
> **Audience.** Four people: Anthropic staff, a senior VP of development, and non-technical colleagues. Keep the language plain. Introduce each technical term once, in one sentence. Avoid jargon such as "Doc Contract" without an explanation.
>
> **Tone.** Calm, confident, a little playful. The demo tickets have jokes in their titles; the slides should not.
>
> **Honesty rules.** Mark anything pre-recorded as "recorded earlier" on the slide. Never imply that a screenshot is a live run. Do not invent numbers; the only figures come from the screenshots listed under "Screenshots to take".
>
> **Leave out.** Code, configuration syntax, API details, and anything about internal credentials or the gateway.
>
> **Placeholders.** Where a scene says `[SCREENSHOT n]`, leave a clearly marked empty frame with the caption given below. The presenter will fill it in.

## The idea in one paragraph

A ticket (the "Value Increment", VI) describes what a change is for. A developer merges the change. A person adds the label `vi:ready-for-docs`. A Claude agent then uses a small set of tools to read the ticket, the change and the current docs, and compares them. If everything agrees, it opens a documentation pull request for a human to review. If the ticket contradicts the product, it stops and asks a question. If the change does not affect any documentation, it says so and opens nothing. The agent never merges documentation.

## Facts to rely on (and limits)

- Verified live, with real Claude and real GitHub Actions: feature docs, a documentation regression, a blocked VI, a no-impact change, and reopening of a VI closed too early.
- A live run takes roughly 50 to 70 seconds.
- Claude's wording differs from run to run. Point at the structure (table, warnings, missing items), never read its sentences as a script.
- Known rough edges: the docs PR shows a CI status that is waiting for approval (GitHub does not auto-run workflows for PRs opened by the Actions bot); finished VIs stay open; there is no automatic check that customer pages avoid internal-only details (the agent is instructed to, but it is not enforced in code).
- Do not claim performance or cost savings. None has been measured.

## Demo sets

Each set has the same four tickets, on separate fictional features, so sets never interfere. The label `demo:set-a` or `demo:set-b` marks them.

| Scene | Set A (team walkthrough) | Set B (jury) |
|---|---|---|
| Live: a default flips, docs follow | #24 crash reporting | #33 heartbeat |
| Regression: old docs disagree with new default | #26 metric export | #35 error tracking |
| Blocked: ticket contradicts the code | #28 trace sampling | #37 profile capture |
| No impact: internal refactor | #30 path helper | #39 time helper |

State at the time of writing: Set A regression, blocked and no-impact tickets were already run live and are the screenshot sources. Set A live (#24) and all of Set B are unrun. Re-check with `gh issue list --label demo:set-a` before presenting.

Each ticket is single-use. After a live run its implementation PR is merged and its labels have changed, so it cannot be live-triggered again. Use a fresh set for every live presentation.

## Before you start (checklist)

- [ ] Repository secret `ANTHROPIC_AUTH_TOKEN`, variables `ANTHROPIC_BASE_URL` and `ANTHROPIC_MODEL`, and the setting "Allow GitHub Actions to create and approve pull requests" are in place.
- [ ] `go test ./...` passes on `main` and the latest run of "Resolve VI documentation" was green.
- [ ] The live ticket (#24 or #33) has label `vi:in-progress` and its implementation PR is merged. It must not yet have `vi:ready-for-docs`.
- [ ] Browser tabs open, in this order: the live ticket, the Actions page of the repository, a finished docs PR from an earlier run, a blocked ticket with its comment, a no-impact ticket.
- [ ] Screenshots taken (see below) and the slide deck generated.
- [ ] A fallback decided: if the live run is slow or fails, switch to the finished docs PR and say "this is from an earlier live run".

## Scenes (about four minutes)

### Scene 1: the problem (0:00 to 0:30)

- **Slide:** "Docs drift from the product" and "an AI that writes docs is not enough".
- **Show:** the title slide only.
- **Say:** Products change every week and the documentation quietly goes stale. Asking a model to "write the docs" is risky, because it will happily state things nobody can back up. We built the opposite: a system that only writes what it can point to evidence for, and asks a human when it cannot.

### Scene 2: trigger it live (0:30 to 1:30)

- **Slide:** "A ticket, a merged change, one label".
- **Show:** the live ticket, then add the label `vi:ready-for-docs` in front of the audience. Switch to the Actions tab and open the running job.
- **Say (while it runs, about a minute):**
  - The ticket says what the change is for, in business terms. The developer's change is already merged.
  - Adding this label is the only human step. It wakes up an agent, a Claude model with three small tools: read the ticket and the change, read only the documentation that is relevant, and submit its conclusion.
  - It does not get the whole repository, only the few files that matter for this feature. That keeps it focused and cheap.
  - Right now it is reading the ticket, comparing it with what the code actually does, and checking what the current documentation says.
  - If the ticket and the code disagree, the code wins as the source of truth for what the product does, and the agent has to stop and ask.
- **If it is slow:** keep narrating and show the previous finished run's summary on the second tab. Do not pretend it is the current run.

### Scene 3: the result (1:30 to 2:30)

- **Slide:** "A pull request a human can trust in a minute".
- **Show:** the docs pull request the run created: the table of documents and why they were chosen, the checklist of what is known and what is missing, the folded evidence section, and the diff of the customer page and the internal page.
- **Say:** Two documents changed, one for customers and one for internal developers, each in its own template. Every statement cites where it came from: the ticket, the code, or a test. Things the agent could not establish, like troubleshooting advice, are listed as missing, not invented. A person reviews this and merges it; the agent is not allowed to.
- **Point at:** the document table, the "missing" items, the "X of Y files read" figure, the folded evidence.

### Scene 4: the guardrail (2:30 to 3:30)

- **Slide:** "When sources disagree, it stops".
- **Show:** [SCREENSHOT 1] the blocked ticket's comment with its conflict table and question, and [SCREENSHOT 2] the ticket's label `vi:needs-clarification`.
- **Say:** Here the ticket asked for one default and the merged code shipped another. The agent did not pick a side and did not write documentation. It showed both sources side by side and asked one question. Then a short, honest note: a ticket closed too early is reopened automatically, so this gate cannot be bypassed by accident.
- **Optional, 15 seconds:** [SCREENSHOT 3] a change to an internal helper. The agent concluded there was no documentation impact and opened nothing. It is as important that it can decide not to write.

### Scene 5: close (3:30 to 4:00)

- **Slide:** "Synchronized, not generated".
- **Show:** the merged docs pull request and the ticket at `vi:done`.
- **Say:** A person merges the documentation, and the ticket closes. Evidence over prompts, humans in the loop, and a system that can say "I don't know" or "I won't". Today's version is a prototype on fictional products; the same idea applies to real tickets and real repositories.

## Screenshots to take

Take them from Set A (already run) for the deck, and keep Set B untouched for the live demo.

1. **[SCREENSHOT 1]** Blocked ticket comment: issue #28, the conflict table and the question. Caption: "The agent stops and shows both sides."
2. **[SCREENSHOT 2]** Same ticket's label area showing `vi:needs-clarification`. Caption: "Status is visible on the ticket."
3. **[SCREENSHOT 3]** No-impact ticket: issue #30 with `vi:done` and the comment that no documentation changed. Caption: "Deciding not to write is a feature."
4. **[SCREENSHOT 4]** Docs PR #32 top of page (table, checklist, validation line). Caption: "Recorded earlier: what a reviewer sees."
5. **[SCREENSHOT 5]** Docs PR #32 folded evidence section expanded. Caption: "Every claim points to a source."
6. **[SCREENSHOT 6]** The job summary of one run in the Actions tab. Caption: "Which model, how much context it read."

## If something goes wrong

| Problem | What to do |
|---|---|
| Label added, nothing starts | Check the Actions tab; if there is no run, remove and re-add the label once. Otherwise switch to the finished example. |
| Run fails or takes over two minutes | Say so, open the previous finished run and docs PR, and label them "from an earlier live run". |
| Docs PR shows a CI check waiting for approval | Explain that GitHub holds checks for PRs opened by automation; the documentation itself is reviewed by a human anyway. |
| Agent wording differs from the screenshots | Expected. Point at the structure, not the sentences. |
| Someone asks "is this really live?" | Yes for the run you triggered; the screenshots are labeled as recorded earlier. |
