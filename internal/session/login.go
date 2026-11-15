package session

import (
	"bytes"
	"context"
	"errors"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/dortanes/coolssh/internal/library"
	"github.com/dortanes/coolssh/internal/secrets"
)

// login authenticates one host of an attempt.
type login struct {
	attempt    *attempt
	ctx        context.Context
	connection library.Connection
	// savedTried is true once the saved password has been offered, so a refusal asks the user.
	savedTried bool
	// asked counts the passwords and answers the user typed for this host.
	asked int
	// typed is the last password the user typed; remember is whether they asked to save it.
	typed    string
	remember bool
}

// methods are OpenSSH's order: keys, then keyboard-interactive, then password.
func (l *login) methods() []ssh.AuthMethod {
	return []ssh.AuthMethod{
		ssh.PublicKeysCallback(l.signers),
		ssh.RetryableAuthMethod(ssh.KeyboardInteractive(l.answer), passwordTries),
		ssh.RetryableAuthMethod(ssh.PasswordCallback(l.password), passwordTries),
	}
}

func (l *login) ask(prompt Prompt) (Answer, error) {
	prompt.Host = l.connection.Name
	answer, err := l.attempt.session.ask(l.ctx, prompt)
	if err != nil {
		return Answer{}, l.attempt.stop(err)
	}
	return answer, nil
}

func (l *login) checkHostKey(address string, remote net.Addr, key ssh.PublicKey) error {
	if l.vouched(key) {
		return nil
	}
	knownHosts := l.attempt.session.owner.options.KnownHosts
	err := knownHosts.check(l.connection.Name, address, remote, key)
	if !errors.Is(err, errUnknownHostKey) {
		if err != nil {
			return l.attempt.stop(err)
		}
		return nil
	}
	if _, err := l.ask(Prompt{Kind: HostKeyPrompt, KeyType: key.Type(), Fingerprint: ssh.FingerprintSHA256(key)}); err != nil {
		return err
	}
	if err := knownHosts.trust(address, key); err != nil {
		return l.attempt.stop(err)
	}
	return nil
}

// vouched reports whether key is the one the host's source vouches for; any other key goes through
// known_hosts as usual.
func (l *login) vouched(key ssh.PublicKey) bool {
	if l.connection.HostKey == "" {
		return false
	}
	trusted, _, _, _, err := ssh.ParseAuthorizedKey([]byte(l.connection.HostKey))
	return err == nil && bytes.Equal(trusted.Marshal(), key.Marshal())
}

// signers are the host's key, then the agent's keys.
func (l *login) signers() ([]ssh.Signer, error) {
	var signers []ssh.Signer
	if key := l.connection.Key; key != nil {
		signer, err := l.keySigner(*key)
		if err != nil {
			return nil, l.attempt.stop(err)
		}
		signers = append(signers, signer)
	}
	if l.attempt.agent != nil {
		// An agent that stopped answering leaves the other methods to try.
		if held, err := l.attempt.agent.Signers(); err == nil {
			signers = append(signers, held...)
		}
	}
	return signers, nil
}

// keySigner opens a key with its saved passphrase, else with one the user types.
func (l *login) keySigner(key library.Key) (ssh.Signer, error) {
	store := l.attempt.session.owner.options.Library
	text, err := store.KeyText(key)
	if err != nil {
		return nil, secretProblem(err, l.connection.Name)
	}
	signer, err := ssh.ParsePrivateKey(text)
	var missing *ssh.PassphraseMissingError
	if !errors.As(err, &missing) {
		if err != nil {
			return nil, problem("library_not_a_key", l.connection.Name)
		}
		return signer, nil
	}
	saved, ok, err := store.Passphrase(key.ID)
	if err != nil {
		return nil, secretProblem(err, l.connection.Name)
	}
	if ok {
		if signer, err := ssh.ParsePrivateKeyWithPassphrase(text, []byte(saved)); err == nil {
			return signer, nil
		}
	}
	for retry := false; ; retry = true {
		answer, err := l.ask(Prompt{Kind: PassphrasePrompt, Key: key.Name, Retry: retry})
		if err != nil {
			return nil, err
		}
		if signer, err := ssh.ParsePrivateKeyWithPassphrase(text, []byte(first(answer.Values))); err == nil {
			return signer, nil
		}
	}
}

// saved is the password saved for the host, offered once.
func (l *login) saved() (string, bool, error) {
	if l.savedTried || l.connection.PasswordItem == "" {
		return "", false, nil
	}
	l.savedTried = true
	password, err := l.attempt.session.owner.options.Library.Password(l.connection.PasswordItem)
	if errors.Is(err, secrets.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, secretProblem(err, l.connection.Name)
	}
	return password, true, nil
}

// typePassword asks the user for the host's password.
func (l *login) typePassword() (string, error) {
	answer, err := l.ask(Prompt{Kind: PasswordPrompt, Retry: l.asked > 0, Savable: l.connection.HostID != ""})
	if err != nil {
		return "", err
	}
	l.asked++
	l.typed, l.remember = first(answer.Values), answer.Remember
	return l.typed, nil
}

func (l *login) password() (string, error) {
	password, ok, err := l.saved()
	if err != nil {
		return "", l.attempt.stop(err)
	}
	if ok {
		return password, nil
	}
	return l.typePassword()
}

// answer replies to the server's keyboard-interactive questions. One hidden question is a password:
// the saved one answers it, else the user types it as one.
func (l *login) answer(name, instruction string, questions []string, echoes []bool) ([]string, error) {
	if len(questions) == 0 {
		return nil, nil
	}
	if len(questions) == 1 && !echoes[0] {
		password, err := l.password()
		if err != nil {
			return nil, err
		}
		return []string{password}, nil
	}
	asked := make([]Question, len(questions))
	for i, text := range questions {
		asked[i] = Question{Text: text, Echo: echoes[i]}
	}
	answer, err := l.ask(Prompt{Kind: QuestionsPrompt, Name: name, Instruction: instruction, Questions: asked, Retry: l.asked > 0})
	if err != nil {
		return nil, err
	}
	l.asked++
	values := make([]string, len(questions))
	copy(values, answer.Values)
	return values, nil
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// secretProblem is the problem of a secret that could not be read: the key file's refusal, else
// the keychain's.
func secretProblem(err error, host string) *Problem {
	var refusal library.Refusal
	if errors.As(err, &refusal) {
		return problem(string(refusal), host)
	}
	return problem("session_keychain", host)
}
