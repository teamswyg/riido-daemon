package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// contextIO carries every side effect of the work-context commands so tests
// can run them without a real daemon, terminal, or PATH.
type contextIO struct {
	stdout        io.Writer
	stderr        io.Writer
	getenv        func(string) string
	dial          func(path string, timeout time.Duration) (net.Conn, error)
	defaultSocket func() (string, error)
	executable    func() (string, error)
	lookPath      func(string) (string, error)
	cliVersion    func() string
}

func defaultContextIO() contextIO {
	return contextIO{
		stdout:        os.Stdout,
		stderr:        os.Stderr,
		getenv:        os.Getenv,
		dial:          dialDaemonSocket,
		defaultSocket: defaultAgentDaemonSocket,
		executable:    os.Executable,
		lookPath:      exec.LookPath,
		cliVersion:    versionLabel,
	}
}

func (c contextIO) environment() contextcli.Environment {
	return contextcli.DetectEnvironment(c.getenv)
}

func (c contextIO) daemonSocket(env contextcli.Environment) (string, error) {
	if env.DaemonSocket != "" {
		return env.DaemonSocket, nil
	}
	return c.defaultSocket()
}

// reportedExit is returned after a work-context command already wrote its
// result or error; main exits with the code without printing again.
type reportedExit struct {
	code contextcli.ExitCode
}

func (e reportedExit) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func (e reportedExit) ExitCode() int {
	return int(e.code)
}
