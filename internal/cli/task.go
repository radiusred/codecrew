package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/radiusred/codecrew/internal/tracker"
)

const taskTemplate = `## Goal
%s

## Requirements
%s

## Plan
%s

## Ask-the-human points
_None identified yet — the implementer adds any before starting._
`

func taskNew(w io.Writer, args []string) error {
	fs := flag.NewFlagSet("task new", flag.ContinueOnError)
	milestoneArg := fs.String("milestone", "", "milestone number to link into (required)")
	title := fs.String("title", "", "task title (required)")
	goal := fs.String("goal", "_To be written._", "what this task delivers")
	requirements := fs.String("requirements", "None directly.", "comma-separated requirement IDs")
	repo := fs.String("repo", "", "spoke repo for the task (default: current repo)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *title == "" || *milestoneArg == "" {
		return fmt.Errorf("task new: --title and --milestone are required")
	}
	c, err := load()
	if err != nil {
		return err
	}
	target := *repo
	if target == "" {
		target = c.current
	}
	n, ok := tracker.MilestoneNumber("M" + strings.TrimPrefix(*milestoneArg, "M") + ":")
	if !ok {
		return fmt.Errorf("bad milestone number %q", *milestoneArg)
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

	body := fmt.Sprintf(taskTemplate, *goal, *requirements, tracker.PlanPlaceholder)
	ref, err := c.t.CreateIssue(target, *title, body, []string{"cc:task"})
	if err != nil {
		return err
	}
	if err := c.t.AddSubIssue(milestone.Ref, ref); err != nil {
		return err
	}
	fmt.Fprintf(w, "created task %s as a sub-issue of %s\n", ref, milestone.Ref)
	return nil
}

func taskStart(w io.Writer, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: codecrew task start <ref>")
	}
	c, err := load()
	if err != nil {
		return err
	}
	ref, err := tracker.ParseRef(args[0], c.current)
	if err != nil {
		return err
	}
	task, err := c.t.Task(ref)
	if err != nil {
		return err
	}
	if task.Closed {
		return refuse("CLOSED", "%s is already closed", ref)
	}
	if !tracker.HasLabel(task, "cc:task") {
		return refuse("NOT_A_TASK", "%s is not labeled cc:task", ref)
	}
	body, err := c.t.IssueBody(ref)
	if err != nil {
		return err
	}
	if !tracker.PlanPresent(body) {
		return refuse("NO_PLAN", "%s has no plan — write the Plan section before starting (SPEC §4)", ref)
	}

	viewer, err := c.t.Viewer()
	if err != nil {
		return err
	}
	if err := c.t.Assign(ref, viewer); err != nil {
		// Bot identities are not always assignable; record the start instead.
		fmt.Fprintf(w, "note: could not assign @%s (%v); recording start as a comment\n", viewer, err)
		if err := c.t.Comment(ref, fmt.Sprintf("**Started by** @%s.", viewer)); err != nil {
			return err
		}
	}

	branch := fmt.Sprintf("task/%d-%s", ref.Number, slug(task.Title))
	if err := c.t.DevelopBranch(ref, branch); err != nil {
		fmt.Fprintf(w, "note: could not create linked branch (%v); create %q manually\n", err, branch)
	} else {
		fmt.Fprintf(w, "linked branch %s created\n", branch)
	}
	fmt.Fprintf(w, "started %s as @%s\n", ref, viewer)
	return nil
}

func checkpoint(w io.Writer, args []string) error {
	refArg, args := splitLeadingRef(args)
	fs := flag.NewFlagSet("checkpoint", flag.ContinueOnError)
	question := fs.String("question", "", "the judgment call a human must make (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if refArg == "" && fs.NArg() == 1 {
		refArg = fs.Arg(0)
	}
	if *question == "" || refArg == "" {
		return fmt.Errorf("usage: codecrew checkpoint <ref> --question \"...\"")
	}
	c, err := load()
	if err != nil {
		return err
	}
	ref, err := tracker.ParseRef(refArg, c.current)
	if err != nil {
		return err
	}
	if err := c.t.Comment(ref, "**Gate raised:** "+*question); err != nil {
		return err
	}
	if err := c.t.AddLabel(ref, tracker.LabelNeedsDecision); err != nil {
		return err
	}
	fmt.Fprintf(w, "gate raised on %s — blocked until a human removes %s\n", ref, tracker.LabelNeedsDecision)
	return nil
}

func taskFinish(w io.Writer, args []string) error {
	refArg, args := splitLeadingRef(args)
	fs := flag.NewFlagSet("task finish", flag.ContinueOnError)
	operatorConfirm := fs.Bool("operator-confirm", false,
		"solo tier: record explicit operator confirmation in place of a non-doer approval")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if refArg == "" && fs.NArg() == 1 {
		refArg = fs.Arg(0)
	}
	if refArg == "" {
		return fmt.Errorf("usage: codecrew task finish <ref> [--operator-confirm]")
	}
	c, err := load()
	if err != nil {
		return err
	}
	ref, err := tracker.ParseRef(refArg, c.current)
	if err != nil {
		return err
	}
	task, err := c.t.Task(ref)
	if err != nil {
		return err
	}
	if task.Closed {
		return refuse("CLOSED", "%s is already closed", ref)
	}
	if tracker.HasLabel(task, tracker.LabelNeedsDecision) {
		return refuse("GATED", "%s has %s raised — a human must resolve it and remove the label", ref, tracker.LabelNeedsDecision)
	}

	prs, err := c.t.ClosingPRs(ref, false)
	if err != nil {
		return err
	}
	if len(prs) == 0 {
		return refuse("NO_PR", "no open PR closes %s", ref)
	}
	pr, err := c.t.PRInfo(ref.Repo, prs[0])
	if err != nil {
		return err
	}
	if pr.ChecksPending {
		return refuse("CHECKS_PENDING", "PR #%d checks still running", pr.Number)
	}
	if !pr.ChecksOK {
		return refuse("CHECKS_FAILING", "PR #%d has failing checks", pr.Number)
	}

	approved := false
	for _, login := range pr.ApprovedBy {
		if login != pr.Author {
			approved = true
		}
	}
	if !approved {
		if !*operatorConfirm {
			return refuse("NO_NONDOER_APPROVAL", "PR #%d has no approving review from a non-author (solo tier: rerun with --operator-confirm)", pr.Number)
		}
		viewer, err := c.t.Viewer()
		if err != nil {
			return err
		}
		if viewer == pr.Author {
			return refuse("SELF_CONFIRM", "--operator-confirm requires an identity other than the PR author @%s", pr.Author)
		}
		prRef := tracker.IssueRef{Repo: pr.Repo, Number: pr.Number}
		msg := fmt.Sprintf("**Operator confirmation:** reviewed and accepted by @%s in place of a formal approval (solo tier, SPEC §6).", viewer)
		if err := c.t.Comment(prRef, msg); err != nil {
			return err
		}
	}

	if err := c.t.MergePR(pr.Repo, pr.Number); err != nil {
		return err
	}
	fmt.Fprintf(w, "merged PR #%d; %s closes via its closing keyword\n", pr.Number, ref)
	return nil
}

// splitLeadingRef pulls a leading positional ref off the args so verbs
// accept both "<ref> --flag" and "--flag <ref>" orders (stdlib flag stops
// parsing at the first positional).
func splitLeadingRef(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	return "", args
}

func slug(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteRune('-')
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > 40 {
		s = strings.Trim(s[:40], "-")
	}
	return s
}
