package main

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestRepositoryContractUsesChannelFiles(t *testing.T) {
	root := filepath.Join("..", "..")
	loaded, err := loadContract(resolvePath(root, "packaging/store/riido_daemon_store_distribution.riido.json"))
	if err != nil {
		t.Fatalf("load repo contract: %v", err)
	}
	for _, want := range []string{
		"riido_daemon_store_distribution/developer-id.riido.json",
		"riido_daemon_store_distribution/mac-app-store.riido.json",
		"riido_daemon_store_distribution/msix-sideload.riido.json",
		"riido_daemon_store_distribution/msix-store.riido.json",
	} {
		if !slices.Contains(loaded.ChannelFiles, want) {
			t.Fatalf("channel_files must include %q: %+v", want, loaded.ChannelFiles)
		}
	}
	if len(loaded.Channels) != len(loaded.ChannelFiles) {
		t.Fatalf("loaded channels=%d channel_files=%d", len(loaded.Channels), len(loaded.ChannelFiles))
	}
}
