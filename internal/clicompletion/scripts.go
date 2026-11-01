package clicompletion

import _ "embed"

//go:embed scripts/riido.bash
var bashScript string

//go:embed scripts/_riido
var zshScript string

//go:embed scripts/riido.fish
var fishScript string

//go:embed scripts/riido.ps1
var powershellScript string

func Script(shell string) (string, bool) {
	switch shell {
	case "bash":
		return bashScript, true
	case "zsh":
		return zshScript, true
	case "fish":
		return fishScript, true
	case "powershell":
		return powershellScript, true
	default:
		return "", false
	}
}
