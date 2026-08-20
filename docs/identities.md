# Identities: running solo, staffing a crew

Roles are contracts, not accounts. Nothing in CodeCrew requires a GitHub App
to exist before a role can be staffed — a role with no identity acts as the
human operator, and the whole protocol runs on `gh auth login` alone. App
identities are infrastructure you add when you want them (attribution) or
need them (enforced independent review). This document covers both ends:
the zero-setup solo path, and minting a proper App identity for a role.

The conceptual model — identity tiers, credential resolution order — is in
[SPEC §5](../SPEC.md); this is the operational companion.

## Running solo

Prerequisites: a GitHub account, `gh` authenticated (`gh auth login`), the
extension installed (`gh extension install radiusred/gh-codecrew`), and a
`.codecrew.yml` in your repo (`hub: self` for a single-repo project).

Everything works. You (or agents running under your auth) create milestones
and tasks, write plans, commit, open PRs, raise and resolve gates. Exactly
one thing degrades: the non-doer review gate. GitHub forbids approving your
own PR, and in pure solo there is no second principal to approve — so
`task finish` accepts `--operator-confirm` instead, and records an explicit
confirmation comment on the PR. When the confirming operator is also the PR
author, the recorded comment says so in as many words: *no independent
principal exists in this project*.

Two honesty notes:

- The confirmation must come from a **human** identity. An identity carrying
  `[bot]`, or one routed to a role in `.codecrew.yml`, is refused
  (`refused[SELF_CONFIRM]`) — agents can never waive review, in any tier.
- Pure solo means no independent verifier, and the record shows it. That is
  the honest floor, not a failure mode; climb a tier when the review gate
  needs teeth.

## Minting a crew member

When a role should act as itself — attributable work, and an approver
distinct from the author — it gets its own **GitHub App**. One App per role,
named for the crew member, not the role (`myorg-coder`, not
`myorg-implementer`: identities outlive role reassignments).

There is no framework tooling for this yet (a manifest-flow
`codecrew identity new` is the intended future); today it is a short manual
ritual per role:

1. **Create the App, owned by the org** (Org → Settings → Developer settings
   → GitHub Apps → New GitHub App). Homepage URL can be anything; deactivate
   webhooks — CodeCrew's Apps only ever act, they never listen.
2. **Grant the minimum repository permissions** for the role's contract.
   `Metadata: read` is always required; beyond that:

   | Role            | Contents      | Issues        | Pull requests  | Checks |
   |-----------------|---------------|---------------|----------------|--------|
   | implementer     | Read & write  | Read & write  | Read & write   | Read   |
   | reviewer        | Read          | Read & write  | Read & write   | Read   |
   | qa              | Read          | Read & write  | Read & write   | Read   |
   | doc-synthesizer | Read & write  | Read & write  | Read & write   | Read   |

   Add `Workflows: read & write` to the implementer if it will ever touch
   `.github/workflows/` files — `Contents` alone cannot push those.
3. **Generate a private key** and store it outside any repo. Convention:
   `~/.config/codecrew/<app-slug>.<date>.private-key.pem`.
4. **Install the App on the org**, scoped to all repositories or at least to
   the hub and every spoke the role will touch. A spoke the App cannot see
   is a spoke the role cannot serve.
5. **Route the role** in the hub's `.codecrew.yml`:
   `roles.<role>.identity: <app-slug>`.

### Acting as the App

Mint a short-lived installation token per invocation — either via the
env-var path from SPEC §5 (`GITHUB_CLIENT_ID` / `GITHUB_PRIVATE_KEY` /
`GITHUB_INSTALLATION_ID`, for orchestrators), or with the bootstrap script
in this repo: `scripts/codecrew-token <app-slug>` — the **full** App slug
(`myorg-coder`, not `coder`). Then:

- `GH_TOKEN=$tok gh codecrew …` / `GH_TOKEN=$tok gh …` for API actions.
- Push over HTTPS as `https://x-access-token:$TOKEN@github.com/owner/repo.git`.
- Commit author is the App's bot user:
  `<slug>[bot] <UID+<slug>[bot]@users.noreply.github.com>`, where UID comes
  from `gh api 'users/<slug>%5Bbot%5D' --jq .id`.

### Known quirks

- **Bot identities are not assignable to issues.** `task start` handles this:
  it records a `**Started by**` comment instead. Expected, not an error.
- **The viewer login carries a `[bot]` suffix** (`myorg-coder[bot]`) while
  the routing table names the bare slug; the CLI normalises this everywhere
  it resolves roles.
- **Approvals from Apps and required-review rules:** whether an App's
  approving PR review satisfies a ruleset's required-review count is a
  platform question to verify before depending on it (see SPEC §5, Platform
  requirements, for the tier-intentionality obligation).
