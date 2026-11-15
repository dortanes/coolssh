package library

import (
	"errors"
	"slices"
	"testing"
)

const netbird = "netbird"

// syncedFolder adds an empty folder taking its hosts from NetBird.
func syncedFolder(t *testing.T, l *Library) string {
	t.Helper()
	id, err := l.CreateFolder("", "mesh")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.SetSource(id, &Source{Kind: netbird}); err != nil {
		t.Fatal(err)
	}
	return id
}

func mustSync(t *testing.T, l *Library, folder string, peers ...Peer) ([]string, bool) {
	t.Helper()
	removed, changed, err := l.SyncPeers(folder, peers)
	if err != nil {
		t.Fatal(err)
	}
	return removed, changed
}

// syncedHost adds a NetBird folder holding one host, web, and returns both.
func syncedHost(t *testing.T, l *Library) (folder, host string) {
	t.Helper()
	folder = syncedFolder(t, l)
	mustSync(t, l, folder, Peer{Key: "k1", Name: "web", Address: "100.64.0.2"})
	return folder, l.Folders[folder].Children[0]
}

func hostNames(l Library, folder string) []string {
	var named []string
	for _, id := range l.Folders[folder].Children {
		named = append(named, l.Hosts[id].Name+"@"+l.Hosts[id].Address)
	}
	return named
}

func TestSyncingPeersAddsUpdatesAndRemovesHostsInNameOrder(t *testing.T) {
	l := Empty()
	folder := syncedFolder(t, &l)
	mustSync(t, &l, folder, Peer{Key: "k1", Name: "web", Address: "100.64.0.2"}, Peer{Key: "k2", Name: "db", Address: "100.64.0.3"})
	if got, want := hostNames(l, folder), []string{"db@100.64.0.3", "web@100.64.0.2"}; !slices.Equal(got, want) {
		t.Fatalf("hosts %v, want %v", got, want)
	}
	db := l.Folders[folder].Children[0]
	removed, _ := mustSync(t, &l, folder, Peer{Key: "k2", Name: "postgres", Address: "100.64.0.3"}, Peer{Key: "k3", Name: "cache", Address: "100.64.0.4"})
	if got, want := hostNames(l, folder), []string{"cache@100.64.0.4", "postgres@100.64.0.3"}; !slices.Equal(got, want) {
		t.Errorf("hosts %v, want %v", got, want)
	}
	if l.Folders[folder].Children[1] != db {
		t.Error("the renamed peer got a new host")
	}
	if len(removed) != 1 {
		t.Errorf("removed %v, want the web host", removed)
	}
}

func TestANewHostKeyFromTheSourceIsAChange(t *testing.T) {
	l := Empty()
	folder := syncedFolder(t, &l)
	peer := Peer{Key: "k1", Name: "web", Address: "100.64.0.2"}
	mustSync(t, &l, folder, peer)
	peer.HostKey = "ssh-ed25519 AAAA"
	if _, changed := mustSync(t, &l, folder, peer); !changed {
		t.Error("a new host key was not a change")
	}
	if got := l.Hosts[l.Folders[folder].Children[0]].HostKey; got != peer.HostKey {
		t.Errorf("host key %q, want the source's", got)
	}
}

func TestSyncingTheSamePeersChangesNothing(t *testing.T) {
	l := Empty()
	folder := syncedFolder(t, &l)
	peers := []Peer{{Key: "k1", Name: "web", Address: "100.64.0.2"}}
	mustSync(t, &l, folder, peers...)
	if _, changed := mustSync(t, &l, folder, peers...); changed {
		t.Error("an identical sync reported a change")
	}
}

func TestASyncedHostKeepsItsSettings(t *testing.T) {
	l := Empty()
	folder, host := syncedHost(t, &l)
	mustSet(t, &l, host, Settings{User: pointer("deploy")})
	mustSync(t, &l, folder, Peer{Key: "k1", Name: "frontend", Address: "100.64.0.2"})
	if got := l.Hosts[host]; got.Name != "frontend" || got.Settings.User == nil || *got.Settings.User != "deploy" {
		t.Errorf("host %+v, want frontend keeping user deploy", got)
	}
}

func TestTheSourceOwnsASyncedFolder(t *testing.T) {
	l, ids := tree(t)
	folder, host := syncedHost(t, &l)
	refused := map[string]error{
		"creating a host in it": func() error { _, err := l.CreateHost(folder, "x"); return err }(),
		"moving a host into it": l.Move([]string{ids["web"]}, folder, 0),
		"moving its host out":   l.Move([]string{host}, "", 0),
		"renaming its host":     l.Rename(host, "x"),
		"readdressing its host": l.SetAddress(host, "x"),
		"deleting its host":     func() error { _, err := l.Delete(host); return err }(),
		"duplicating its host":  func() error { _, err := l.Duplicate(host, "x"); return err }(),
	}
	for what, err := range refused {
		if !errors.Is(err, errSynced) {
			t.Errorf("%s: got %v, want the synced refusal", what, err)
		}
	}
}

func TestChoosingASourceIsRefused(t *testing.T) {
	l, ids := tree(t)
	syncedFolder(t, &l)
	other, err := l.CreateFolder("", "second")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.SetSource(ids["prod"], &Source{Kind: "other"}); !errors.Is(err, errSourceNotEmpty) {
		t.Errorf("a folder with hosts: got %v", err)
	}
	if err := l.SetSource(other, &Source{Kind: netbird}); !errors.Is(err, errSourceTaken) {
		t.Errorf("a second NetBird folder: got %v", err)
	}
}

func TestTurningTheSourceOffKeepsOrdinaryHosts(t *testing.T) {
	l := Empty()
	folder, host := syncedHost(t, &l)
	if err := l.SetSource(folder, nil); err != nil {
		t.Fatal(err)
	}
	if err := l.Rename(host, "mine"); err != nil {
		t.Errorf("renaming a former synced host: %v", err)
	}
	if err := l.check(); err != nil {
		t.Errorf("the library no longer checks: %v", err)
	}
}

func TestAFileWithAHostAddedToASyncedFolderIsRefused(t *testing.T) {
	l := Empty()
	folder := syncedFolder(t, &l)
	stray := Host{ID: "stray", Name: "stray"}
	l.Hosts[stray.ID] = stray
	l.setChildren(folder, []string{stray.ID})
	if err := l.check(); err == nil {
		t.Error("check accepted an ordinary host in a synced folder")
	}
}
