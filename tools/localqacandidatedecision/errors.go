package main

import "errors"

var errMissingCandidateInput = errors.New("local QA candidate decision requires -candidate-in")
