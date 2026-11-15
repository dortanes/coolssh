package library

import (
	"errors"
	"testing"
)

func pointer[T any](value T) *T {
	return &value
}

var defaults = Settings{
	User:             pointer("local"),
	Port:             pointer(22),
	KeyID:            pointer(""),
	Password:         pointer(false),
	JumpHostID:       pointer(""),
	AgentForwarding:  pointer(false),
	KeepaliveSeconds: pointer(30),
}

func mustSet(t *testing.T, l *Library, id string, settings Settings) {
	t.Helper()
	if err := l.SetSettings(id, settings); err != nil {
		t.Fatal(err)
	}
}

func mustInherit(t *testing.T, l *Library, id string) Inherited {
	t.Helper()
	inherited, err := l.Inherited(id, defaults)
	if err != nil {
		t.Fatal(err)
	}
	return inherited
}

func TestTheNearestFolderThatSetsASettingWins(t *testing.T) {
	l, ids := tree(t)
	mustSet(t, &l, ids["prod"], Settings{User: pointer("deploy"), Port: pointer(2222)})
	mustSet(t, &l, ids["db"], Settings{User: pointer("postgres")})
	inherited := mustInherit(t, &l, ids["primary"])
	if *inherited.Values.User != "postgres" || inherited.Origins.User != ids["db"] {
		t.Errorf("user %q from %q, want postgres from db", *inherited.Values.User, inherited.Origins.User)
	}
	if *inherited.Values.Port != 2222 || inherited.Origins.Port != ids["prod"] {
		t.Errorf("port %d from %q, want 2222 from prod", *inherited.Values.Port, inherited.Origins.Port)
	}
	if *inherited.Values.KeepaliveSeconds != 30 || inherited.Origins.KeepaliveSeconds != "" {
		t.Errorf("keepalive %d from %q, want the default", *inherited.Values.KeepaliveSeconds, inherited.Origins.KeepaliveSeconds)
	}
}

func TestTheDefaultKeyServesHostsNoFolderGivesAKey(t *testing.T) {
	l, ids := tree(t)
	fallback, err := l.AddKey(Key{Name: "everywhere"})
	if err != nil {
		t.Fatal(err)
	}
	work, err := l.AddKey(Key{Name: "work"})
	if err != nil {
		t.Fatal(err)
	}
	if err := l.SetDefaultKey(fallback); err != nil {
		t.Fatal(err)
	}
	mustSet(t, &l, ids["db"], Settings{KeyID: &work})
	if inherited := mustInherit(t, &l, ids["bastion"]); *inherited.Values.KeyID != fallback || inherited.Origins.KeyID != "" {
		t.Errorf("bastion takes key %q from %q, want the default key", *inherited.Values.KeyID, inherited.Origins.KeyID)
	}
	if inherited := mustInherit(t, &l, ids["primary"]); *inherited.Values.KeyID != work {
		t.Errorf("primary takes key %q, want db's", *inherited.Values.KeyID)
	}
	if err := l.DeleteKey(fallback); err != nil {
		t.Fatal(err)
	}
	if l.DefaultKeyID != "" {
		t.Error("a deleted key is still the default")
	}
	if err := l.SetDefaultKey("missing"); err == nil {
		t.Error("an unknown key became the default")
	}
}

func TestMovingAHostChangesWhatItInherits(t *testing.T) {
	l, ids := tree(t)
	mustSet(t, &l, ids["db"], Settings{User: pointer("postgres")})
	if err := l.Move([]string{ids["primary"]}, "", 0); err != nil {
		t.Fatal(err)
	}
	inherited := mustInherit(t, &l, ids["primary"])
	if *inherited.Values.User != "local" || inherited.Origins.User != "" {
		t.Errorf("user %q from %q after the move, want the default", *inherited.Values.User, inherited.Origins.User)
	}
}

func TestSettingsOutOfBoundsAreRefused(t *testing.T) {
	l, ids := tree(t)
	tests := []struct {
		name     string
		item     string
		settings Settings
		want     error
	}{
		{"port zero", "web", Settings{Port: pointer(0)}, errPortRange},
		{"port too high", "web", Settings{Port: pointer(65536)}, errPortRange},
		{"negative keepalive", "web", Settings{KeepaliveSeconds: pointer(-1)}, errKeepaliveRange},
		{"unknown key", "web", Settings{KeyID: pointer("missing")}, errUnknownKey},
		{"own jump host", "web", Settings{JumpHostID: pointer(ids["web"])}, errOwnJumpHost},
		{"a folder as jump host", "web", Settings{JumpHostID: pointer(ids["prod"])}, errNotAHost},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := l.SetSettings(ids[test.item], test.settings); !errors.Is(err, test.want) {
				t.Errorf("SetSettings = %v, want %v", err, test.want)
			}
		})
	}
}

func TestSettingsKeepTheSavedPasswordAndUnsetABlankUser(t *testing.T) {
	l, ids := tree(t)
	if err := l.setPassword(ids["web"], pointer(true)); err != nil {
		t.Fatal(err)
	}
	if err := l.SetSettings(ids["web"], Settings{User: pointer("  "), Password: pointer(false)}); err != nil {
		t.Fatal(err)
	}
	settings := l.Hosts[ids["web"]].Settings
	if settings.User != nil {
		t.Errorf("a blank user was kept as %q", *settings.User)
	}
	if settings.Password == nil || !*settings.Password {
		t.Error("the saved password was dropped by a settings change")
	}
}

func TestTheUsersDefaultsComeBeforeTheBuiltInOnes(t *testing.T) {
	l, ids := tree(t)
	if err := l.SetDefaults(Settings{User: pointer("admin"), Port: pointer(2222)}); err != nil {
		t.Fatal(err)
	}
	mustSet(t, &l, ids["prod"], Settings{Port: pointer(22)})
	inherited := mustInherit(t, &l, ids["primary"])
	if *inherited.Values.User != "admin" || inherited.Origins.User != "" {
		t.Errorf("user %q from %q, want the default admin", *inherited.Values.User, inherited.Origins.User)
	}
	if *inherited.Values.Port != 22 {
		t.Errorf("port %d, want the folder's 22 over the default", *inherited.Values.Port)
	}
	if *inherited.Values.KeepaliveSeconds != 30 {
		t.Errorf("keepalive %d, want the built-in 30", *inherited.Values.KeepaliveSeconds)
	}
}

func TestDefaultsKeepOnlyWhatEveryHostCanShare(t *testing.T) {
	l, ids := tree(t)
	err := l.SetDefaults(Settings{User: pointer("  "), JumpHostID: pointer(ids["bastion"]), Password: pointer(true)})
	if err != nil {
		t.Fatal(err)
	}
	if l.Defaults.User != nil || l.Defaults.JumpHostID != nil || l.Defaults.Password != nil {
		t.Errorf("defaults %+v, want none set", l.Defaults)
	}
	if err := l.SetDefaults(Settings{Port: pointer(70000)}); !errors.Is(err, errPortRange) {
		t.Errorf("port 70000: got %v", err)
	}
}
