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
