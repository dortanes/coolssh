package desktop

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/dortanes/coolssh/internal/library"
)

// LibraryEvent carries the whole library to every window each time it changes.
const LibraryEvent = "library"

func init() {
	application.RegisterEvent[library.Library](LibraryEvent)
}

// refuse is the error a window shows for a failed request: the catalog's wording of the library's
// refusal, else of failed, with the cause logged.
func (a *App) refuse(err error, failed string) error {
	var refusal library.Refusal
	if errors.As(err, &refusal) {
		return errors.New(a.text(string(refusal)))
	}
	a.application.Logger.Error(failed, "error", err)
	return errors.New(a.text(failed))
}

// published tells every window about a library that changed, or words why it did not.
func (a *App) published(updated library.Library, err error, failed string) error {
	if err != nil {
		return a.refuse(err, failed)
	}
	a.application.Event.Emit(LibraryEvent, updated)
	a.followSources(updated)
	return nil
}

func (a *App) edit(edit func(*library.Library) error) error {
	updated, err := a.options.Library.Change(edit)
	return a.published(updated, err, "library_not_saved")
}

// Library returns the whole library.
func (s *Service) Library() library.Library {
	return s.app.options.Library.Library()
}

// Inherited answers what a host or folder would take for each setting it leaves unset, and from
// which folder.
func (s *Service) Inherited(id string) (library.Inherited, error) {
	inherited, err := s.app.options.Library.Inherited(id)
	if err != nil {
		return library.Inherited{}, s.app.refuse(err, "library_not_saved")
	}
	return inherited, nil
}

// CreateFolder adds a folder at the end of parent, the root when empty, and returns its id.
func (s *Service) CreateFolder(parent, name string) (string, error) {
	var id string
	return id, s.app.edit(func(l *library.Library) (err error) {
		id, err = l.CreateFolder(parent, name)
		return err
	})
}

// CreateHost adds a host at the end of parent, the root when empty, and returns its id.
func (s *Service) CreateHost(parent, name string) (string, error) {
	var id string
	return id, s.app.edit(func(l *library.Library) (err error) {
		id, err = l.CreateHost(parent, name)
		return err
	})
}

// Rename gives a host or folder a new name.
func (s *Service) Rename(id, name string) error {
	return s.app.edit(func(l *library.Library) error { return l.Rename(id, name) })
}

// SetTags gives a host its tags.
func (s *Service) SetTags(id string, tags []string) error {
	return s.app.edit(func(l *library.Library) error { return l.SetTags(id, tags) })
}

// SetIcon gives a folder one of the icons.
func (s *Service) SetIcon(id string, icon library.Icon) error {
	return s.app.edit(func(l *library.Library) error { return l.SetIcon(id, icon) })
}

// SetHostIcon sets the icon the hosts in a folder are drawn with; empty takes the folder's above.
func (s *Service) SetHostIcon(id string, icon library.Icon) error {
	return s.app.edit(func(l *library.Library) error { return l.SetHostIcon(id, icon) })
}

// SetExpanded records whether a folder shows what it holds.
func (s *Service) SetExpanded(id string, expanded bool) error {
	return s.app.edit(func(l *library.Library) error { return l.SetExpanded(id, expanded) })
}

// SetAddress gives a host the name or address it is reached at.
func (s *Service) SetAddress(id, address string) error {
	return s.app.edit(func(l *library.Library) error { return l.SetAddress(id, address) })
}

// SetSettings replaces a host's or folder's connection settings.
func (s *Service) SetSettings(id string, settings library.Settings) error {
	return s.app.edit(func(l *library.Library) error { return l.SetSettings(id, settings) })
}

// Move puts items into parent, the root when empty, at index among its children.
func (s *Service) Move(ids []string, parent string, index int) error {
	return s.app.edit(func(l *library.Library) error { return l.Move(ids, parent, index) })
}

// Duplicate puts a copy of a host, named name, right after it and returns the copy's id.
func (s *Service) Duplicate(id, name string) (string, error) {
	var copied string
	return copied, s.app.edit(func(l *library.Library) (err error) {
		copied, err = l.Duplicate(id, name)
		return err
	})
}

// Delete removes a host, or a folder with everything inside it, and their saved passwords.
func (s *Service) Delete(id string) error {
	updated, err := s.app.options.Library.Delete(id)
	return s.app.published(updated, err, "library_not_saved")
}

// SetPassword saves a host's or folder's password in the keyring.
func (s *Service) SetPassword(id, password string) error {
	updated, err := s.app.options.Library.SetPassword(id, password)
	return s.app.published(updated, err, "secrets_not_saved")
}

// ClearPassword forgets a host's or folder's password.
func (s *Service) ClearPassword(id string) error {
	updated, err := s.app.options.Library.ClearPassword(id)
	return s.app.published(updated, err, "secrets_not_saved")
}

// PickKeyFile asks for a private key file in the system's file dialog, starting in the user's
// .ssh folder, and returns its path; empty when the user cancels.
func (s *Service) PickKeyFile() (string, error) {
	dialog := s.app.application.Dialog.OpenFile().
		SetTitle(s.app.text("keys_pick_title")).
		CanChooseFiles(true).
		ShowHiddenFiles(true)
	if home, err := os.UserHomeDir(); err == nil {
		dialog.SetDirectory(filepath.Join(home, ".ssh"))
	}
	if window := s.app.application.Window.Current(); window != nil {
		dialog.AttachToWindow(window)
	}
	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return "", s.app.refuse(err, "keys_not_picked")
	}
	return path, nil
}

// AddKeyFile adds the private key at path, named after the file, and returns its id. A
// passphrase, when given, opens an encrypted key and is saved in the keyring.
func (s *Service) AddKeyFile(path, passphrase string) (string, error) {
	id, updated, err := s.app.options.Library.AddKeyFile(path, filepath.Base(path), passphrase)
	return id, s.app.published(updated, err, "secrets_not_saved")
}

// AddKeyText adds a pasted private key named name and returns its id; its text, and a passphrase
// when given, are saved in the keyring.
func (s *Service) AddKeyText(name, text, passphrase string) (string, error) {
	id, updated, err := s.app.options.Library.AddKeyText(name, text, passphrase)
	return id, s.app.published(updated, err, "secrets_not_saved")
}

// SetDefaultKey makes a key the one hosts use when neither they nor their folders set one; empty
// sets none.
func (s *Service) SetDefaultKey(id string) error {
	return s.app.edit(func(l *library.Library) error { return l.SetDefaultKey(id) })
}

// RenameKey gives a key a new name.
func (s *Service) RenameKey(id, name string) error {
	return s.app.edit(func(l *library.Library) error { return l.RenameKey(id, name) })
}

// DeleteKey removes a key and its secrets; hosts and folders that used it inherit their key again.
func (s *Service) DeleteKey(id string) error {
	updated, err := s.app.options.Library.DeleteKey(id)
	return s.app.published(updated, err, "secrets_not_saved")
}

// BuiltInDefaults returns the settings a host takes when nothing, the user's defaults included,
// sets one.
func (s *Service) BuiltInDefaults() library.Settings {
	return s.app.options.Library.BuiltInDefaults()
}

// SetDefaults replaces the user's defaults for user, port, agent forwarding and keepalive.
func (s *Service) SetDefaults(defaults library.Settings) error {
	return s.app.edit(func(l *library.Library) error { return l.SetDefaults(defaults) })
}
