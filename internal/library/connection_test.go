package library

import (
	"errors"
	"testing"
)

func mustAddress(t *testing.T, l *Library, id, address string) {
	t.Helper()
	if err := l.SetAddress(id, address); err != nil {
		t.Fatal(err)
	}
}

func TestAConnectionTakesTheHostsOwnSettingsOverInheritedOnes(t *testing.T) {
	l, ids := tree(t)
	mustAddress(t, &l, ids["primary"], "db.example.com")
	mustSet(t, &l, ids["prod"], Settings{User: pointer("deploy"), Port: pointer(2222)})
	mustSet(t, &l, ids["primary"], Settings{Port: pointer(2200)})
	connection, err := l.Connection(ids["primary"], defaults)
	if err != nil {
		t.Fatal(err)
	}
	if connection.Address != "db.example.com" || connection.User != "deploy" || connection.Port != 2200 {
		t.Errorf("got %s@%s:%d, want deploy@db.example.com:2200", connection.User, connection.Address, connection.Port)
	}
	if connection.KeepaliveSeconds != 30 || connection.Key != nil || connection.Jump != nil {
		t.Errorf("got keepalive %d, key %v, jump %v; want the defaults", connection.KeepaliveSeconds, connection.Key, connection.Jump)
	}
}

func TestAConnectionNamesTheItemWhosePasswordItUses(t *testing.T) {
	l, ids := tree(t)
	mustAddress(t, &l, ids["primary"], "db.example.com")
	if err := l.setPassword(ids["prod"], pointer(true)); err != nil {
		t.Fatal(err)
	}
	connection, err := l.Connection(ids["primary"], defaults)
	if err != nil {
		t.Fatal(err)
	}
	if connection.PasswordItem != ids["prod"] {
		t.Errorf("password item %q, want prod", connection.PasswordItem)
	}
	if err := l.setPassword(ids["primary"], pointer(true)); err != nil {
		t.Fatal(err)
	}
	connection, _ = l.Connection(ids["primary"], defaults)
	if connection.PasswordItem != ids["primary"] {
		t.Errorf("password item %q, want the host's own", connection.PasswordItem)
	}
}

func TestAConnectionResolvesItsJumpHostWithTheJumpHostsSettings(t *testing.T) {
	l, ids := tree(t)
	mustAddress(t, &l, ids["primary"], "db.internal")
	mustAddress(t, &l, ids["bastion"], "bastion.example.com")
	mustSet(t, &l, ids["prod"], Settings{JumpHostID: pointer(ids["bastion"]), User: pointer("deploy")})
	mustSet(t, &l, ids["bastion"], Settings{User: pointer("jump")})
	connection, err := l.Connection(ids["primary"], defaults)
	if err != nil {
		t.Fatal(err)
	}
	if connection.Jump == nil || connection.Jump.Address != "bastion.example.com" || connection.Jump.User != "jump" {
		t.Fatalf("jump %+v, want jump@bastion.example.com", connection.Jump)
	}
}

func TestAConnectionIsRefused(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(t *testing.T, l *Library, ids map[string]string)
		want    error
	}{
		{"without an address", func(*testing.T, *Library, map[string]string) {}, errNoAddress},
		{"through a loop of jump hosts", func(t *testing.T, l *Library, ids map[string]string) {
			mustAddress(t, l, ids["web"], "web.example.com")
			mustAddress(t, l, ids["bastion"], "bastion.example.com")
			mustSet(t, l, ids["web"], Settings{JumpHostID: pointer(ids["bastion"])})
			mustSet(t, l, ids["bastion"], Settings{JumpHostID: pointer(ids["web"])})
		}, errJumpLoop},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l, ids := tree(t)
			c.prepare(t, &l, ids)
			if _, err := l.Connection(ids["web"], defaults); !errors.Is(err, c.want) {
				t.Errorf("got %v, want %v", err, c.want)
			}
		})
	}
}

func TestAnAddressConnectionTakesTheAddressOverTheDefaults(t *testing.T) {
	l, _ := tree(t)
	l.Defaults = Settings{User: pointer("admin")}
	cases := []struct {
		target, user, address string
		port                  int
	}{
		{"host.example.com", "admin", "host.example.com", 22},
		{"deploy@host.example.com", "deploy", "host.example.com", 22},
		{"deploy@host.example.com:2222", "deploy", "host.example.com", 2222},
		{"[2001:db8::1]:2200", "admin", "2001:db8::1", 2200},
		{"2001:db8::1", "admin", "2001:db8::1", 22},
	}
	for _, c := range cases {
		connection, err := l.AddressConnection(c.target, defaults)
		if err != nil {
			t.Errorf("%s: %v", c.target, err)
			continue
		}
		if connection.User != c.user || connection.Address != c.address || connection.Port != c.port {
			t.Errorf("%s: got %s@%s:%d", c.target, connection.User, connection.Address, connection.Port)
		}
		if connection.Name != c.target || connection.HostID != "" || connection.PasswordItem != "" {
			t.Errorf("%s: got name %q, host %q, password %q", c.target, connection.Name, connection.HostID, connection.PasswordItem)
		}
	}
	for target, want := range map[string]error{
		"host.example.com:70000": errPortRange,
		"deploy@":                errNoAddress,
		"two words":              errNoAddress,
	} {
		if _, err := l.AddressConnection(target, defaults); !errors.Is(err, want) {
			t.Errorf("%s: %v, want %v", target, err, want)
		}
	}
}
