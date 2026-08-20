package tracker

import (
	"fmt"
	"strings"

	"github.com/radiusred/codecrew/internal/gh"
)

// GitHub implements Tracker over the gh CLI.
type GitHub struct{}

func (GitHub) OpenMilestones(hub string) ([]Milestone, error) {
	var issues []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}
	path := fmt.Sprintf("repos/%s/issues?labels=cc:milestone&state=open&per_page=100", hub)
	if err := gh.JSON(&issues, "api", path); err != nil {
		return nil, err
	}
	milestones := make([]Milestone, 0, len(issues))
	for _, is := range issues {
		milestones = append(milestones, Milestone{
			Ref:   IssueRef{Repo: hub, Number: is.Number},
			Title: is.Title,
			Tasks: ParseTaskRefs(is.Body, hub),
		})
	}
	return milestones, nil
}

const taskQuery = `
query($owner: String!, $repo: String!, $num: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $num) {
      title
      state
      assignees(first: 10) { nodes { login } }
      labels(first: 20) { nodes { name } }
      closedByPullRequestsReferences(first: 10, includeClosedPrs: false) {
        nodes { state }
      }
    }
  }
}`

func (GitHub) Task(ref IssueRef) (Task, error) {
	owner, repo, ok := strings.Cut(ref.Repo, "/")
	if !ok {
		return Task{}, fmt.Errorf("bad repo ref %q", ref.Repo)
	}
	var resp struct {
		Data struct {
			Repository struct {
				Issue *struct {
					Title     string `json:"title"`
					State     string `json:"state"`
					Assignees struct {
						Nodes []struct {
							Login string `json:"login"`
						} `json:"nodes"`
					} `json:"assignees"`
					Labels struct {
						Nodes []struct {
							Name string `json:"name"`
						} `json:"nodes"`
					} `json:"labels"`
					ClosedByPullRequestsReferences struct {
						Nodes []struct {
							State string `json:"state"`
						} `json:"nodes"`
					} `json:"closedByPullRequestsReferences"`
				} `json:"issue"`
			} `json:"repository"`
		} `json:"data"`
	}
	err := gh.JSON(&resp, "api", "graphql",
		"-f", "query="+taskQuery,
		"-f", "owner="+owner,
		"-f", "repo="+repo,
		"-F", fmt.Sprintf("num=%d", ref.Number))
	if err != nil {
		return Task{}, err
	}
	issue := resp.Data.Repository.Issue
	if issue == nil {
		return Task{}, fmt.Errorf("%s: issue not found", ref)
	}
	t := Task{
		Ref:    ref,
		Title:  issue.Title,
		Closed: issue.State == "CLOSED",
	}
	for _, a := range issue.Assignees.Nodes {
		t.Assignees = append(t.Assignees, a.Login)
	}
	for _, l := range issue.Labels.Nodes {
		t.Labels = append(t.Labels, l.Name)
	}
	for _, pr := range issue.ClosedByPullRequestsReferences.Nodes {
		if pr.State == "OPEN" {
			t.OpenLinkedPR = true
		}
	}
	return t, nil
}
