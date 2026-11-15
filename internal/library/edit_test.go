package library

import (
	"errors"
	"slices"
	"testing"
)

// tree builds: root [prod [db [primary], web], bastion].
func tree(t *testing.T) (l Library, ids map[string]string) {
	t.Helper()
	l = Empty()
	ids = map[string]string{}
	add := func(name, parent string, create func(string, string) (string, error)) {
		id, err := create(parent, name)
		if err != nil {
			t.Fatal(err)
		}
		ids[name] = id
	}
	add("prod", "", l.CreateFolder)
	add("db", ids["prod"], l.CreateFolder)
	add("primary", ids["db"], l.CreateHost)
	add("web", ids["prod"], l.CreateHost)
	add("bastion", "", l.CreateHost)
	return l, ids
}

func names(l Library, parent string) []string {
	var named []string
	for _, id := range l.children(parent) {
		if folder, ok := l.Folders[id]; ok {
			named = append(named, folder.Name)
		} else {
			named = append(named, l.Hosts[id].Name)
		}
	}
	return named
}

func TestCreatedItemsGoAtTheEndOfTheirFolder(t *testing.T) {
	l, ids := tree(t)
	if got := names(l, ""); !slices.Equal(got, []string{"prod", "bastion"}) {
		t.Errorf("root = %q", got)
	}
	if got := names(l, ids["prod"]); !slices.Equal(got, []string{"db", "web"}) {
		t.Errorf("prod = %q", got)
	}
	if _, err := l.CreateHost(ids["bastion"], "inside a host"); !errors.Is(err, errNotAFolder) {
		t.Errorf("a host inside a host: %v", err)
	}
	if _, err := l.CreateFolder("", "  "); !errors.Is(err, errEmptyName) {
		t.Errorf("a blank name: %v", err)
	}
}

func TestMovePlacesItemsWhereTheyAreDropped(t *testing.T) {
	tests := []struct {
		name   string
		moving []string
		parent string
		index  int
		want   map[string][]string
	}{
		{"onto a folder", []string{"bastion"}, "prod", 2, map[string][]string{"": {"prod"}, "prod": {"db", "web", "bastion"}}},
		{"between items", []string{"bastion"}, "prod", 1, map[string][]string{"prod": {"db", "bastion", "web"}}},
		{"to the root", []string{"primary"}, "", 0, map[string][]string{"": {"primary", "prod", "bastion"}, "db": nil}},
		{"down within a folder", []string{"db"}, "prod", 2, map[string][]string{"prod": {"web", "db"}}},
		{"a folder with its contents", []string{"db"}, "", 2, map[string][]string{"": {"prod", "bastion", "db"}, "db": {"primary"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l, ids := tree(t)
			var moving []string
			for _, name := range test.moving {
				moving = append(moving, ids[name])
			}
			if err := l.Move(moving, ids[test.parent], test.index); err != nil {
				t.Fatal(err)
			}
			for parent, want := range test.want {
				if got := names(l, ids[parent]); !slices.Equal(got, want) {
					t.Errorf("%q holds %q, want %q", parent, got, want)
				}
			}
		})
	}
}

func TestAFolderNeverMovesIntoItself(t *testing.T) {
	for _, target := range []string{"prod", "db"} {
		l, ids := tree(t)
		before := l.clone()
		if err := l.Move([]string{ids["prod"]}, ids[target], 0); !errors.Is(err, errMoveIntoItself) {
			t.Errorf("prod into %s: %v", target, err)
		}
		if !slices.Equal(l.Root, before.Root) || !slices.Equal(l.Folders[ids["prod"]].Children, before.Folders[ids["prod"]].Children) {
			t.Errorf("a refused move into %s changed the tree", target)
		}
	}
}

func TestDeletingAFolderTakesEverythingInIt(t *testing.T) {
	l, ids := tree(t)
	jump := ids["primary"]
	if err := l.SetSettings(ids["bastion"], Settings{JumpHostID: &jump}); err != nil {
		t.Fatal(err)
	}
	removed, err := l.Delete(ids["prod"])
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(removed)
	want := []string{ids["prod"], ids["db"], ids["primary"], ids["web"]}
	slices.Sort(want)
	if !slices.Equal(removed, want) {
		t.Errorf("removed %q, want %q", removed, want)
	}
	if got := names(l, ""); !slices.Equal(got, []string{"bastion"}) {
		t.Errorf("root = %q", got)
	}
	if len(l.Folders) != 0 || len(l.Hosts) != 1 {
		t.Errorf("%d folders and %d hosts left", len(l.Folders), len(l.Hosts))
	}
	if l.Hosts[ids["bastion"]].Settings.JumpHostID != nil {
		t.Error("a host still jumps through a deleted host")
	}
}

func TestDuplicateCopiesAHostRightAfterItWithoutItsPassword(t *testing.T) {
	l, ids := tree(t)
	user, saved := "deploy", true
	if err := l.SetSettings(ids["web"], Settings{User: &user}); err != nil {
		t.Fatal(err)
	}
	if err := l.SetAddress(ids["web"], " web.example.com "); err != nil {
		t.Fatal(err)
	}
	if err := l.setPassword(ids["web"], &saved); err != nil {
		t.Fatal(err)
	}
	id, err := l.Duplicate(ids["web"], "web copy")
	if err != nil {
		t.Fatal(err)
	}
	if got := names(l, ids["prod"]); !slices.Equal(got, []string{"db", "web", "web copy"}) {
		t.Errorf("prod = %q", got)
	}
	copied := l.Hosts[id]
	if copied.Address != "web.example.com" || *copied.Settings.User != "deploy" || copied.Settings.Password != nil {
		t.Errorf("copy = %+v", copied)
	}
	if _, err := l.Duplicate(ids["prod"], "prod copy"); !errors.Is(err, errNotAHost) {
		t.Errorf("duplicating a folder: %v", err)
	}
}

func TestFolderIconsComeFromTheSet(t *testing.T) {
	l, ids := tree(t)
	if err := l.SetIcon(ids["prod"], DatabaseIcon); err != nil {
		t.Fatal(err)
	}
	if err := l.SetIcon(ids["prod"], "rocket"); !errors.Is(err, errUnknownIcon) {
		t.Errorf("an unknown icon: %v", err)
	}
	if err := l.SetIcon(ids["web"], ServerIcon); !errors.Is(err, errNotAFolder) {
		t.Errorf("an icon on a host: %v", err)
	}
}

func TestHostIconsComeFromTheHostSet(t *testing.T) {
	l, ids := tree(t)
	if err := l.SetHostIcon(ids["prod"], DatabaseIcon); err != nil {
		t.Fatal(err)
	}
	if err := l.SetHostIcon(ids["prod"], FolderIcon); !errors.Is(err, errUnknownIcon) {
		t.Errorf("the folder icon: %v", err)
	}
	if err := l.SetHostIcon(ids["prod"], ""); err != nil || l.Folders[ids["prod"]].HostIcon != "" {
		t.Errorf("back to inheriting: %v", err)
	}
	if err := l.SetHostIcon(ids["web"], ServerIcon); !errors.Is(err, errNotAFolder) {
		t.Errorf("a host icon on a host: %v", err)
	}
}
