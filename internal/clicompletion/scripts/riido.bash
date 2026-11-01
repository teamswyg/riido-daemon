# Riido completion for Bash 3.2+. Source this file in the current shell.
_riido_completion() {
  local LC_ALL=C candidate prefix current left flag
  local -a candidates
  prefix="${COMP_LINE:0:COMP_POINT}"
  candidates=()
  while IFS= read -r candidate; do
    if [[ "$candidate" == --*=* && "$COMP_WORDBREAKS" == *"="* ]]; then
      current="${COMP_WORDS[COMP_CWORD]}"
      flag="${candidate%%=*}"
      left="${current%%=*}"
      if [[ "$left" == \""$flag" || "$left" == \'"$flag" ]]; then
        : # A wholly quoted option has no unquoted '=' replacement boundary.
      elif [[ "$current" != *=* || "$left" == "$flag" ]]; then
        candidate="${candidate#*=}"
      else
        continue # Leave mixed quoting in the option name unchanged.
      fi
    fi
    candidates[${#candidates[@]}]="$candidate"
  done < <(printf '%s\0' shell "$prefix" | command riido completion __complete 2>/dev/null)
  COMPREPLY=("${candidates[@]}")
}
complete -F _riido_completion riido
