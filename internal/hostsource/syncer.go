package hostsource

import (
	"context"
	"maps"
	"slices"
	"sync"
)

// Folder is a synced folder as the syncer follows it.
type Folder struct {
	ID   string
	Kind Kind
	// Groups narrows the folder to peers in any of these groups, by id; empty for every peer.
	Groups []string
	Token  string
}

func (f Folder) equal(other Folder) bool {
	return f.ID == other.ID && f.Kind == other.Kind && f.Token == other.Token && slices.Equal(f.Groups, other.Groups)
}

// State is what a window shows of a folder's sync.
type State struct {
	Status Status `json:"status"`
	// Hosts is how many hosts the last sync gave the folder.
	Hosts int `json:"hosts"`
	// Links is how this device reaches each of the folder's peers, by peer key, as last synced.
	Links map[string]Link `json:"links"`
}

// Syncer runs one watch per synced folder and hands each folder's peers, narrowed to its groups,
// to apply. It is safe for concurrent use.
type Syncer struct {
	sources map[Kind]Source
	order   []Kind
	apply   func(folder string, peers []Peer)
	changed func()

	mu      sync.Mutex
	watches map[string]*watch
	states  map[string]State
}

type watch struct {
	folder Folder
	cancel context.CancelFunc
}

// NewSyncer follows no folder until Follow. apply receives a folder's peers on every sync, and
// changed is called, with no lock held, whenever a folder's state changes.
func NewSyncer(sources []Source, apply func(folder string, peers []Peer), changed func()) *Syncer {
	byKind := map[Kind]Source{}
	var order []Kind
	for _, source := range sources {
		byKind[source.Kind()] = source
		order = append(order, source.Kind())
	}
	return &Syncer{sources: byKind, order: order, apply: apply, changed: changed, watches: map[string]*watch{}, states: map[string]State{}}
}

// Source returns the source of a kind; false for a kind the syncer does not know.
func (s *Syncer) Source(kind Kind) (Source, bool) {
	source, ok := s.sources[kind]
	return source, ok
}

// Follow makes the syncer watch exactly folders: it starts a watch for a new folder, restarts
// one whose source, groups or token changed, and stops one no longer synced.
func (s *Syncer) Follow(folders []Folder) {
	wanted := map[string]Folder{}
	for _, folder := range folders {
		if _, ok := s.sources[folder.Kind]; ok {
			wanted[folder.ID] = folder
		}
	}
	s.mu.Lock()
	for id, running := range s.watches {
		if folder, ok := wanted[id]; !ok || !folder.equal(running.folder) {
			running.cancel()
			delete(s.watches, id)
			delete(s.states, id)
		}
	}
	for id, folder := range wanted {
		if _, ok := s.watches[id]; !ok {
			s.start(folder)
		}
	}
	s.mu.Unlock()
	s.changed()
}

// Restart watches a followed folder afresh: the source is read again from the start, groups
// included. A folder not followed is left alone.
func (s *Syncer) Restart(id string) {
	s.mu.Lock()
	running, ok := s.watches[id]
	if ok {
		running.cancel()
		// The hosts and links stay shown until the source answers again.
		kept := s.states[id]
		s.start(running.folder)
		kept.Status = Connecting
		s.states[id] = kept
	}
	s.mu.Unlock()
	if ok {
		s.changed()
	}
}

// start watches folder; s.mu is held.
func (s *Syncer) start(folder Folder) {
	ctx, cancel := context.WithCancel(context.Background())
	running := &watch{folder: folder, cancel: cancel}
	s.watches[folder.ID] = running
	s.states[folder.ID] = State{Status: Connecting}
	config := Config{Token: folder.Token, WithGroups: len(folder.Groups) > 0}
	go s.sources[folder.Kind].Watch(ctx, config, func(update Update) {
		s.report(ctx, running, update)
	})
}

// report applies one update of a running watch, unless the watch stopped meanwhile.
func (s *Syncer) report(ctx context.Context, running *watch, update Update) {
	var peers []Peer
	if update.Status == Synced {
		peers = narrow(update.Peers, running.folder.Groups)
	}
	s.mu.Lock()
	if ctx.Err() != nil || s.watches[running.folder.ID] != running {
		s.mu.Unlock()
		return
	}
	state := s.states[running.folder.ID]
	state.Status = update.Status
	if update.Status == Synced {
		state.Hosts = len(peers)
		state.Links = make(map[string]Link, len(peers))
		for _, peer := range peers {
			state.Links[peer.Key] = peer.Link
		}
	}
	s.states[running.folder.ID] = state
	s.mu.Unlock()
	if update.Status == Synced {
		s.apply(running.folder.ID, peers)
	}
	s.changed()
}

// narrow keeps the peers in any of groups; every peer when groups is empty.
func narrow(peers []Peer, groups []string) []Peer {
	if len(groups) == 0 {
		return peers
	}
	return slices.DeleteFunc(slices.Clone(peers), func(peer Peer) bool {
		return !slices.ContainsFunc(peer.Groups, func(group string) bool { return slices.Contains(groups, group) })
	})
}

// States are the sync states of the followed folders, by folder id.
func (s *Syncer) States() map[string]State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return maps.Clone(s.states)
}

// Availability is whether a source can be chosen on this device.
type Availability struct {
	Kind Kind `json:"kind"`
	// Installed is true when the source's client is on this device.
	Installed bool `json:"installed"`
}

// Available lists every source, in the order they were given, with whether its client is
// installed.
func (s *Syncer) Available() []Availability {
	listed := make([]Availability, len(s.order))
	for i, kind := range s.order {
		listed[i] = Availability{Kind: kind, Installed: s.sources[kind].Installed()}
	}
	return listed
}
