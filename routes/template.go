package routes

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"icyphox.sh/legit/config"
	"icyphox.sh/legit/git"
)

func Write404(w http.ResponseWriter, c config.Config) {
	w.WriteHeader(404)
	tpath := filepath.Join(c.Template.Dir, "404.html")
	t := template.Must(template.ParseFiles(tpath))
	t.Execute(w, nil)
}

func Write500(w http.ResponseWriter, c config.Config) {
	w.WriteHeader(500)
	tpath := filepath.Join(c.Template.Dir, "500.html")
	t := template.Must(template.ParseFiles(tpath))
	t.Execute(w, nil)
}

func (d *deps) listFiles(files []git.NiceTree, data map[string]any, w http.ResponseWriter) {
	tpath := filepath.Join(d.c.Template.Dir, "*")
	t := template.Must(template.ParseGlob(tpath))

	data["files"] = files
	data["meta"] = d.c.Meta

	if err := t.ExecuteTemplate(w, "repo", data); err != nil {
		log.Println(err)
		return
	}
}

func (d *deps) showFile(content string, data map[string]any, w http.ResponseWriter) {
	tpath := filepath.Join(d.c.Template.Dir, "*")
	t := template.Must(template.ParseGlob(tpath))

	// TODO: Process content here.

	data["content"] = content
	data["meta"] = d.c.Meta

	if err := t.ExecuteTemplate(w, "file", data); err != nil {
		log.Println(err)
		return
	}
}
