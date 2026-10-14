package main

import (
	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
	"github.com/teamswyg/riido-daemon/internal/project"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (o *mwsdOptions) applyDefaults() error {
	if o.socketPath == "" {
		socketPath, err := mwsdbridge.DefaultSocketPath()
		if err != nil {
			return err
		}
		o.socketPath = socketPath
	}
	if o.command != mwsdCommandSync {
		return nil
	}
	return o.applySyncDefaults()
}

func (o *mwsdOptions) applySyncDefaults() error {
	if o.statePath == "" {
		statePath, err := project.DefaultStatePath()
		if err != nil {
			return err
		}
		o.statePath = statePath
	}
	if o.taskDBPath == "" {
		taskDBPath, err := taskdb.DefaultTaskDBPath()
		if err != nil {
			return err
		}
		o.taskDBPath = taskDBPath
	}
	return nil
}
