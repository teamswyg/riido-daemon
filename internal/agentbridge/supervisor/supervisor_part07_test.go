package supervisor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func TestSupervisorKeepsOpenClawAndCursorNativeConfigInstructionOnly(t *testing.T) {
	for _, provider := range []bridge.Provider{"openclaw", "cursor"} {
		t.Run(string(provider), func(t *testing.T) {
			source := controlplane.NewMemorySource()
			source.Enqueue(bridge.TaskRequest{
				ID:                       "t-" + string(provider) + "-native-config",
				Provider:                 provider,
				Prompt:                   "hello",
				AllowExperimentalRuntime: true,
				Metadata: map[string]string{
					MetadataWorkspaceID: "ws-1",
				},
			})

			reporter := newReporterProbe()
			fake := process.NewFake()
			running := process.NewFakeRunning()
			fake.NextRunning = running
			rt := startNamedRuntime(t, fake, "rt-"+string(provider), string(provider))
			actor, err := New(Config{
				DaemonID:            "daemon-1",
				Runtime:             rt,
				Source:              source,
				Reporter:            reporter,
				Workdir:             workdir.NewFSAdapter(t.TempDir()),
				PollEvery:           10 * time.Millisecond,
				HeartbeatEvery:      time.Hour,
				PolicyBundleVersion: policy.DefaultLocalPolicyBundleVersion,
				PolicyBundle:        policy.DefaultLocalPolicyBundle(),
				RuntimeTrustTier:    policy.TrustTierHost,
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := actor.Start(context.Background()); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = actor.Stop(ctx)
			})

			select {
			case <-reporter.started:
			case <-time.After(2 * time.Second):
				t.Fatal("task was not claimed")
			}

			select {
			case <-running.StartedRecv():
			case <-time.After(2 * time.Second):
				t.Fatal("provider process was not spawned")
			}

			go func() {
				running.EmitStdout([]byte("ok"))
				running.EmitExit(0, nil)
			}()
			var res agentbridge.Result
			select {
			case res = <-reporter.results:
				if res.Status != agentbridge.ResultCompleted {
					t.Fatalf("result: %+v", res)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("result was not reported")
			}

			if hasEnvPrefix(running.Command().Env, "TEST_NATIVE_CONFIG_HOME=") {
				t.Fatalf("native config home metadata must be omitted for %s: %+v", provider, running.Command())
			}
			manifest := readNativeConfigManifest(t, filepath.Join(res.Workdir, workdir.NativeConfigManifestPath))
			if manifest.ProviderKind != string(provider) ||
				manifest.PrimaryInstructionFile != "AGENTS.md" ||
				manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly ||
				manifest.ConfigHomeDir != "" ||
				len(manifest.ProviderSettingsFiles) != 0 ||
				len(manifest.HookFiles) != 0 {
				t.Fatalf("native config manifest = %+v", manifest)
			}
			if len(manifest.GeneratedFiles) != 2 ||
				!containsString(manifest.GeneratedFiles, "AGENTS.md") ||
				!containsString(manifest.GeneratedFiles, workdir.NativeConfigManifestPath) {
				t.Fatalf("generated files = %+v", manifest.GeneratedFiles)
			}
			for _, blocked := range []string{
				".cursor/settings.json",
				".cursor/rules",
				".openclaw/settings.json",
				".openclaw/config.json",
			} {
				if _, err := os.Stat(filepath.Join(res.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
					t.Fatalf("provider-native artifact %s should be absent, stat err=%v", blocked, err)
				}
			}
		})
	}
}
