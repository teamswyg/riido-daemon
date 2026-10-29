package agentbridge

import (
	"path/filepath"
	"runtime"
	"strings"
)

func contextCLIInstruction(provider, executable string) string {
	section := TaskContextCLIContractInstruction()
	if provider != "codex" || !filepath.IsAbs(executable) {
		return section
	}
	// This value comes from daemon runtime configuration, never assignment Env.
	// Quote it for the native shell so spaces, apostrophes and substitutions stay
	// literal. A login shell may replace PATH but cannot redirect this command.
	command := quoteContextCLIExecutable(executable, runtime.GOOS)
	section = strings.ReplaceAll(section, "riido ", command+" ")
	return section + "\n- Invoke the daemon-owned executable shown above directly. Do not replace it with a PATH lookup or a different riido installation. Native sandbox restrictions still apply; do not bypass a denied socket."
}

func quoteContextCLIExecutable(executable, goos string) string {
	if goos == "windows" {
		return "& '" + strings.ReplaceAll(executable, "'", "''") + "'"
	}
	return "'" + strings.ReplaceAll(executable, "'", "'\"'\"'") + "'"
}
