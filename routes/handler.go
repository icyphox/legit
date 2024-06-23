package routes

import (
	"net/http"

	"git.icyphox.sh/legit/config"
)

// Checks for gitprotocol-http(5) specific smells; if found, passes
// the request on to the git http service, else render the web frontend.
func (d *deps) Multiplex(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("rest")

	if r.URL.RawQuery == "service=git-receive-pack" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no pushing allowed!"))
		return
	}

	if path == "info/refs" &&
		r.URL.RawQuery == "service=git-upload-pack" &&
		r.Method == "GET" {
		d.InfoRefs(w, r)
	} else if path == "git-upload-pack" && r.Method == "POST" {
		d.UploadPack(w, r)
	} else if r.Method == "GET" {
		d.RepoIndex(w, r)
	}
}

func Handlers(c *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	d := deps{c}

	mux.HandleFunc("GET /", d.Index)
	mux.HandleFunc("GET /static/{file}", d.ServeStatic)
	mux.HandleFunc("GET /{name}", d.Multiplex)
	mux.HandleFunc("POST /{name}", d.Multiplex)
	mux.HandleFunc("GET /{name}/tree/{ref}/{rest...}", d.RepoTree)
	mux.HandleFunc("GET /{name}/blob/{ref}/{rest...}", d.FileContent)
	mux.HandleFunc("GET /{name}/log/{ref}", d.Log)
	mux.HandleFunc("GET /{name}/archive/{file}", d.Archive)
	mux.HandleFunc("GET /{name}/commit/{ref}", d.Diff)
	mux.HandleFunc("GET /{name}/refs/{$}", d.Refs)
	mux.HandleFunc("GET /{name}/{rest...}", d.Multiplex)
	mux.HandleFunc("POST /{name}/{rest...}", d.Multiplex)

	return mux
}
