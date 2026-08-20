// Package tracker defines the backend interface, shaped by the workflow
// verbs rather than by any tracker's feature set (SPEC.md §10), and the pure
// protocol logic: task-ref parsing and state inference.
package tracker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IssueRef identifies an issue by repo and number.
type IssueRef struct {
	Repo   string // owner/repo
	Number int
}

func (r IssueRef) String() string {
	return fmt.Sprintf("%s#%d", r.Repo, r.Number)
}

// Milestone is a cc:milestone tracking issue in the hub.
type Milestone struct {
	Ref   IssueRef
	Title string
	Tasks []IssueRef
}

// Task is a cc:task issue in a spoke.
type Task struct {
	Ref          IssueRef
	Title        string
	Closed       bool
	Assignees    []string
	Labels       []string
	OpenLinkedPR bool
}

// State is an inferred task lifecycle state (SPEC.md §4).
type State string

const (
	Ready      State = "ready"
	InProgress State = "in progress"
	Gated      State = "gated"
	InReview   State = "in review"
	Done       State = "done"
)

// LabelNeedsDecision marks a raised human gate.
const LabelNeedsDecision = "cc:needs-decision"

// Tracker is the backend interface. GitHub is the only implementation; the
// seam exists so a future backend stays possible.
type Tracker interface {
	// OpenMilestones returns the open cc:milestone issues in the hub repo.
	OpenMilestones(hub string) ([]Milestone, error)
	// Task fetches one task issue.
	Task(ref IssueRef) (Task, error)
}

var taskRefPattern = regexp.MustCompile(`(?m)^\s*-\s*\[[ xX]\]\s*(?:([\w.-]+/[\w.-]+)#(\d+)|#(\d+))`)

// ParseTaskRefs extracts task references from a milestone issue body's task
// list. Qualified refs (owner/repo#N) and short refs (#N, resolved against
// hubRepo) are both accepted.
func ParseTaskRefs(body, hubRepo string) []IssueRef {
	var refs []IssueRef
	for _, m := range taskRefPattern.FindAllStringSubmatch(body, -1) {
		if m[1] != "" {
			n, _ := strconv.Atoi(m[2])
			refs = append(refs, IssueRef{Repo: m[1], Number: n})
		} else {
			n, _ := strconv.Atoi(m[3])
			refs = append(refs, IssueRef{Repo: hubRepo, Number: n})
		}
	}
	return refs
}

// InferState derives a task's lifecycle state from tracker signals, most
// terminal first: Done > Gated > In review > In progress > Ready.
func InferState(t Task) State {
	switch {
	case t.Closed:
		return Done
	case hasLabel(t, LabelNeedsDecision):
		return Gated
	case t.OpenLinkedPR:
		return InReview
	case len(t.Assignees) > 0:
		return InProgress
	default:
		return Ready
	}
}

func hasLabel(t Task, name string) bool {
	for _, l := range t.Labels {
		if strings.EqualFold(l, name) {
			return true
		}
	}
	return false
}
