package routes

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/alexedwards/flow"
	gogit "github.com/go-git/go-git/v5"
	"icyphox.sh/legit/config"
	"icyphox.sh/legit/git"
)

type deps struct {
	c *config.Config
}

func (d *deps) Repo(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	treePath := flow.Param(r.Context(), "...")

	name = filepath.Clean(name)
	// TODO: remove .git
	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		Write404(w, *d.c)
		return
	}

	files, err := git.FilesAtHead(repo, treePath)
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
	}

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
