package main

import "errors"

func behaviorError(message string) error {
	return errors.New(message)
}
