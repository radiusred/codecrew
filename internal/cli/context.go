package cli

import (
	"fmt"

	"github.com/radiusred/codecrew/internal/config"
	"github.com/radiusred/codecrew/internal/gh"
	"github.com/radiusred/codecrew/internal/tracker"
)

// ctx is everything a verb needs: the resolved project topology and the
// tracker backend.
type ctx struct {
	cfg     *config.Config
	current string // owner/repo the command runs in
	hub     string // owner/repo of the hub
	t       tracker.Tracker
}

func load() (*ctx, error) {
	cfg, err := config.Load(".")
	if err != nil {
		return nil, err
	}
	current, err := gh.CurrentRepo()
	if err != nil {
		return nil, err
	}
	return &ctx{
		cfg:     cfg,
		current: current,
		hub:     cfg.HubRepo(current),
		t:       tracker.GitHub{},
	}, nil
}

// refusal is a blocked gate: a machine-readable code plus a human detail.
// Verbs exit nonzero with "refused[CODE]: detail" so agents can act on the
// specific unmet condition (SPEC.md §6).
type refusal struct {
	Code   string
	Detail string
}

func (r refusal) Error() string {
	return fmt.Sprintf("refused[%s]: %s", r.Code, r.Detail)
}

func refuse(code, format string, args ...any) error {
	return refusal{Code: code, Detail: fmt.Sprintf(format, args...)}
}
