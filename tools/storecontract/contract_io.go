package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadContract(path string) (contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contract{}, fmt.Errorf("read contract %s: %w", path, err)
	}
	var loaded contract
	if err := json.Unmarshal(data, &loaded); err != nil {
		return contract{}, fmt.Errorf("decode contract %s: %w", path, err)
	}
	return loaded, nil
}
