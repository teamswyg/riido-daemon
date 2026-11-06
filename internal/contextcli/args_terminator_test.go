package contextcli

import (
	"reflect"
	"testing"
)

func TestParseArgsOptionTerminator(t *testing.T) {
	t.Run("dash ID after flags", func(t *testing.T) {
		args, err := ParseArgs([]string{"--json", "--team-id", "team-one", "--", "-meeting-fixture"}, ArgSpec{Values: []string{"--team-id"}, MaxPositionals: 1})
		if err != nil || !args.JSON || args.Value("--team-id") != "team-one" || !reflect.DeepEqual(args.Positionals, []string{"-meeting-fixture"}) {
			t.Fatalf("terminated positional rejected or changed: %+v %v", args, err)
		}
	})
	t.Run("remaining options are literal", func(t *testing.T) {
		args, err := ParseArgs([]string{"--", "--help", "--json", "--profile=staging", "--"}, ArgSpec{MaxPositionals: 4})
		if err != nil || args.Help || args.JSON || args.HasValue("--profile") || !reflect.DeepEqual(args.Positionals, []string{"--help", "--json", "--profile=staging", "--"}) {
			t.Fatalf("literal arguments changed options: %+v %v", args, err)
		}
	})
	for _, raw := range [][]string{{"--", "one", "two"}, {"--unknown", "--", "one"}} {
		if _, err := ParseArgs(raw, ArgSpec{MaxPositionals: 1}); err == nil || err.Code != CodeUsage {
			t.Fatal("terminator weakened argument validation", err)
		}
	}
	for _, raw := range [][]string{{"--", "--token=synthetic"}, {"--", "--api-key", "synthetic"}} {
		if _, err := ParseArgs(raw, ArgSpec{MaxPositionals: 2}); err == nil || err.Code != CodeSecretInArgv {
			t.Fatal("terminator bypassed secret preflight", err)
		}
	}
}
