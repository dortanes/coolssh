// Package library holds the user's hosts in nested folders, the keys they use, and the connection
// settings folders pass down to what they hold.
package library

import (
	"crypto/rand"
	"slices"
	"strings"
)

// fileVersion is the version of the library this package writes and reads.
const fileVersion = 1

// Icon is a folder's or a host's icon, one of a fixed set the windows draw.
type Icon string

const (
	FolderIcon    Icon = "folder"
	MonitorIcon   Icon = "monitor"
	ServerIcon    Icon = "server"
	DatabaseIcon  Icon = "database"
	CloudIcon     Icon = "cloud"
	GlobeIcon     Icon = "globe"
	HomeIcon      Icon = "home"
	BriefcaseIcon Icon = "briefcase"
	CodeIcon      Icon = "code"
	ShieldIcon    Icon = "shield"
	TerminalIcon  Icon = "terminal"
	FlaskIcon     Icon = "flask"
	BoxIcon       Icon = "box"
)

var folderIcons = []Icon{
	FolderIcon, ServerIcon, DatabaseIcon, CloudIcon, GlobeIcon, HomeIcon,
	BriefcaseIcon, CodeIcon, ShieldIcon, TerminalIcon, FlaskIcon, BoxIcon,
}

var hostIcons = []Icon{
	MonitorIcon, ServerIcon, DatabaseIcon, CloudIcon, GlobeIcon, HomeIcon,
	BriefcaseIcon, CodeIcon, ShieldIcon, TerminalIcon, FlaskIcon, BoxIcon,
}

// Folder holds hosts and folders in order, and settings they inherit.
type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon Icon   `json:"icon"`
	// HostIcon is the icon the hosts inside are drawn with; empty to take the folder's above.
	HostIcon Icon     `json:"hostIcon,omitempty"`
	Expanded bool     `json:"expanded"`
	Children []string `json:"children"`
	Settings Settings `json:"settings"`
	// Source is where the folder takes its hosts from; nil for a folder the user fills.
	Source *Source `json:"source,omitempty"`
}

// Host is a machine the user connects to.
type Host struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	// Settings left unset are inherited from the folders above the host.
	Settings Settings `json:"settings"`
	// Peer is the key of the peer a synced host stands for; empty for a host the user added.
	Peer string `json:"peer,omitempty"`
	// HostKey is the SSH host key a synced host's source vouches for, trusted alongside
	// known_hosts; empty for none.
	HostKey string `json:"hostKey,omitempty"`
	// Tags are the user's words for finding the host.
	Tags []string `json:"tags,omitempty"`
}

// Library is the whole library: the root's items in order, every folder and host by id, and the
// keys. It is also the file's format and what the windows receive.
type Library struct {
	Version int               `json:"version"`
	Root    []string          `json:"root"`
	Folders map[string]Folder `json:"folders"`
	Hosts   map[string]Host   `json:"hosts"`
	Keys    []Key             `json:"keys"`
	// DefaultKeyID names the key a host uses when neither it nor a folder above it sets one; empty
	// for none.
	DefaultKeyID string `json:"defaultKeyId,omitempty"`
	// Defaults are the user's own defaults for user, port, agent forwarding and keepalive, which a
	// host takes when neither it nor a folder above it sets one; unset ones fall to the built-in.
	Defaults Settings `json:"defaults"`
}

// Empty is a library with nothing in it.
func Empty() Library {
	return Library{Version: fileVersion, Root: []string{}, Folders: map[string]Folder{}, Hosts: map[string]Host{}, Keys: []Key{}}
}

// Refusal is why the library refused a change: a key of the catalogs the windows word it from.
type Refusal string

func (r Refusal) Error() string {
	return string(r)
}

const (
	errUnknownItem     Refusal = "library_unknown_item"
	errNotAFolder      Refusal = "library_not_a_folder"
	errNotAHost        Refusal = "library_not_a_host"
	errMoveIntoItself  Refusal = "library_move_into_itself"
	errEmptyName       Refusal = "library_empty_name"
	errUnknownIcon     Refusal = "library_unknown_icon"
	errPortRange       Refusal = "library_port_range"
	errKeepaliveRange  Refusal = "library_keepalive_range"
	errUnknownKey      Refusal = "library_unknown_key"
	errOwnJumpHost     Refusal = "library_own_jump_host"
	errNotAKey         Refusal = "library_not_a_key"
	errWrongPassphrase Refusal = "library_wrong_passphrase"
	errUnreadableKey   Refusal = "library_unreadable_key"
	errNoAddress       Refusal = "library_no_address"
	errJumpLoop        Refusal = "library_jump_loop"
	errUnknownKeyKind  Refusal = "library_unknown_key_kind"
	errNoPublicKey     Refusal = "library_no_public_key"
	errNoKey           Refusal = "library_no_key"
	errTagTooLong      Refusal = "library_tag_too_long"
)

var refusals = []Refusal{
	errUnknownItem, errNotAFolder, errNotAHost, errMoveIntoItself, errEmptyName, errUnknownIcon,
	errPortRange, errKeepaliveRange, errUnknownKey, errOwnJumpHost, errNotAKey, errWrongPassphrase,
	errUnreadableKey, errNoAddress, errJumpLoop, errSynced, errSourceNotEmpty, errSourceTaken,
	errUnknownKeyKind, errNoPublicKey, errNoKey, errTagTooLong,
}

// Refusals lists every Refusal, each a key the catalogs word.
func Refusals() []Refusal {
	return slices.Clone(refusals)
}

func newID() string {
	return rand.Text()
}

func (l *Library) clone() Library {
	copied := Library{
		Version:      l.Version,
		Root:         slices.Clone(l.Root),
		Folders:      make(map[string]Folder, len(l.Folders)),
		Hosts:        make(map[string]Host, len(l.Hosts)),
		Keys:         slices.Clone(l.Keys),
		DefaultKeyID: l.DefaultKeyID,
		Defaults:     l.Defaults.clone(),
	}
	for id, folder := range l.Folders {
		folder.Children = slices.Clone(folder.Children)
		folder.Settings = folder.Settings.clone()
		folder.Source = folder.Source.clone()
		copied.Folders[id] = folder
	}
	for id, host := range l.Hosts {
		host.Settings = host.Settings.clone()
		host.Tags = slices.Clone(host.Tags)
		copied.Hosts[id] = host
	}
	return copied
}

// children returns the ids directly inside parent, the root when parent is empty.
func (l *Library) children(parent string) []string {
	if parent == "" {
		return l.Root
	}
	return l.Folders[parent].Children
}

func (l *Library) setChildren(parent string, ids []string) {
	if parent == "" {
		l.Root = ids
		return
	}
	folder := l.Folders[parent]
	folder.Children = ids
	l.Folders[parent] = folder
}

// parentOf returns the folder directly holding id, empty for the root.
func (l *Library) parentOf(id string) string {
	for folderID, folder := range l.Folders {
		if slices.Contains(folder.Children, id) {
			return folderID
		}
	}
	return ""
}

// ancestors returns the folders above id, the nearest first.
func (l *Library) ancestors(id string) []string {
	var above []string
	for parent := l.parentOf(id); parent != ""; parent = l.parentOf(parent) {
		above = append(above, parent)
	}
	return above
}

func (l *Library) exists(id string) bool {
	_, folder := l.Folders[id]
	_, host := l.Hosts[id]
	return folder || host
}

func (l *Library) isFolder(id string) bool {
	_, ok := l.Folders[id]
	return ok
}

// checkParent refuses a parent that is neither the root nor a folder the user fills.
func (l *Library) checkParent(parent string) error {
	if parent != "" && !l.isFolder(parent) {
		return errNotAFolder
	}
	if l.synced(parent) {
		return errSynced
	}
	return nil
}

func checkName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errEmptyName
	}
	return name, nil
}
