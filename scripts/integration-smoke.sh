#!/usr/bin/env bash
# Probe provider CLIs, then run each available opt-in TestIntegration.
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

provider_specs=(
    "claude claude ${RIIDO_CLAUDE_PATH:-}"
    "codex codex ${RIIDO_CODEX_PATH:-}"
    "openclaw openclaw ${RIIDO_OPENCLAW_PATH:-}"
    "cursor cursor-agent ${RIIDO_CURSOR_PATH:-}"
)

resolve_exec() {
    local default_name="$1" override="$2"
    if [ -n "$override" ] && [ -x "$override" ]; then
        echo "$override"
        return 0
    fi
    command -v "$default_name" 2>/dev/null
}

probe_version() {
    "$1" --version 2>/dev/null | head -1 | tr -d '\r' || true
}

printf '%-12s  %-9s  %-46s  %s\n' "provider" "available" "executable" "version"
printf '%-12s  %-9s  %-46s  %s\n' "--------" "---------" "----------" "-------"

available_providers=()
probe() {
    local name="$1" default_exe="$2" env_override="$3" exe="" version=""
    if exe=$(resolve_exec "$default_exe" "$env_override"); then
        version=$(probe_version "$exe")
        printf '%-12s  %-9s  %-46s  %s\n' "$name" "yes" "$exe" "${version:-<no --version output>}"
        available_providers+=("$name")
        return
    fi
    printf '%-12s  %-9s  %-46s  %s\n' "$name" "no" "(not on PATH)" "-"
}

for spec in "${provider_specs[@]}"; do
    read -r name default_exe env_override <<<"$spec"
    probe "$name" "$default_exe" "$env_override"
done
echo

if [ ${#available_providers[@]} -eq 0 ]; then
    echo "no provider CLIs found on PATH; nothing to integration-test."
    echo "this is an operator-environment skip, not a provider PASS."
    exit 0
fi

fail_count=0
for provider in "${available_providers[@]}"; do
    echo "=== integration: $provider ==="
    if AGENTBRIDGE_INTEGRATION=1 go test "./internal/provider/$provider" -race -run TestIntegration -v; then
        echo "=== integration: $provider PASS ==="
    else
        fail_count=$((fail_count + 1))
        echo "=== integration: $provider FAILED ==="
    fi
    echo
done

if [ "$fail_count" -gt 0 ]; then
    echo "integration-smoke: $fail_count provider(s) failed"
    exit 1
fi
echo "integration-smoke: all available providers passed"
