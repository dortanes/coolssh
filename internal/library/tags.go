package library

import (
	"strings"
	"unicode/utf8"
)

// maxTagLength bounds a tag, in characters: a word or two, not a note.
const maxTagLength = 32

// SetTags gives a host its tags, each trimmed and kept once ignoring case, in the order given;
// empty ones are dropped. A synced host takes tags too: they are the user's, not its source's.
func (l *Library) SetTags(id string, tags []string) error {
	host, ok := l.Hosts[id]
	if !ok {
		return errNotAHost
	}
	kept := []string{}
	seen := map[string]bool{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[strings.ToLower(tag)] {
			continue
		}
		if utf8.RuneCountInString(tag) > maxTagLength {
			return errTagTooLong
		}
		seen[strings.ToLower(tag)] = true
		kept = append(kept, tag)
	}
	host.Tags = kept
	l.Hosts[id] = host
	return nil
}
