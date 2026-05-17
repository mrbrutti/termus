package main

import (
	"fmt"
	"strings"
)

func printCompletion(shell string) error {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "zsh":
		fmt.Print(zshCompletionScript())
		return nil
	case "bash":
		fmt.Print(bashCompletionScript())
		return nil
	default:
		return fmt.Errorf("unknown shell %q (expected zsh or bash)", shell)
	}
}

func zshCompletionScript() string {
	return `#compdef termus
_termus() {
  local -a opts
  opts=(
    '-track[play an authored track]:track id:->track_ids'
    '--track[play an authored track]:track id:->track_ids'
    '--algo[play a procedural style]:style: '
    '--sf2-strategy[soundfont strategy]:strategy:(single pro max)'
    '--listen-mode[listening mode]:mode:(endless album-side hour-stream radio)'
    '--out[render to wav]:file:_files'
    '--playlist-out[render playlist to directory]:directory:_files -/'
    '--stems[export stems]'
    '--midi[export midi]'
    '--debug[show debug inspector]'
    '--help[show help]'
  )
  _arguments -s $opts && return 0
  case $state in
    track_ids)
      local -a tracks
      tracks=("${(@f)$($words[1] --complete-track-prefix "$PREFIX" 2>/dev/null)}")
      _describe 'tracks' tracks
      ;;
  esac
}
_termus "$@"
`
}

func bashCompletionScript() string {
	return `_termus_complete() {
  local cur prev words cword
  _init_completion || return
  case "$prev" in
    -track|--track)
      COMPREPLY=( $( compgen -W "$("$1" --complete-track-prefix "$cur" 2>/dev/null)" -- "$cur" ) )
      return
      ;;
  esac
  COMPREPLY=( $( compgen -W "--track --algo --sf2-strategy --listen-mode --out --playlist-out --stems --midi --debug --help" -- "$cur" ) )
}
complete -F _termus_complete termus ./termus
`
}
