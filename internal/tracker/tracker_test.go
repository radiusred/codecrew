package tracker

import (
	"reflect"
	"testing"
)

func TestParseTaskRefs(t *testing.T) {
	body := `## Goal
Something.

## Tasks
- [ ] radiusred/codecrew#3 — lint workflow
- [x] radiusred/other-repo#12 — done thing
- [ ] #7 — short ref resolves to hub
- not a task item #99
* [ ] #100 — wrong bullet marker is still not matched by spec format

## Gates
- [ ] this gate line has no issue ref
`
	got := ParseTaskRefs(body, "radiusred/hub")
	want := []IssueRef{
		{Repo: "radiusred/codecrew", Number: 3},
		{Repo: "radiusred/other-repo", Number: 12},
		{Repo: "radiusred/hub", Number: 7},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseTaskRefs = %v, want %v", got, want)
	}
}

func TestParseTaskRefsEmpty(t *testing.T) {
	if refs := ParseTaskRefs("## Tasks\n\n_none yet_\n", "o/r"); len(refs) != 0 {
		t.Errorf("expected no refs, got %v", refs)
	}
}

func TestInferState(t *testing.T) {
	cases := []struct {
		name string
		task Task
		want State
	}{
		{"open unassigned is ready", Task{}, Ready},
		{"assigned is in progress", Task{Assignees: []string{"cody"}}, InProgress},
		{"open PR is in review", Task{Assignees: []string{"cody"}, OpenLinkedPR: true}, InReview},
		{"needs-decision gates even in review", Task{Labels: []string{"cc:task", "cc:needs-decision"}, OpenLinkedPR: true}, Gated},
		{"closed is done regardless", Task{Closed: true, Labels: []string{"cc:needs-decision"}, OpenLinkedPR: true}, Done},
	}
	for _, c := range cases {
		if got := InferState(c.task); got != c.want {
			t.Errorf("%s: InferState = %q, want %q", c.name, got, c.want)
		}
	}
}
