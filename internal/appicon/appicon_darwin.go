//go:build darwin && cgo

package appicon

/*
#cgo CFLAGS: -fobjc-arc
#cgo LDFLAGS: -framework AppKit
#include <stdlib.h>
#include "appicon_darwin.h"
*/
import "C"

func draw(side int) []byte {
	var length C.int
	bytes := C.coolssh_appicon_png(C.int(side), &length)
	if bytes == nil {
		return nil
	}
	defer C.free(bytes)
	return C.GoBytes(bytes, length)
}
