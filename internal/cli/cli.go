// Package cli dispatches the codecrew workflow verbs (SPEC.md §6).
package cli

import (
	"fmt"
	"os"
)

const usage = `usage: codecrew <verb>

verbs:
  status    where the project is: open milestones, task states, raised gates
`

// Run executes one verb.
func Run(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("no verb given")
	}
	switch args[0] {
	case "status":
		return status(os.Stdout)
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown verb %q", args[0])
	}
}
