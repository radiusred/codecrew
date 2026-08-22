package cli

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	codecrew "github.com/radiusred/gh-codecrew"
)

const hubConfigScaffold = `codecrew: "0.1"
hub: self

# Role routing: who holds each role (SPEC §5). Declare all four at
# onboarding. ~ routes the role to you, the human operator; replace it with
# a GitHub App slug or a username to delegate. (A value with a slash is
# reserved for team routing.)
roles:
  implementer: { identity: ~ }
  reviewer: { identity: ~ }
  qa: { identity: ~ }
  doc-synthesizer: { identity: ~ }
`

const roadmapScaffold = `# Roadmap

| Milestone | Goal | Tracking issue | Status |
|-----------|------|----------------|--------|
`

const agentsScaffold = `# Agents

This repository is part of a CodeCrew project — coordination state lives in
GitHub issues and PRs, per the protocol at
https://github.com/radiusred/gh-codecrew (SPEC.md).

- ` + "`.codecrew.yml`" + ` names the hub; the hub's ` + "`roles/`" + ` holds the role
  contracts. Read the contract for the role you were dispatched as before
  doing anything else.
- ` + "`gh codecrew status`" + ` shows where the project is; ` + "`gh codecrew help`" + `
  lists the workflow verbs. Blocked gates refuse with
  ` + "`refused[CODE]: detail`" + ` — act on the code, don't work around it.
- Plans before commits, decisions recorded when made, and the verifier is
  never the doer.
`

// scaffold writes the greenfield files into dir. Hub mode (hub == "self")
// writes the full set; spoke mode writes only the pointer. Existing files
// are never touched — they are reported as skipped.
func scaffold(dir, hub string, contracts fs.FS) (written, skipped []string, err error) {
	files := map[string]string{
		".codecrew.yml": hubConfigScaffold,
	}
	if hub != "self" {
		files[".codecrew.yml"] = fmt.Sprintf("codecrew: \"0.1\"\nhub: %s\n", hub)
	} else {
		files["ROADMAP.md"] = roadmapScaffold
		files["AGENTS.md"] = agentsScaffold
		entries, err := fs.ReadDir(contracts, "roles")
		if err != nil {
			return nil, nil, err
		}
		for _, e := range entries {
			data, err := fs.ReadFile(contracts, "roles/"+e.Name())
			if err != nil {
				return nil, nil, err
			}
			files[filepath.Join("roles", e.Name())] = string(data)
		}
	}
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		if _, err := os.Stat(path); err == nil {
			skipped = append(skipped, rel)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return written, skipped, err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return written, skipped, err
		}
		written = append(written, rel)
	}
	slices.Sort(written)
	slices.Sort(skipped)
	return written, skipped, nil
}

// initCmd scaffolds a new CodeCrew repo: hub mode by default, spoke mode
// with --hub owner/repo (pointer only — contracts and roadmap live in the
// hub).
func initCmd(w io.Writer, args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	hub := flags.String("hub", "self", "hub repo this spoke points at (default: this repo is the hub)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	written, skipped, err := scaffold(".", *hub, codecrew.Roles)
	if err != nil {
		return err
	}
	for _, f := range written {
		fmt.Fprintf(w, "wrote %s\n", f)
	}
	for _, f := range skipped {
		fmt.Fprintf(w, "kept existing %s\n", f)
	}
	if *hub == "self" {
		fmt.Fprintln(w, "\nnext: declare who holds each role in .codecrew.yml (~ = you),")
		fmt.Fprintln(w, "commit, then `codecrew milestone new --title \"...\" --goal \"...\"`")
	} else {
		fmt.Fprintf(w, "\nnext: commit the pointer; tasks for this spoke attach to milestones in %s\n", *hub)
	}
	return nil
}
