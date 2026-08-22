package config

import "testing"

func TestParse(t *testing.T) {
	cfg, err := Parse([]byte(`
codecrew: "0.1"
hub: self
roles:
  implementer:
    harness: claude-code
    model: claude-fable-5
    identity: radiusred-cody
  reviewer:
    harness: codex
    identity: ~
`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Hub != "self" {
		t.Errorf("Hub = %q, want self", cfg.Hub)
	}
	if got := cfg.Roles["implementer"].Identity; got != "radiusred-cody" {
		t.Errorf("implementer identity = %q", got)
	}
	if got := cfg.Roles["reviewer"].Identity; got != "" {
		t.Errorf("nil identity should parse as empty, got %q", got)
	}
}

func TestParseMissingHub(t *testing.T) {
	if _, err := Parse([]byte(`codecrew: "0.1"`)); err == nil {
		t.Error("expected error for missing hub")
	}
}

func TestRoleFor(t *testing.T) {
	cfg := &Config{Roles: map[string]Role{
		"implementer": {Identity: "radiusred-cody"},
		"qa":          {Identity: "radiusred-testy"},
		"reviewer":    {}, // identity: ~ — human operator during bootstrap
	}}
	cases := []struct {
		login, want string
	}{
		{"radiusred-cody[bot]", "implementer"},
		{"radiusred-testy[bot]", "qa"},
		{"radiusred-testy", "qa"},
		{"davison", ""},
		{"", ""}, // empty login must not match the reviewer's empty identity
	}
	for _, c := range cases {
		if got := cfg.RoleFor(c.login); got != c.want {
			t.Errorf("RoleFor(%q) = %q, want %q", c.login, got, c.want)
		}
	}
	if got := (&Config{}).RoleFor("radiusred-cody[bot]"); got != "" {
		t.Errorf("no roles configured should resolve to \"\", got %q", got)
	}
}

func TestHoldsRole(t *testing.T) {
	routed := &Config{Roles: map[string]Role{
		"implementer": {Identity: "radiusred-cody"},
		"qa":          {Identity: "radiusred-testy"},
	}}
	human := &Config{Roles: map[string]Role{
		"implementer": {Identity: "radiusred-cody"},
		"qa":          {Identity: "alice"},
	}}
	unrouted := &Config{Roles: map[string]Role{
		"implementer": {Identity: "radiusred-cody"},
		"qa":          {}, // identity: ~ — held by the human operator
	}}
	empty := &Config{}
	cases := []struct {
		name  string
		cfg   *Config
		login string
		want  bool
	}{
		{"routed: its bot holds", routed, "radiusred-testy[bot]", true},
		{"routed: bare slug holds", routed, "radiusred-testy", true},
		{"routed: a human does not", routed, "davison", false},
		{"routed: another crew bot does not", routed, "radiusred-cody[bot]", false},
		{"routed to a human: that human holds", human, "alice", true},
		{"routed to a human: another human does not", human, "bob", false},
		{"unrouted: the operator holds", unrouted, "davison", true},
		{"unrouted: a crew bot does not", unrouted, "radiusred-cody[bot]", false},
		{"unrouted: an unrelated bot does not", unrouted, "somebot[bot]", false},
		{"unrouted: an identity routed elsewhere does not", unrouted, "radiusred-cody", false},
		{"no roles at all: the operator holds", empty, "davison", true},
		{"empty login never holds", unrouted, "", false},
	}
	for _, c := range cases {
		if got := c.cfg.HoldsRole(c.login, "qa"); got != c.want {
			t.Errorf("%s: HoldsRole(%q) = %v, want %v", c.name, c.login, got, c.want)
		}
	}
}

func TestHubRepo(t *testing.T) {
	self := &Config{Hub: "self"}
	if got := self.HubRepo("radiusred/spoke"); got != "radiusred/spoke" {
		t.Errorf("self hub = %q, want current repo", got)
	}
	remote := &Config{Hub: "radiusred/hub"}
	if got := remote.HubRepo("radiusred/spoke"); got != "radiusred/hub" {
		t.Errorf("named hub = %q, want radiusred/hub", got)
	}
}
