//go:build openbsd
// +build openbsd

package main

import (
	"golang.org/x/sys/unix"
)

func Unveil(path string, perms string) error {
	return unix.Unveil(path, perms)
}

func UnveilBlock() error {
	return unix.UnveilBlock()
}

func UnveilPaths(paths []string, perms string) error {
	for _, path := range paths {
		err := Unveil(path, perms)
		if err != nil {
			return err
		}
	}
	return UnveilBlock()
}
