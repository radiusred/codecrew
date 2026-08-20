package cli

import (
	"fmt"
	"io"

	"github.com/radiusred/codecrew/internal/config"
	"github.com/radiusred/codecrew/internal/gh"
	"github.com/radiusred/codecrew/internal/tracker"
)

func status(w io.Writer) error {
	cfg, err := config.Load(".")
	if err != nil {
		return err
	}
	current, err := gh.CurrentRepo()
	if err != nil {
		return err
	}
	hub := cfg.HubRepo(current)

	var t tracker.Tracker = tracker.GitHub{}
	milestones, err := t.OpenMilestones(hub)
	if err != nil {
		return err
	}
	if len(milestones) == 0 {
		fmt.Fprintf(w, "no open milestones in %s\n", hub)
		return nil
	}

	var gated []tracker.Task
	for _, m := range milestones {
		fmt.Fprintf(w, "%s (%s)\n", m.Title, m.Ref)
		if len(m.Tasks) == 0 {
			fmt.Fprintln(w, "  no tasks yet")
		}
		for _, ref := range m.Tasks {
			task, err := t.Task(ref)
			if err != nil {
				return err
			}
			state := tracker.InferState(task)
			who := ""
			if state == tracker.InProgress || state == tracker.InReview {
				if len(task.Assignees) > 0 {
					who = " @" + task.Assignees[0]
				}
			}
			fmt.Fprintf(w, "  [%-11s] %-28s %s%s\n", state, ref, task.Title, who)
			if state == tracker.Gated {
				gated = append(gated, task)
			}
		}
		fmt.Fprintln(w)
	}

	if len(gated) == 0 {
		fmt.Fprintln(w, "gates raised: none")
	} else {
		fmt.Fprintln(w, "gates raised:")
		for _, task := range gated {
			fmt.Fprintf(w, "  %s — %s\n", task.Ref, task.Title)
		}
	}
	return nil
}
