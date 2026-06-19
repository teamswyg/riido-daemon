package main

import "os"

func tempDir() (string, error) {
	return os.MkdirTemp("", "riido-exec-path-*")
}
