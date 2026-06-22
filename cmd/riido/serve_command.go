package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func runServe(args []string) error {
	socketPath, err := riidoapi.DefaultSocketPath()
	if err != nil {
		return err
	}
	transport := riidoapi.LocalTransportUnixSocket
	taskDBPath, err := taskdb.DefaultTaskDBPath()
	if err != nil {
		return err
	}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--socket":
			index++
			if index >= len(args) {
				return fmt.Errorf("--socket requires a path")
			}
			socketPath = args[index]
		case "--transport":
			index++
			if index >= len(args) {
				return fmt.Errorf("--transport requires a value")
			}
			transport = riidoapi.LocalTransport(args[index])
		case "--task-db":
			index++
			if index >= len(args) {
				return fmt.Errorf("--task-db requires a path")
			}
			taskDBPath = args[index]
		case "--help", "-h":
			printUsage()
			return nil
		default:
			return fmt.Errorf("unknown argument: %s", args[index])
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	fmt.Fprintf(os.Stderr, "riido serve transport=%s socket=%s task_db=%s\n", transport, socketPath, taskDBPath)
	return riidoapi.NewServer(riidoapi.Config{
		AppVersion: versionLabel(),
		SocketPath: socketPath,
		TaskDBPath: taskDBPath,
		Transport:  transport,
	}).Serve(ctx)
}
