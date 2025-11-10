#!/bin/bash
# bash completion for shine CLI
# Install: Copy to /etc/bash_completion.d/shine or source in ~/.bashrc

_shine_completions() {
  local cur prev opts
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"

  # Top-level commands
  if [ $COMP_CWORD -eq 1 ]; then
    # Get command names from JSON output
    if command -v jq &> /dev/null; then
      opts=$(shine help --json names 2>/dev/null | jq -r '.[]' | tr '\n' ' ')
    else
      opts="start stop status reload logs help"
    fi
    COMPREPLY=($(compgen -W "$opts" -- "$cur"))
    return 0
  fi

  # Subcommand completions
  case "${COMP_WORDS[1]}" in
    help)
      if [ $COMP_CWORD -eq 2 ]; then
        if command -v jq &> /dev/null; then
          opts=$(shine help --json names 2>/dev/null | jq -r '.[]' | tr '\n' ' ')
          opts="$opts topics list categories"
        else
          opts="topics list categories start stop status reload logs"
        fi
        COMPREPLY=($(compgen -W "$opts" -- "$cur"))
      fi
      ;;
    logs)
      if [ $COMP_CWORD -eq 2 ]; then
        # Complete with common panel IDs
        opts="shinectl panel-0 panel-1"
        COMPREPLY=($(compgen -W "$opts" -- "$cur"))
      fi
      ;;
  esac
}

complete -F _shine_completions shine
