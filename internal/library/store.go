package library

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/dortanes/coolssh/internal/jsonfile"
	"github.com/dortanes/coolssh/internal/secrets"
)

// Store holds the library in effect, the file it is saved to and the keyring its secrets live
// in. It is safe for concurrent use.
type Store struct {
	path     string
	secrets  secrets.Secrets
	defaults Settings
	mu       sync.Mutex
	library  Library
}

// Open reads path, a missing file being an empty library; an unreadable or invalid one leaves an
// empty library in effect, untouched until saved, with an error. defaults must set every setting.
func Open(path string, secrets secrets.Secrets, defaults Settings) (*Store, error) {
	s := &Store{path: path, secrets: secrets, defaults: defaults, library: Empty()}
	loaded := Empty()
	err := jsonfile.Read(path, &loaded)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err == nil {
		err = loaded.check()
	}
	if err != nil {
		return s, fmt.Errorf("library %s: %w", path, err)
	}
	s.library = loaded
	return s, nil
}

// Library returns a copy of the library in effect.
func (s *Store) Library() Library {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.library.clone()
}

// BuiltInDefaults are the settings a host takes when nothing, the user's defaults included, sets
// one.
func (s *Store) BuiltInDefaults() Settings {
	return s.defaults.clone()
}

// Inherited answers what id would take for each setting it leaves unset, and from where.
func (s *Store) Inherited(id string) (Inherited, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.library.Inherited(id, s.defaults)
}

// Change applies edit to a copy of the library, saves it and puts it in effect; when edit fails
// or the file cannot be written, nothing changes.
func (s *Store) Change(edit func(*Library) error) (Library, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.library.clone()
	if err := edit(&next); err != nil {
		return Library{}, err
	}
	if err := jsonfile.Write(s.path, next); err != nil {
		return Library{}, fmt.Errorf("library %s: %w", s.path, err)
	}
	s.library = next
	return next.clone(), nil
}

// Delete removes a host or a folder with everything inside it, and the secrets of all of them.
func (s *Store) Delete(id string) (Library, error) {
	var removed []string
	library, err := s.Change(func(l *Library) error {
		var err error
		removed, err = l.Delete(id)
		return err
	})
	if err != nil {
		return Library{}, err
	}
	return library, s.forget(removed)
}

// forget deletes the secrets of items no longer in the library.
func (s *Store) forget(removed []string) error {
	var errs []error
	for _, gone := range removed {
		errs = append(errs, s.secrets.Delete(passwordSecret(gone)), s.secrets.Delete(sourceTokenSecret(gone)))
	}
	return errors.Join(errs...)
}

// SyncPeers makes a synced folder hold exactly peers and saves the library when that changed it;
// changed is false, and nothing is written, when the folder already matched.
func (s *Store) SyncPeers(id string, peers []Peer) (library Library, changed bool, err error) {
	s.mu.Lock()
	next := s.library.clone()
	removed, changed, err := next.SyncPeers(id, peers)
	if err != nil || !changed {
		s.mu.Unlock()
		return Library{}, false, err
	}
	if err := jsonfile.Write(s.path, next); err != nil {
		s.mu.Unlock()
		return Library{}, false, fmt.Errorf("library %s: %w", s.path, err)
	}
	s.library = next
	library = next.clone()
	s.mu.Unlock()
	return library, true, s.forget(removed)
}

// SetSourceToken saves the access token a synced folder's source reads groups with.
func (s *Store) SetSourceToken(id, token string) (Library, error) {
	if err := s.secrets.Set(sourceTokenSecret(id), token); err != nil {
		return Library{}, err
	}
	return s.Change(func(l *Library) error { return l.setSourceToken(id, true) })
}

// ClearSourceToken forgets a synced folder's access token.
func (s *Store) ClearSourceToken(id string) (Library, error) {
	library, err := s.Change(func(l *Library) error { return l.setSourceToken(id, false) })
	if err != nil {
		return Library{}, err
	}
	return library, s.secrets.Delete(sourceTokenSecret(id))
}

// SourceToken returns a synced folder's access token; empty when none is saved.
func (s *Store) SourceToken(id string) (string, error) {
	token, err := s.secrets.Get(sourceTokenSecret(id))
	if errors.Is(err, secrets.ErrNotFound) {
		return "", nil
	}
	return token, err
}

// SetPassword saves password for a host or folder in the keyring.
func (s *Store) SetPassword(id, password string) (Library, error) {
	if err := s.secrets.Set(passwordSecret(id), password); err != nil {
		return Library{}, err
	}
	saved := true
	return s.Change(func(l *Library) error { return l.setPassword(id, &saved) })
}

// ClearPassword forgets a host's or folder's password; it inherits one again.
func (s *Store) ClearPassword(id string) (Library, error) {
	library, err := s.Change(func(l *Library) error { return l.setPassword(id, nil) })
	if err != nil {
		return Library{}, err
	}
	return library, s.secrets.Delete(passwordSecret(id))
}

// AddKeyFile adds the private key at path, named name; the library keeps the path. A passphrase,
// when given, opens an encrypted key and is saved in the keyring.
func (s *Store) AddKeyFile(path, name, passphrase string) (string, Library, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return "", Library{}, errUnreadableKey
	}
	key, err := inspectKey(text, passphrase)
	if err != nil {
		return "", Library{}, err
	}
	key.Name, key.Source, key.Path = name, FileKey, path
	return s.addKey(key, "", passphrase)
}

// AddKeyText adds a pasted private key, named name; its text is saved in the keyring. A
// passphrase, when given, opens an encrypted key and is saved with it.
func (s *Store) AddKeyText(name, text, passphrase string) (string, Library, error) {
	key, err := inspectKey([]byte(text), passphrase)
	if err != nil {
		return "", Library{}, err
	}
	key.Name, key.Source = name, StoredKey
	return s.addKey(key, text, passphrase)
}

func (s *Store) addKey(key Key, text, passphrase string) (string, Library, error) {
	var id string
	library, err := s.Change(func(l *Library) error {
		var err error
		id, err = l.AddKey(key)
		return err
	})
	if err != nil {
		return "", Library{}, err
	}
	if text != "" {
		if err := s.secrets.Set(keySecret(id), text); err != nil {
			return "", Library{}, s.undoKey(id, err)
		}
	}
	if passphrase != "" {
		if err := s.secrets.Set(passphraseSecret(id), passphrase); err != nil {
			return "", Library{}, s.undoKey(id, err)
		}
	}
	return id, library, nil
}

// undoKey removes a key whose secrets could not be saved, and returns why they could not.
func (s *Store) undoKey(id string, cause error) error {
	_, err := s.DeleteKey(id)
	return errors.Join(cause, err)
}

// DeleteKey removes a key and its secrets; hosts and folders that used it inherit their key again.
func (s *Store) DeleteKey(id string) (Library, error) {
	library, err := s.Change(func(l *Library) error { return l.DeleteKey(id) })
	if err != nil {
		return Library{}, err
	}
	return library, errors.Join(s.secrets.Delete(keySecret(id)), s.secrets.Delete(passphraseSecret(id)))
}

func sourceTokenSecret(id string) string {
	return "source/" + id
}

func passwordSecret(id string) string {
	return "password/" + id
}

func keySecret(id string) string {
	return "key/" + id
}

func passphraseSecret(id string) string {
	return "key/" + id + "/passphrase"
}

// SetSource makes a folder take its hosts from source; nil keeps its hosts as ordinary ones and
// forgets the source's token.
func (s *Store) SetSource(id string, source *Source) (Library, error) {
	library, err := s.Change(func(l *Library) error { return l.SetSource(id, source) })
	if err != nil || source != nil {
		return library, err
	}
	return library, s.secrets.Delete(sourceTokenSecret(id))
}
