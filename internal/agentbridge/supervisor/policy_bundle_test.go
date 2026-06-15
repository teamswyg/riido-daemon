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

func TestSupervisorAppliesPolicyBundleToClaudeNativeHooks(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:       "t-claude-hooks",
		Provider: "claude",
		Prompt:   "hello",
		Metadata: map[string]string{
			MetadataWorkspaceID: "ws-1",
		},
	})

	reporter := newReporterProbe()
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	rt := startNamedRuntime(t, fake, "rt-claude", "claude")
	bundle := policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.no-hooks.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {AllowedSurfaces: policy.AllowedSurfaceSet{}},
		},
	}
	actor, err := New(Config{
		DaemonID:            "daemon-1",
		Runtime:             rt,
		Source:              source,
		Reporter:            reporter,
		Workdir:             workdir.NewFSAdapter(t.TempDir()),
		PollEvery:           10 * time.Millisecond,
		HeartbeatEvery:      time.Hour,
		PolicyBundleVersion: bundle.Version,
		PolicyBundle:        bundle,
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

	manifest := readNativeConfigManifest(t, filepath.Join(res.Workdir, workdir.NativeConfigManifestPath))
	if manifest.ProviderKind != "claude" || manifest.HookMode != workdir.NativeConfigHookModeInstructionOnly {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	for _, blocked := range []string{".claude/settings.json", ".riido/hooks/claude-audit-hook.sh"} {
		if containsString(manifest.GeneratedFiles, blocked) {
			t.Fatalf("manifest generated files must not include %q: %+v", blocked, manifest.GeneratedFiles)
		}
		if _, err := os.Stat(filepath.Join(res.Workdir, filepath.FromSlash(blocked))); !os.IsNotExist(err) {
			t.Fatalf("blocked hook artifact %s should be absent, stat err=%v", blocked, err)
		}
	}
}

func TestSupervisorAppliesPolicyBundleToCodexConfigHome(t *testing.T) {
	source := controlplane.NewMemorySource()
	source.Enqueue(bridge.TaskRequest{
		ID:                       "t-codex-home",
		Provider:                 "codex",
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
	rt := startNamedRuntime(t, fake, "rt-codex", "codex")
	bundle := policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.no-codex-home.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {AllowedSurfaces: policy.AllowedSurfaceSet{}},
		},
	}
	actor, err := New(Config{
		DaemonID:            "daemon-1",
		Runtime:             rt,
		Source:              source,
		Reporter:            reporter,
		Workdir:             workdir.NewFSAdapter(t.TempDir()),
		PollEvery:           10 * time.Millisecond,
		HeartbeatEvery:      time.Hour,
		PolicyBundleVersion: bundle.Version,
		PolicyBundle:        bundle,
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
		t.Fatalf("native config home metadata should be omitted when policy denies it: %+v", running.Command())
	}
	manifest := readNativeConfigManifest(t, filepath.Join(res.Workdir, workdir.NativeConfigManifestPath))
	if manifest.ProviderKind != "codex" || manifest.ConfigHomeDir != "" {
		t.Fatalf("native config manifest = %+v", manifest)
	}
	if containsString(manifest.GeneratedFiles, ".codex/config.toml") {
		t.Fatalf("manifest generated files must not include Codex config home artifact: %+v", manifest.GeneratedFiles)
	}
	if _, err := os.Stat(filepath.Join(res.Workdir, ".codex", "config.toml")); !os.IsNotExist(err) {
		t.Fatalf("Codex config home artifact should be absent, stat err=%v", err)
	}
}
