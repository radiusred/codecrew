package cli

import (
	"fmt"
	"io"

	"github.com/radiusred/gh-codecrew/internal/config"
)

// holder resolves a role name against a routing table: the routed identity,
// "~" when operator-held (explicitly or because no table is declared), or
// an error for a name absent from a declared table.
func holder(roles map[string]config.Role, name string) (string, error) {
	role, ok := roles[name]
	if !ok && len(roles) > 0 {
		return "", fmt.Errorf("role %q is not in the routing table", name)
	}
	if role.Identity == "" {
		return "~", nil
	}
	return role.Identity, nil
}

// roleHolder prints the identity a role routes to — an App slug, a human
// username, or "~" when the role is operator-held. Script-consumable
// (`--reviewer $(codecrew role reviewer)`), and correct from a pointer-only
// spoke because resolution falls back to the hub's routing table.
func roleHolder(w io.Writer, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: codecrew role <name>")
	}
	c, err := load()
	if err != nil {
		return err
	}
	id, err := holder(c.rolesConfig().Roles, args[0])
	if err != nil {
		return err
	}
	fmt.Fprintln(w, id)
	return nil
}
