# M3: First spoke

Tracking issue: [#25](https://github.com/radiusred/gh-codecrew/issues/25) ·
Closed 2026-08-21 · Synthesized by the doc-synthesizer role
([radiusred-wordy](https://github.com/apps/radiusred-wordy)) from the records
gathered by `codecrew milestone close 3` (two: a Decision on #38, a Deviation
on radiusred/www#41), the task plans and merged PR descriptions, and QA's
verdicts on #25.

## Goal and outcome

Prove the protocol works outside its own repository, and close the discipline
gaps M2 exposed. Delivered in full: gate resolutions are now gathered as
Decision records and enforced at `task finish` (`refused[GATE_UNRECORDED]`),
`task start` skips linked branches for no-commit roles, `milestone close`
refuses unless every requirement carries a satisfied QA verdict — and
[radiusred/www](https://github.com/radiusred/www) became the first spoke,
delivering a real public artifact end-to-end through the installed extension:
[a blog post introducing CodeCrew, published by the protocol it
describes](https://www.radiusred.uk/blog/posts/2026-08-20-this-post-was-delivered-by-the-framework-it-introduces/).
Supporting work with no requirement ID made the hub model and platform
footprint explicit in the SPEC ([#29](https://github.com/radiusred/gh-codecrew/issues/29)
/ PR [#37](https://github.com/radiusred/gh-codecrew/pull/37)) and made
pure-solo operation real in both code and docs
([#38](https://github.com/radiusred/gh-codecrew/issues/38) / PR
[#39](https://github.com/radiusred/gh-codecrew/pull/39), which added
[docs/identities.md](../identities.md)).

## Decisions

- **`--operator-confirm` refuses crew identities; a human operator may
  confirm, including their own PR (pure solo).** The old guard refused
  viewer == author, which contradicted SPEC §5 tier 1 ("works on `gh auth
  login` alone"); the new guard is an identity-class check — a `[bot]`
  suffix or an identity routed to a role can never waive review, in any
  tier, while a human operator may, with the recorded comment stating when
  author and operator are the same principal. Trade-off: a workable tier-1
  solo path that keeps the trail, versus losing the hard viewer≠author
  guarantee — accepted because in pure solo no distinct principal exists,
  and refusal only pushes the operator outside the CLI, losing the record
  entirely. Rejected: documenting "solo requires one App identity"
  (contradicts SPEC §5), and deferring via a protocol gate (the operator was
  present and decided).
  ([Decision on #38](https://github.com/radiusred/gh-codecrew/issues/38#issuecomment-5361983688)
  — the milestone's one gathered Decision; delivered by PR
  [#39](https://github.com/radiusred/gh-codecrew/pull/39).)
- **An unrecorded gate resolution blocks `task finish`.** Belt and braces
  for M3-R1: the gatherer accepts `**Gate resolved:**` as a Decision record,
  `checkpoint` teaches the convention in the gate comment itself, and `task
  finish` refuses with `refused[GATE_UNRECORDED]` when a raised gate lacks a
  later resolution record — the exact failure that lost M2's rename decision.
  Rejected: gather-only (non-blocking), on M2's protocol-hardening lesson.
  (Ask-the-human point in the
  [#26 plan](https://github.com/radiusred/gh-codecrew/issues/26) / PR
  [#30](https://github.com/radiusred/gh-codecrew/pull/30); no Decision
  comment.)
- **The no-commit role set is hardcoded (`qa`, `reviewer`), not
  configurable.** SPEC §7 fixes the v1 role vocabulary and the role
  contracts define commit rights ("Never fix what you find", "Never rewrite
  the implementer's code"); the routing config stays advisory routing only.
  Rejected: a per-role config flag.
  (Ask-the-human point in the
  [#27 plan](https://github.com/radiusred/gh-codecrew/issues/27) / PR
  [#32](https://github.com/radiusred/gh-codecrew/pull/32); no Decision
  comment.)
- **`untestable` blocks close, same as `not satisfied`.** Matches M3-R3's
  "satisfied (or superseded)" wording. The accepted tension: QA's contract
  legitimately allows `untestable`, so a genuinely untestable requirement
  deadlocks its milestone until the operator re-scopes it and QA re-verdicts.
  Rejected: untestable-passes-with-loud-output.
  (Ask-the-human point in the
  [#28 plan](https://github.com/radiusred/gh-codecrew/issues/28) / PR
  [#33](https://github.com/radiusred/gh-codecrew/pull/33); no Decision
  comment.)
- **The blog post's byline is Cody, breaking the all-Wordy precedent.**
  Deliberate: the implementer role delivered the post, and the byline records
  who did the work. Flagged as an ask-the-human point on
  [radiusred/www#41](https://github.com/radiusred/www/issues/41) and
  approved through PR review; no Decision comment. (Recorded without further
  editorial comment by the byline's previous sole holder.)

## QA (second dispatch — first clean pass)

`radiusred-testy[bot]` executed independently against M3's requirements
([verdicts on #25](https://github.com/radiusred/gh-codecrew/issues/25#issuecomment-5363174641)),
and for the first time every requirement was satisfied on the first pass (M2
needed a remedy loop):

- **M3-R1 satisfied** — built from `main` at `975602e`, ran the suite and
  focused record probes, and independently traced the close path: trimmed
  `**Gate resolved:**` comments become Decision records, and gathering
  traverses every milestone task plus its closing PRs.
- **M3-R2 satisfied** — the executed branch-selection path resolves
  `radiusred-testy[bot]` to `qa` and bypasses branch creation with the
  explicit no-linked-branch result; no live `task start` probe, since no
  open `cc:task` allowed a safe one.
- **M3-R3 satisfied** — verified by running the gate itself: before posting
  any verdict, testy ran a freshly built `milestone close 3`; it exited 1,
  made no mutation, and returned `refused[VERDICT_MISSING]` naming all four
  requirements. The gate blocked its own milestone's close until QA had
  spoken — the loop M2 ran on trust, now refused as code.
- **M3-R4 satisfied** — verified radiusred/www#41 is a closed sub-issue of
  this milestone, the pointer file on `radiusred/www@main`, human approval
  and green CI on PR radiusred/www#42, the findings and Deviation records on
  the issue, and the deployed post returning HTTP 200.

QA changed no files, commits, or branches, per its contract.

## The spoke run: five findings

The plan on radiusred/www#41 named the spoke findings as M3-R4's real yield
([findings comment](https://github.com/radiusred/www/issues/41#issuecomment-5360605551);
finding 5 surfaced in the
[Deviation](https://github.com/radiusred/www/issues/41#issuecomment-5360655887)):

1. **Spokes need Issues enabled.** `task new --repo radiusred/www` failed
   with HTTP 410 until the operator enabled them. Candidate: a clearer
   refusal, and a spoke-onboarding note in the docs.
2. **A checkless spoke passes `task finish`'s CI gate vacuously.** www had
   no `pull_request` workflows, so SPEC §8's deterministic-gates layer was
   silently absent. (Remedied on www as a side effect of the deviation
   below; the general hardening — warn or refuse on zero checks — remains
   open.)
3. **Auto-merge is a repo setting the hub habit relied on.** `gh pr merge
   --auto` failed on www. Not protocol friction — `task finish` is the
   protocol's merge verb and works without it — but the hub's
   arm-auto-merge convenience doesn't travel.
4. **What worked cross-repo, first try:** `task new` created the spoke issue
   and attached it as a sub-issue of the hub milestone; `task start` from a
   hub checkout created the linked branch in the spoke; `codecrew status`
   from the spoke checkout resolved the full milestone through the two-line
   pointer file. (Role resolution's hub-config fallback was *not* exercised
   this run — `task start` ran from the hub checkout, which carries the
   roles table locally.)
5. **An org ruleset can require a check a spoke cannot report.** "Lint
   commit messages" is required on every repo in the org, but only the hub
   carried the workflow that reports it, so the required check never ran and
   the PR was blocked indefinitely. This produced the milestone's one
   deviation, below.

## Deviations

One recorded, on the spoke
([Deviation on radiusred/www#41](https://github.com/radiusred/www/issues/41#issuecomment-5360655887)):
the commit-message lint workflow (`.github/workflows/ci.yml` +
`commitlint.config.mjs`) was added to www, unplanned, as a third commit on
PR [radiusred/www#42](https://github.com/radiusred/www/pull/42) — the org
ruleset required a status check nothing in the spoke could report (finding
5), blocking merge indefinitely. The addition unblocked the PR and
incidentally remedied finding 2: www is no longer checkless, so `task
finish`'s CI gate is real there now. Consolidating the workflow at org level
is filed as [#35](https://github.com/radiusred/gh-codecrew/issues/35).

## Requirement outcomes

| Requirement | Delivered by | Outcome |
|-------------|--------------|---------|
| M3-R1 — gate resolutions gathered as decisions | [#26](https://github.com/radiusred/gh-codecrew/issues/26) / PR #30 | Done; QA satisfied |
| M3-R2 — role-aware `task start` branches | [#27](https://github.com/radiusred/gh-codecrew/issues/27) / PR #32 | Done; QA satisfied |
| M3-R3 — verdict-loop close gate | [#28](https://github.com/radiusred/gh-codecrew/issues/28) / PR #33 | Done; QA satisfied by running the gate itself |
| M3-R4 — a spoke proven end-to-end | [radiusred/www#41](https://github.com/radiusred/www/issues/41) / PR radiusred/www#42 | Done; QA satisfied; post live |

## Protocol-discipline gaps observed

- **Judgment calls still live in plans and PR bodies.** `milestone close 3`
  gathered two records, but three decisions of record — blocking
  `GATE_UNRECORDED` (#26), the hardcoded no-commit set (#27),
  untestable-blocks-close (#28) — plus the byline call on radiusred/www#41
  exist only as ask-the-human points in plans and PR descriptions, invisible
  to the gatherer. M3-R1 closed this gap for gates; the plan-time analog
  remains. Candidate: operator approval of a plan's ask-the-human points is
  a decision — record it as one.
- **Nothing forces a spoke to be gate-capable at enrolment.** Findings 1, 2,
  and 5 share a shape: the spoke lacked something (Issues, checks, a
  required-check reporter) and the protocol discovered it mid-flight, by
  failure. Candidate: a spoke-onboarding verb or checklist verifying Issues,
  checks, and org-ruleset compatibility before the first task.
- **PR descriptions go stale; issue comments don't.** PR radiusred/www#42's
  body says "Deviations: none," written before the deviation happened; the
  record landed on the issue afterward, where the gatherer caught it. Not a
  lost record — but a reminder that the issue-comment trail is the
  authoritative record, and PR bodies are snapshots.
