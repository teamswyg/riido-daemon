package cursor

import (
	"fmt"
)

func profileArgs(profile Profile, prompt string) ([]string, error) {
	switch profile {
	case ProfileRootPrint:
		return []string{"-p", prompt, "--output-format", "stream-json"}, nil
	case ProfileAgentSubcommand:
		return []string{"agent", "-p", prompt, "--output-format", "stream-json"}, nil
	case ProfileLegacyChat:
		return []string{"chat", "-p", prompt, "--output-format", "stream-json"}, nil
	default:
		return nil, fmt.Errorf("cursor: unknown profile %q (allowed: root-print, agent-subcommand, legacy-chat)", profile)
	}
}
