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
		{Author: "human", Body: "**Gate resolved:** rename the repo\n**Trade-off:** redirects vs a second repo"},
	}
	records := ExtractRecords(src, comments)
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}
	if records[0].Kind != "Decision" || records[1].Kind != "Deviation" {
		t.Errorf("kinds = %s, %s", records[0].Kind, records[1].Kind)
	}
	if records[2].Kind != "Decision" {
		t.Errorf("gate resolution kind = %s, want Decision", records[2].Kind)
	}
	if records[0].Source != "o/r#4" {
		t.Errorf("source = %q", records[0].Source)
	}
}

func TestUnresolvedGates(t *testing.T) {
	gate := Comment{Author: "cody", Body: "**Gate raised:** rename or split?"}
	resolved := Comment{Author: "human", Body: "**Gate resolved:** rename"}
	decision := Comment{Author: "human", Body: "**Decision:** rename"}
	chat := Comment{Author: "human", Body: "thinking about it"}
	cases := []struct {
		name     string
		comments []Comment
		want     int
	}{
		{"no gates", []Comment{chat, decision}, 0},
		{"gate then resolution", []Comment{gate, resolved}, 0},
		{"gate then decision counts", []Comment{gate, decision}, 0},
		{"gate with only chat after", []Comment{gate, chat}, 1},
		{"resolution before gate does not count", []Comment{resolved, gate}, 1},
		{"second gate after resolution is unresolved", []Comment{gate, resolved, gate}, 1},
		{"one trailing resolution covers earlier gates", []Comment{gate, gate, resolved}, 0},
	}
	for _, c := range cases {
		if got := len(UnresolvedGates(c.comments)); got != c.want {
			t.Errorf("%s: got %d unresolved, want %d", c.name, got, c.want)
		}
	}
}

func TestRequirementIDs(t *testing.T) {
	body := "## Goal\nX\n\n## Requirements\n" +
		"- **M3-R1** — gate resolutions are gathered\n" +
		"- **M3-R2** — task start is role-aware\n" +
		"- **M3-R1** — duplicated ID is deduped\n\n" +
		"## Gates\n- **M3-R9** — outside the Requirements section, ignored\n"
	got := RequirementIDs(body)
	want := []string{"M3-R1", "M3-R2"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("RequirementIDs = %v, want %v", got, want)
	}
	if ids := RequirementIDs("## Goal\nno requirements section"); len(ids) != 0 {
		t.Errorf("missing section should yield none, got %v", ids)
	}
}

func TestParseVerdicts(t *testing.T) {
	comments := []Comment{
		// M2's actual shape: bullet list, em-dash inside the bold, trailing period.
		{Author: "testy", Body: "## QA verdict\n\n- **M2-R1 — satisfied.** built and ran it\n- **M2-R2 — not satisfied.** stale claim\n"},
		{Author: "testy", Body: "**M2-R2 — satisfied.** This verdict supersedes my earlier one."},
		{Author: "cody", Body: "**M2-R4 — satisfied.** (not QA — caller filters by author)"},
		{Author: "human", Body: "- **M2-R3** — a requirement definition line must not parse as a verdict"},
	}
	got := ParseVerdicts(comments)
	want := []Verdict{
		{ID: "M2-R1", State: "satisfied", Author: "testy"},
		{ID: "M2-R2", State: "not satisfied", Author: "testy"},
		{ID: "M2-R2", State: "satisfied", Author: "testy"},
		{ID: "M2-R4", State: "satisfied", Author: "cody"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d verdicts %v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("verdict[%d] = %v, want %v", i, got[i], want[i])
		}
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
