package create

import "testing"

func TestCommandIncludesExpectedSubcommands(t *testing.T) {
	cmd := Command()

	got := map[string]bool{}
	for _, child := range cmd.Children {
		got[child.Name] = true
	}

	for _, name := range []string{"local", "localandgithub", "githubfromlocal"} {
		if !got[name] {
			t.Fatalf("expected subcommand %q to be present", name)
		}
	}

	if got["githubfromcurdir"] {
		t.Fatal("unexpected deprecated subcommand githubfromcurdir")
	}
}
