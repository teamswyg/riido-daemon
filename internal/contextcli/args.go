package contextcli

import (
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultLimit   = 20
	MaxLimit       = 100
	DefaultTimeout = 30 * time.Second
	maxTimeout     = 10 * time.Minute
)

// secretFlags are rejected everywhere: secrets never travel in argv.
var secretFlags = []string{
	"--token", "--access-token", "--refresh-token", "--secret", "--password", "--api-key",
}

var (
	commonValueFlags = []string{"--format", "--profile", "--workspace", "--team", "--limit", "--cursor", "--timeout"}
	commonBoolFlags  = []string{"--json", "--no-input", "--no-color", "--help", "-h"}
)

// ArgSpec declares the command-specific flags and positional arity.
type ArgSpec struct {
	Values         []string
	Bools          []string
	MaxPositionals int
}

// Options are the flags shared by every work-context command.
type Options struct {
	JSON      bool
	Help      bool
	NoInput   bool
	NoColor   bool
	Profile   string
	Workspace string
	Team      string
	Cursor    string
	Limit     int
	Timeout   time.Duration
}

// Args is the parsed command line.
type Args struct {
	Options

	Positionals []string
	values      map[string]string
	bools       map[string]bool
}

// Value returns the value of a command-specific value flag.
func (a Args) Value(flag string) string { return a.values[flag] }

// Bool reports whether a command-specific boolean flag was set.
func (a Args) Bool(flag string) bool { return a.bools[flag] }

// ParseArgs parses args against spec. Secrets passed as flags fail with
// SECRET_IN_ARGV before anything else is interpreted.
func ParseArgs(args []string, spec ArgSpec) (Args, *CLIError) {
	if err := rejectSecretArgs(args); err != nil {
		return Args{}, err
	}
	parsed := Args{
		Options: Options{Limit: DefaultLimit, Timeout: DefaultTimeout},
		values:  map[string]string{},
		bools:   map[string]bool{},
	}
	for i := 0; i < len(args); i++ {
		name, inline, hasInline := strings.Cut(args[i], "=")
		if !strings.HasPrefix(name, "-") {
			parsed.Positionals = append(parsed.Positionals, args[i])
			continue
		}
		switch {
		case slices.Contains(commonBoolFlags, name) || slices.Contains(spec.Bools, name):
			if hasInline {
				return Args{}, Failf(CodeUsage, "%s 옵션은 값을 받지 않습니다", name)
			}
			parsed.setBool(name)
		case slices.Contains(commonValueFlags, name) || slices.Contains(spec.Values, name):
			value := inline
			if !hasInline {
				i++
				if i >= len(args) {
					return Args{}, Failf(CodeUsage, "%s 옵션에 값이 필요합니다", name)
				}
				value = args[i]
			}
			if err := parsed.setValue(name, value); err != nil {
				return Args{}, err
			}
		default:
			return Args{}, Failf(CodeUsage, "알 수 없는 옵션: %s", name)
		}
	}
	if len(parsed.Positionals) > spec.MaxPositionals {
		return Args{}, Failf(CodeUsage, "인자가 너무 많습니다: %s", strings.Join(parsed.Positionals, " "))
	}
	return parsed, nil
}

func rejectSecretArgs(args []string) *CLIError {
	for _, arg := range args {
		name, _, _ := strings.Cut(arg, "=")
		if slices.Contains(secretFlags, name) {
			return Fail(CodeSecretInArgv,
				name+" 옵션은 지원하지 않습니다. 비밀값은 명령 인자로 받지 않습니다",
				"헤드리스 로그인은 riido auth login --with-token 으로 stdin에 입력하세요")
		}
	}
	return nil
}

func (a *Args) setBool(name string) {
	switch name {
	case "--json":
		a.JSON = true
	case "--help", "-h":
		a.Help = true
	case "--no-input":
		a.NoInput = true
	case "--no-color":
		a.NoColor = true
	default:
		a.bools[name] = true
	}
}

func (a *Args) setValue(name, value string) *CLIError {
	switch name {
	case "--format":
		return a.setFormat(value)
	case "--profile":
		a.Profile = value
	case "--workspace":
		a.Workspace = value
	case "--team":
		a.Team = value
	case "--cursor":
		a.Cursor = value
	case "--limit":
		return a.setLimit(value)
	case "--timeout":
		return a.setTimeout(value)
	default:
		a.values[name] = value
	}
	return nil
}

func (a *Args) setFormat(value string) *CLIError {
	switch value {
	case "json":
		a.JSON = true
	case "text":
		a.JSON = false
	default:
		return Failf(CodeUsage, "--format은 json 또는 text만 받습니다: %s", value)
	}
	return nil
}

func (a *Args) setLimit(value string) *CLIError {
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > MaxLimit {
		return Failf(CodeUsage, "--limit은 1에서 %d 사이 정수여야 합니다: %s", MaxLimit, value)
	}
	a.Limit = limit
	return nil
}

func (a *Args) setTimeout(value string) *CLIError {
	timeout, err := time.ParseDuration(value)
	if err != nil {
		seconds, convErr := strconv.Atoi(value)
		if convErr != nil {
			return Failf(CodeUsage, "--timeout 형식이 올바르지 않습니다: %s", value)
		}
		timeout = time.Duration(seconds) * time.Second
	}
	if timeout <= 0 || timeout > maxTimeout {
		return Failf(CodeUsage, "--timeout은 0보다 크고 %s 이하여야 합니다: %s", maxTimeout, value)
	}
	a.Timeout = timeout
	return nil
}
