package toolargs

const (
	// RedactedValue is stored when an argument key or value identifies
	// sensitive material. The original value must not be preserved in
	// ToolRef.Args.
	RedactedValue = "[redacted]"

	maxArgs       = 32
	maxDepth      = 4
	maxValueRunes = 256
)

var sensitiveKeyTokens = []string{
	"api_key",
	"apikey",
	"authorization",
	"bearer",
	"credential",
	"password",
	"private_key",
	"secret",
	"token",
}
