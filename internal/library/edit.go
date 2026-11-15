package library

import (
	"slices"
	"strings"
)

// CreateFolder adds an empty, expanded folder at the end of parent and returns its id.
func (l *Library) CreateFolder(parent, name string) (string, error) {
	return l.insert(parent, name, func(id, name string) {
		l.Folders[id] = Folder{ID: id, Name: name, Icon: FolderIcon, Expanded: true, Children: []string{}}
	})
}

// CreateHost adds a host without an address at the end of parent and returns its id.
func (l *Library) CreateHost(parent, name string) (string, error) {
	return l.insert(parent, name, func(id, name string) {
		l.Hosts[id] = Host{ID: id, Name: name}
	})
}

// insert has add record a new item under a new id and name, puts it at the end of parent and
// returns its id.
func (l *Library) insert(parent, name string, add func(id, name string)) (string, error) {
	if err := l.checkParent(parent); err != nil {
		return "", err
	}
	name, err := checkName(name)
	if err != nil {
		return "", err
	}
	id := newID()
	add(id, name)
	l.setChildren(parent, append(slices.Clone(l.children(parent)), id))
	return id, nil
}

// Rename gives a host or folder a new name.
func (l *Library) Rename(id, name string) error {
	if l.Hosts[id].Peer != "" {
		return errSynced
	}
	name, err := checkName(name)
	if err != nil {
		return err
	}
	if folder, ok := l.Folders[id]; ok {
		folder.Name = name
		l.Folders[id] = folder
		return nil
	}
	if host, ok := l.Hosts[id]; ok {
		host.Name = name
		l.Hosts[id] = host
		return nil
	}
	return errUnknownItem
}

// SetIcon gives a folder one of the icons.
func (l *Library) SetIcon(id string, icon Icon) error {
	folder, ok := l.Folders[id]
	if !ok {
		return errNotAFolder
	}
	if !slices.Contains(folderIcons, icon) {
		return errUnknownIcon
	}
	folder.Icon = icon
	l.Folders[id] = folder
	return nil
}

// SetHostIcon sets the icon the hosts in a folder are drawn with; empty takes the folder's above.
func (l *Library) SetHostIcon(id string, icon Icon) error {
	folder, ok := l.Folders[id]
	if !ok {
		return errNotAFolder
	}
	if icon != "" && !slices.Contains(hostIcons, icon) {
		return errUnknownIcon
	}
	folder.HostIcon = icon
	l.Folders[id] = folder
	return nil
}

// SetExpanded records whether a folder shows what it holds.
func (l *Library) SetExpanded(id string, expanded bool) error {
	folder, ok := l.Folders[id]
	if !ok {
		return errNotAFolder
	}
	folder.Expanded = expanded
	l.Folders[id] = folder
	return nil
}

// SetAddress gives a host the name or address it is reached at.
func (l *Library) SetAddress(id, address string) error {
	host, ok := l.Hosts[id]
	if !ok {
		return errNotAHost
	}
	if host.Peer != "" {
		return errSynced
	}
	host.Address = strings.TrimSpace(address)
	l.Hosts[id] = host
	return nil
}

// contains reports whether id is folder itself or lies anywhere inside it.
func (l *Library) contains(folder, id string) bool {
	return id == folder || slices.Contains(l.ancestors(id), folder)
}

// Move puts items, in the given order, into parent at index among its children as they are
// before the move. A folder never moves into itself or a folder inside it.
func (l *Library) Move(ids []string, parent string, index int) error {
	if err := l.checkParent(parent); err != nil {
		return err
	}
	for _, id := range ids {
		if !l.exists(id) {
			return errUnknownItem
		}
		if l.Hosts[id].Peer != "" {
			return errSynced
		}
		if l.isFolder(id) && parent != "" && l.contains(id, parent) {
			return errMoveIntoItself
		}
	}
	target := l.children(parent)
	index = min(max(index, 0), len(target))
	for _, id := range target[:index] {
		if slices.Contains(ids, id) {
			index--
		}
	}
	for _, id := range ids {
		from := l.parentOf(id)
		l.setChildren(from, slices.DeleteFunc(slices.Clone(l.children(from)), func(child string) bool {
			return child == id
		}))
	}
	l.setChildren(parent, slices.Insert(slices.Clone(l.children(parent)), index, ids...))
	return nil
}

// Delete removes a host or a folder with everything inside it, and returns every removed id.
// Hosts that used a removed host as their jump host inherit it again.
func (l *Library) Delete(id string) ([]string, error) {
	if !l.exists(id) {
		return nil, errUnknownItem
	}
	if l.Hosts[id].Peer != "" {
		return nil, errSynced
	}
	parent := l.parentOf(id)
	l.setChildren(parent, slices.DeleteFunc(slices.Clone(l.children(parent)), func(child string) bool {
		return child == id
	}))
	removed := l.subtree(id)
	for _, gone := range removed {
		delete(l.Folders, gone)
		delete(l.Hosts, gone)
	}
	l.forgetJumpHosts(removed)
	return removed, nil
}

// forgetJumpHosts has the hosts and folders that jumped through a removed host inherit their
// jump host again.
func (l *Library) forgetJumpHosts(removed []string) {
	if len(removed) == 0 {
		return
	}
	l.forEachSettings(func(settings *Settings) {
		if settings.JumpHostID != nil && slices.Contains(removed, *settings.JumpHostID) {
			settings.JumpHostID = nil
		}
	})
}

// subtree returns id and every id inside it.
func (l *Library) subtree(id string) []string {
	ids := []string{id}
	for _, child := range l.Folders[id].Children {
		ids = append(ids, l.subtree(child)...)
	}
	return ids
}

// Duplicate puts a copy of a host, named name, right after it and returns the copy's id. A saved
// password stays with the original.
func (l *Library) Duplicate(id, name string) (string, error) {
	host, ok := l.Hosts[id]
	if !ok {
		return "", errNotAHost
	}
	if host.Peer != "" {
		return "", errSynced
	}
	name, err := checkName(name)
	if err != nil {
		return "", err
	}
	copied := Host{ID: newID(), Name: name, Address: host.Address, Settings: host.Settings.clone(), Tags: slices.Clone(host.Tags)}
	copied.Settings.Password = nil
	l.Hosts[copied.ID] = copied
	parent := l.parentOf(id)
	siblings := slices.Clone(l.children(parent))
	l.setChildren(parent, slices.Insert(siblings, slices.Index(siblings, id)+1, copied.ID))
	return copied.ID, nil
}

// forEachSettings calls change with the settings of every folder and host.
func (l *Library) forEachSettings(change func(*Settings)) {
	for id, folder := range l.Folders {
		change(&folder.Settings)
		l.Folders[id] = folder
	}
	for id, host := range l.Hosts {
		change(&host.Settings)
		l.Hosts[id] = host
	}
}
