#!/usr/bin/env python3
import json
import sys

manifest_path = "docs/30-architecture/runtime-upgrade-flow.riido.json"
with open(manifest_path, encoding="utf-8") as f:
    manifest = json.load(f)

matches = [
    row
    for row in manifest.get("native_config", [])
    if "Q-WS-006" in row.get("decision_refs", [])
    and row.get("name") == "dirty workdir reinjection threshold is zero"
    and row.get("status") in {"implemented", "reserved"}
]

if not matches:
    print("Q-WS-006 must live in runtime-upgrade-flow.riido.json", file=sys.stderr)
    sys.exit(1)
