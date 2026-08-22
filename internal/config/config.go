// Package config locates and parses the .codecrew.yml pointer file that
// every repo in a CodeCrew project carries (SPEC.md §3, §5).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Role is one entry of the advisory role routing table.
type Role struct {
	Harness  string `yaml:"harness"`
	Model    string `yaml:"model"`
	Identity string `yaml:"identity"`
}

// Config is the parsed .codecrew.yml.
type Config struct {
	Codecrew string          `yaml:"codecrew"`
	Hub      string          `yaml:"hub"`
	Roles    map[string]Role `yaml:"roles"`

	// Dir is the directory the pointer file was found in.
	Dir string `yaml:"-"`
}

// Load walks upward from dir until it finds a .codecrew.yml and parses it.
func Load(dir string) (*Config, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	for {
		path := filepath.Join(dir, ".codecrew.yml")
		if data, err := os.ReadFile(path); err == nil {
			cfg, err := Parse(data)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			cfg.Dir = dir
			return cfg, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("no .codecrew.yml found (not a CodeCrew repo?)")
		}
		dir = parent
	}
}

// Parse decodes pointer-file content.
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Hub == "" {
		return nil, fmt.Errorf("missing required field: hub")
	}
	return &cfg, nil
}

// RoleFor resolves a viewer login to the role name whose identity it acts
// as, normalising GitHub's "[bot]" suffix (the viewer of an App token is
// "<slug>[bot]" while the routing table names the App slug). Unmatched or
// empty logins resolve to "" — an empty identity (unrouted role) never
// matches.
func (c *Config) RoleFor(login string) string {
	login = strings.TrimSuffix(login, "[bot]")
	if login == "" {
		return ""
	}
	for name, role := range c.Roles {
		if role.Identity == login {
			return name
		}
	}
	return ""
}

// HoldsRole reports whether login holds the named role. A routed role is
// held exactly by its identity — a GitHub App slug or a human's username.
// An unrouted role (no identity) is held by the human operator: any human
// login not routed to a different role. Solo is a routing configuration,
// not a degraded tier — every role is always staffed (SPEC §5, decided at
// the gate on gh-codecrew#42).
func (c *Config) HoldsRole(login, role string) bool {
	bare := strings.TrimSuffix(login, "[bot]")
	if bare == "" {
		return false
	}
	if id := c.Roles[role].Identity; id != "" {
		return bare == id
	}
	return bare == login && c.RoleFor(login) == ""
}

// HubRepo resolves the hub to an owner/repo string. current is the
// owner/repo of the repository the command runs in, used when hub is "self".
func (c *Config) HubRepo(current string) string {
	if c.Hub == "self" {
		return current
	}
	return c.Hub
}
