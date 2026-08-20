# Working in a CodeCrew project

This repository follows the CodeCrew protocol. If you are an agent dispatched
to work here:

1. **Find the hub.** Read `.codecrew.yml`; `hub: self` means this repo is the
   hub, otherwise follow `hub: owner/repo`.
2. **Load your role contract** from the hub's `roles/` directory — you were
   dispatched as one of: [implementer](roles/implementer.md),
   [reviewer](roles/reviewer.md), [qa](roles/qa.md),
   [doc-synthesizer](roles/doc-synthesizer.md). If no role was named, you are
   the implementer.
3. **Resolve your identity** per the contract: orchestrator-injected env vars
   (`GITHUB_CLIENT_ID`, `GITHUB_PRIVATE_KEY`, `GITHUB_INSTALLATION_ID`) first,
   then the role's app key in `~/.config/codecrew/` via
   `scripts/codecrew-token`, then the operator's own `gh` auth.
4. **The protocol is [SPEC.md](SPEC.md)** in the hub. The short version:
   state lives in GitHub issues and PRs, not in files; plan in the task issue
   before the first commit; atomic conventional commits referencing the task
   (`(#123)`); record decisions and deviations as structured comments the
   moment they happen; never verify or approve your own work; stop for a
   human at ask-the-human points and whenever `cc:needs-decision` is raised.

Commit messages follow conventional commits and are linted in CI.
