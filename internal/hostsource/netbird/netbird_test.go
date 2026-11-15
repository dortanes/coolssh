package netbird

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/netbirdio/netbird/client/proto"
	nbapi "github.com/netbirdio/netbird/shared/management/http/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/dortanes/coolssh/internal/hostsource"
)

const token = "nbp_test"

// fakeDaemon streams the statuses a test sends it, as NetBird's daemon does.
type fakeDaemon struct {
	proto.UnimplementedDaemonServiceServer
	statuses chan *proto.StatusResponse
	current  *proto.StatusResponse
}

func (d *fakeDaemon) Status(context.Context, *proto.StatusRequest) (*proto.StatusResponse, error) {
	return d.current, nil
}

func (d *fakeDaemon) SubscribeStatus(_ *proto.StatusRequest, stream grpc.ServerStreamingServer[proto.StatusResponse]) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case status := <-d.statuses:
			if err := stream.Send(status); err != nil {
				return err
			}
		}
	}
}

// fakeManagement serves the management API's peers and groups to the test token only.
func fakeManagement(t *testing.T) *httptest.Server {
	t.Helper()
	serve := func(body any) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Token "+token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(body)
		}
	}
	mux := http.NewServeMux()
	mux.Handle("/api/peers", serve([]nbapi.Peer{
		{Ip: "100.64.0.2", Groups: []nbapi.GroupMinimum{{Id: "servers", Name: "Servers"}}},
		{Ip: "100.64.0.3", Groups: []nbapi.GroupMinimum{{Id: "people", Name: "People"}}},
	}))
	mux.Handle("/api/groups", serve([]nbapi.Group{{Id: "servers", Name: "Servers"}, {Id: "people", Name: "People"}}))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func connected(management string) *proto.StatusResponse {
	return &proto.StatusResponse{
		Status: daemonConnected,
		FullStatus: &proto.FullStatus{
			ManagementState: &proto.ManagementState{URL: management, Connected: true},
			Peers: []*proto.PeerState{
				{PubKey: "key-web", Fqdn: "web.netbird.cloud", IP: "100.64.0.2", SshHostKey: []byte("ssh-ed25519 AAAA\n"), ConnStatus: "Connected", Relayed: true, Latency: durationpb.New(12 * time.Millisecond)},
				{PubKey: "key-laptop", Fqdn: "laptop.netbird.cloud", IP: "100.64.0.3"},
			},
		},
	}
}

// newSource is a Source talking to a fake daemon in memory.
func newSource(t *testing.T, daemon *fakeDaemon) *Source {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	proto.RegisterDaemonServiceServer(server, daemon)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	return &Source{dial: dialer(listener), api: newAPI()}
}

// dialer connects to a daemon listening in memory.
func dialer(listener *bufconn.Listener) func() (*grpc.ClientConn, error) {
	return func() (*grpc.ClientConn, error) {
		return grpc.NewClient("passthrough:///daemon",
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
}

func watch(t *testing.T, source *Source, config hostsource.Config) <-chan hostsource.Update {
	t.Helper()
	updates := make(chan hostsource.Update, 16)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go source.Watch(ctx, config, func(update hostsource.Update) { updates <- update })
	return updates
}

func next(t *testing.T, updates <-chan hostsource.Update) hostsource.Update {
	t.Helper()
	select {
	case update := <-updates:
		return update
	case <-time.After(5 * time.Second):
		t.Fatal("no update")
		return hostsource.Update{}
	}
}

func TestPeersAreNamedAfterTheirFQDNAndReachedAtTheirIP(t *testing.T) {
	daemon := &fakeDaemon{statuses: make(chan *proto.StatusResponse, 1)}
	updates := watch(t, newSource(t, daemon), hostsource.Config{})
	daemon.statuses <- connected("https://management.example.com:443")
	update := next(t, updates)
	if update.Status != hostsource.Synced || len(update.Peers) != 2 {
		t.Fatalf("update %+v, want two synced peers", update)
	}
	if web := update.Peers[0]; web.Name != "web" || web.Address != "100.64.0.2" || web.Key != "key-web" || web.HostKey != "ssh-ed25519 AAAA" {
		t.Errorf("peer %+v, want web at 100.64.0.2 with its host key", web)
	}
	want := hostsource.Link{Status: hostsource.LinkConnected, Relayed: true, LatencyMs: 12}
	if web, laptop := update.Peers[0].Link, update.Peers[1].Link; web != want || laptop.Status != hostsource.LinkIdle {
		t.Errorf("links %+v and %+v, want web connected over a relay and laptop idle", web, laptop)
	}
}

func TestAGroupFilterGetsEachPeersGroups(t *testing.T) {
	management := fakeManagement(t)
	daemon := &fakeDaemon{statuses: make(chan *proto.StatusResponse, 1)}
	updates := watch(t, newSource(t, daemon), hostsource.Config{Token: token, WithGroups: true})
	daemon.statuses <- connected(management.URL)
	update := next(t, updates)
	if update.Status != hostsource.Synced || !slices.Equal(update.Peers[0].Groups, []string{"servers"}) {
		t.Errorf("update %+v, want web in servers", update)
	}
}

func TestARefusedTokenIsReported(t *testing.T) {
	management := fakeManagement(t)
	daemon := &fakeDaemon{statuses: make(chan *proto.StatusResponse, 1)}
	updates := watch(t, newSource(t, daemon), hostsource.Config{Token: "wrong", WithGroups: true})
	daemon.statuses <- connected(management.URL)
	if update := next(t, updates); update.Status != hostsource.Unauthorized {
		t.Errorf("status %s, want unauthorized", update.Status)
	}
}

func TestADaemonOffItsNetworkIsDisconnected(t *testing.T) {
	daemon := &fakeDaemon{statuses: make(chan *proto.StatusResponse, 1)}
	updates := watch(t, newSource(t, daemon), hostsource.Config{})
	daemon.statuses <- &proto.StatusResponse{Status: "NeedsLogin"}
	if update := next(t, updates); update.Status != hostsource.Disconnected {
		t.Errorf("status %s, want disconnected", update.Status)
	}
}

func TestAMissingDaemonIsNotRunning(t *testing.T) {
	listener := bufconn.Listen(1)
	listener.Close()
	source := &Source{dial: dialer(listener), api: newAPI()}
	if update := next(t, watch(t, source, hostsource.Config{})); update.Status != hostsource.NotRunning {
		t.Errorf("status %s, want not running", update.Status)
	}
}

func TestGroupsAreListedWithTheToken(t *testing.T) {
	management := fakeManagement(t)
	daemon := &fakeDaemon{current: connected(management.URL)}
	groups, err := newSource(t, daemon).Groups(context.Background(), hostsource.Config{Token: token})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 || groups[0] != (hostsource.Group{ID: "servers", Name: "Servers"}) {
		t.Errorf("groups %+v", groups)
	}
}
