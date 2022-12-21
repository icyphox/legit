//go:build openbsd
// +build openbsd

package main

import (
	"golang.org/x/sys/unix"
	"log"
)

func Unveil(path string, perms string) error {
	log.Printf("unveil: \"%s\", %s", path, perms)
	return unix.Unveil(path, perms)
}

func UnveilBlock() error {
	log.Printf("unveil: block")
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
