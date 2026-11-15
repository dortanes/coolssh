package library

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestTagsAreTrimmedAndKeptOnceIgnoringCase(t *testing.T) {
	l, ids := tree(t)
	if err := l.SetTags(ids["web"], []string{" prod ", "Web", "", "PROD", "eu"}); err != nil {
		t.Fatal(err)
	}
	if got := l.Hosts[ids["web"]].Tags; !slices.Equal(got, []string{"prod", "Web", "eu"}) {
		t.Errorf("tags = %q", got)
	}
	copied, err := l.Duplicate(ids["web"], "web copy")
	if err != nil {
		t.Fatal(err)
	}
	if got := l.Hosts[copied].Tags; !slices.Equal(got, []string{"prod", "Web", "eu"}) {
		t.Errorf("the copy's tags = %q", got)
	}
	if err := l.SetTags(ids["web"], []string{strings.Repeat("x", maxTagLength+1)}); !errors.Is(err, errTagTooLong) {
		t.Errorf("a long tag: %v", err)
	}
	if err := l.SetTags(ids["prod"], []string{"prod"}); !errors.Is(err, errNotAHost) {
		t.Errorf("tags on a folder: %v", err)
	}
}

func TestACopyOfTheLibraryOwnsItsTags(t *testing.T) {
	l, ids := tree(t)
	if err := l.SetTags(ids["web"], []string{"prod"}); err != nil {
		t.Fatal(err)
	}
	l.clone().Hosts[ids["web"]].Tags[0] = "changed"
	if got := l.Hosts[ids["web"]].Tags; !slices.Equal(got, []string{"prod"}) {
		t.Errorf("tags after changing a copy = %q", got)
	}
}
