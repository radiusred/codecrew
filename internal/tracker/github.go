package tracker

import (
	"fmt"
	"strings"

	"github.com/radiusred/codecrew/internal/gh"
)

// GitHub implements Tracker over the gh CLI.
type GitHub struct{}

func (g GitHub) OpenMilestones(hub string) ([]Milestone, error) {
	milestones, err := listMilestones(hub, "open")
	if err != nil {
		return nil, err
	}
	for i := range milestones {
		tasks, err := g.SubIssues(milestones[i].Ref)
		if err != nil {
			return nil, err
		}
		milestones[i].Tasks = tasks
	}
	return milestones, nil
}

func listMilestones(hub, state string) ([]Milestone, error) {
	var issues []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}
	path := fmt.Sprintf("repos/%s/issues?labels=cc:milestone&state=%s&per_page=100", hub, state)
	if err := gh.JSON(&issues, "api", path); err != nil {
		return nil, err
	}
	milestones := make([]Milestone, 0, len(issues))
	for _, is := range issues {
		milestones = append(milestones, Milestone{
			Ref:   IssueRef{Repo: hub, Number: is.Number},
			Title: is.Title,
		})
	}
	return milestones, nil
}

func (GitHub) AddSubIssue(parent, child IssueRef) error {
	var issue struct {
		ID int64 `json:"id"`
	}
	if err := gh.JSON(&issue, "api", fmt.Sprintf("repos/%s/issues/%d", child.Repo, child.Number)); err != nil {
		return err
	}
	_, err := gh.Run("api", "-X", "POST",
		fmt.Sprintf("repos/%s/issues/%d/sub_issues", parent.Repo, parent.Number),
		"-F", fmt.Sprintf("sub_issue_id=%d", issue.ID))
	return err
}

func (GitHub) SubIssues(parent IssueRef) ([]IssueRef, error) {
	var subs []struct {
		Number        int    `json:"number"`
		RepositoryURL string `json:"repository_url"`
	}
	path := fmt.Sprintf("repos/%s/issues/%d/sub_issues?per_page=100", parent.Repo, parent.Number)
	if err := gh.JSON(&subs, "api", path); err != nil {
		return nil, err
	}
	refs := make([]IssueRef, 0, len(subs))
	for _, s := range subs {
		repo := parent.Repo
		if i := strings.Index(s.RepositoryURL, "/repos/"); i >= 0 {
			repo = s.RepositoryURL[i+len("/repos/"):]
		}
		refs = append(refs, IssueRef{Repo: repo, Number: s.Number})
	}
	return refs, nil
}

func (GitHub) AllMilestoneTitles(hub string) ([]string, error) {
	milestones, err := listMilestones(hub, "all")
	if err != nil {
		return nil, err
	}
	titles := make([]string, len(milestones))
	for i, m := range milestones {
		titles[i] = m.Title
	}
	return titles, nil
}

func (GitHub) IssueBody(ref IssueRef) (string, error) {
	var issue struct {
		Body string `json:"body"`
	}
	err := gh.JSON(&issue, "api", fmt.Sprintf("repos/%s/issues/%d", ref.Repo, ref.Number))
	return issue.Body, err
}

func (GitHub) CreateIssue(repo, title, body string, labels []string) (IssueRef, error) {
	args := []string{"api", "-X", "POST", fmt.Sprintf("repos/%s/issues", repo),
		"-f", "title=" + title, "-f", "body=" + body}
	for _, l := range labels {
		args = append(args, "-f", "labels[]="+l)
	}
	var created struct {
		Number int `json:"number"`
	}
	if err := gh.JSON(&created, args...); err != nil {
		return IssueRef{}, err
	}
	return IssueRef{Repo: repo, Number: created.Number}, nil
}

func (GitHub) Comment(ref IssueRef, body string) error {
	_, err := gh.Run("api", "-X", "POST",
		fmt.Sprintf("repos/%s/issues/%d/comments", ref.Repo, ref.Number), "-f", "body="+body)
	return err
}

func (GitHub) AddLabel(ref IssueRef, label string) error {
	_, err := gh.Run("api", "-X", "POST",
		fmt.Sprintf("repos/%s/issues/%d/labels", ref.Repo, ref.Number), "-f", "labels[]="+label)
	return err
}

func (GitHub) Assign(ref IssueRef, login string) error {
	_, err := gh.Run("api", "-X", "POST",
		fmt.Sprintf("repos/%s/issues/%d/assignees", ref.Repo, ref.Number), "-f", "assignees[]="+login)
	return err
}

// Viewer resolves the current login. Installation tokens cannot call REST
// /user, so GraphQL viewer (which resolves to the bot user) is tried next.
func (GitHub) Viewer() (string, error) {
	var user struct {
		Login string `json:"login"`
	}
	if err := gh.JSON(&user, "api", "user"); err == nil {
		return user.Login, nil
	}
	var resp struct {
		Data struct {
			Viewer struct {
				Login string `json:"login"`
			} `json:"viewer"`
		} `json:"data"`
	}
	if err := gh.JSON(&resp, "api", "graphql", "-f", "query=query { viewer { login } }"); err != nil {
		return "", fmt.Errorf("cannot resolve current identity: %w", err)
	}
	return resp.Data.Viewer.Login, nil
}

func (GitHub) DevelopBranch(ref IssueRef, name string) error {
	_, err := gh.Run("issue", "develop", fmt.Sprint(ref.Number),
		"--repo", ref.Repo, "--name", name)
	return err
}

func (GitHub) ClosingPRs(ref IssueRef, includeClosed bool) ([]int, error) {
	owner, repo, ok := strings.Cut(ref.Repo, "/")
	if !ok {
		return nil, fmt.Errorf("bad repo ref %q", ref.Repo)
	}
	var resp struct {
		Data struct {
			Repository struct {
				Issue struct {
					Refs struct {
						Nodes []struct {
							Number int `json:"number"`
						} `json:"nodes"`
					} `json:"closedByPullRequestsReferences"`
				} `json:"issue"`
			} `json:"repository"`
		} `json:"data"`
	}
	query := fmt.Sprintf(`
query($owner: String!, $repo: String!, $num: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $num) {
      closedByPullRequestsReferences(first: 20, includeClosedPrs: %t) {
        nodes { number }
      }
    }
  }
}`, includeClosed)
	err := gh.JSON(&resp, "api", "graphql",
		"-f", "query="+query,
		"-f", "owner="+owner,
		"-f", "repo="+repo,
		"-F", fmt.Sprintf("num=%d", ref.Number))
	if err != nil {
		return nil, err
	}
	var numbers []int
	for _, n := range resp.Data.Repository.Issue.Refs.Nodes {
		numbers = append(numbers, n.Number)
	}
	return numbers, nil
}

func (GitHub) PRInfo(repo string, number int) (PR, error) {
	var view struct {
		State  string `json:"state"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Reviews []struct {
			State  string `json:"state"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"reviews"`
	}
	err := gh.JSON(&view, "pr", "view", fmt.Sprint(number), "--repo", repo,
		"--json", "state,author,reviews")
	if err != nil {
		return PR{}, err
	}
	pr := PR{
		Repo:   repo,
		Number: number,
		Author: strings.TrimPrefix(view.Author.Login, "app/"),
		Open:   view.State == "OPEN",
	}
	latest := map[string]string{}
	for _, r := range view.Reviews {
		if r.State == "APPROVED" || r.State == "CHANGES_REQUESTED" || r.State == "DISMISSED" {
			latest[r.Author.Login] = r.State
		}
	}
	for login, state := range latest {
		if state == "APPROVED" {
			pr.ApprovedBy = append(pr.ApprovedBy, login)
		}
	}
	var checks []struct {
		Bucket string `json:"bucket"`
	}
	if err := gh.JSONLoose(&checks, "pr", "checks", fmt.Sprint(number), "--repo", repo, "--json", "bucket"); err != nil {
		return PR{}, err
	}
	pr.ChecksOK = true
	for _, c := range checks {
		switch c.Bucket {
		case "pass", "skipping":
		case "pending":
			pr.ChecksPending = true
			pr.ChecksOK = false
		default:
			pr.ChecksOK = false
		}
	}
	return pr, nil
}

func (GitHub) MergePR(repo string, number int) error {
	_, err := gh.Run("pr", "merge", fmt.Sprint(number), "--repo", repo, "--rebase")
	return err
}

func (GitHub) CloseIssue(ref IssueRef, comment string) error {
	_, err := gh.Run("issue", "close", fmt.Sprint(ref.Number),
		"--repo", ref.Repo, "--comment", comment)
	return err
}

func (GitHub) Comments(ref IssueRef) ([]Comment, error) {
	var raw []struct {
		Body string `json:"body"`
		URL  string `json:"html_url"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	path := fmt.Sprintf("repos/%s/issues/%d/comments?per_page=100", ref.Repo, ref.Number)
	if err := gh.JSON(&raw, "api", path); err != nil {
		return nil, err
	}
	comments := make([]Comment, len(raw))
	for i, c := range raw {
		comments[i] = Comment{Author: c.User.Login, Body: c.Body, URL: c.URL}
	}
	return comments, nil
}

func (GitHub) HasMilestoneDoc(repo string, n int) (bool, error) {
	var entries []struct {
		Name string `json:"name"`
	}
	err := gh.JSON(&entries, "api", fmt.Sprintf("repos/%s/contents/docs/milestones", repo))
	if err != nil {
		// A missing directory means no docs yet, not a failure.
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			return false, nil
		}
		return false, err
	}
	prefix := fmt.Sprintf("%d-", n)
	for _, e := range entries {
		if strings.HasPrefix(e.Name, prefix) && strings.HasSuffix(e.Name, ".md") {
			return true, nil
		}
	}
	return false, nil
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
