package main

import (
	"fmt"
	"net/url"
)

func taskEndpoint(base, taskID, suffix string) string {
	return fmt.Sprintf("%s/tasks/%s%s", base, url.PathEscape(taskID), suffix)
}
