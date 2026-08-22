# Your first milestone

This is the whole path, start to finish, for one person and a brand-new
project: install CodeCrew, scaffold a repo, and drive one milestone from
opening to audited close. No GitHub Apps, no configuration beyond what the
tool writes for you, no prior knowledge assumed. Every command is one you
can paste.

Two ideas carry everything else. First, **project state lives in GitHub** —
issues, sub-issues, labels, and PRs — so there is nothing to keep in sync
and nothing to lose. Second, **the verbs refuse**: when a gate isn't met,
the CLI exits with `refused[CODE]: detail` telling you exactly what's
missing. A refusal is the guide-rail working, not an error. This guide walks
you into each one you'll meet, on purpose.

## 0. Prerequisites

A GitHub account, [`gh`](https://cli.github.com/) installed and signed in
(`gh auth login`), and `git`. That's all — you are the whole crew today
(SPEC §5 calls this solo routing; [identities.md](identities.md) covers
delegating roles to agents later).

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

- More requirements per milestone, more tasks per requirement — the loop is
  the same shape at every size.
- CI required checks give `task finish`'s gate real teeth ([SPEC
  §8](../SPEC.md)).
- When you want a second pair of eyes — human or agent — route a role:
  [identities.md](identities.md). The protocol doesn't change; only the
  routing table does.
