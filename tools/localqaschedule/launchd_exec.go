package main

import (
	"os"
	"os/exec"
	"runtime"
)

var (
	launchdGOOS    = runtime.GOOS
	launchdGetuid  = os.Getuid
	launchdCommand = exec.Command
)
