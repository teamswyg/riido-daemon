#!/usr/bin/env bash
# Probe local PATH for supported provider CLIs, print the detected matrix, then
# run each available provider's opt-in TestIntegration.

set -u
set -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

resolve_exec() {
    local default_name="$1"
    local override="$2"
    if [ -n "$override" ] && [ -x "$override" ]; then
        echo "$override"
        return 0
    fi
    local found
    if found=$(command -v "$default_name" 2>/dev/null); then
        echo "$found"
        return 0
    fi
    return 1
}

probe_version() {
    local exe="$1"
    "$exe" --version 2>/dev/null | head -1 | tr -d '\r' || true
}

printf '%-12s  %-9s  %-46s  %s\n' "provider" "available" "executable" "version"
printf '%-12s  %-9s  %-46s  %s\n' "--------" "---------" "----------" "-------"

available_providers=()

probe() {
    local name="$1"
    local default_exe="$2"
    local env_override="$3"

    local exe=""
    if exe=$(resolve_exec "$default_exe" "$env_override"); then
        local version
        version=$(probe_version "$exe")
        printf '%-12s  %-9s  %-46s  %s\n' "$name" "yes" "$exe" "${version:-<no --version output>}"
        available_providers+=("$name")
    else
        printf '%-12s  %-9s  %-46s  %s\n' "$name" "no" "(not on PATH)" "-"
    fi
}

probe claude       claude       "${RIIDO_CLAUDE_PATH:-}"
probe codex        codex        "${RIIDO_CODEX_PATH:-}"
probe openclaw     openclaw     "${RIIDO_OPENCLAW_PATH:-}"
probe cursor       cursor-agent "${RIIDO_CURSOR_PATH:-}"

echo

if [ ${#available_providers[@]} -eq 0 ]; then
    echo "no provider CLIs found on PATH; nothing to integration-test."
    echo "this is an operator-environment skip, not a provider PASS."
    exit 0
fi

fail_count=0
for provider in "${available_providers[@]}"; do
    echo "=== integration: $provider ==="
    if ! AGENTBRIDGE_INTEGRATION=1 go test "./internal/provider/$provider" -race -run TestIntegration -v; then
        fail_count=$((fail_count + 1))
        echo "=== integration: $provider FAILED ==="
    else
        echo "=== integration: $provider PASS ==="
    fi
    echo
done

if [ "$fail_count" -gt 0 ]; then
    echo "integration-smoke: $fail_count provider(s) failed"
    exit 1
fi
echo "integration-smoke: all available providers passed"
