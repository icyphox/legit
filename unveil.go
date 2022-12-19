//go:build openbsd
// +build openbsd

// Doesn't do anything yet.

package main

/*
#include <stdlib.h>
#include <unistd.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func Unveil(path string, perms string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	cperms := C.CString(perms)
	defer C.free(unsafe.Pointer(cperms))

	rv, err := C.unveil(cpath, cperms)
	if rv != 0 {
		return fmt.Errorf("unveil(%s, %s) failure (%d)", path, perms, err)
	}
	return nil
}
