package desktop

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/dortanes/coolssh/internal/hostsource"
	"github.com/dortanes/coolssh/internal/library"
)

// SourcesEvent carries every synced folder's sync state, by folder id, each time one changes.
const SourcesEvent = "sources"

// groupsTimeout bounds listing a source's groups for the folder editor.
const groupsTimeout = 15 * time.Second

func init() {
	application.RegisterEvent[map[string]hostsource.State](SourcesEvent)
}

// sourceTokens caches the synced folders' access tokens read from the keyring.
type sourceTokens struct {
	mu     sync.Mutex
	tokens map[string]string
}

func (a *App) openSources() {
	a.tokens = &sourceTokens{tokens: map[string]string{}}
	a.syncer = hostsource.NewSyncer(a.options.Sources, a.syncPeers, a.publishSources)
	a.followSources(a.options.Library.Library())
}

// syncPeers puts a folder's peers in the library and tells the windows when that changed it.
func (a *App) syncPeers(folder string, peers []hostsource.Peer) {
	held := make([]library.Peer, len(peers))
	for i, peer := range peers {
		held[i] = library.Peer{Key: peer.Key, Name: peer.Name, Address: peer.Address, HostKey: peer.HostKey}
	}
	updated, changed, err := a.options.Library.SyncPeers(folder, held)
	switch {
	case err != nil:
		a.application.Logger.Error("the synced hosts were not saved", "folder", folder, "error", err)
	case changed:
		a.application.Event.Emit(LibraryEvent, updated)
	}
}

// followSources has the syncer watch exactly the library's synced folders.
func (a *App) followSources(l library.Library) {
	var folders []hostsource.Folder
	for id, folder := range l.Folders {
		if folder.Source == nil {
			continue
		}
		synced := hostsource.Folder{ID: id, Kind: hostsource.Kind(folder.Source.Kind), Groups: folder.Source.Groups}
		if folder.Source.Token {
			synced.Token = a.sourceToken(id)
		}
		folders = append(folders, synced)
	}
	a.syncer.Follow(folders)
}

// sourceToken returns a synced folder's token, reading the keyring until a read succeeds.
func (a *App) sourceToken(folder string) string {
	a.tokens.mu.Lock()
	defer a.tokens.mu.Unlock()
	if token, ok := a.tokens.tokens[folder]; ok {
		return token
	}
	token, err := a.options.Library.SourceToken(folder)
	if err != nil {
		a.application.Logger.Error("the source token was not read", "folder", folder, "error", err)
		return ""
	}
	a.tokens.tokens[folder] = token
	return token
}

func (a *App) rememberToken(folder, token string) {
	a.tokens.mu.Lock()
	defer a.tokens.mu.Unlock()
	a.tokens.tokens[folder] = token
}

func (a *App) publishSources() {
	a.application.Event.Emit(SourcesEvent, a.syncer.States())
}

// Sources lists every host source with whether its client is installed on this device.
func (s *Service) Sources() []hostsource.Availability {
	return s.app.syncer.Available()
}

// SourceStates returns every synced folder's sync state, by folder id.
func (s *Service) SourceStates() map[string]hostsource.State {
	return s.app.syncer.States()
}

// Resync reads a synced folder's source again from the start.
func (s *Service) Resync(folder string) {
	s.app.syncer.Restart(folder)
}

// SetSource makes a folder take its hosts from a source, narrowed to groups when any are given;
// an empty kind keeps its hosts as ordinary ones.
func (s *Service) SetSource(folder, kind string, groups []string) error {
	var source *library.Source
	if kind != "" {
		source = &library.Source{Kind: kind, Groups: groups}
	}
	updated, err := s.app.options.Library.SetSource(folder, source)
	if err == nil && source == nil {
		s.app.rememberToken(folder, "")
	}
	return s.app.published(updated, err, "library_not_saved")
}

// SetSourceToken saves the access token a folder's source reads groups with.
func (s *Service) SetSourceToken(folder, token string) error {
	updated, err := s.app.options.Library.SetSourceToken(folder, token)
	if err == nil {
		s.app.rememberToken(folder, token)
	}
	return s.app.published(updated, err, "secrets_not_saved")
}

// ClearSourceToken forgets a folder's access token; its group filter no longer applies.
func (s *Service) ClearSourceToken(folder string) error {
	updated, err := s.app.options.Library.ClearSourceToken(folder)
	if err == nil {
		s.app.rememberToken(folder, "")
	}
	return s.app.published(updated, err, "secrets_not_saved")
}

// SourceGroups lists the groups a synced folder can narrow its hosts to, read with its token.
func (s *Service) SourceGroups(folder string) ([]hostsource.Group, error) {
	held, ok := s.app.options.Library.Library().Folders[folder]
	if !ok || held.Source == nil {
		return nil, s.app.refuse(errors.New("not a synced folder"), "source_groups_unavailable")
	}
	source, ok := s.app.syncer.Source(hostsource.Kind(held.Source.Kind))
	if !ok {
		return nil, s.app.refuse(errors.New("unknown source"), "source_groups_unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), groupsTimeout)
	defer cancel()
	groups, err := source.Groups(ctx, hostsource.Config{Token: s.app.sourceToken(folder)})
	if errors.Is(err, hostsource.ErrUnauthorized) {
		return nil, errors.New(s.app.text("source_unauthorized"))
	}
	if err != nil {
		return nil, s.app.refuse(err, "source_groups_unavailable")
	}
	return groups, nil
}
