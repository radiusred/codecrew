# Founding decisions

Decisions made during CodeCrew's design phase (August 2026), recorded in the
format the protocol itself prescribes. Once the hub is bootstrapped, decisions
are captured as issue/PR comments and synthesized into milestone documents;
these predate the first issue, so they live here.

---

**Decision:** GitHub is the only backend, behind an interface shaped by the
workflow verbs.
**Trade-off:** Portability to other trackers versus a small, honest v1 surface.
**Rejected:** A Jira adapter, or designing the interface against Jira's
constraints now. Jira is an expert system that takes any shape through
configuration; abstracting over it produces a lowest-common-denominator
interface that serves nobody. The verb-shaped interface (with tracker and
review-surface as conceptually separate ports) is the seam if a second backend
is ever justified.

---

**Decision:** Everything is an issue. The canonical milestone object is a
tracking issue in the hub; GitHub Projects v2 is never canonical.
**Trade-off:** No native board/roadmap UI versus a uniform, universally
accessible protocol.
**Rejected:** Projects v2 as the roadmap spine. Its GraphQL-only API and
org-token requirements would couple every agent harness to GitHub's most
awkward API, while issues are readable and writable by any harness through
`gh` or REST. A read-only Projects mirror may be layered on later as a human
dashboard, derived from issue state.

---

**Decision:** Hub-and-spokes topology, with no single-repo mode — a
single-repo project is the degenerate case where the hub is the spoke.
**Trade-off:** A pointer file and qualified references carried even by solo
projects, versus zero migration when a project grows multi-repo.
**Rejected:** A distinct single-repo mode (growth becomes a representation
migration) and pure per-repo autonomy with no hub (cross-repo milestones
fragment into naming conventions; milestone documents have no home).

---

**Decision:** Task issues live in the spoke whose code they change, not in
the hub.
**Trade-off:** Milestone membership must be maintained by cross-repo links,
versus GitHub's native traceability working unmodified.
**Rejected:** Centralizing all issues in the hub. That severs closing
keywords, PR linkbacks, CODEOWNERS, and repo-scoped permissions — traceability
would have to be rebuilt by convention, which is the corpus-maintenance trap
CodeCrew exists to escape.

---

**Decision:** Documentation is synthesis at milestone close, from decisions
and deviations captured in structured comments at the moment they occur.
**Trade-off:** A small in-flight discipline (the comment convention) versus a
maintained documentation corpus.
**Rejected:** Reconstructing the "why" from raw history at milestone close —
lossy and prone to confabulation — and GSD-style per-phase document sets,
whose maintenance cost is the ceremony being discarded.

---

**Decision:** Live tracker state is mutable and unpoliced; the durable audit
trail is what lands in git (commits, merged PR descriptions, committed
milestone documents).
**Trade-off:** A team can edit or delete agent-produced records versus zero
enforcement machinery.
**Rejected:** Snapshotting or hashing tracker content to detect tampering.
Teams that rewrite the record either have good reason or pay the price; the
milestone document already gives a point-in-time snapshot in git history for
free.

---

**Decision:** Harness-neutral form factor — a protocol document, role
contracts, and a thin CLI. CodeCrew is not a Claude Code plugin, and it does
not dispatch agents.
**Trade-off:** No deep integration with any one harness versus operability by
Claude Code, Codex, Gemini CLI, and orchestrators like Paperclip alike, with
GitHub itself as the inter-agent message bus.
**Rejected:** A plugin/skill package for a single harness, and building an
agent dispatcher — dispatching is the operator's or orchestrator's job, and
execution scaffolding is the part of frameworks that ages badly.

---

**Decision:** The CLI is written in Go.
**Trade-off:** Rust's stronger type system and memory guarantees versus
readability for the humans reviewing agent-written code.
**Rejected:** Rust. Both produce single static cross-compiled binaries, so
users are unaffected either way; the CLI is a thin I/O wrapper that never
exercises Rust's strengths. Go aligns with the ecosystem (`gh` is Go, GitHub
ships `go-gh`) and a binary named `gh-codecrew` is distributable via
`gh extension install` for free.

---

**Decision:** v1 wraps the `gh` CLI rather than speaking the REST API
directly.
**Trade-off:** A runtime dependency on `gh` being installed versus
authentication, base URLs, and enterprise quirks solved for free.
**Rejected:** Direct REST for v1 — it buys independence from `gh` at the cost
of owning an auth story, in exactly the multi-harness environments where auth
friction hurts most. The backend interface keeps direct REST possible later.
