package detectutil

import "runtime"

func pathEnvKey() string {
	if runtime.GOOS == "windows" {
		return "Path"
	}
	return "PATH"
}
