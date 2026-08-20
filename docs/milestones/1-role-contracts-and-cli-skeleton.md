# M1: Role contracts and CLI skeleton

Tracking issue: [#1](https://github.com/radiusred/codecrew/issues/1) ·
Closed 2026-08-20 · Synthesized by the doc-synthesizer role
([radiusred-wordy](https://github.com/apps/radiusred-wordy)) from the records
gathered by `codecrew milestone close 1`.

## Goal and outcome

Prove the CodeCrew protocol by using it to build CodeCrew's own first
increment: the four role contracts and a working `codecrew` CLI covering the
full SPEC §6 verb set. The milestone delivered everything it promised, and
every task in it was executed under the protocol it was building — task
issues with plans before first commits, agent-authored PRs under app
identities, non-doer approvals, and this document as the closing gate.

## Decisions

The project's pre-tracker design decisions (backend choice, topology,
everything-is-an-issue, form factor, language, and more) are recorded with
their trade-offs in [docs/founding-decisions.md](../founding-decisions.md) —
they predate the first issue. Decisions recorded in-tracker during M1:

- **Reuse the org's existing GitHub Apps as per-role agent identities**
  (cody/testy/wordy), minting short-lived installation tokens from
  locally-held keys. Rejected: machine user accounts (consume a seat, share
  the self-approval problem) and fresh purpose-made apps (the existing fleet
  already mapped one-to-one onto the roles). Reviewer identity stays with the
  human operator until agent review is automated.
  ([Decision on #6](https://github.com/radiusred/codecrew/pull/6#issuecomment-5350328251))

Decisions made during M1 but recorded in PR descriptions and SPEC changes
rather than as Decision comments (a discipline gap — see below):

- **Identity tiers** (PR [#8](https://github.com/radiusred/codecrew/pull/8)):
  app identities are optional infrastructure, required only where a
  GitHub-native approval must come from a party with no human account —
  enforced-review-with-one-human, or agent-reviews-agent. Solo operation
  degrades `task finish` to an explicit operator confirmation. Now SPEC §5/§6.
- **Machine-readable refusals** (PR
  [#12](https://github.com/radiusred/codecrew/pull/12)): blocked gates exit
  nonzero with `refused[CODE]: detail` so agents can act on the specific
  unmet condition rather than parse prose.
- **Credential resolution order** (PR #8, from the org's existing `gh-cli`
  skill): orchestrator-injected env vars, then local private key, then the
  operator's own `gh` auth — one convention across solo, team, and
  Paperclip-dispatched operation.

## Deviations

- **Early merges bypassed the review requirement.** PRs
  [#2](https://github.com/radiusred/codecrew/pull/2) and
  [#4](https://github.com/radiusred/codecrew/pull/4) were admin-merged
  because a single GitHub identity authored everything and GitHub forbids
  self-approval. The deviation was recorded at the time
  ([Deviation on #4](https://github.com/radiusred/codecrew/pull/4#issuecomment-5350256585))
  and resolved by the app-identity decision above; every merge from PR #6
  onward carried a genuine non-doer approval.
- **Bot identities are not assignable** (GitHub returns 403), so `task start`
  records a "Started by" comment instead of setting the assignee. The
  in-progress state inference therefore under-reports for bot-run tasks — an
  accepted v1 limitation, noted for a future protocol revision.

## Requirement outcomes

| Requirement | Delivered by | Outcome |
|-------------|--------------|---------|
| M1-R1 — role contracts under `roles/` | [#7](https://github.com/radiusred/codecrew/issues/7) / PR #8 | Done |
| M1-R2 — static Go CLI wrapping `gh`, verb-shaped backend seam | [#9](https://github.com/radiusred/codecrew/issues/9) / PR #10 | Done |
| M1-R3 — `status` with inferred states and raised gates | #9 / PR #10 | Done |
| M1-R4 — `milestone new` / `task new` templated and linked | [#11](https://github.com/radiusred/codecrew/issues/11) / PR #12 | Done |
| M1-R5 — `task start` / `checkpoint` / `task finish` gate enforcement | #11 / PR #12 | Done |
| M1-R6 — this document, synthesized from recorded material | [#13](https://github.com/radiusred/codecrew/issues/13) | Done on merge |

Gate verification: CI (commitlint + Go build/test) green on every merged PR;
`codecrew milestone close 1` was executed by the built binary and correctly
refused with `DOC_MISSING` until this document existed — the milestone is
closed by the tool it delivered.

## Protocol-discipline gaps observed

Recorded so M2 can act on them, per the doc-synthesizer contract:

- **Decision capture under-used.** Only two structured records existed for a
  milestone that made many decisions; most rationale lived in PR descriptions
  and the founding document. The convention works — the gathered records were
  exactly the right raw material — but implementers should reach for
  `**Decision:**` comments more readily.
- **The QA role was never dispatched.** No independent behavioural
  verification ran against M1's requirements; review plus CI stood in for it.
  First QA dispatch (testy under a Codex harness) is a natural M2 candidate.
- **Requirement links in milestone-issue task entries** are maintained by
  hand; `task new` links the issue but does not annotate which requirements
  it covers in the milestone body.
