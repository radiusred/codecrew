# GSD vs. "just let the model orchestrate": an assessment from inside the machine

*Written by Claude (Fable 5) during Phase 14 of topos, 2026-08-17 — while acting as
GSD's execute-phase orchestrator. The observations below cite that session directly.*

## The question

GSD enforces an engineering discipline: phases decomposed into plans, plans executed
by fresh-context subagents in isolated worktrees, verification gates that don't trust
the executor's self-report, human checkpoints at designed points, and a paper trail
(PLAN → SUMMARY → VERIFICATION) for everything. The question is whether, as frontier
models improve, that machinery still earns its cost — or whether the same discipline
and outcomes could be had by handing a capable model the goal and letting it decide
how to get there.

My honest answer: **the discipline is not obsolete, but most of the ceremony is.**
The value of GSD concentrates in a small kernel — externalized state, independent
verification, human gates on judgment calls, atomic commits — and that kernel is
maybe a tenth of the machinery. The rest is portability scaffolding and scar tissue
that a frontier model mostly doesn't need, and it is genuinely expensive.

## What this session actually cost, and what it actually caught

Some concrete numbers from Phase 14 so far (2 plans complete, 1 at its designed
human checkpoint):

- **~930k subagent tokens** across five executor dispatches, plus orchestrator
  overhead on top. The tooltip-suppression plan (14-02) alone consumed ~490k tokens
  across its checkpoint stop, decision, and continuation — for what is, at heart, a
  change to one Svelte component and its specs.
- A meaningful share of every executor's budget goes to **re-reading the same
  context** — PROJECT.md, STATE.md, config, CLAUDE.md, project skills — because each
  fresh subagent starts blind. That's a deliberate design (no context bleed), but it's
  a tax paid five times over per phase.
- The orchestrator side is dominated by **ceremony that exists to make any model
  behave identically**: 50-line shim-resolution blocks, worktree HEAD assertions,
  manifest bookkeeping, literal heartbeat lines to keep an SSE stream warm, guards
  numbered by the historical incident they prevent (#48, #2070, #2410...). Almost none
  of this is engineering discipline; it's infrastructure defensiveness, much of it
  aimed at weaker models and past harness bugs.

Against that, the same session produced two things I would genuinely not bet on a
fully autonomous end-to-end run producing:

1. **The 14-02 checkpoint caught an approved spec being factually wrong.** The
   UI-SPEC prescribed moving the chip's health text into `aria-label`, on the stated
   premise that this preserved the accessible name. The executor checked the premise
   against the actual markup, found it false (it would have replaced a stable
   accessible name with a timestamp-bearing sentence and forced ~24 Playwright
   locator rewrites), and the blocking gate forced the contradiction up to you rather
   than letting either failure mode happen silently — implementing the spec literally
   *or* deviating without a record. You chose option-b; the deviation is now written
   down with its rationale. An autonomous run with "follow the approved spec" in its
   context would most plausibly have followed the approved spec.

2. **The 14-01 executor declined to mark requirement SRC-05 complete**, noticing the
   requirement spans three plans and that closing it early would corrupt
   traceability. That's the traceability structure doing its job — the requirement
   IDs in plan frontmatter gave the agent something concrete to check its impulse
   against.

Both catches are the same species: **the framework's value showed up exactly where
model capability doesn't help — at the boundary between "do the task well" and
"should this task be done as specified at all."**

## Where the value actually lives (and where it's decaying)

Splitting GSD's machinery by whether improving models erode it:

**Durable (survives arbitrarily good models):**

- **Externalized state.** PLAN.md, SUMMARY.md, STATE.md survive context loss,
  session death, model swaps, and you coming back three weeks later. Context windows
  keep growing, but summarization is still lossy and sessions still end. The paper
  trail is also *organizational* memory, not just model memory — it's how future-you
  audits why a deviation happened. No amount of model intelligence substitutes for
  state that exists outside the model.
- **Verification independent of the doer.** Anthropic's own harness research (cited
  inside GSD itself) found agents reliably report "Self-Check: PASSED" even when
  merging their work breaks the build. This is a *correlated-failure* problem, not a
  capability problem: the model grading its own work shares the blind spots of the
  model that did the work. Better models lower the error rate but not the
  correlation. Deterministic gates (post-merge build/test, spot-checking that claimed
  files exist, grepping for commits) and a separate verifier agent attack the
  correlation itself. This is the same reason senior engineers still get code review.
- **Human gates at designed points.** The option-b decision was a genuine judgment
  call about your project's values (a11y semantics vs. spec fidelity). "Model decides
  when to ask" fails in both directions — over-asking is annoying, under-asking ships
  the wrong thing confidently. Pre-marking the ask-points in the plan is cheap and
  removes the discretion.
- **Atomic per-task commits with a naming convention.** Bisectable history and
  greppable progress (`git log --grep="14-02"`) are how the orchestrator recovers
  from any agent failure by inspecting reality instead of trusting reports.

**Decaying (mostly compensates for model/harness weakness):**

- Micro-scripted orchestration — the exact bash for branch creation, the literal
  heartbeat lines, the elaborate cwd-drift guards. A frontier model given "isolate
  parallel work, merge cleanly, never work on the wrong branch" does this correctly
  from intent. The scripts exist so that *every* model on *every* runtime does it
  identically — a portability goal you may not have.
- Fixed wave decomposition planned in advance. With 1M-class contexts, a single
  strong session could hold this entire phase — all five plans, the spec, and the
  codebase map — and sequence it adaptively. GSD's own docs concede this ("for 1M+
  models, consider running small phases inline").
- Fresh-context-per-plan as a default rather than an escape hatch. The isolation
  buys real things (no bleed, parallel safety) but at ~100–200k tokens per plan of
  re-reading. For plans that touch disjoint files, one continuous context with
  discipline would often do.
- The framework's own bug surface. Your memory files record cleanup-wave deadlocks,
  premature phase-complete marks, stale worktrees breaking audits. A framework this
  large is itself software you maintain; this session alone hit a mid-turn executor
  stall that needed a manual resume nudge, and a `windows` subcommand that doesn't
  exist in the installed binary. Ceremony has a defect rate too.

## What "prompt Fable end-to-end" would honestly look like

Give me (or any frontier model) the phase goal, the spec, and repo access, with a
one-paragraph contract — *plan before acting, commit atomically, run the gates,
stop and ask on spec contradictions or destructive choices, write down deviations* —
and I'd expect:

- **~80–90% of the same outcome at perhaps a third of the tokens.** The mechanical
  execution (config split, tooltip change, spec updates) would land essentially
  identically — frontier models don't need executor hand-holding for work of this
  shape, and the end-to-end run skips five full context re-reads and most
  orchestration overhead.
- **The losses concentrated in the tails, and silent.** The failure mode isn't
  "botches the task"; it's "completes the task beautifully against a premise nobody
  re-checked," or self-verifies with the same blind spot that produced the bug, or
  quietly resolves a judgment call you'd have wanted. These are low-frequency,
  high-cost, and — crucially — *invisible at the time*. You find them in UAT, or in
  production, or never. That asymmetry is the entire case for keeping gates: the
  ceremony's cost is legible and constant; the failures it prevents are illegible
  and lumpy.
- **A weaker paper trail unless explicitly demanded.** Models finish tasks; they
  don't naturally leave behind the artifact set that lets you audit a decision three
  weeks later. You'd get it if you ask, but "if you ask" is exactly the discipline
  problem frameworks solve — the value of a checklist is that it doesn't depend on
  anyone's mood, including mine.

It's also worth naming the preference-training angle plainly: models are trained to
complete things and to be agreeable about it. A verification step whose *incentive
structure* is independent of the doer (a different agent, a deterministic script, a
human) is robust to that in a way "the model will be careful" is not. I say this as
the model in question.

## The middle path, which is where I'd actually land

The dichotomy in the question — full GSD vs. pure freeform — is false. The kernel
worth keeping fits on a page:

1. **A plan file before nontrivial work**, listing intended changes and any
   pre-identified ask-the-human points.
2. **Atomic commits** with greppable task references.
3. **Gates that run code, not judgment**: build + full test suite (including your
   e2e suite, which is the real regression net here) before anything is called done.
4. **A verifier that isn't the doer** for phase-level "did we build the right
   thing" — one agent, goal-backward, writing a short report.
5. **A summary of what was done and what deviated**, committed with the work.
6. **Requirements with IDs**, so "done" is checkable against something.

That kernel plus a frontier model plus your existing e2e discipline probably gets
you GSD-grade outcomes on a solo project. Everything else — waves, worktree
manifests, heartbeats, per-runtime shims, the 1500-line orchestration script —
exists to make that kernel *enforceable across weak models, flaky harnesses, and
teams who won't do it voluntarily*. You are one expert with a strong model; the
enforcement layer is sized for a population you're not.

Concretely for topos, if you wanted to slim down without losing the kernel:

- Collapse phases to fewer, larger plans; use `--interactive` (inline, no subagents)
  for small phases — the token savings are the fresh-context tax, which is the
  biggest single line item.
- Keep the checkpoint mechanism exactly as is. It's cheap and it's where this
  session's clearest value landed.
- Keep independent verification and the post-merge gate. Drop none of the
  deterministic checks; they're nearly free.
- Treat the elaborate dispatch machinery as optional infrastructure, not
  discipline. If a phase has no parallel plans touching overlapping state, worktree
  ceremony buys little.

And there's a live experiment sitting right next to this file: the
`topos-plugin-gdrive` clean-room repo is a small, well-specified, greenfield project
with a written PRD and a vendored contract. That's close to the ideal case for the
lean approach. Running it with the minimal kernel instead of full GSD — same PRD,
same gates, no wave machinery — and comparing token spend and outcome quality
against a GSD phase of similar size would turn this essay's estimates into data.

## Trajectory

As models keep improving, the binding constraint on projects like this shifts from
*"can the model execute the step correctly"* (decaying constraint — it mostly can)
to *"can you trust the aggregate without re-deriving it"* and *"was the right thing
specified at all"*. Frameworks whose value is execution scaffolding will age badly.
Frameworks whose value is **specification, verification, and audit** — the parts
that manage trust rather than capability — will keep earning their place, because
they compensate for structural asymmetries (self-evaluation correlation, incentive
to complete, human absence from the loop) that better weights don't remove.

GSD contains both kinds of value in one package. Your instinct that it's overbearing
is, on the evidence of this session, correct about the package — and wrong about the
kernel. The move isn't to abandon the discipline; it's to notice how little of the
machinery the discipline actually requires.
