package hostsource

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"
)

// fakeSource reports whatever the test sends it, and the config each watch started with.
type fakeSource struct {
	updates chan Update
	configs chan Config
}

func newFakeSource() *fakeSource {
	return &fakeSource{updates: make(chan Update), configs: make(chan Config, 8)}
}

func (f *fakeSource) Kind() Kind      { return "fake" }
func (f *fakeSource) Installed() bool { return true }

func (f *fakeSource) Watch(ctx context.Context, config Config, report func(Update)) {
	f.configs <- config
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-f.updates:
			report(update)
		}
	}
}

func (f *fakeSource) Groups(context.Context, Config) ([]Group, error) {
	return nil, nil
}

// applied records the peers each folder was given.
type applied struct {
	mu    sync.Mutex
	peers map[string][]string
}

func (a *applied) apply(folder string, peers []Peer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	names := []string{}
	for _, peer := range peers {
		names = append(names, peer.Name)
	}
	a.peers[folder] = names
}

func (a *applied) of(folder string) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.peers[folder]
}

func newSyncer(t *testing.T, source *fakeSource) (*Syncer, *applied) {
	t.Helper()
	a := &applied{peers: map[string][]string{}}
	syncer := NewSyncer([]Source{source}, a.apply, func() {})
	t.Cleanup(func() { syncer.Follow(nil) })
	return syncer, a
}

// following is a syncer following folder, with the config its watch started with.
func following(t *testing.T, folder Folder) (*fakeSource, *Syncer, *applied, Config) {
	t.Helper()
	source := newFakeSource()
	syncer, a := newSyncer(t, source)
	syncer.Follow([]Folder{folder})
	return source, syncer, a, <-source.configs
}

func awaitState(t *testing.T, syncer *Syncer, folder string, status Status, hosts int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if state := syncer.States()[folder]; state.Status == status && state.Hosts == hosts {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("state %+v, want %s with %d hosts", syncer.States()[folder], status, hosts)
}

var peers = []Peer{
	{Key: "1", Name: "web", Groups: []string{"servers"}, Link: Link{Status: LinkConnected}},
	{Key: "2", Name: "laptop", Groups: []string{"people"}},
	{Key: "3", Name: "db", Groups: []string{"servers", "data"}},
}

func TestAFolderGetsTheSourcesPeers(t *testing.T) {
	source, syncer, a, _ := following(t, Folder{ID: "mesh", Kind: "fake"})
	source.updates <- Update{Status: Synced, Peers: peers}
	awaitState(t, syncer, "mesh", Synced, 3)
	if got := a.of("mesh"); len(got) != 3 {
		t.Errorf("applied %v, want every peer", got)
	}
}

func TestAGroupFilterNarrowsThePeers(t *testing.T) {
	source, syncer, a, config := following(t, Folder{ID: "mesh", Kind: "fake", Groups: []string{"servers"}, Token: "t"})
	if !config.WithGroups || config.Token != "t" {
		t.Errorf("config %+v, want groups read with the token", config)
	}
	source.updates <- Update{Status: Synced, Peers: peers}
	awaitState(t, syncer, "mesh", Synced, 2)
	if links := syncer.States()["mesh"].Links; len(links) != 2 || links["1"].Status != LinkConnected {
		t.Errorf("links %+v, want the two servers' links", links)
	}
	if got := a.of("mesh"); !slices.Equal(got, []string{"web", "db"}) {
		t.Errorf("applied %v, want the servers", got)
	}
}

func TestAStoppedSourceKeepsTheHostsItLastGave(t *testing.T) {
	source, syncer, a, _ := following(t, Folder{ID: "mesh", Kind: "fake"})
	source.updates <- Update{Status: Synced, Peers: peers}
	source.updates <- Update{Status: NotRunning}
	awaitState(t, syncer, "mesh", NotRunning, 3)
	if got := a.of("mesh"); len(got) != 3 {
		t.Errorf("applied %v after the source stopped", got)
	}
}

func TestChangingAFoldersFilterRestartsItsWatch(t *testing.T) {
	source, syncer, _, _ := following(t, Folder{ID: "mesh", Kind: "fake"})
	syncer.Follow([]Folder{{ID: "mesh", Kind: "fake"}})
	syncer.Follow([]Folder{{ID: "mesh", Kind: "fake", Groups: []string{"servers"}}})
	select {
	case config := <-source.configs:
		if !config.WithGroups {
			t.Errorf("restarted with %+v, want groups", config)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the watch was not restarted")
	}
	if len(source.configs) != 0 {
		t.Error("following the same folder again restarted its watch")
	}
}

func TestRestartingAFolderReadsTheSourceAgain(t *testing.T) {
	source, syncer, _, _ := following(t, Folder{ID: "mesh", Kind: "fake"})
	syncer.Restart("mesh")
	select {
	case <-source.configs:
	case <-time.After(2 * time.Second):
		t.Fatal("the watch was not restarted")
	}
	if state := syncer.States()["mesh"]; state.Status != Connecting {
		t.Errorf("state %+v right after a restart, want connecting", state)
	}
}

func TestAFolderNoLongerSyncedIsForgotten(t *testing.T) {
	_, syncer, _, _ := following(t, Folder{ID: "mesh", Kind: "fake"})
	syncer.Follow(nil)
	if _, ok := syncer.States()["mesh"]; ok {
		t.Error("a forgotten folder still has a state")
	}
}
