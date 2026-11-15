package session

import (
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/dortanes/coolssh/internal/library"
)

// Problem is why a session closed: a key of the catalogs and the values its message names. Every
// session message names the host as {host}.
type Problem struct {
	Code   string
	Values map[string]string
}

func (p *Problem) Error() string {
	keys := slices.Sorted(maps.Keys(p.Values))
	parts := []string{p.Code}
	for _, key := range keys {
		parts = append(parts, key+"="+p.Values[key])
	}
	return strings.Join(parts, " ")
}

// problem is the Problem code names about host, with values given as name, value pairs.
func problem(code, host string, values ...string) *Problem {
	named := map[string]string{"host": host}
	for i := 0; i+1 < len(values); i += 2 {
		named[values[i]] = values[i+1]
	}
	return &Problem{Code: code, Values: named}
}

var problems = []string{
	"session_unreachable", "session_authentication_failed", "session_host_key_changed",
	"session_host_key_revoked", "session_not_answering", "session_lost", "session_failed",
	"session_shell_refused", "session_keychain", "session_key_installed", "session_key_not_installed",
}

// Problems lists every session code, each a key the catalogs word.
func Problems() []string {
	return slices.Clone(problems)
}

// errDeclined reports that the user declined what the session asked; the session closes without a
// problem.
var errDeclined = errors.New("declined")

// asProblem is err as the Problem a session closes with: itself, a library refusal, or failed
// with the error's text.
func asProblem(err error, host string) *Problem {
	var found *Problem
	if errors.As(err, &found) {
		return found
	}
	var refusal library.Refusal
	if errors.As(err, &refusal) {
		return problem(string(refusal), host)
	}
	return problem("session_failed", host, "reason", err.Error())
}
