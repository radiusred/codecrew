package cli

import (
	"fmt"
	"io"
)

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
	roles := c.rolesConfig().Roles
	role, ok := roles[args[0]]
	if !ok && len(roles) > 0 {
		return fmt.Errorf("role %q is not in the routing table", args[0])
	}
	if role.Identity == "" {
		fmt.Fprintln(w, "~")
		return nil
	}
	fmt.Fprintln(w, role.Identity)
	return nil
}
