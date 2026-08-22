package cli

import (
	"testing"

	"github.com/radiusred/gh-codecrew/internal/config"
)

func TestHolder(t *testing.T) {
	table := map[string]config.Role{
		"implementer": {Identity: "radiusred-cody"},
		"reviewer":    {Identity: "davison"},
		"qa":          {}, // identity: ~ — operator-held
	}
	cases := []struct {
		name    string
		roles   map[string]config.Role
		role    string
		want    string
		wantErr bool
	}{
		{"routed app", table, "implementer", "radiusred-cody", false},
		{"routed human", table, "reviewer", "davison", false},
		{"explicitly operator-held", table, "qa", "~", false},
		{"absent from a declared table", table, "doc-synthesizer", "", true},
		{"no table declared: any role is operator-held", nil, "reviewer", "~", false},
	}
	for _, c := range cases {
		got, err := holder(c.roles, c.role)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", c.name, err, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("%s: holder = %q, want %q", c.name, got, c.want)
		}
	}
}
