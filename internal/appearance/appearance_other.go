//go:build !darwin || !cgo

package appearance

import "github.com/dortanes/coolssh/internal/preferences"

func apply(preferences.Appearance) {}
