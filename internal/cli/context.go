package cli

import (
	"fmt"

	"github.com/radiusred/gh-codecrew/internal/config"
	"github.com/radiusred/gh-codecrew/internal/gh"
	"github.com/radiusred/gh-codecrew/internal/tracker"
)

// ctx is everything a verb needs: the resolved project topology and the
// tracker backend.
type ctx struct {
	cfg     *config.Config
	current string // owner/repo the command runs in
	hub     string // owner/repo of the hub
	t       tracker.Tracker

	roles *config.Config // memoized routing table (local or hub)
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

// rolesConfig returns the config whose routing table governs role
// resolution. Spokes carry only the pointer config (SPEC §5), so when the
// local file has no roles the hub's .codecrew.yml is fetched (memoized); an
// unreadable hub config degrades to the local one rather than failing the
// verb — routing is advisory.
func (c *ctx) rolesConfig() *config.Config {
	if c.roles != nil {
		return c.roles
	}
	c.roles = c.cfg
	if len(c.cfg.Roles) == 0 {
		if data, err := c.t.FileContent(c.hub, ".codecrew.yml"); err == nil {
			if hubCfg, err := config.Parse(data); err == nil {
				c.roles = hubCfg
			}
		}
	}
	return c.roles
}

// roleFor resolves a viewer login to its role name via the routing table.
func (c *ctx) roleFor(login string) string {
	return c.rolesConfig().RoleFor(login)
}

// holdsRole reports whether login holds the named role (config.HoldsRole
// over the governing routing table).
func (c *ctx) holdsRole(login, role string) bool {
	return c.rolesConfig().HoldsRole(login, role)
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
