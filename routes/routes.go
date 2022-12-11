package routes

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/alexedwards/flow"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"icyphox.sh/legit/config"
	"icyphox.sh/legit/git"
)

type deps struct {
	c *config.Config
}

func (d *deps) RepoIndex(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	name = filepath.Clean(name)
	// TODO: remove .git
	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		Write404(w, *d.c)
		return
	}

	head, err := repo.Head()
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	files, err := git.FilesAtRef(repo, head.Hash(), "")
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	d.listFiles(files, w)
	return
}

func (d *deps) RepoTree(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	treePath := flow.Param(r.Context(), "...")
	ref := flow.Param(r.Context(), "ref")

	name = filepath.Clean(name)
	// TODO: remove .git
	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		Write404(w, *d.c)
		return
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	files, err := git.FilesAtRef(repo, *hash, treePath)
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	d.listFiles(files, w)
	return
}

func (d *deps) FileContent(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	treePath := flow.Param(r.Context(), "...")
	ref := flow.Param(r.Context(), "ref")

	name = filepath.Clean(name)
	// TODO: remove .git
	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		Write404(w, *d.c)
		return
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	contents, err := git.FileContentAtRef(repo, *hash, treePath)
	d.showFile(contents, w)
	return
}
