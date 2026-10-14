package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadContract(path string) (contract, error) {
	var loaded contract
	if err := decodeJSONFile(path, &loaded); err != nil {
		return contract{}, err
	}
	for _, channelFile := range loaded.ChannelFiles {
		var channel channel
		shardPath := filepath.Join(filepath.Dir(path), channelFile)
		if err := decodeJSONFile(shardPath, &channel); err != nil {
			return contract{}, err
		}
		loaded.Channels = append(loaded.Channels, channel)
	}
	return loaded, nil
}

func decodeJSONFile(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read contract %s: %w", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("decode contract %s: %w", path, err)
	}
	return nil
}
