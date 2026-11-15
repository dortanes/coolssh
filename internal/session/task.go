package session

import "fmt"

// Task is a command a session runs instead of a shell, its output shown in the session's tab.
type Task struct {
	Command string
	// Done and Failed are the codes the session closes with when the command exits with status 0,
	// and otherwise.
	Done, Failed string
}

// InstallKeyTask adds an authorized_keys line to the user's authorized_keys on a POSIX host
// unless the file holds it already. The line is a key type and base64, so it quotes safely.
func InstallKeyTask(line string) Task {
	return Task{
		Command: fmt.Sprintf(
			"umask 077 && mkdir -p ~/.ssh && touch ~/.ssh/authorized_keys && "+
				"{ grep -qxF '%[1]s' ~/.ssh/authorized_keys || echo '%[1]s' >> ~/.ssh/authorized_keys; }",
			line,
		),
		Done:   "session_key_installed",
		Failed: "session_key_not_installed",
	}
}
