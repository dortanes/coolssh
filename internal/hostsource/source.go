// Package hostsource keeps synced folders in step with the mesh networks their hosts come from.
// Each network is a Source; adding one is a package implementing Source, registered in main.
package hostsource

import (
	"context"
	"errors"
)

// Kind names a source, as a synced folder records it.
type Kind string

// Peer is a machine a source's client knows.
type Peer struct {
	// Key identifies the peer across renames.
	Key     string
	Name    string
	Address string
	// Groups are the ids of the groups the peer is in; empty when the source did not read them.
	Groups []string
	// HostKey is the SSH host key the source vouches for at the peer's address, in authorized_keys
	// form; empty when it knows none.
	HostKey string
	Link    Link
}

// LinkStatus is whether this device is connected to a peer over the mesh.
type LinkStatus string

const (
	LinkConnected  LinkStatus = "connected"
	LinkConnecting LinkStatus = "connecting"
	// LinkIdle is a peer the client connects to only once traffic needs it.
	LinkIdle LinkStatus = "idle"
)

// Link is how this device reaches a peer over the mesh, as its client last reported.
type Link struct {
	Status LinkStatus `json:"status"`
	// Relayed is true when traffic goes through a relay rather than straight to the peer.
	Relayed bool `json:"relayed"`
	// LatencyMs is the round trip to the peer in milliseconds; 0 when not measured.
	LatencyMs int `json:"latencyMs"`
}

// Group is a group of peers a folder can narrow its hosts to.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Status is where a folder's sync stands.
type Status string

const (
	Connecting Status = "connecting"
	Synced     Status = "synced"
	// NotRunning is a source whose client is installed and not running.
	NotRunning Status = "not-running"
	// Disconnected is a client that runs and is not connected to its network.
	Disconnected Status = "disconnected"
	// Unauthorized is an access token the source refused.
	Unauthorized Status = "unauthorized"
	// GroupsUnavailable is a group filter whose groups could not be read.
	GroupsUnavailable Status = "groups-unavailable"
)

// ErrUnauthorized reports an access token the source refused.
var ErrUnauthorized = errors.New("access token refused")

// Config is what a source needs to watch for one folder.
type Config struct {
	// Token is the access token the user saved for the folder; empty for none.
	Token string
	// WithGroups asks for each peer's groups, which a group filter needs.
	WithGroups bool
}

// Update is what a source reports: its status and, while Synced, every peer.
type Update struct {
	Status Status
	Peers  []Peer
}

// Source is a mesh network whose client lists the peers this device can reach.
type Source interface {
	Kind() Kind
	// Installed reports whether the source's client is on this device.
	Installed() bool
	// Watch calls report with the source's state and peers, at once and on every change, until
	// ctx ends. It never returns early: an unreachable client is reported and retried.
	Watch(ctx context.Context, config Config, report func(Update))
	// Groups lists the groups a folder can narrow its hosts to; ErrUnauthorized for a token the
	// source refused.
	Groups(ctx context.Context, config Config) ([]Group, error)
}
