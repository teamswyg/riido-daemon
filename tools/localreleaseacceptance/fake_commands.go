package main

import "os"

func writeExecutable(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o755)
}

func fakeCurlScript() string {
	return `#!/bin/sh
out=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) shift; out="$1" ;;
    http*) url="$1" ;;
  esac
  shift || true
done
[ -n "$out" ] || exit 2
base="${url##*/}"
cp "$INSTALL_FIXTURE_DIR/$base" "$out"
`
}

func fakeInstallScript() string {
	return `#!/bin/sh
mode="$2"
source="$3"
target="$4"
if [ -e "$target" ]; then
  echo present > "$INSTALL_MARKER"
else
  echo absent > "$INSTALL_MARKER"
fi
cp "$source" "$target"
chmod "$mode" "$target"
`
}

func fakeUnameScript() string {
	return `#!/bin/sh
case "$1" in
  -s) echo Darwin ;;
  -m) echo arm64 ;;
  *) exit 1 ;;
esac
`
}
