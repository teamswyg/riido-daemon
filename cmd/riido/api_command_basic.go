package main

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func runAPIStatus(args []string, config apiCLIConfig) error {
	config, err := parseAPIConnectionArgs(args, config)
	if err != nil {
		if isCLIHelp(err) {
			return nil
		}
		return err
	}
	var status riidoapi.Status
	if err := requestAPI(config, 5*time.Second, "status", nil, &status); err != nil {
		return err
	}
	return printJSON(status)
}

func runAPITasks(args []string, config apiCLIConfig) error {
	config, err := parseAPIConnectionArgs(args, config)
	if err != nil {
		if isCLIHelp(err) {
			return nil
		}
		return err
	}
	var db taskdb.TaskDB
	if err := requestAPI(config, 5*time.Second, "tasks", nil, &db); err != nil {
		return err
	}
	return printJSON(db)
}
