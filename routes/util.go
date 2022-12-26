package routes

import (
	"os"
	"path/filepath"
	"strings"
)

const defaultDescription = `Unnamed repository; edit this file 'description' to name the repository.`

func getDescription(path string) (desc string) {
	data, err := os.ReadFile(filepath.Join(path, ".git", "description"))
	if err != nil {
		return ""
	}

	desc = strings.TrimSpace(string(data))
	if desc == defaultDescription {
		return ""
	}

	return
}

func (d *deps) isIgnored(name string) bool {
	for _, i := range d.c.Repo.Ignore {
		if name == i {
			return true
		}
	}

	return false
}
