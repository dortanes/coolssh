//go:build darwin && cgo

package appearance

/*
#cgo LDFLAGS: -framework AppKit
#include "appearance_darwin.h"
*/
import "C"

import "github.com/dortanes/coolssh/internal/preferences"

func apply(chosen preferences.Appearance) {
	switch chosen {
	case preferences.LightAppearance:
		C.coolssh_appearance_set(C.coolssh_appearance_light)
	case preferences.DarkAppearance:
		C.coolssh_appearance_set(C.coolssh_appearance_dark)
	default:
		C.coolssh_appearance_set(C.coolssh_appearance_system)
	}
}
