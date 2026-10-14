package claude

// MaxLineBytes is the largest stream-json line accepted on stdout or stderr.
// Claude can emit very large tool results; the adapter stays bounded at 10 MB.
const MaxLineBytes = 10 * 1024 * 1024

var (
	stdoutStreamPrefixes = []string{"stdout: ", "STDOUT: "}
	stderrStreamPrefixes = []string{"stderr: ", "STDERR: "}
)
