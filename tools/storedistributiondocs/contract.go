package main

import "path/filepath"

func loadManifest(repo, rel string) (manifest, error) {
	var m manifest
	err := readJSON(repoPath(repo, rel), &m)
	return m, err
}

func loadContract(repo, rel string) (contract, error) {
	var c contract
	path := repoPath(repo, rel)
	if err := readJSON(path, &c); err != nil {
		return contract{}, err
	}
	base := filepath.Dir(path)
	for _, file := range c.ChannelFiles {
		var item channel
		if err := readJSON(filepath.Join(base, file), &item); err != nil {
			return contract{}, err
		}
		c.Channels = append(c.Channels, item)
	}
	return c, nil
}

func channelByID(c contract, id string) channel {
	for _, item := range c.Channels {
		if item.ID == id {
			return item
		}
	}
	return channel{ID: id}
}
