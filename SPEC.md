# CodeCrew Protocol Specification

Version 0.1 (draft) — 2026-08-20

CodeCrew is a lightweight framework for agent-driven software delivery. It keeps
the auditability and reproducible discipline of heavyweight frameworks like GSD
while discarding their ceremony, and it uses the tools teams already have —
GitHub issues, PRs, and CI — as its state store and message bus instead of
maintaining a parallel documentation corpus.

The design rationale is recorded in [docs/founding-decisions.md](docs/founding-decisions.md)
and the analysis that motivated the project in
[docs/gsd-vs-frontier-orchestration.md](docs/gsd-vs-frontier-orchestration.md).

## 1. What CodeCrew is

Three parts, in order of importance:

1. **A protocol** — conventions for how project state is represented in GitHub:
   what a milestone, task, plan, decision, deviation, gate, and summary look
   like, and how agents and humans transact over them. This document is the
   protocol. Any agent that can read and write GitHub issues can participate.
2. **Role contracts** — short, harness-neutral prompt files defining what each
   role (implementer, reviewer, QA, doc-synthesizer) reads, writes, and is
   accountable for. Any frontier-model harness (Claude Code, Codex, Gemini CLI,
   …) can load them; orchestrators such as Paperclip assign them to agents.
3. **A thin CLI** — `codecrew`, a single static Go binary wrapping `gh`. It
   implements the workflow verbs so agents don't each reimplement the
   conventions, and it runs the deterministic gates as code, not judgment. It
   is the only real software in the framework, and it is deliberately small.

CodeCrew does **not** dispatch or schedule agents. A human operator, a cron
job, or an orchestrator does that; CodeCrew defines what a dispatched agent
reads and writes.

## 2. Principles

The kernel of discipline CodeCrew enforces — everything else is left to the
model's judgment:

- **Externalized state.** All coordination state lives in GitHub, sharded by
  task. There is no global state file and nothing that two agents can merge-
  conflict over; GitHub serializes concurrent writes server-side.
- **Plan before nontrivial work.** A task's plan — intended changes,
  requirement IDs covered, pre-identified ask-the-human points — is written
  into the task issue before the first commit.
- **Atomic commits with greppable task references.** History is bisectable and
  progress is inspectable from reality (`git log --grep`), not from reports.
- **Gates run code, not judgment.** Build and full test suite (CI required
  checks) must be green before anything is called done. CodeCrew reads check
  status; it never defines CI.
- **The verifier is never the doer.** Every task is reviewed by a role other
  than the one that implemented it — ideally a different model on a different
  harness — because self-evaluation shares the blind spots of the work itself.
- **Human gates at designed points.** Ask-the-human points are pre-marked in
  plans, and any participant can raise one mid-task. A raised gate blocks
  progression until a human resolves it.
- **Decisions are captured at decision time.** Trade-offs and rejected
  alternatives are recorded in a structured comment the moment they are made,
  not reconstructed later.
- **Requirements have IDs**, so "done" is checkable against something.
- **Documentation is synthesis, not maintenance.** The only documents the
  framework produces are per-milestone "why" documents, compiled at milestone
  close from the decisions and deviations recorded during the work.
- **Intentional platform footprint.** The protocol knows which GitHub
  features it depends on and what plan tier they need (§5, Platform
  requirements); a dependency that would exclude free-plan adopters is taken
  knowingly or not at all.

## 3. Topology: hub and spokes

A CodeCrew project is a **hub** repository and one or more **spoke**
repositories.

**There is no single-repo mode.** A single-repo project is the degenerate case
where N=1 and the hub *is* the spoke. Representations are identical in both
cases, so growing from one repo to many requires no migration.

**The hub** holds: the roadmap, milestone tracking issues, synthesized
milestone documents (`docs/milestones/`), role contracts (`roles/`), and the
CodeCrew configuration.

**Spokes** hold: task issues, the PRs that implement them, and their own CI
gates. Task issues live in the repo whose code they change, so GitHub's native
traceability (closing keywords, PR linkbacks, CODEOWNERS, repo-scoped
permissions) works without convention.

**Every repo carries a pointer file, `.codecrew.yml`**, so an agent dropped
into any repo can find the coordination point. In the hub it declares
`hub: self`; in a spoke it names the hub (`hub: owner/repo`).

### Choosing a hub

The hub is wherever the pointer files say it is — any repo the project owns
can serve. The decision ladder:

1. **`hub: self`** for a single-repo project. The degenerate case above;
   nothing to decide.
2. **The org's `.github` repo** for a multi-repo org whose delivery record
   can be public. The hub's cargo — roadmap, milestone issues, milestone
   documents, role contracts — is org-level coordination state, which is what
   `.github` exists to hold. The caveat is visibility: `.github` repos are
   public by convention, and milestone issues carry goals, decisions, and
   gate discussions, so this choice publishes the delivery narrative even
   when spoke repos are private.
3. **A dedicated hub repo** when the tracking must stay private, or when one
   org runs several projects that should not share a delivery stream.

Three constraints make every rung workable:

- **One hub per spoke.** The pointer file has a single `hub:` field; a repo
  belongs to one delivery stream at a time.
- **One milestone number-line per hub.** Milestone numbering scans the hub's
  milestone issues, so a hub is one serialized stream of work. An org running
  several concurrent projects wants a hub per project, not interleaved
  numbering in a shared one.
- **Multiple hubs per org coexist.** All protocol state — milestone issues,
  numbering, labels, the roadmap — is hub-scoped. The only org-global pieces
  are the role App identities, and those are meant to be shared: the same
  implementer identity can serve every hub in the org.

**The hub is not the framework's repo.** `radiusred/gh-codecrew` is both the
distribution point for the CLI (`gh extension install radiusred/gh-codecrew`)
and this project's own hub — a dogfooding coincidence, not a pattern.
Adopters install the extension from it and point their `.codecrew.yml` at a
repo they own; the protocol requires write access to the hub (labels,
sub-issue attachment, milestone close), so a hub you don't control is not a
hub.

### Growth and restructuring

- **Adding a spoke:** create the repo, add a `.codecrew.yml` naming the hub.
  Nothing moves, nothing converts.
- **Cross-repo references** use the qualified form `owner/repo#123`. Short
  `#123` references remain valid forever within their own repo because issues
  never leave it.
- **Hub changes happen at milestone boundaries.** The protocol's only live
  state is the open milestone and its open sub-issues; everything else is
  inert record that nothing re-reads. So a hub move (splitting a project onto
  its own hub, extracting the hub from a repo that outgrew hub-is-spoke) is:
  close the current milestone in the old hub as normal, repoint each spoke's
  `.codecrew.yml` in a one-line PR, and start the next milestone in the new
  hub — a fresh number-line. History stays where it happened: closed
  milestone issues, milestone documents, and decision trails are
  point-in-time snapshots, and moving them rewrites the audit trail. Leave a
  tombstone row in the old hub's ROADMAP naming where the stream continued.
  Worst case this closes a milestone slightly early with requirements carried
  forward — cheap, and it keeps every milestone's record gathered from one
  hub and synthesized into one document.
- **Mid-flight transfer is the discouraged exception.** GitHub issue transfer
  preserves comments and leaves redirects, but drops label associations and
  is unverified for sub-issue links — moving an *open* milestone risks the
  record the boundary rule keeps intact. Reach for it only when a boundary
  close is genuinely impossible.
- **Repo splits:** open task issues are transferred with GitHub's issue
  transfer. Transfer drops label associations, so conventions must be cheap to
  re-apply (re-label, re-attach the sub-issue link to the milestone).

## 4. State model

### Roadmap

`ROADMAP.md`, committed in the hub. A short ordered list of milestones with a
one-line goal each and links to their tracking issues. It is the only
forward-looking document the framework maintains.

### Milestone

A **tracking issue in the hub**, labeled `cc:milestone`. This is the canonical
milestone object — GitHub's repo-scoped Milestone feature is never canonical,
even in single-repo projects. Body structure:

```markdown
## Goal
One paragraph.

## Requirements
- **M1-R1** — <requirement>
- **M1-R2** — <requirement>

## Gates
What "done" means beyond CI: e2e suites, manual UAT, sign-offs.
```

The milestone's tasks are attached as **GitHub sub-issues**, not listed in
the body: the tracking issue shows live per-task states and progress
natively, and nothing is hand-maintained — a checkbox list rots the moment
someone forgets to tick it (state is inferred, not bookkept).

Requirement IDs are `M<milestone>-R<n>` and are referenced by task plans and
PR descriptions. A requirement may span multiple tasks; it is only marked
complete when every task covering it is closed.

### Task

An **issue in the spoke whose code it changes**, labeled `cc:task`, attached
to its milestone as a sub-issue. Body structure:

```markdown
## Goal
What this task delivers.

## Requirements
M1-R2, M1-R4

## Plan
Intended changes, in enough detail that a deviation is detectable.

## Ask-the-human points
Pre-identified judgment calls that must stop for a human. "None" is valid.
```

The plan is written or updated by the implementer **before the first commit**.
Trivial tasks may have trivial plans; they may not have absent ones.

### Task lifecycle

State is inferred, not bookkept:

| State       | Signal                                        |
|-------------|-----------------------------------------------|
| Ready       | Issue open, unassigned                        |
| In progress | Assignee set                                  |
| Gated       | `cc:needs-decision` label present             |
| In review   | Linked PR open                                |
| Done        | PR merged, issue closed via closing keyword   |

The label set is deliberately tiny: `cc:milestone`, `cc:task`,
`cc:needs-decision`. Anything inferable from GitHub state is inferred.

### Decisions and deviations

Recorded as issue or PR comments **at the moment they occur**, using a
structured prefix so synthesis can find them:

```markdown
**Decision:** <what was decided>
**Trade-off:** <what was weighed>
**Rejected:** <the alternative and why>
```

```markdown
**Deviation:** <what differs from the plan>
**Why:** <rationale>
```

A decision made at a human gate uses the resolution form and is gathered as
a Decision:

```markdown
**Gate resolved:** <what was decided>
**Trade-off:** <what was weighed, if anything>
```

A deviation that changes what a requirement means, or an ask-the-human point
being reached, raises a human gate (§8) instead of being silently recorded.

### Summary

The PR description, finalized at merge time: what was done, which requirements
it satisfies, and any deviations (linking the deviation comments). The merged
PR *is* the task's summary artifact; no separate file is written.

### Commits

Atomic, one logical change each, every commit referencing the task issue
(`(#123)` suffix — the local number suffices because commits live in the same
repo as their task issue). Linear history (rebase merging) is recommended.

### Milestone document

`docs/milestones/<id>-<slug>.md`, committed to the hub when the milestone
closes. It is the "why" document: the architectural, pattern, and technology
choices made during the milestone, their trade-offs and rejected alternatives —
synthesized from the Decision and Deviation comments recorded during the work,
never reconstructed from raw history. It lands via a normal PR and passes the
same review gate as code.

## 5. Configuration

`.codecrew.yml` in every repo. Spokes need only the pointer; the hub carries
the full configuration:

```yaml
codecrew: "0.1"
hub: self                # spokes: owner/repo

# Advisory role routing, read by whoever dispatches agents.
roles:
  implementer: { harness: claude-code, model: claude-fable-5, identity: my-org-coder }
  reviewer:    { harness: codex, identity: my-org-reviewer }
  qa:          { harness: codex, identity: my-org-qa }
  doc-synthesizer: { harness: claude-code, identity: my-org-docs }
```

Role routing is **advisory**: CodeCrew does not dispatch agents, so the config
is a contract for the orchestrator (or human) that does.

**Every role is always staffed — solo is a routing configuration, not a
reduced protocol.** The routing table says *who* holds each role: a GitHub
App (an agent acting as itself), a specific human (their GitHub username),
or — with no identity (`~`) — the human operator. An orchestrator dispatches
sub-agents or delegates to other harnesses per this table; a solo operator
embodies whatever routes to `~`. The hub's config should declare all four
roles at project onboarding, each routed explicitly; an orchestrator finding
no routing table should prompt for one rather than assume. `codecrew init`
(§6) scaffolds exactly this — the table with every role routed `~` — so
onboarding starts explicit. The CLI tolerates an absent table (every role is
then operator-held) but says so in its output.

`identity` names the **GitHub App** — or the GitHub username of a specific
human — the role acts as. A value containing a slash is reserved: it will
name a GitHub team (`org/team-slug`), routing the role to any member
([#44](https://github.com/radiusred/gh-codecrew/issues/44)). Each agent-staffed role gets its own app identity
so agent-authored work is attributable and distinct from the operator —
without this, every PR is authored by the operator's account and GitHub's
prohibition on self-approval makes the non-doer review gate unsatisfiable.
App private keys live outside the repo (by convention
`~/.config/codecrew/<slug>.*.private-key.pem`); short-lived installation
tokens are minted per invocation.

Where a verb gates on a role — `milestone close` counts only the qa role
holder's verdicts — "holder" means the routed identity, or, for an unrouted
role, any human identity not routed to a different role. Crew identities can
never stand in for a role they are not routed to.

### Identity tiers

App identities are optional infrastructure, required only where a
GitHub-native approval must come from a party with no human account:

1. **Solo** — one human, agents act under the operator's own auth. The whole
   protocol works on `gh auth login` alone; the review gate degrades to
   explicit operator confirmation (§6) because GitHub forbids self-approval.
   The confirmation must come from a human identity — a `[bot]` or routed
   role identity is refused — and when the confirmer is also the PR author,
   the recorded comment states that no independent principal exists. See
   [docs/identities.md](docs/identities.md) for the operational path.
2. **Multi-human team** — a colleague satisfies the review gate natively; app
   identities add only attribution (distinguishing an agent's work from its
   dispatcher's).
3. **Enforced review with one human, or agent-reviews-agent** — app
   identities are required: the author and approver must be distinct
   principals, and each automated role needs its own.

Credential resolution is uniform across tiers: orchestrator-injected env vars
(`GITHUB_CLIENT_ID` / `GITHUB_PRIVATE_KEY` / `GITHUB_INSTALLATION_ID`), then a
locally-held private key, then the operator's `gh` auth.

### Platform requirements

Everything the protocol *requires* is available on every GitHub plan,
public or private: issues and sub-issues, labels, comments, pull requests,
rebase merging, GitHub Apps, and CI check reading. A free-plan solo operator
loses nothing.

Features CodeCrew benefits from but deliberately does **not** require,
because they are plan- or visibility-gated on private repos:

- **Branch rulesets / required status checks and reviews** — free on public
  repos; private repos need a paid plan. Without them, `task finish` *is*
  the enforcement: it refuses on red checks and missing approval even where
  GitHub wouldn't block the merge button.
- **Auto-merge** — a convenience some hubs enable; `task finish` merges
  through the API and never depends on it.
- **Actions minutes** — private repos have quotas; the protocol reads check
  results but never defines CI, so a spoke's CI budget is its own affair.

The standing obligation: any future protocol dependency on a GitHub feature
states its plan availability — public/private differences included — at the
moment it is introduced.

## 6. Workflow verbs

The CLI's surface, and therefore the backend interface. Every verb is safe to
run by any role from any repo in the project (the pointer file resolves the
hub).

| Verb | What it does |
|------|--------------|
| `codecrew status` | Where the project is: open milestones, task states, raised gates. |
| `codecrew init [--hub owner/repo]` | Scaffolds a new repo: hub mode writes `.codecrew.yml` with the full `~`-routed roles table, the ROADMAP.md seed, the role contracts (embedded at the installed release), and an AGENTS.md entry point; spoke mode writes the two-line pointer. Idempotent — existing files are kept and reported. |
| `codecrew milestone new` | Creates a milestone tracking issue in the hub from the template; updates ROADMAP.md. |
| `codecrew task new --milestone <id> --repo <spoke>` | Creates a task issue in the spoke from the template; attaches it to the milestone as a sub-issue. |
| `codecrew task start <ref>` | Assigns the caller's identity, verifies a plan is present (refuses to start a planless nontrivial task), creates the working branch — unless the caller's role routing resolves to a role whose contract forbids commits (`qa`, `reviewer`), which get no branch. |
| `codecrew checkpoint <ref> --question "…"` | Raises a human gate: posts the question as a comment, applies `cc:needs-decision`. |
| `codecrew role <name>` | Prints the identity holding a role — an App slug or username, or `~` for the operator (§5). Script-consumable; resolves from the hub's routing table when run in a spoke. The implementer uses it to request review from the reviewer role's holder at PR creation (CODEOWNERS-driven requests coexist — requested reviewers union). |
| `codecrew task finish <ref>` | The gatekeeper: verifies a PR exists, CI checks are green, an approving review exists from a non-doer, and deviations referenced in the PR body have recorded comments — then merges (rebase) and closes. Refuses otherwise, with the specific unmet condition. In a solo-tier project (§5) where author and operator are the same principal, the non-doer approval degrades to an explicit operator confirmation, recorded as a PR comment; the confirming identity must be human — crew identities (`[bot]` suffix or routed role) are refused with `refused[SELF_CONFIRM]`. |
| `codecrew milestone close <id>` | Verifies all tasks closed and every requirement's latest QA verdict is `satisfied` (`refused[VERDICT_MISSING]` / `refused[VERDICT_UNSATISFIED]` otherwise; only verdicts from the qa role's holder count — its routed identity, or the human operator when the role is unrouted (§5) — and a later verdict supersedes an earlier one); gathers every Decision/Deviation comment across the milestone's tasks into raw material for the doc-synthesizer; refuses to close until the milestone document PR is merged. |

Verbs exit nonzero with a machine-readable reason when a gate blocks them, so
agents can act on the refusal rather than parse prose.

## 7. Roles

Role contracts live in the hub under `roles/`, one short markdown file each,
loadable by any harness (and referenced from `AGENTS.md` for harnesses that
read it natively). Roles are contracts, not accounts: no GitHub App needs to
exist for a role to be staffed — every role can act as the human operator
(§5, and [docs/identities.md](docs/identities.md) for both the solo path and
App creation). v1 roles:

- **implementer** — writes the plan into the task issue, does the work in
  atomic commits, records decisions/deviations as they happen, opens the PR,
  finalizes the summary. Never approves their own PR.
- **reviewer** — reviews the PR against the plan and requirement IDs; the
  conversation is ordinary PR review comments and requested changes. Ideally a
  different model/harness from the implementer, per role routing.
- **qa** — exercises the built thing against the milestone's gates and the
  requirements' intent (not just the tests the implementer wrote); reports
  findings as issue/PR comments.
- **doc-synthesizer** — at milestone close, compiles the recorded decisions
  and deviations into the milestone document and opens its PR.

The inter-agent protocol is **GitHub itself** — issue comments, PR reviews,
labels. There is no other message bus, so any two harnesses interoperate by
construction.

## 8. Verification and gates

Three independent layers, attacking different failure modes:

1. **Deterministic gates** — each spoke's CI required checks. Owned by the
   repo, read by `task finish`. Catch what code can catch.
2. **Independent review** — a non-doer approval required to merge. Catches
   correlated self-evaluation failure: the model grading its own work shares
   the blind spots of the model that did the work.
3. **Human gates** — pre-marked ask-the-human points plus ad-hoc
   `checkpoint`s. Catch the boundary cases where capability doesn't help:
   "should this task be done as specified at all." A `cc:needs-decision` label
   blocks `task finish` until a human removes it, and the resolution must be
   recorded as a `**Gate resolved:**` comment (§4) so it is gathered into the
   milestone record as a Decision — `task finish` refuses
   (`refused[GATE_UNRECORDED]`) while a raised gate has no resolution comment,
   even after the label is removed.

## 9. Environments

CodeCrew is harness-neutral by construction: participation requires only the
ability to run a CLI and read/write GitHub. Supported shapes:

- **Solo operator** — one human, one harness, hub-is-spoke.
- **Mixed-model team** — role routing assigns different models to different
  roles (e.g. Claude implements, Codex reviews); they converse through PRs.
- **Orchestrated company** (e.g. Paperclip) — the orchestrator maps its agents
  to CodeCrew roles via the routing config and dispatches them; CodeCrew
  defines what each one reads and writes.

## 10. The CLI

- **Go**, single static binary (`CGO_ENABLED=0`), cross-compiled for
  linux/mac/windows. Also installable as a `gh` extension (`gh codecrew …`),
  since a gh extension is just a binary named `gh-codecrew`.
- **Wraps `gh`** in v1 — auth, base URLs, and enterprise quirks come for free.
- **Backend interface** is shaped by the workflow verbs (§6), not by GitHub's
  feature set. The GitHub adapter is the only v1 implementation; the interface
  keeps two ports conceptually separate even though GitHub fuses them — the
  **tracker** (issues, labels, comments) and the **review surface** (PRs,
  reviews, checks) — so a future backend pairing a different tracker with a
  git host stays possible.
- Clear errors over mystery: if the hub declares a spoke the caller's token
  cannot see, the CLI says exactly that.

## 11. Auditability and mutability

Live tracker state is mutable, and CodeCrew does not police edits — teams that
rewrite the record either have good reason or pay the price. The durable audit
trail is what lands in git: atomic commits with task references, merged PR
descriptions, and the committed milestone documents, which are point-in-time
snapshots of the decisions that mattered. The immutable record is in git; the
working record is in the tracker.

## 12. Non-goals

CodeCrew deliberately does not have:

- wave orchestration, worktree manifests, dispatch scripts, or heartbeats —
  execution scaffolding ages badly; frontier models isolate and merge
  correctly from intent
- a global state file — state is sharded by task and mediated by GitHub
- agent dispatching or scheduling — that is the operator's or orchestrator's
  job
- a maintained documentation corpus — documents are synthesized at milestone
  boundaries, from records captured at decision time
- a Jira adapter — GitHub only, with the backend interface as the seam if
  that ever changes
- configuration for behavior the model can decide well — the framework
  manages trust (specification, verification, audit), not capability

## 13. v1 scope and open questions

**v1 delivers:** this protocol; the four role contracts; the CLI with the §6
verbs and the GitHub adapter; CodeCrew's own hub bootstrapped with it.

**Open questions, deferred until the bootstrap surfaces evidence:**

- Whether `task finish` should also verify requirement-coverage claims in the
  PR body against the milestone's requirement list, or leave that to review.
- An optional read-only GitHub Projects mirror as a human dashboard, derived
  from issue state.
- Issue/PR templates shipped as `.github/` templates versus created by the
  CLI at `new` time.
