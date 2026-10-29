package runtimeactor

import (
	"testing"
	"time"
)

func TestCapabilityRefreshUpdatesModelsAndRetainsOnFailure(t *testing.T) {
	a := &Actor{cfg: Config{Now: time.Now, Models: []RuntimeModel{{ModelID: "old"}}}}
	capability := Capability{Provider: "fake", CapabilityFingerprint: "fingerprint"}
	for _, models := range [][]RuntimeModel{{{ModelID: "new", IsDefault: true}}, nil} {
		refresh := newCapabilityRefresh(1)
		refresh.probes["fake"] = &capabilityProbe{generation: 1, fingerprint: "fingerprint", cancel: func() {}}
		refresh.finish(a, capabilityProbeResult{provider: "fake", generation: 1, fingerprint: "fingerprint", capability: capability, models: models}, []Capability{capability}, map[string]time.Time{}, nil)
		if len(a.cfg.Models) != 1 || a.cfg.Models[0].ModelID != "new" {
			t.Fatalf("models=%+v", a.cfg.Models)
		}
	}
}
