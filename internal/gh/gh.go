// Package gh is a thin exec wrapper over the GitHub CLI. Wrapping gh (rather
// than speaking REST directly) is a founding decision: auth, base URLs, and
// enterprise quirks come for free (docs/founding-decisions.md).
package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// Run executes gh with args and returns stdout.
func Run(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := bytes.TrimSpace(stderr.Bytes())
		if len(msg) == 0 {
			return nil, fmt.Errorf("gh %s: %w", args[0], err)
		}
		return nil, fmt.Errorf("gh %s: %s", args[0], msg)
	}
	return stdout.Bytes(), nil
}

// JSON executes gh and unmarshals its stdout into v.
func JSON(v any, args ...string) error {
	out, err := Run(args...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(out, v); err != nil {
		return fmt.Errorf("gh %s: unexpected output: %w", args[0], err)
	}
	return nil
}

// JSONLoose is JSON for gh commands that exit nonzero while still printing a
// valid result (e.g. `gh pr checks` with failing checks). The exit error is
// ignored when stdout unmarshals cleanly.
func JSONLoose(v any, args ...string) error {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if jsonErr := json.Unmarshal(stdout.Bytes(), v); jsonErr == nil {
		return nil
	}
	if runErr != nil {
		return fmt.Errorf("gh %s: %s", args[0], bytes.TrimSpace(stderr.Bytes()))
	}
	return fmt.Errorf("gh %s: unexpected output", args[0])
}

// CurrentRepo returns the owner/repo of the repository in the working
// directory, per gh's own resolution of the origin remote.
func CurrentRepo() (string, error) {
	var repo struct {
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := JSON(&repo, "repo", "view", "--json", "nameWithOwner"); err != nil {
		return "", err
	}
	return repo.NameWithOwner, nil
}
