//go:build !openbsd
// +build !openbsd

// Stub functions for GOOS that don't support unix.Unveil()

package main

func Unveil(path string, perms string) error {
	return nil
}

func UnveilBlock() error {
	return nil
}

func UnveilPaths(paths []string, perms string) error {
	return nil
}
