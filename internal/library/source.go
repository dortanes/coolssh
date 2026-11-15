package library

import (
	"cmp"
	"slices"
)

// Source is where a synced folder takes its hosts from.
type Source struct {
	// Kind names the source, as hostsource does: "netbird".
	Kind string `json:"kind"`
	// Groups narrows the hosts to peers in any of these groups, by id; empty for every peer.
	Groups []string `json:"groups,omitempty"`
	// Token is true where an access token for the source is saved, in the keyring.
	Token bool `json:"token,omitempty"`
}

// Peer is a host as its source reports it.
type Peer struct {
	// Key identifies the peer across renames.
	Key     string
	Name    string
	Address string
	// HostKey is the SSH host key the source vouches for, in authorized_keys form; empty for none.
	HostKey string
}

const (
	errSynced         Refusal = "library_synced"
	errSourceNotEmpty Refusal = "library_source_not_empty"
	errSourceTaken    Refusal = "library_source_taken"
)

func (s *Source) clone() *Source {
	if s == nil {
		return nil
	}
	copied := *s
	copied.Groups = slices.Clone(s.Groups)
	return &copied
}

// synced reports whether id is a folder that takes its hosts from a source, or a host such a
// folder holds; the source owns both.
func (l *Library) synced(id string) bool {
	if folder, ok := l.Folders[id]; ok {
		return folder.Source != nil
	}
	return l.Hosts[id].Peer != ""
}

// SetSource makes an empty folder take its hosts from source, one folder per kind; nil keeps its
// hosts as ordinary ones. The token flag stays as it is saved.
func (l *Library) SetSource(id string, source *Source) error {
	folder, ok := l.Folders[id]
	if !ok {
		return errNotAFolder
	}
	if source == nil {
		folder.Source = nil
		for _, child := range folder.Children {
			host := l.Hosts[child]
			host.Peer = ""
			l.Hosts[child] = host
		}
		l.Folders[id] = folder
		return nil
	}
	source = source.clone()
	if folder.Source == nil && len(folder.Children) > 0 {
		return errSourceNotEmpty
	}
	for other, held := range l.Folders {
		if other != id && held.Source != nil && held.Source.Kind == source.Kind {
			return errSourceTaken
		}
	}
	source.Token = folder.Source != nil && folder.Source.Token
	folder.Source = source
	l.Folders[id] = folder
	return nil
}

// setSourceToken records whether an access token is saved for a synced folder.
func (l *Library) setSourceToken(id string, saved bool) error {
	folder, ok := l.Folders[id]
	if !ok || folder.Source == nil {
		return errNotAFolder
	}
	folder.Source = folder.Source.clone()
	folder.Source.Token = saved
	l.Folders[id] = folder
	return nil
}

// placePeer gives peer the host hostID, a new one when hostID is empty, and returns the host's id
// and whether that changed it.
func (l *Library) placePeer(hostID string, peer Peer) (string, bool) {
	if hostID == "" {
		hostID = newID()
	}
	host, known := l.Hosts[hostID]
	if known && host.Name == peer.Name && host.Address == peer.Address && host.HostKey == peer.HostKey {
		return hostID, false
	}
	host.ID, host.Name, host.Address, host.Peer, host.HostKey = hostID, peer.Name, peer.Address, peer.Key, peer.HostKey
	l.Hosts[hostID] = host
	return hostID, true
}

// SyncPeers makes a synced folder hold exactly peers, in name order, and returns the hosts it
// removed; changed is false when the folder already matched.
func (l *Library) SyncPeers(id string, peers []Peer) (removed []string, changed bool, err error) {
	folder, ok := l.Folders[id]
	if !ok || folder.Source == nil {
		return nil, false, errNotAFolder
	}
	byKey := map[string]string{}
	for _, child := range folder.Children {
		byKey[l.Hosts[child].Peer] = child
	}
	peers = slices.SortedFunc(slices.Values(peers), func(a, b Peer) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Key, b.Key))
	})
	children := make([]string, 0, len(peers))
	for _, peer := range peers {
		hostID, placed := l.placePeer(byKey[peer.Key], peer)
		changed = changed || placed
		delete(byKey, peer.Key)
		children = append(children, hostID)
	}
	for _, gone := range byKey {
		removed = append(removed, gone)
		delete(l.Hosts, gone)
	}
	l.forgetJumpHosts(removed)
	changed = changed || !slices.Equal(children, folder.Children)
	folder.Children = children
	l.Folders[id] = folder
	return removed, changed, nil
}
