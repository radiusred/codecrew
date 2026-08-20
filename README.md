# CodeCrew

A lightweight framework for agent-driven software delivery: the auditability
and reproducible discipline of heavyweight frameworks, without the ceremony.
Project state lives in the tools teams already use — GitHub issues, PRs, and
CI — and the only documents the framework produces are per-milestone records
of the decisions that shaped the system and why.

CodeCrew is three things:

1. **A protocol** ([SPEC.md](SPEC.md)) — conventions for representing
   milestones, tasks, plans, decisions, deviations, and gates in GitHub, and
   how agents and humans transact over them.
2. **Role contracts** ([roles/](roles/)) — harness-neutral prompt files for
   the implementer, reviewer, qa, and doc-synthesizer roles, loadable by any
   agent (Claude Code, Codex, Gemini CLI, or an orchestrator's company).
3. **A CLI** — `codecrew`, a single static Go binary wrapping `gh`, providing
   the workflow verbs with gates enforced as code.

## Status

The protocol is specified and the CLI works: `status`, `milestone new/close`,
`task new/start/finish`, and `checkpoint` are all implemented, with
machine-readable refusals (`refused[CODE]: detail`) when a gate blocks.
Milestone 1 was built *with* the protocol it delivered — agent-authored PRs
under GitHub App identities, non-doer review, deterministic CI gates, and a
synthesized closing document — and was closed by the binary itself. See
[docs/milestones/](docs/milestones/) for the per-milestone records.

Not yet here: a QA agent dispatch history, and any backend other than GitHub.

## Install and use

```sh
gh extension install radiusred/gh-codecrew   # precompiled, all platforms
gh codecrew status         # open milestones, inferred task states, raised gates
gh codecrew help           # the full verb list

# or build from source (single static binary; requires gh on PATH):
go build ./cmd/codecrew
```

Every repo in a CodeCrew project carries a `.codecrew.yml` pointing at the
hub (this repo is its own hub: `hub: self`). Agents dispatched into a
CodeCrew repo start at [AGENTS.md](AGENTS.md). Agent identities are GitHub
Apps; short-lived tokens come from `scripts/codecrew-token` (see SPEC §5 for
the identity tiers — a solo operator needs nothing but `gh auth login`).

## Documents

- [SPEC.md](SPEC.md) — the protocol
- [docs/founding-decisions.md](docs/founding-decisions.md) — design decisions
  and their trade-offs
- [docs/milestones/](docs/milestones/) — the per-milestone "why" records
- [docs/gsd-vs-frontier-orchestration.md](docs/gsd-vs-frontier-orchestration.md)
  — the analysis that motivated the project

Licensed under [Apache 2.0](LICENSE).
