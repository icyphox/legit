package routes

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git.icyphox.sh/legit/git"
	"github.com/alexedwards/flow"
	"github.com/dustin/go-humanize"
)

func isGoModule(gr *git.GitRepo) bool {
	_, err := gr.FileContent("go.mod")
	return err == nil
}

func getDescription(path string) (desc string) {
	db, err := os.ReadFile(filepath.Join(path, "description"))
	if err == nil {
		desc = string(db)
	} else {
		desc = ""
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

type repository struct {
	Name        string
	Category    string
	Path        string
	Slug        string
	Description string
	LastCommit  string
}

type entry struct {
	Name         string
	Repositories []*repository
}

type entries struct {
	Children []*entry
	c        map[string]*entry
}

func (ent *entries) Add(r repository) {
	if r.Category == "" {
		ent.Children = append(ent.Children, &entry{
			Name:         r.Name,
			Repositories: []*repository{&r},
		})
		return
	}
	t, ok := ent.c[r.Category]
	if !ok {
		t := &entry{
			Name:         r.Category,
			Repositories: []*repository{&r},
		}
		ent.c[r.Category] = t
		ent.Children = append(ent.Children, t)
		return
	}
	t.Repositories = append(t.Repositories, &r)
}

func (d *deps) getAllRepos() (*entries, error) {
	entries := &entries{
		Children: []*entry{},
		c:        map[string]*entry{},
	}
	max := strings.Count(d.c.Repo.ScanPath, string(os.PathSeparator)) + 2

	err := filepath.WalkDir(d.c.Repo.ScanPath, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if de.IsDir() {
			// Check if we've exceeded our recursion depth
			if strings.Count(path, string(os.PathSeparator)) > max {
				return fs.SkipDir
			}

			if d.isIgnored(path) {
				return fs.SkipDir
			}

			// A bare repo should always have at least a HEAD file, if it
			// doesn't we can continue recursing
			if _, err := os.Lstat(filepath.Join(path, "HEAD")); err == nil {
				repo, err := git.Open(path, "")
				if err != nil {
					log.Println(err)
				} else {
					relpath, _ := filepath.Rel(d.c.Repo.ScanPath, path)
					category := strings.Split(relpath, string(os.PathSeparator))[0]
					r := repository{
						Name:        filepath.Base(path),
						Category:    category,
						Path:        path,
						Slug:        relpath,
						Description: getDescription(path),
					}
					if c, err := repo.LastCommit(); err == nil {
						r.LastCommit = humanize.Time(c.Author.When)
					}
					entries.Add(r)
					// Since we found a Git repo, we don't want to recurse
					// further
					return fs.SkipDir
				}
			}
		}
		return nil
	})
	sort.Slice(entries.Children, func(i, j int) bool {
		return entries.Children[i].Name < entries.Children[j].Name
	})
	return entries, err
}

func repoPath(ctx context.Context) string {
	return filepath.Join(
		filepath.Clean(flow.Param(ctx, "category")),
		filepath.Clean(flow.Param(ctx, "name")),
	)
}
