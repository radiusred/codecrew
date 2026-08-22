package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/radiusred/gh-codecrew/internal/tracker"
)

// Tasks have no body section: they are attached as GitHub sub-issues, which
// the milestone issue tracks natively.
const milestoneTemplate = `## Goal
%s

## Requirements
_Add requirements as **M%d-R1** — ... lines._

## Gates
_What "done" means beyond CI: e2e suites, manual UAT, sign-offs._
`

func milestoneNew(w io.Writer, args []string) error {
	fs := flag.NewFlagSet("milestone new", flag.ContinueOnError)
	title := fs.String("title", "", "milestone title (required)")
	goal := fs.String("goal", "_To be written._", "one-paragraph goal")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *title == "" {
		return fmt.Errorf("milestone new: --title is required")
	}
	c, err := load()
	if err != nil {
		return err
	}
	titles, err := c.t.AllMilestoneTitles(c.hub)
	if err != nil {
		return err
	}
	n := tracker.NextMilestoneNumber(titles)
	fullTitle := fmt.Sprintf("M%d: %s", n, *title)
	body := fmt.Sprintf(milestoneTemplate, *goal, n)
	ref, err := c.t.CreateIssue(c.hub, fullTitle, body, []string{"cc:milestone"})
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "created milestone %s (%s)\n", fullTitle, ref)

	row := fmt.Sprintf("| M%d | %s | [#%d](https://github.com/%s/issues/%d) | Open |",
		n, *title, ref.Number, ref.Repo, ref.Number)
	if c.current == c.hub {
		if appendRoadmapRow(filepath.Join(c.cfg.Dir, "ROADMAP.md"), row) == nil {
			fmt.Fprintln(w, "ROADMAP.md updated locally — commit it via your next PR")
			return nil
		}
	}
	fmt.Fprintf(w, "add to the hub's ROADMAP.md:\n%s\n", row)
	return nil
}

// appendRoadmapRow adds the row after the roadmap table's last row.
func appendRoadmapRow(path, row string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	last := -1
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "|") {
			last = i
		}
	}
	if last < 0 {
		return fmt.Errorf("no table in %s", path)
	}
	lines = append(lines[:last+1], append([]string{row}, lines[last+1:]...)...)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func milestoneClose(w io.Writer, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: codecrew milestone close <n>")
	}
	n, err := strconv.Atoi(strings.TrimPrefix(args[0], "M"))
	if err != nil {
		return fmt.Errorf("bad milestone number %q", args[0])
	}
	c, err := load()
	if err != nil {
		return err
	}
	milestones, err := c.t.OpenMilestones(c.hub)
	if err != nil {
		return err
	}
	var milestone *tracker.Milestone
	for i := range milestones {
		if got, ok := tracker.MilestoneNumber(milestones[i].Title); ok && got == n {
			milestone = &milestones[i]
			break
		}
	}
	if milestone == nil {
		return refuse("NOT_FOUND", "no open milestone M%d in %s", n, c.hub)
	}

	// Gate 1: every task closed.
	var open []string
	for _, ref := range milestone.Tasks {
		task, err := c.t.Task(ref)
		if err != nil {
			return err
		}
		if !task.Closed {
			open = append(open, fmt.Sprintf("%s (%s)", ref, tracker.InferState(task)))
		}
	}
	if len(open) > 0 {
		return refuse("OPEN_TASKS", "tasks not closed: %s", strings.Join(open, ", "))
	}

	// Gate 2: every requirement carries a satisfied QA verdict (a later
	// verdict supersedes an earlier one; only the qa role's holder counts —
	// its routed identity, or the human operator when the role is unrouted).
	body, err := c.t.IssueBody(milestone.Ref)
	if err != nil {
		return err
	}
	comments, err := c.t.Comments(milestone.Ref)
	if err != nil {
		return err
	}
	latest := map[string]string{}
	for _, v := range tracker.ParseVerdicts(comments) {
		if c.holdsRole(v.Author, "qa") {
			latest[v.ID] = v.State
		}
	}
	var missing, unsatisfied []string
	for _, id := range tracker.RequirementIDs(body) {
		state, ok := latest[id]
		switch {
		case !ok:
			missing = append(missing, id)
		case state != "satisfied":
			unsatisfied = append(unsatisfied, fmt.Sprintf("%s (%s)", id, state))
		}
	}
	if len(missing) > 0 {
		return refuse("VERDICT_MISSING", "no QA verdict on %s for: %s — dispatch QA (roles/qa.md)", milestone.Ref, strings.Join(missing, ", "))
	}
	if len(unsatisfied) > 0 {
		return refuse("VERDICT_UNSATISFIED", "latest QA verdict not satisfied for: %s — remedy and re-dispatch QA", strings.Join(unsatisfied, ", "))
	}
	if len(latest) > 0 && c.rolesConfig().Roles["qa"].Identity == "" {
		fmt.Fprintln(w, "note: qa is unrouted — verdicts counted from the human operator holding the role; declare role routing in the hub's .codecrew.yml at onboarding (SPEC §5)")
	}

	// Gather Decision/Deviation raw material for the doc-synthesizer.
	records, err := gatherRecords(c, milestone)
	if err != nil {
		return err
	}

	// Gate 2: the synthesized milestone document is merged.
	hasDoc, err := c.t.HasMilestoneDoc(c.hub, n)
	if err != nil {
		return err
	}
	if !hasDoc {
		fmt.Fprintf(w, "raw material for docs/milestones/%d-<slug>.md (%d records):\n\n", n, len(records))
		writeRecords(w, records)
		return refuse("DOC_MISSING", "docs/milestones/%d-*.md not on the default branch of %s — dispatch the doc-synthesizer, merge its PR, rerun", n, c.hub)
	}

	comment := fmt.Sprintf("Closed by `codecrew milestone close %d`: all %d tasks done, milestone document merged.", n, len(milestone.Tasks))
	if err := c.t.CloseIssue(milestone.Ref, comment); err != nil {
		return err
	}
	fmt.Fprintf(w, "closed %s (%s)\n", milestone.Title, milestone.Ref)
	return nil
}

// gatherRecords collects Decision/Deviation comments from every task issue
// in the milestone and each task's PRs (open and closed).
func gatherRecords(c *ctx, m *tracker.Milestone) ([]tracker.Record, error) {
	var records []tracker.Record
	for _, ref := range m.Tasks {
		comments, err := c.t.Comments(ref)
		if err != nil {
			return nil, err
		}
		records = append(records, tracker.ExtractRecords(ref, comments)...)
		prs, err := c.t.ClosingPRs(ref, true)
		if err != nil {
			return nil, err
		}
		for _, num := range prs {
			prRef := tracker.IssueRef{Repo: ref.Repo, Number: num}
			prComments, err := c.t.Comments(prRef)
			if err != nil {
				return nil, err
			}
			records = append(records, tracker.ExtractRecords(prRef, prComments)...)
		}
	}
	return records, nil
}

func writeRecords(w io.Writer, records []tracker.Record) {
	for _, r := range records {
		fmt.Fprintf(w, "---\n%s on %s by @%s (%s)\n\n%s\n\n", r.Kind, r.Source, r.Author, r.URL, r.Body)
	}
}
