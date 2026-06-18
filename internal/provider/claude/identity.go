package claude

import providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"

// Name is the canonical adapter identifier.
const Name = string(providercatalog.KindClaude)

// DefaultExecutable is the binary name resolved on $PATH when no
// explicit executable is configured.
const DefaultExecutable = "claude"
