package tracker

import "testing"

func TestParseRef(t *testing.T) {
	cases := []struct {
		in   string
		want IssueRef
		ok   bool
	}{
		{"12", IssueRef{"o/def", 12}, true},
		{"#12", IssueRef{"o/def", 12}, true},
		{"radiusred/codecrew#7", IssueRef{"radiusred/codecrew", 7}, true},
		{"nonsense", IssueRef{}, false},
		{"owner/repo", IssueRef{}, false},
	}
	for _, c := range cases {
		got, err := ParseRef(c.in, "o/def")
		if c.ok && (err != nil || got != c.want) {
			t.Errorf("ParseRef(%q) = %v, %v; want %v", c.in, got, err, c.want)
		}
		if !c.ok && err == nil {
			t.Errorf("ParseRef(%q) should fail", c.in)
		}
	}
}

func TestNextMilestoneNumber(t *testing.T) {
	titles := []string{"M1: First", "M3: Skipped ahead", "not a milestone", "M2: Second"}
	if got := NextMilestoneNumber(titles); got != 4 {
		t.Errorf("NextMilestoneNumber = %d, want 4", got)
	}
	if got := NextMilestoneNumber(nil); got != 1 {
		t.Errorf("NextMilestoneNumber(empty) = %d, want 1", got)
	}
}

func TestPlanPresent(t *testing.T) {
	planless := "## Goal\nX\n\n## Plan\n" + PlanPlaceholder + "\n\n## Ask-the-human points\nNone."
	if PlanPresent(planless) {
		t.Error("placeholder plan should not count as present")
	}
	if PlanPresent("## Goal\nX\n\n## Ask-the-human points\nNone.") {
		t.Error("missing Plan section should not count as present")
	}
	planned := "## Goal\nX\n\n## Plan\n- change the thing\n\n## Ask-the-human points\nNone."
	if !PlanPresent(planned) {
		t.Error("real plan should count as present")
	}
}

func TestExtractRecords(t *testing.T) {
	src := IssueRef{"o/r", 4}
	comments := []Comment{
		{Author: "cody", Body: "**Decision:** use X\n**Trade-off:** Y\n**Rejected:** Z"},
		{Author: "human", Body: "just a chat comment mentioning **Deviation:** midway"},
		{Author: "cody", Body: "  **Deviation:** skipped W\n**Why:** unnecessary"},
	}
	records := ExtractRecords(src, comments)
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0].Kind != "Decision" || records[1].Kind != "Deviation" {
		t.Errorf("kinds = %s, %s", records[0].Kind, records[1].Kind)
	}
	if records[0].Source != "o/r#4" {
		t.Errorf("source = %q", records[0].Source)
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
