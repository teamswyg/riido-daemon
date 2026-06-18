package detectutil

import "os"

func processPATH() string {
	return os.Getenv("PATH")
}
