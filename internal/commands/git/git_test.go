package git

import "testing"

func TestCommandIncludesExpectedSubcommands(t *testing.T) {
	cmd := Command()

	got := map[string]bool{}
	for _, child := range cmd.Children {
		got[child.Name] = true
	}

	for _, name := range []string{"repo", "push"} {
		if !got[name] {
			t.Fatalf("expected subcommand %q to be present", name)
		}
	}
}
