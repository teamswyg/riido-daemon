package main

import (
	"bufio"
	"fmt"
	"os"
)

func printDaemonLogTail(flags daemonLogsFlags) error {
	lines, err := readDaemonLogLines(flags.logFile)
	if err != nil {
		return err
	}
	from := 0
	if len(lines) > flags.lines {
		from = len(lines) - flags.lines
	}
	for _, line := range lines[from:] {
		fmt.Println(line)
	}
	return nil
}

func readDaemonLogLines(logFile string) ([]string, error) {
	f, err := os.Open(logFile)
	if err != nil {
		return nil, daemonWrapf(ErrDaemonIO, "logs.open", err, "open log")
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, daemonWrapf(ErrDaemonIO, "logs.scan", err, "scan log")
	}
	return lines, nil
}
