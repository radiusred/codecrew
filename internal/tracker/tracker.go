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

// Comment is one issue or PR comment.
type Comment struct {
	Author string
	Body   string
	URL    string
}

// PR is the review-surface state task finish gates on.
type PR struct {
	Repo          string
	Number        int
	Author        string
	Open          bool
	ChecksPending bool
	ChecksOK      bool
	ApprovedBy    []string
}

// Tracker is the backend interface, shaped by the workflow verbs. GitHub is
// the only implementation; the seam exists so a future backend stays possible.
type Tracker interface {
	// OpenMilestones returns the open cc:milestone issues in the hub repo.
	OpenMilestones(hub string) ([]Milestone, error)
	// AllMilestoneTitles returns titles of every cc:milestone issue, open or
	// closed, for milestone-number derivation.
	AllMilestoneTitles(hub string) ([]string, error)
	// Task fetches one task issue.
	Task(ref IssueRef) (Task, error)
	// IssueBody fetches an issue's body text.
	IssueBody(ref IssueRef) (string, error)
	// CreateIssue opens an issue and returns its ref.
	CreateIssue(repo, title, body string, labels []string) (IssueRef, error)
	// UpdateBody replaces an issue's body.
	UpdateBody(ref IssueRef, body string) error
	// Comment posts an issue (or PR) comment.
	Comment(ref IssueRef, body string) error
	// AddLabel applies a label.
	AddLabel(ref IssueRef, label string) error
	// Assign assigns a login to an issue.
	Assign(ref IssueRef, login string) error
	// Viewer returns the login the current credentials act as.
	Viewer() (string, error)
	// DevelopBranch creates a branch linked to the issue.
	DevelopBranch(ref IssueRef, name string) error
	// ClosingPRs returns numbers of PRs that will close (or closed) the
	// issue, in the issue's own repo.
	ClosingPRs(ref IssueRef, includeClosed bool) ([]int, error)
	// PRInfo fetches the gate-relevant state of one PR.
	PRInfo(repo string, number int) (PR, error)
	// MergePR rebase-merges a PR.
	MergePR(repo string, number int) error
	// CloseIssue closes an issue with a closing comment.
	CloseIssue(ref IssueRef, comment string) error
	// Comments lists issue (or PR) comments.
	Comments(ref IssueRef) ([]Comment, error)
	// HasMilestoneDoc reports whether docs/milestones/<n>-*.md exists on the
	// default branch of repo.
	HasMilestoneDoc(repo string, n int) (bool, error)
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

// HasLabel reports whether the task carries the label.
func HasLabel(t Task, name string) bool { return hasLabel(t, name) }

var refPattern = regexp.MustCompile(`^(?:([\w.-]+/[\w.-]+))?#?(\d+)$`)

// ParseRef parses "12", "#12", or "owner/repo#12"; bare and short forms
// resolve against defaultRepo.
func ParseRef(s, defaultRepo string) (IssueRef, error) {
	m := refPattern.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return IssueRef{}, fmt.Errorf("bad issue ref %q (want N, #N, or owner/repo#N)", s)
	}
	repo := m[1]
	if repo == "" {
		repo = defaultRepo
	}
	n, _ := strconv.Atoi(m[2])
	return IssueRef{Repo: repo, Number: n}, nil
}

var milestoneTitle = regexp.MustCompile(`^M(\d+)\s*:`)

// MilestoneNumber extracts n from a milestone title of the form "M<n>: ...".
func MilestoneNumber(title string) (int, bool) {
	m := milestoneTitle.FindStringSubmatch(strings.TrimSpace(title))
	if m == nil {
		return 0, false
	}
	n, _ := strconv.Atoi(m[1])
	return n, true
}

// NextMilestoneNumber derives the next milestone number from existing
// milestone issue titles.
func NextMilestoneNumber(titles []string) int {
	max := 0
	for _, t := range titles {
		if n, ok := MilestoneNumber(t); ok && n > max {
			max = n
		}
	}
	return max + 1
}

// PlanPlaceholder is the Plan section content task new writes; task start
// refuses while it is still in place.
const PlanPlaceholder = "_To be written by the implementer before the first commit._"

// PlanPresent reports whether the task body's Plan section has real content.
func PlanPresent(body string) bool {
	content := section(body, "## Plan")
	content = strings.ReplaceAll(content, PlanPlaceholder, "")
	return strings.TrimSpace(content) != ""
}

// section returns the text between the given heading and the next "## ".
func section(body, heading string) string {
	_, rest, found := strings.Cut(body, heading)
	if !found {
		return ""
	}
	if i := strings.Index(rest, "\n## "); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// AppendTask inserts a task-list entry into the milestone body's Tasks
// section, preserving surrounding content.
func AppendTask(body string, ref IssueRef, title string) string {
	entry := fmt.Sprintf("- [ ] %s — %s", ref, title)
	_, rest, found := strings.Cut(body, "## Tasks")
	if !found {
		return strings.TrimRight(body, "\n") + "\n\n## Tasks\n" + entry + "\n"
	}
	sectionEnd := len(rest)
	if i := strings.Index(rest, "\n## "); i >= 0 {
		sectionEnd = i
	}
	head := body[:len(body)-len(rest)]
	sec := strings.TrimRight(rest[:sectionEnd], "\n")
	tail := rest[sectionEnd:]
	return head + sec + "\n" + entry + "\n" + tail
}

// Record is one Decision or Deviation captured in a comment.
type Record struct {
	Kind   string // "Decision" or "Deviation"
	Source string // the issue/PR ref the comment was found on
	Author string
	Body   string
	URL    string
}

// ExtractRecords finds Decision/Deviation comments per the SPEC §4
// convention (body starts with **Decision:** or **Deviation:**).
func ExtractRecords(source IssueRef, comments []Comment) []Record {
	var records []Record
	for _, c := range comments {
		trimmed := strings.TrimSpace(c.Body)
		var kind string
		switch {
		case strings.HasPrefix(trimmed, "**Decision:**"):
			kind = "Decision"
		case strings.HasPrefix(trimmed, "**Deviation:**"):
			kind = "Deviation"
		default:
			continue
		}
		records = append(records, Record{
			Kind:   kind,
			Source: source.String(),
			Author: c.Author,
			Body:   trimmed,
			URL:    c.URL,
		})
	}
	return records
}
