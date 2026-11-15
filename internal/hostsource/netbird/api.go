package netbird

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	nbapi "github.com/netbirdio/netbird/shared/management/http/api"

	"github.com/dortanes/coolssh/internal/hostsource"
)

// api reads the management server's REST API with a personal access token.
type api struct {
	client *http.Client
}

func newAPI() *api {
	return &api{client: &http.Client{Timeout: 15 * time.Second}}
}

// peerGroups returns the ids of each peer's groups, by the peer's NetBird IP.
func (a *api) peerGroups(ctx context.Context, management, token string) (map[string][]string, error) {
	var peers []nbapi.Peer
	if err := a.get(ctx, management, "/api/peers", token, &peers); err != nil {
		return nil, err
	}
	byAddress := make(map[string][]string, len(peers))
	for _, peer := range peers {
		for _, group := range peer.Groups {
			byAddress[peer.Ip] = append(byAddress[peer.Ip], group.Id)
		}
	}
	return byAddress, nil
}

func (a *api) groups(ctx context.Context, management, token string) ([]hostsource.Group, error) {
	var groups []nbapi.Group
	if err := a.get(ctx, management, "/api/groups", token, &groups); err != nil {
		return nil, err
	}
	listed := make([]hostsource.Group, len(groups))
	for i, group := range groups {
		listed[i] = hostsource.Group{ID: group.Id, Name: group.Name}
	}
	return listed, nil
}

// get decodes the JSON at path on the management server's host into into.
func (a *api) get(ctx context.Context, management, path, token string, into any) error {
	base, err := url.Parse(management)
	if err != nil {
		return fmt.Errorf("management url %q: %w", management, err)
	}
	if base.Host == "" {
		return fmt.Errorf("management url %q has no host", management)
	}
	endpoint := url.URL{Scheme: "https", Host: base.Host, Path: path}
	if base.Scheme == "http" {
		endpoint.Scheme = "http"
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Token "+token)
	request.Header.Set("Accept", "application/json")
	response, err := a.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	switch {
	case response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden:
		return hostsource.ErrUnauthorized
	case response.StatusCode != http.StatusOK:
		return fmt.Errorf("%s: %s", path, response.Status)
	}
	return json.NewDecoder(response.Body).Decode(into)
}
