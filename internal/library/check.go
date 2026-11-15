package library

import (
	"fmt"
	"slices"
)

// check refuses a library read from disk that this package would never write: another version,
// an item listed twice, missing, out of the tree, or settings out of bounds.
func (l *Library) check() error {
	if l.Version != fileVersion {
		return fmt.Errorf("version %d, want %d", l.Version, fileVersion)
	}
	placed := map[string]bool{}
	if err := l.place("", placed); err != nil {
		return err
	}
	if len(placed) != len(l.Folders)+len(l.Hosts) {
		return fmt.Errorf("%d items outside the tree", len(l.Folders)+len(l.Hosts)-len(placed))
	}
	if err := l.checkKeys(); err != nil {
		return err
	}
	if err := l.checkSources(); err != nil {
		return err
	}
	if err := checkBounds(l.Defaults); err != nil {
		return fmt.Errorf("defaults: %w", err)
	}
	for id := range placed {
		if err := l.checkSettings(id, l.settingsOf(id)); err != nil {
			return fmt.Errorf("item %s: %w", id, err)
		}
	}
	return nil
}

// place walks the tree from parent, recording every item and refusing one met twice or unknown.
func (l *Library) place(parent string, placed map[string]bool) error {
	for _, id := range l.children(parent) {
		if placed[id] {
			return fmt.Errorf("item %s listed twice", id)
		}
		if !l.exists(id) {
			return fmt.Errorf("item %s listed but missing", id)
		}
		placed[id] = true
		if l.isFolder(id) {
			if err := l.place(id, placed); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Library) checkKeys() error {
	for index, key := range l.Keys {
		if slices.ContainsFunc(l.Keys[:index], func(earlier Key) bool { return earlier.ID == key.ID }) {
			return fmt.Errorf("key %s listed twice", key.ID)
		}
	}
	if l.DefaultKeyID != "" && !l.hasKey(l.DefaultKeyID) {
		return fmt.Errorf("default key %s missing", l.DefaultKeyID)
	}
	return nil
}

// checkSources refuses a synced folder holding anything but its source's hosts, a synced host
// outside one, and two folders syncing one kind.
func (l *Library) checkSources() error {
	kinds := map[string]bool{}
	for id, folder := range l.Folders {
		if folder.Source == nil {
			continue
		}
		if kinds[folder.Source.Kind] {
			return fmt.Errorf("source %s synced twice", folder.Source.Kind)
		}
		kinds[folder.Source.Kind] = true
		for _, child := range folder.Children {
			if l.Hosts[child].Peer == "" {
				return fmt.Errorf("synced folder %s holds %s", id, child)
			}
		}
	}
	for id, host := range l.Hosts {
		if host.Peer != "" && !l.synced(l.parentOf(id)) {
			return fmt.Errorf("synced host %s outside a synced folder", id)
		}
	}
	return nil
}

func (l *Library) settingsOf(id string) Settings {
	if folder, ok := l.Folders[id]; ok {
		return folder.Settings
	}
	return l.Hosts[id].Settings
}
