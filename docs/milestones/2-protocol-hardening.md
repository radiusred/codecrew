# M2: Protocol hardening

Tracking issue: [#15](https://github.com/radiusred/gh-codecrew/issues/15) ·
Closed 2026-08-20 · Synthesized by the doc-synthesizer role
([radiusred-wordy](https://github.com/apps/radiusred-wordy)) from the records
gathered by `codecrew milestone close 2`, the gate trail on #19, and QA's
verdicts on #15.

## Goal and outcome

Harden the protocol against manual-bookkeeping rot and prove the remaining
roles. Delivered in full: task tracking moved to GitHub sub-issues with the
hand-maintained list deleted, the README gained a protocol-enforced freshness
mechanism (and was immediately caught stale by it — see QA below), the QA
role ran for the first time under a Codex harness, and `codecrew` became an
installable gh extension — which required renaming the repository.

## Decisions

- **Track milestone tasks as GitHub sub-issues; the milestone body carries no
  task list.** Trade-off: dependence on a newer GitHub API surface versus
  zero hand-maintained tracking state. Rejected: the markdown checkbox list
  (rotted in M1 the first time a tick was forgotten), GitHub's retired
  tasklist beta, and Projects-as-tracker. Body parsing was deleted rather
  than kept as a fallback — one representation, no drift.
  ([Decision on #16](https://github.com/radiusred/gh-codecrew/issues/16#issuecomment-5357032988))
- **Rename the repository to `gh-codecrew`** rather than maintain a separate
  distribution-only artifact repo (or defer the extension requirement). This
  was M2's first live human gate: `codecrew checkpoint` raised
  `cc:needs-decision` on #19 with the options, and the operator chose the
  rename; GitHub's redirects preserved every existing URL, remote, and issue
  reference. ([gate and resolution on #19](https://github.com/radiusred/gh-codecrew/issues/19#issuecomment-5357164380))
- **README freshness is a doc-synthesizer obligation at milestone close**,
  not a harness hook — a hook in one operator's tooling would silently not
  exist for other harnesses or teammates; the protocol obligation rides the
  reviewed milestone PR. (PR
  [#21](https://github.com/radiusred/gh-codecrew/pull/21); undocumented as a
  Decision comment — inferred from the PR description.)

## QA (first dispatch)

`radiusred-testy[bot]`, driven by a Codex CLI harness, executed
independently against M2's requirements
([verdicts on #15](https://github.com/radiusred/gh-codecrew/issues/15),
findings on [#18](https://github.com/radiusred/gh-codecrew/issues/18)):

- **M2-R1 satisfied** — verified by building from `main`, running `status`,
  and inspecting the sub-issues REST surface and the code paths directly.
- **M2-R2 not satisfied at time of audit** — the README claimed "Not yet
  here: a QA agent dispatch history," a claim the QA dispatch itself had
  falsified. Remedied in this document's PR; the verdict's recursion is worth
  savoring: the freshness mechanism's first catch was the sentence denying
  the catcher existed.
- **M2-R4 satisfied** — installed the published extension, ran it live,
  verified all five release artifacts and their digests.

QA changed no files, commits, or branches, per its contract.

## Deviations

None recorded during M2. (The M1 deviation pattern — bot identities not
being assignable — remained in effect and visible in every `task start`.)

## Requirement outcomes

| Requirement | Delivered by | Outcome |
|-------------|--------------|---------|
| M2-R1 — sub-issue task tracking, no hand-maintained lists | [#16](https://github.com/radiusred/gh-codecrew/issues/16) / PR #20 | Done; QA satisfied |
| M2-R2 — README reflects reality + freshness obligation | [#17](https://github.com/radiusred/gh-codecrew/issues/17) / PR #21, remedied by [#23](https://github.com/radiusred/gh-codecrew/issues/23) | Done after QA-found staleness fixed; re-verified post-merge |
| M2-R3 — first QA dispatch with per-requirement verdicts | [#18](https://github.com/radiusred/gh-codecrew/issues/18) | Done |
| M2-R4 — gh extension installable | [#19](https://github.com/radiusred/gh-codecrew/issues/19) / PR #22, release v0.1.0 | Done; QA satisfied |

## Protocol-discipline gaps observed

- **Gate resolutions are not Decision-formatted.** The #19 rename decision
  was recorded as "Gate resolved (operator)" prose, so `milestone close`'s
  record gathering missed it (1 record gathered for a 2-decision milestone).
  Candidate fix: the resolution comment convention should reuse the
  `**Decision:**` structure, or the gatherer should also collect gate
  resolutions.
- **QA tasks get linked branches they never use.** `task start` creates a
  development branch even for roles that must not commit; harmless clutter,
  but a `--no-branch` flag (or role-aware default) would be cleaner.
- **Verdict-remedy loop is manual.** QA's not-satisfied verdict was remedied
  in the close PR and re-verified by a targeted re-dispatch, but nothing in
  the protocol forces the re-verification; a future `milestone close` gate
  could require a satisfied (or superseded) verdict per requirement.
