package routes

import (
	"html/template"
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

	d.renderFiles(files, w)
	return
}

func (d *deps) RepoFiles(w http.ResponseWriter, r *http.Request) {
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

	d.renderFiles(files, w)
	return
}

func (d *deps) renderFiles(files []git.NiceTree, w http.ResponseWriter) {
	tpath := filepath.Join(d.c.Template.Dir, "*")
	t := template.Must(template.ParseGlob(tpath))

	data := make(map[string]interface{})
	data["files"] = files
	data["meta"] = d.c.Meta

	if err := t.ExecuteTemplate(w, "repo", data); err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}
}
