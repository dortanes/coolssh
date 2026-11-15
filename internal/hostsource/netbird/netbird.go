// Package netbird is the NetBird host source: peers from the local client's daemon, and, for a
// group filter, their groups from NetBird's management API.
package netbird

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/netbirdio/netbird/client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dortanes/coolssh/internal/hostsource"
)

// Kind is NetBird's name as a synced folder records it.
const Kind hostsource.Kind = "netbird"

// The daemon's status, and a peer's connection, while connected and while connecting.
const (
	daemonConnected  = "Connected"
	daemonConnecting = "Connecting"
)

// linkStatus maps the daemon's name for a peer's connection; an unknown one reads as idle.
func linkStatus(daemon string) hostsource.LinkStatus {
	switch daemon {
	case daemonConnected:
		return hostsource.LinkConnected
	case daemonConnecting:
		return hostsource.LinkConnecting
	}
	return hostsource.LinkIdle
}

// The wait between attempts to reach a daemon that is not running, doubling up to the longest.
const (
	firstRetry   = time.Second
	longestRetry = 30 * time.Second
)

// groupsAge is how long the peers' groups are trusted before they are read again.
const groupsAge = time.Minute

// Source reads NetBird's peers from the daemon of the client on this device.
type Source struct {
	// dial connects to the daemon; the real one is the platform's socket or pipe.
	dial func() (*grpc.ClientConn, error)
	api  *api
	// clients are the places NetBird's client is installed at on this platform.
	clients []string
}

// New reads peers from the daemon this platform's NetBird client runs.
func New() *Source {
	return &Source{
		dial: func() (*grpc.ClientConn, error) {
			target, options := daemonTarget()
			return grpc.NewClient(target, append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))...)
		},
		api:     newAPI(),
		clients: installedClients,
	}
}

func (s *Source) Kind() hostsource.Kind {
	return Kind
}

// Installed reports whether the daemon is reachable or NetBird's client is at a known place.
func (s *Source) Installed() bool {
	if daemonPresent() {
		return true
	}
	for _, path := range s.clients {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	_, err := exec.LookPath("netbird")
	return err == nil
}

// Watch follows the daemon's status stream, reconnecting with backoff while the daemon is down.
func (s *Source) Watch(ctx context.Context, config hostsource.Config, report func(hostsource.Update)) {
	wait := firstRetry
	for ctx.Err() == nil {
		if s.follow(ctx, config, report) {
			wait = firstRetry
		}
		if ctx.Err() != nil {
			return
		}
		report(hostsource.Update{Status: hostsource.NotRunning})
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		wait = min(2*wait, longestRetry)
	}
}

// follow reports every status the daemon streams until the stream ends; true when it got any.
func (s *Source) follow(ctx context.Context, config hostsource.Config, report func(hostsource.Update)) bool {
	conn, err := s.dial()
	if err != nil {
		return false
	}
	defer conn.Close()
	stream, err := proto.NewDaemonServiceClient(conn).SubscribeStatus(ctx, &proto.StatusRequest{GetFullPeerStatus: true})
	if err != nil {
		return false
	}
	groups := &peerGroups{api: s.api, token: config.Token}
	received := false
	for {
		status, err := stream.Recv()
		if err != nil {
			return received
		}
		received = true
		report(s.update(ctx, status, config, groups))
	}
}

// update is what one status from the daemon means for the folder.
func (s *Source) update(ctx context.Context, status *proto.StatusResponse, config hostsource.Config, groups *peerGroups) hostsource.Update {
	switch status.GetStatus() {
	case daemonConnected:
	case daemonConnecting:
		return hostsource.Update{Status: hostsource.Connecting}
	default:
		return hostsource.Update{Status: hostsource.Disconnected}
	}
	full := status.GetFullStatus()
	peers := make([]hostsource.Peer, 0, len(full.GetPeers()))
	for _, peer := range full.GetPeers() {
		peers = append(peers, hostsource.Peer{
			Key:     peer.GetPubKey(),
			Name:    strings.SplitN(peer.GetFqdn(), ".", 2)[0],
			Address: peer.GetIP(),
			// The key of NetBird's own SSH server, which answers on port 22 while it is on.
			HostKey: strings.TrimSpace(string(peer.GetSshHostKey())),
			Link: hostsource.Link{
				Status:    linkStatus(peer.GetConnStatus()),
				Relayed:   peer.GetRelayed(),
				LatencyMs: int(peer.GetLatency().AsDuration().Milliseconds()),
			},
		})
	}
	if !config.WithGroups {
		return hostsource.Update{Status: hostsource.Synced, Peers: peers}
	}
	byAddress, err := groups.of(ctx, full.GetManagementState().GetURL(), peers)
	switch {
	case errors.Is(err, hostsource.ErrUnauthorized):
		return hostsource.Update{Status: hostsource.Unauthorized}
	case err != nil:
		return hostsource.Update{Status: hostsource.GroupsUnavailable}
	}
	for i := range peers {
		peers[i].Groups = byAddress[peers[i].Address]
	}
	return hostsource.Update{Status: hostsource.Synced, Peers: peers}
}

// Groups lists the account's groups with the folder's token, from the management server the
// daemon is connected to.
func (s *Source) Groups(ctx context.Context, config hostsource.Config) ([]hostsource.Group, error) {
	management, err := s.managementURL(ctx)
	if err != nil {
		return nil, err
	}
	return s.api.groups(ctx, management, config.Token)
}

// managementURL asks the daemon which management server it is connected to.
func (s *Source) managementURL(ctx context.Context) (string, error) {
	conn, err := s.dial()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	status, err := proto.NewDaemonServiceClient(conn).Status(ctx, &proto.StatusRequest{GetFullPeerStatus: true})
	if err != nil {
		return "", err
	}
	return status.GetFullStatus().GetManagementState().GetURL(), nil
}

// peerGroups holds the peers' groups read from the management API, read again when the peers
// change or the groups grow old.
type peerGroups struct {
	api       *api
	token     string
	byAddress map[string][]string
	addresses map[string]bool
	read      time.Time
}

func (g *peerGroups) of(ctx context.Context, management string, peers []hostsource.Peer) (map[string][]string, error) {
	addresses := map[string]bool{}
	for _, peer := range peers {
		addresses[peer.Address] = true
	}
	fresh := g.byAddress != nil && time.Since(g.read) < groupsAge && sameKeys(addresses, g.addresses)
	if fresh {
		return g.byAddress, nil
	}
	byAddress, err := g.api.peerGroups(ctx, management, g.token)
	if err != nil {
		return nil, err
	}
	g.byAddress, g.addresses, g.read = byAddress, addresses, time.Now()
	return byAddress, nil
}

func sameKeys(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for key := range a {
		if !b[key] {
			return false
		}
	}
	return true
}
