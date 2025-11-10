#compdef shine
# zsh completion for shine CLI
# Install: Copy to ~/.zfunc/_shine and add to fpath

_shine() {
  local line state

  _arguments -C \
    '1: :->command' \
    '*::arg:->args'

  case $state in
    command)
      local -a commands
      IFS=$'\n' commands=($(shine help --json 2>/dev/null | jq -r '.[] | "\(.name):\(.synopsis)"'))
      _describe 'command' commands
      ;;
    args)
      case $line[1] in
        help)
          local -a topics
          topics=(
            'topics:List all available help topics'
            'list:Show all commands with descriptions'
            'categories:Show commands organized by category'
          )
          # Add individual commands
          IFS=$'\n' topics+=($(shine help --json names 2>/dev/null | jq -r '.[] | "\(.):Show help for \(.)"'))
          _describe 'help topic' topics
          ;;
        logs)
          # Complete with panel IDs (could be enhanced to query actual panels)
          _values 'panel-id' 'shinectl' 'panel-0' 'panel-1'
          ;;
      esac
      ;;
  esac
}

_shine
