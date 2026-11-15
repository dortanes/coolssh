package library

import "strings"

// The bounds of the numeric settings.
const (
	minPort      = 1
	maxPort      = 65535
	maxKeepalive = 3600
)

// Settings are how a host is reached. A nil setting is inherited from the nearest folder above
// that sets it, else the default.
type Settings struct {
	User *string `json:"user,omitempty"`
	Port *int    `json:"port,omitempty"`
	// KeyID names a library key; empty sets no key.
	KeyID *string `json:"keyId,omitempty"`
	// Password is true where a password is saved for the item, in the keyring.
	Password *bool `json:"password,omitempty"`
	// JumpHostID names the host to connect through; empty sets none.
	JumpHostID      *string `json:"jumpHostId,omitempty"`
	AgentForwarding *bool   `json:"agentForwarding,omitempty"`
	// KeepaliveSeconds is how often an idle session is kept alive; 0 turns it off.
	KeepaliveSeconds *int `json:"keepaliveSeconds,omitempty"`
}

// Origins names, for each setting, the folder it is inherited from, empty for the default.
type Origins struct {
	User             string `json:"user"`
	Port             string `json:"port"`
	KeyID            string `json:"keyId"`
	Password         string `json:"password"`
	JumpHostID       string `json:"jumpHostId"`
	AgentForwarding  string `json:"agentForwarding"`
	KeepaliveSeconds string `json:"keepaliveSeconds"`
}

// Inherited is what an item would take for each setting it leaves unset, and from where. Every
// setting of Values is set.
type Inherited struct {
	Values  Settings `json:"values"`
	Origins Origins  `json:"origins"`
}

func (s Settings) clone() Settings {
	return Settings{
		User:             clonePointer(s.User),
		Port:             clonePointer(s.Port),
		KeyID:            clonePointer(s.KeyID),
		Password:         clonePointer(s.Password),
		JumpHostID:       clonePointer(s.JumpHostID),
		AgentForwarding:  clonePointer(s.AgentForwarding),
		KeepaliveSeconds: clonePointer(s.KeepaliveSeconds),
	}
}

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

// SetSettings replaces an item's settings. Whether a password is saved changes only with the
// password itself, so the item keeps its own; a blank user is unset.
func (l *Library) SetSettings(id string, settings Settings) error {
	settings = settings.clone()
	if settings.User != nil && strings.TrimSpace(*settings.User) == "" {
		settings.User = nil
	}
	if err := l.checkSettings(id, settings); err != nil {
		return err
	}
	if folder, ok := l.Folders[id]; ok {
		settings.Password = folder.Settings.Password
		folder.Settings = settings
		l.Folders[id] = folder
		return nil
	}
	host := l.Hosts[id]
	settings.Password = host.Settings.Password
	host.Settings = settings
	l.Hosts[id] = host
	return nil
}

// setPassword records whether a password is saved for an item; nil inherits it again.
func (l *Library) setPassword(id string, saved *bool) error {
	if folder, ok := l.Folders[id]; ok {
		folder.Settings.Password = saved
		l.Folders[id] = folder
		return nil
	}
	if host, ok := l.Hosts[id]; ok {
		host.Settings.Password = saved
		l.Hosts[id] = host
		return nil
	}
	return errUnknownItem
}

// checkSettings refuses settings outside their bounds or naming what the library does not hold.
func (l *Library) checkSettings(id string, settings Settings) error {
	if !l.exists(id) {
		return errUnknownItem
	}
	if err := checkBounds(settings); err != nil {
		return err
	}
	if settings.KeyID != nil && *settings.KeyID != "" && !l.hasKey(*settings.KeyID) {
		return errUnknownKey
	}
	if settings.JumpHostID == nil || *settings.JumpHostID == "" {
		return nil
	}
	if *settings.JumpHostID == id {
		return errOwnJumpHost
	}
	if _, ok := l.Hosts[*settings.JumpHostID]; !ok {
		return errNotAHost
	}
	return nil
}

// checkBounds refuses a port or keepalive outside its bounds.
func checkBounds(settings Settings) error {
	if settings.Port != nil && (*settings.Port < minPort || *settings.Port > maxPort) {
		return errPortRange
	}
	if settings.KeepaliveSeconds != nil && (*settings.KeepaliveSeconds < 0 || *settings.KeepaliveSeconds > maxKeepalive) {
		return errKeepaliveRange
	}
	return nil
}

// SetDefaults replaces the user's defaults; only user, port, agent forwarding and keepalive are kept.
func (l *Library) SetDefaults(settings Settings) error {
	defaults := Settings{
		User:             clonePointer(settings.User),
		Port:             clonePointer(settings.Port),
		AgentForwarding:  clonePointer(settings.AgentForwarding),
		KeepaliveSeconds: clonePointer(settings.KeepaliveSeconds),
	}
	if defaults.User != nil && strings.TrimSpace(*defaults.User) == "" {
		defaults.User = nil
	}
	if err := checkBounds(defaults); err != nil {
		return err
	}
	l.Defaults = defaults
	return nil
}

// Inherited answers each setting id leaves unset from the nearest folder that sets it, else the
// library's default key and the user's defaults, else the built-in defaults.
func (l *Library) Inherited(id string, defaults Settings) (Inherited, error) {
	if !l.exists(id) {
		return Inherited{}, errUnknownItem
	}
	var inherited Inherited
	for _, folder := range l.ancestors(id) {
		inherited.take(l.Folders[folder].Settings, folder)
	}
	l.fallBack(&inherited, defaults)
	return inherited, nil
}

// fallBack fills what is still unset from the default key, the user's defaults and defaults.
func (l *Library) fallBack(inherited *Inherited, defaults Settings) {
	if l.DefaultKeyID != "" {
		inherited.take(Settings{KeyID: &l.DefaultKeyID}, "")
	}
	inherited.take(l.Defaults, "")
	inherited.take(defaults, "")
}

// take fills every setting still unset from settings, recording origin as where it came from.
func (i *Inherited) take(settings Settings, origin string) {
	inherit(&i.Values.User, &i.Origins.User, settings.User, origin)
	inherit(&i.Values.Port, &i.Origins.Port, settings.Port, origin)
	inherit(&i.Values.KeyID, &i.Origins.KeyID, settings.KeyID, origin)
	inherit(&i.Values.Password, &i.Origins.Password, settings.Password, origin)
	inherit(&i.Values.JumpHostID, &i.Origins.JumpHostID, settings.JumpHostID, origin)
	inherit(&i.Values.AgentForwarding, &i.Origins.AgentForwarding, settings.AgentForwarding, origin)
	inherit(&i.Values.KeepaliveSeconds, &i.Origins.KeepaliveSeconds, settings.KeepaliveSeconds, origin)
}

func inherit[T any](value **T, origin *string, from *T, folder string) {
	if *value == nil && from != nil {
		*value = clonePointer(from)
		*origin = folder
	}
}
