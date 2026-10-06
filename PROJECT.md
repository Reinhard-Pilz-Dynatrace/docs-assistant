# Project: Evidence-Grounded Documentation Automation

## Background

This is a hackathon prototype exploring automated generation and maintenance
of customer-facing and internal developer documentation from source-code
changes.

The real environment uses product source repositories, Jira Value Increments
(VIs) / Epics, documentation repositories, CI pipelines, and LLM-based agents.

For this prototype we intentionally implement the workflow entirely in GitHub:

- GitHub repository = product source repository
- GitHub Issue = mock VI / Epic
- GitHub Pull Request = product change
- GitHub Actions = automation / CI
- Markdown under `/docs/customer` = customer-facing documentation
- Markdown under `/docs/internal` = internal developer documentation

The prototype should nevertheless be architected so that GitHub Issues could
later be replaced by Jira and GitHub Actions by another CI system.

The purpose is NOT simply to demonstrate that an LLM can write Markdown.

The purpose is to demonstrate a system that can determine:

1. What changed in the product?
2. Is the change customer-visible?
3. Which existing documentation is affected?
4. What customer/business context is available for the change?
5. Which claims can be supported by evidence?
6. Which information is missing or contradictory?
7. What documentation change should therefore be proposed?

The key principle is:

> Generate documentation from evidence, not merely from a prompt.


# Desired demo

We want a visually understandable GitHub-native end-to-end demo.

The ideal flow is:

    GitHub Issue (VI) + linked implementation PR
        |
        v
    implementation PR merged
        |
        v
    VI labeled `vi:ready-for-docs`
        |
        v
    GitHub Action invokes live Claude ResolutionAgent
        |
        +----> inspect VI, merged change, code/config/tests
        +----> compare VI claims with product behavior
        +----> select docs and load audience templates
        |
        +---- conflict / blocking gap ----> VI: "Needs Clarification"
        |
        v
    Doc Contract + tool trace
        |
        v
    validate evidence and template requirements
        |
        v
    docs PR for human review
        |
        v
    VI: "Docs Review" -> "Done" after docs PR is merged

A reviewer should be able to navigate:

    VI Issue and workflow labels
       -> implementation PR
       -> live Claude tool-use trace
       -> docs PR or explicit "Needs Clarification" decision
       -> human review -> VI resolved

The demo should be understandable without explaining the implementation first.


# Mock product

Create a small but believable fake product repository resembling a large
enterprise observability agent.

Do NOT attempt to reproduce proprietary Dynatrace code.

Create a fictional "Observability Agent" with a feature such as process
monitoring.

A reasonable repository structure might be:

        cmd/
            docs-assistant/
                main.go
        internal/
            assistant/
                ...
        product/
            process_monitoring/
                config.yaml
                process_monitor.go
                process_monitor_test.go
        docs/
            customer/
                process-monitoring.md
            internal/
                process-monitoring.md

        mappings/
            docs-map.yaml

    tools/
      ...

    .github/
      workflows/
      ISSUE_TEMPLATE/

Implement the prototype in Go. Prefer the standard library and only small,
necessary dependencies; favor simplicity and demo reliability over product
realism.

The mock feature should expose customer-visible configuration that can be
changed through a PR.

Example initial configuration:

    process_monitoring:
      enabled: true
      scan_interval_seconds: 30

A demo feature could add:

    include_container_processes: true

A second demo change could modify:

    scan_interval_seconds: 30 -> 60

Existing documentation should contain enough information for the automation
to discover that documentation is affected.


# Mock Value Increment

Represent a VI using a GitHub Issue.

Provide a GitHub Issue template containing structured sections such as:

- User problem
- Target audience
- User goal
- Business value
- Scope
- Entry point
- Expected customer outcome
- Limitations
- Documentation hints

Example:

    Title:
    VI: Container process monitoring

    User problem:
    Customers need visibility into processes running inside containers.

    Target audience:
    Platform engineers

    User goal:
    Include containerized processes in process monitoring.

    Scope:
    Add `include_container_processes`.

    Limitations:
    Linux only.

    Expected customer outcome:
    Container processes become visible through process monitoring.

Implementation PRs should reference their VI issue.

Use mutually exclusive GitHub Issue labels as the demo's visible VI
workflow states. This keeps the state on the ticket and gives GitHub Actions
a direct `issues.labeled` trigger without a webhook service or extra Project
v2 credentials:

- `vi:in-progress`: implementation is underway.
- `vi:ready-for-docs`: implementation PR is merged; start the agent.
- `vi:needs-clarification`: the agent found a contradiction or blocking gap.
- `vi:docs-review`: a docs PR is open and awaiting human review.
- `vi:done`: docs PR is merged and the VI is resolved.

Trigger the GitHub Action when `vi:ready-for-docs` is added, and use the
GitHub Issue, linked PR, and merge commit as inputs. The agent swaps the
state label to `vi:needs-clarification` or `vi:docs-review` and explains its
decision on the VI. GitHub does not provide a synchronous hook to veto issue
closure; if a VI is closed before its docs gate is satisfied, automation can
reopen it and explain the unresolved condition. Treat this as a visible
workflow guard, not an unbreakable permission lock.

Do not tightly couple the domain model to GitHub Issues.

Conceptually use a provider abstraction:

    WorkItemProvider
        GitHubIssueProvider + workflow labels
        future: JiraProvider


# Core concept: Doc Contract

The central intermediate artifact should be a structured representation of
what the automation knows.

Do NOT directly go from:

    code diff -> LLM prompt -> Markdown

Instead produce a machine-readable Doc Contract.

For example:

    {
      "change": {
        "feature": "process_monitoring",
        "customer_visible": true,
                "documentation_audiences": ["customer", "internal-developer"],
        "settings": [...]
      },

    "work_item_context": {
        "user_goal": "...",
        "audience": "...",
        "expected_outcome": "...",
        "limitations": [...]
      },

      "affected_docs": [
        {
                    "path": "docs/customer/process-monitoring.md",
                    "audience": "customer",
                    "type": "product-guide",
                    "reason": "settings-schema mapping"
                },
                {
                    "path": "docs/internal/process-monitoring.md",
                    "audience": "internal-developer",
                    "type": "developer-guide",
                    "reason": "settings-schema mapping"
        }
      ],

      "evidence": [...],

      "conflicts": [...],

      "missing_information": [...]
    }

The exact schema may be improved during implementation.

Persist this artifact during the GitHub Action so that it can be inspected
during the demo.


# Evidence and provenance

Claims in every proposed document should have provenance. Keep the audience
on each claim and document so internal-only details are not copied into
customer-facing documentation.

Useful evidence types include:

- CODE
- SCHEMA
- TEST
- VI
- EXISTING_DOCS
- INFERRED

The automation must distinguish between:

1. facts directly supported by evidence,
2. reasonable derived information,
3. information that cannot safely be established.

It is preferable to report:

    Expected outcome: UNKNOWN

than to invent an expected outcome.

Generated documentation must not silently turn unsupported assumptions into
facts.


# Missing information

Missing customer context is an important OUTPUT of the system.

For example:

    Documentation context
    ---------------------
    ✓ Customer goal              [VI]
    ✓ Configuration              [CODE]
    ✓ Platform limitation        [VI]
    ? Developer operational note [MISSING]

The system should make these gaps visible in the generated PR or PR comment.


# Conflicting information

The system should detect obvious contradictions between sources.

This is an important demo scenario.

Example:

 Checked-in product code/configuration says:

    scan_interval_seconds = 60

 The VI says:

    default scan interval = 30

The system should NOT silently choose whichever source the LLM prefers.

Instead produce something similar to:

    SOURCE CONFLICT

    Code/schema:
    scan_interval_seconds = 60

    VI:
    Default scan interval = 30 seconds.

    Proposed action:
    Move VI to "Needs Clarification". Do not create a docs PR or mark the VI
    Done until the VI is corrected or the product behavior is changed.

For a conflict with existing documentation only, product code/configuration
is authoritative: propose the correction and keep the conflict visible in
the docs PR. For a conflict between the VI and product behavior, do not
silently choose one; block the VI pending human resolution.


# Documentation impact detection

Do not give the entire repository to the generation stage.

Implement a simple mechanism for mapping code concepts to documentation.

For the prototype this can use a checked-in mapping file, e.g.:

    process_monitoring:
      schema:
        - process_monitoring
        - include_container_processes
        - scan_interval_seconds

            docs:
                - path: docs/customer/process-monitoring.md
                    audience: customer
                    type: product-guide
                    template: docs/templates/customer-product-guide.md
                - path: docs/internal/process-monitoring.md
                    audience: internal-developer
                    type: developer-guide
                    template: docs/templates/internal-developer-guide.md

The important behavior is:

        changed code
                -> changed concepts / identifiers
                -> affected documentation
                -> minimal relevant context

The automation should report why a documentation page was selected.


# Documentation templates

Check in one concise template per audience. Templates define required
sections and recommendations, not facts the agent may invent:

- `docs/templates/customer-product-guide.md`: Overview, Configuration,
  Defaults, Limitations, and Troubleshooting.
- `docs/templates/internal-developer-guide.md`: Behavior and ownership,
  Code/configuration locations, Tests and verification, and Operational
  notes.

Every proposed page must retain the required headings. Each section must be
grounded in evidence or explicitly say that the information is not
established and appear in the missing-information report. A missing fact
that makes a VI claim contradictory or prevents an accurate required section
from being written is a blocking gap; move the VI to `Needs Clarification`.


# Context efficiency

Measure how much context is selected for documentation generation.

Where reasonably possible report:

- number of repository files
- number of selected files
- approximate characters/tokens in full candidate context
- approximate characters/tokens in selected context
- percentage reduction

Exact token accounting is not required if unavailable.
A clearly labelled estimate is acceptable.

Example output:

    Context selection

    Repository candidate context: ~74,000 tokens
    Selected context:             ~3,200 tokens
    Reduction:                    ~95.7%

Do NOT hard-code impressive numbers.
Measure them from the repository.


# LLM integration

The final demo must use a live Claude model; do not use a mocked LLM in the
demo. Keep a single ResolutionAgent boundary:

    ResolutionAgent
        ClaudeResolutionAgent (demo)
        TestResolutionAgent (unit tests only)

The agent should use bounded tools to retrieve the linked VI, inspect the
product change, select and read mapped documentation, and validate its
proposal. It should produce a Doc Contract with evidence references, missing
context, conflicts, and audience-specific documentation proposals. Do not
pass the entire repository to the agent. Generate or update each selected
document for its declared audience; do not expose internal-only details in a
customer proposal.

The live run must visibly show Claude choosing and invoking bounded tools,
then reasoning over their results to produce the Doc Contract and
audience-specific proposals. The tool trace and Action summary must identify
the actual provider/model. Test doubles may be used for deterministic unit
tests, but must not be presented as live agent behavior. Store the Claude
credential as a GitHub Actions secret; never commit it. A live end-to-end
rehearsal requires that secret to be configured.

Keep evidence collection, docs mapping, context measurement, contract-shape
checks, and validation deterministic. Checked-in product code and
configuration are authoritative for actual behavior, settings, and defaults;
tests corroborate behavior but do not override the implementation. The VI is
authoritative for intended outcome, audience, and business context, not for
what the product currently does. If those disagree, the agent must flag the
conflict and block resolution instead of treating the VI as product truth.
Unsupported claims must remain unknown or unresolved.


# GitHub PR experience

The output should be pleasant to demo.

Prefer a concise GitHub PR comment/report such as:

    ## Documentation Impact Report

    ### Product change

    ✓ New `include_container_processes` setting
    ✓ `scan_interval_seconds`: 30 -> 60

    ### Affected documentation

        ✓ `docs/customer/process-monitoring.md` [customer / product guide]
            Reason: settings-schema mapping
        ✓ `docs/internal/process-monitoring.md` [internal developer / guide]
            Reason: settings-schema mapping

        ### Proposed updates

        ✓ Customer guide: configuration and customer-visible limitation
        ✓ Developer guide: implementation and verification guidance

    ### Customer context

    ✓ User goal               [VI]
    ✓ Platform limitation     [VI]
    ✓ Configuration           [CODE]
    ? Troubleshooting         [MISSING]

    ### Evidence

    ...

    ### Context selection

    4 relevant files selected from 37
    Estimated context reduction: 91%

    ### Validation

    ✓ Markdown valid
    ✓ Configuration values match schema
    ⚠ One customer-information gap remains

Use GitHub-native presentation where practical:
job summaries, PR comments, checks, artifacts, etc.


# Documentation PR

When there are no blocking contradictions or gaps, the agent creates a docs
PR and moves the VI to `Docs Review`.

The generated PR should explain, for each changed document and audience:

- what product change triggered it,
- which VI supplied business/customer context,
- why these docs were selected,
- what evidence supports the change,
- whether anything remains unresolved.

Human review remains mandatory.

The prototype must NOT automatically merge documentation.

After the docs PR is merged, mark the VI `Done`. If required information is
missing or the VI contradicts product behavior, do not create a completion
PR or mark the VI done; move it to `Needs Clarification` and post the exact
evidence and question. If a user closes the VI before the gate is satisfied,
reopen it with an explanatory comment where repository permissions allow.


# Validation

Implement lightweight deterministic validation.

Possible checks:

- Markdown syntax/basic linting
- referenced configuration setting exists
- documented default agrees with schema
- required sections for each document type exist
- generated docs do not claim unsupported configuration
- customer-facing docs do not expose internal-only details
- Doc Contract schema is valid

If useful, demonstrate a self-correction loop:

    generated docs
        -> validation fails
        -> generator receives validation result
        -> corrected proposal
        -> validation succeeds

Do not make this a requirement if it threatens demo reliability.


# Demo scenarios

Prepare at least TWO reproducible scenarios.


## Scenario A: successful customer-visible feature

VI describes container process monitoring.

Implementation PR adds:

    include_container_processes

Expected result:

- change identified as customer-visible
- linked VI retrieved
- customer and internal developer docs identified separately
- Doc Contract produced
- audience-appropriate changes proposed for both docs
- evidence shown
- validation succeeds


## Scenario B: documentation regression / conflict

Implementation changes:

    scan_interval_seconds: 30 -> 60

Both existing documentation pages still say:

    The default scan interval is 30 seconds.

Expected result:

- both affected docs identified with their audiences and types
- contradiction detected
- evidence identifies schema/code as source
- correction proposed for both docs
- conflict/change is visible in report


# Optional Scenario C: no docs impact

Make an internal refactoring change that does not alter customer-facing or
developer-documented behavior.

Expected result:

    No documentation impact detected for either audience.

No docs PR should be created.

This is valuable because the system should demonstrate that it can decide
NOT to generate documentation.


# Architecture principles

Optimize for:

1. Demoability
2. Deterministic behavior
3. Clear intermediate artifacts
4. Evidence over hallucination
5. Minimal context
6. Replaceable integrations
7. Human review
8. Simple code

Avoid:

- unnecessary frameworks
- complex databases
- vector databases unless genuinely necessary
- elaborate multi-agent architectures
- excessive abstractions
- UI development
- production-scale security infrastructure
- trying to simulate an entire enterprise repository


# Integration boundaries

Keep these conceptual boundaries clean:

    SourceControlProvider
        GitHub
        future: Bitbucket

    WorkItemProvider
        GitHub Issues
        future: Jira

    CIProvider
        GitHub Actions
        future: Jenkins

    ResolutionAgent
        ClaudeResolutionAgent
        TestResolutionAgent (tests only)

The prototype itself should use GitHub-native implementations.


# README

Create an excellent README.

It should explain:

- the problem
- why "just ask an LLM to write docs" is insufficient
- the architecture
- the Doc Contract
- evidence/grounding
- how customer-facing and internal developer documentation are handled
- context minimization
- the Go implementation and local test command
- demo scenarios
- how to run locally
- how to configure GitHub Actions
- how to reproduce the demo

Include a simple Mermaid architecture diagram if appropriate.


# Implementation strategy for today's demo

The goal is to show a live Claude agent participating in resolution of a VI,
not merely writing Markdown. Use one bounded ResolutionAgent with tools, not
a multi-agent framework. Unit tests may use test doubles, but the jury demo
and final end-to-end rehearsal must call Claude for real.

1. Create the smallest believable Go product fixture: process-monitoring
   configuration and tests, customer and internal developer guides, both
   required templates, an audience-tagged docs mapping, and a VI issue
   template. Prepare a consistent feature VI and a contradictory VI.
2. Define the mutually exclusive `vi:*` workflow labels and trigger on the
    `issues.labeled` event for `vi:ready-for-docs`. Link the VI to its
    implementation PR and require the PR to be merged before adding that
    label.
3. Implement narrow tools for retrieving the VI/status, inspecting the
   linked PR and merged code/configuration/tests, selecting docs and
    templates, swapping the VI workflow label, and opening a docs PR. Use
    least-privilege issue, contents, and pull-request permissions.
4. Implement Claude tool calling and the Doc Contract. The live trace should
   show: inspect VI -> gather source evidence -> compare VI claims with
   product behavior -> select audience docs/templates -> propose and
   validate updates, or move the VI to `Needs Clarification` with evidence.
5. Test deterministic parsing, mapping, contract, template, and validation
   behavior with Go tests and test doubles. Then, when the Claude secret is
   available, rehearse both scenarios against the live model: the consistent
   VI produces a docs PR and enters `Docs Review`; the contradictory VI is
   blocked with a clear question and no completion PR.
6. Add the README and short demo script. Verify the real issue-label event,
    GitHub Action, live Claude tool trace, docs PR, and premature closure
    recovery using `gh`. Never auto-merge docs.

The Claude credential is a pre-demo dependency for live end-to-end
validation, not a reason to defer the Go implementation or deterministic
tests. Do not substitute a simulated agent in the jury demo.


# Definition of done

The prototype is successful if I can demonstrate:

1. A GitHub Issue representing a VI with mutually exclusive workflow labels.
2. A PR implementing that VI.
3. Automation triggered when `vi:ready-for-docs` is added after its
    implementation PR is merged.
4. Detection of customer-visible and internal developer documentation impact.
5. Identification of relevant docs by audience without loading the whole
    repository.
6. A visible Doc Contract.
7. Evidence/provenance for important information.
8. Detection of missing or contradictory information, with the VI moved to
    `Needs Clarification` when resolution is blocked.
9. A docs PR proposing audience-appropriate customer and internal developer
    documentation changes.
10. Automated validation.
11. A second scenario showing documentation regression/conflict.
12. A visible live Claude tool-use trace showing how the VI is resolved into
    an evidence-backed docs proposal or a justified block.
13. Customer and internal developer docs conforming to their required
    templates.
14. A prematurely closed VI is reopened with the unresolved gate explained,
    where repository permissions allow.
15. Ideally, an internal-only change that correctly produces no docs PR.

The demo should make the following statement credible:

> This is not an AI that writes documentation.
> It is a system that keeps customer-facing product knowledge and internal
> developer guidance synchronized with the product.