# Riido completion for fish. Source this file or save to completions/riido.fish.
function __riido_complete
    # Raw prefix: no variable, wildcard or command expansion while completing.
    set -l prefix (commandline --cut-at-cursor --current-process | string collect --allow-empty)
    printf '%s\0' shell "$prefix" | command riido completion __complete 2>/dev/null
end
complete --command riido --erase
complete --command riido --no-files --arguments '(__riido_complete)'
