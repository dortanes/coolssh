// Package catalog holds the copy the Go host shows, read from the frontend's message catalogs, and
// resolves which supported language the user reads.
package catalog

import (
	"fmt"
	"io/fs"
	"strings"

	"go.yaml.in/yaml/v3"
	"golang.org/x/text/language"
)

// Source is the language every key is written in first; it answers when nothing else matches.
var Source = language.English

// Supported lists the languages with a catalog, Source first. Each catalog file is named after the
// tag's string form, which the frontend uses as its locale.
var Supported = []language.Tag{Source, language.Russian}

// Catalogs holds every supported language's messages by key.
type Catalogs struct {
	messages map[language.Tag]map[string]string
	matcher  language.Matcher
}

// Load reads `<tag>.yaml` for every supported language from files; a message's key joins its path
// with underscores, as the frontend's catalog plugin does: `keys: {empty: …}` is `keys_empty`.
func Load(files fs.FS) (*Catalogs, error) {
	c := &Catalogs{messages: map[language.Tag]map[string]string{}, matcher: language.NewMatcher(Supported)}
	for _, tag := range Supported {
		name := tag.String() + ".yaml"
		raw, err := fs.ReadFile(files, name)
		if err != nil {
			return nil, fmt.Errorf("catalog %s: %w", name, err)
		}
		tree := map[string]any{}
		if err := yaml.Unmarshal(raw, &tree); err != nil {
			return nil, fmt.Errorf("catalog %s: %w", name, err)
		}
		messages := map[string]string{}
		if err := flatten("", tree, messages); err != nil {
			return nil, fmt.Errorf("catalog %s: %w", name, err)
		}
		c.messages[tag] = messages
	}
	return c, nil
}

func flatten(prefix string, group map[string]any, messages map[string]string) error {
	for name, value := range group {
		if prefix != "" {
			name = prefix + "_" + name
		}
		switch value := value.(type) {
		case string:
			messages[name] = value
		case map[string]any:
			if err := flatten(name, value, messages); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s is neither a message nor a group", name)
		}
	}
	return nil
}

// Has reports whether every supported language's catalog has a message for key.
func (c *Catalogs) Has(key string) bool {
	for _, tag := range Supported {
		if c.messages[tag][key] == "" {
			return false
		}
	}
	return true
}

// Resolve picks the language the user reads: the chosen one when it names a supported language,
// else the first supported one among the system's preferred languages, else Source.
func (c *Catalogs) Resolve(chosen string, preferred []language.Tag) language.Tag {
	for _, tag := range Supported {
		if tag.String() == chosen {
			return tag
		}
	}
	_, index, confidence := c.matcher.Match(preferred...)
	if confidence == language.No {
		return Source
	}
	return Supported[index]
}

// Text is a message in a supported language, the Source message when that catalog lacks the key.
// A key missing from Source is a bug in the caller.
func (c *Catalogs) Text(tag language.Tag, key string) string {
	if text, ok := c.messages[tag][key]; ok {
		return text
	}
	if text, ok := c.messages[Source][key]; ok {
		return text
	}
	panic(fmt.Sprintf("catalog: no message %q", key))
}

// Format is Text with each `{name}` placeholder replaced by its value in values.
func (c *Catalogs) Format(tag language.Tag, key string, values map[string]string) string {
	pairs := make([]string, 0, 2*len(values))
	for name, value := range values {
		pairs = append(pairs, "{"+name+"}", value)
	}
	return strings.NewReplacer(pairs...).Replace(c.Text(tag, key))
}
