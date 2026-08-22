# Your first milestone

This is the whole path, start to finish, for the most common setup: **one
human, one coding agent to hand** (a Claude Code, Codex, or similar session),
and a brand-new project. Install CodeCrew, scaffold a repo, and drive one
milestone from opening to audited close. No GitHub Apps, no configuration
beyond what the tool writes for you, no prior knowledge assumed — and no
paid GitHub plan: everything here works on the free tier.

Two ideas carry everything else. First, **project state lives in GitHub** —
issues, sub-issues, labels, and PRs — so there is nothing to keep in sync
and nothing to lose. Second, **the verbs refuse**: when a gate isn't met,
the CLI exits with `refused[CODE]: detail` telling you exactly what's
missing. A refusal is the guide-rail working, not an error. This guide walks
you into each one you'll meet, on purpose.

## You and your agent

The expected division of labour: **your agent runs the loop, you make the
calls**. The agent acts as orchestrator and implementer in one session — it
runs the verbs, writes plans into task issues, does the work, opens the
PRs, and stops at every gate; you review, resolve raised gates, confirm
merges, and give verdicts. Point the agent at this page (and `AGENTS.md`,
which `init` writes) and ask it to drive; the refusals keep it honest even
when you're not watching.

Every command below is equally pasteable by a human — the protocol is fully
human-operable end to end, and running one milestone by hand is a fine way
to learn what your agent is doing. But the framework earns its keep when
the agent does the typing and the record proves what happened.

## 0. Prerequisites

A GitHub account, [`gh`](https://cli.github.com/) installed and signed in
(`gh auth login`), `git`, and (recommended) your coding agent of choice.
You and your agent are the whole crew today (SPEC §5 calls this solo
routing; the [What next](#what-next) ladder covers expanding it).

## 1. Install and scaffold

```sh
gh extension install radiusred/gh-codecrew

gh repo create my-project --private --clone   # or use an existing repo
cd my-project
gh codecrew init
```

`init` writes your `.codecrew.yml` (the project's pointer file, with a
routing table declaring all four roles — `~` means "held by you"), a
`ROADMAP.md` seed, the four role contracts under `roles/`, and an
`AGENTS.md` entry point for any agent you later dispatch. Commit and push:

```sh
git add -A && git commit -m "chore: scaffold codecrew" && git push
```

## 2. Open a milestone

```sh
gh codecrew milestone new --title "Walking skeleton" \
  --goal "A deployed hello-world proving the delivery pipeline end to end."
```

This creates the milestone tracking issue — the canonical milestone object —
and adds a row to `ROADMAP.md` (commit it with your next PR). Now edit the
issue's **Requirements** section on GitHub, giving each requirement a bold
ID the tooling can check "done" against:

```markdown
## Requirements
- **M1-R1** — visiting the app's URL returns a greeting
```

## 3. Open a task, plan it, start it

```sh
gh codecrew task new --milestone 1 --title "Serve the greeting" --requirements M1-R1
```

The task issue is attached to the milestone as a sub-issue — `gh codecrew
status` now shows your project. Try to start it:

```sh
gh codecrew task start 3   # your task's issue number
```

It refuses: `refused[NO_PLAN]`. Plans come before work — edit the task
issue's **Plan** section with your intended changes (a trivial task may have
a trivial plan; it may not have an absent one), then run `task start` again.
It assigns you and creates a linked branch.

## 4. Do the work

```sh
git fetch && git switch task/3-serve-the-greeting
# ...write the code, in commits that reference the task...
git commit -m "feat: serve the greeting (#3)"
git push
gh pr create --fill --body "Closes #3"
```

The PR description is the task's summary artifact — say what was done and
which requirement it satisfies. If a decision came up mid-task (a trade-off,
a rejected alternative), record it as an issue or PR comment starting
`**Decision:**` *at the moment you make it* — these comments are collected
automatically at milestone close.

## 5. Finish the task

```sh
gh codecrew task finish 3
```

Refused again — no approving review from a non-author, and you can't approve
your own PR. Solo has an honest degradation:

```sh
gh codecrew task finish 3 --operator-confirm
```

This records an explicit confirmation comment on the PR — stating that
author and operator are the same principal — then rebase-merges. The task
issue closes via its `Closes #3` keyword. (When your repo has CI, `task
finish` also refuses on failing or pending checks; add CI early so this
gate has teeth.)

## 6. Verdict — you hold the qa role

Try to close the milestone:

```sh
gh codecrew milestone close 1
```

`refused[VERDICT_MISSING]` — every requirement needs a QA verdict, and the
qa role is unrouted, which means it's yours. Change hats: build what was
merged, exercise the requirement's *intent* (not just what the code's own
tests prove), and post a comment on the **milestone issue**, one line per
requirement, in exactly this form:

```markdown
**M1-R1 — satisfied.** Deployed from main, visited the URL, got the greeting.
```

The evidence is yours, from your own execution — `roles/qa.md` is the
contract you just performed.

## 7. The milestone document

```sh
gh codecrew milestone close 1
```

One refusal left: `refused[DOC_MISSING]` — and above it, the CLI prints
every `**Decision:**` and `**Deviation:**` comment recorded during the
milestone. That's the raw material. Write `docs/milestones/1-<slug>.md` from
it — the "why" record: what was decided, what was traded off, what deviated
from plan — flip the ROADMAP row to Done, and merge it like any other change
(a PR, `--operator-confirm` if you like the discipline).

## 8. Close

```sh
gh codecrew milestone close 1
```

All gates pass: tasks closed, verdicts satisfied, document merged. The
milestone issue closes with the record comment, and everything above — the
plan, the decisions, the confirmation, the verdicts, the document — is
permanent, linked, and public to whoever you choose.

## What next

The loop above is the same shape at every size — more requirements per
milestone, more tasks per requirement, CI required checks to give `task
finish`'s gate real teeth ([SPEC §8](../SPEC.md)). The interesting growth is
in the crew, and it's a ladder:

1. **Split the roles across sessions.** Your agent doing the work and then
   verdicting its own milestone shares one context's blind spots. The cheap
   fix: have your harness launch a sub-agent per role, each briefed with the
   role's contract from `roles/` — a fresh-context QA reading `roles/qa.md`
   probes what the implementer's context wouldn't. Or cross model families:
   dispatch another LLM through its own CLI for the reviewer or qa seat.
   Same identities, same commands — just different eyes.
2. **Give crew members their own identities.** When you want the record to
   show *which* agent did what — and GitHub itself to enforce that the
   approver isn't the author — route roles to GitHub App identities (or
   other humans, by username): [identities.md](identities.md). The protocol
   doesn't change; only the routing table does.
3. **Full orchestration platforms** — an orchestrator dispatching the whole
   crew against the routing table, webhooks instead of polling. Works today
   through the same seams; dedicated docs are on the backlog.
