package desktop

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/dortanes/coolssh/internal/library"
	"github.com/dortanes/coolssh/internal/session"
)

// TrustedKeysEvent carries the host keys trusted in the app each time they change.
const TrustedKeysEvent = "trusted-keys"

func init() {
	application.RegisterEvent[[]session.TrustedKey](TrustedKeysEvent)
}

func (a *App) followTrustedKeys() {
	a.options.KnownHosts.Follow(func() {
		trusted, err := a.options.KnownHosts.Trusted()
		if err != nil {
			a.application.Logger.Error("the trusted host keys were not read", "error", err)
			return
		}
		a.application.Event.Emit(TrustedKeysEvent, trusted)
	})
}

// GenerateKey makes a new key of kind named name, kept in the keyring with its passphrase when one
// is given, and returns its id.
func (s *Service) GenerateKey(name string, kind library.KeyKind, passphrase string) (string, error) {
	id, updated, err := s.app.options.Library.GenerateKey(name, kind, passphrase)
	return id, s.app.published(updated, err, "secrets_not_saved")
}

// PublicKey returns a key's public half as a line for authorized_keys.
func (s *Service) PublicKey(id string) (string, error) {
	line, err := s.app.options.Library.PublicKey(id)
	if err != nil {
		return "", s.app.refuse(err, "keys_public_unreadable")
	}
	return line, nil
}

// InstallKey opens a session that adds the key a host uses to the host's authorized_keys, and
// returns its id; a host without a key is refused.
func (s *Service) InstallKey(hostID string) (string, error) {
	line, err := s.app.options.Library.HostPublicKey(hostID)
	if err != nil {
		return "", s.app.refuse(err, "keys_public_unreadable")
	}
	id, err := s.app.sessions.OpenTask(hostID, session.InstallKeyTask(line))
	if err != nil {
		return "", s.app.refuse(err, "session_not_opened")
	}
	return id, nil
}

// TrustedKeys lists the host keys trusted in the app.
func (s *Service) TrustedKeys() ([]session.TrustedKey, error) {
	trusted, err := s.app.options.KnownHosts.Trusted()
	if err != nil {
		return nil, s.app.refuse(err, "keys_trusted_unreadable")
	}
	return trusted, nil
}

// ForgetTrustedKey stops trusting a host key, so the next connection to host asks again.
func (s *Service) ForgetTrustedKey(host, fingerprint string) error {
	if err := s.app.options.KnownHosts.Forget(host, fingerprint); err != nil {
		return s.app.refuse(err, "keys_trusted_not_saved")
	}
	return nil
}
