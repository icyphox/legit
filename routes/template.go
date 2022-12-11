package routes

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"icyphox.sh/legit/config"
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

func funcMap() template.FuncMap {
	return template.FuncMap{
		"prettyMode": func(mode uint32) string {
			return os.FileMode(mode).String()
		},
	}
}
