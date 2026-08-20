// Package cli dispatches the codecrew workflow verbs (SPEC.md §6).
package cli

import (
	"fmt"
	"os"
)

const usage = `usage: codecrew <verb>

verbs:
  status                                     where the project is
  milestone new --title T [--goal G]         create a milestone tracking issue
  milestone close <n>                        close a milestone (gates: tasks closed, doc merged)
  task new --milestone N --title T           create a task issue, linked into the milestone
           [--repo owner/repo] [--goal G] [--requirements IDs]
  task start <ref>                           assign, verify plan, create linked branch
  task finish <ref> [--operator-confirm]     enforce gates, then rebase-merge
  checkpoint <ref> --question "..."          raise a human gate (cc:needs-decision)

Blocked gates exit nonzero with "refused[CODE]: detail".
`

// Run executes one verb.
func Run(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("no verb given")
	}
	verb, rest := args[0], args[1:]
	switch verb {
	case "status":
		return status(os.Stdout)
	case "milestone", "task":
		if len(rest) == 0 {
			fmt.Fprint(os.Stderr, usage)
			return fmt.Errorf("%s: missing subcommand", verb)
		}
		switch verb + " " + rest[0] {
		case "milestone new":
			return milestoneNew(os.Stdout, rest[1:])
		case "milestone close":
			return milestoneClose(os.Stdout, rest[1:])
		case "task new":
			return taskNew(os.Stdout, rest[1:])
		case "task start":
			return taskStart(os.Stdout, rest[1:])
		case "task finish":
			return taskFinish(os.Stdout, rest[1:])
		default:
			fmt.Fprint(os.Stderr, usage)
			return fmt.Errorf("unknown subcommand %q", verb+" "+rest[0])
		}
	case "checkpoint":
		return checkpoint(os.Stdout, rest)
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown verb %q", verb)
	}
}
