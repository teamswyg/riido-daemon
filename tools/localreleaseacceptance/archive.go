package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
)

const daemonScript = `#!/bin/sh
case "${1:-}" in
  version|--version) echo "riido version v-local-qa" ;;
  *) echo "riido local qa binary" ;;
esac
`

func writeDaemonArchive(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	gz := gzip.NewWriter(file)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	if err := tw.WriteHeader(&tar.Header{Name: "riido", Mode: 0o755, Size: int64(len(daemonScript))}); err != nil {
		return err
	}
	_, err = tw.Write([]byte(daemonScript))
	return err
}
