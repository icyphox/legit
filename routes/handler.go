package routes

import (
	"embed"
	"net/http"

	"git.icyphox.sh/legit/config"
	"github.com/alexedwards/flow"
)

var StaticFiles *embed.FS

// Checks for gitprotocol-http(5) specific smells; if found, passes
// the request on to the git http service, else render the web frontend.
func (d *deps) Multiplex(w http.ResponseWriter, r *http.Request) {
	path := flow.Param(r.Context(), "...")

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

func Handlers(c *config.Config) *flow.Mux {
	mux := flow.New()
	d := deps{c}

	mux.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		d.Write404(w)
	})

	mux.HandleFunc("/", d.Index, "GET")

	if c.Dirs.Static != "" {
		//read from file system
		mux.HandleFunc("/static/:file", d.ServeStatic, "GET")
	} else {
		//read embedded static directory
		mux.Handle("/static/...", http.FileServer(http.FS(StaticFiles)), "GET")
	}

	mux.HandleFunc("/:name", d.Multiplex, "GET", "POST")
	mux.HandleFunc("/:name/tree/:ref/...", d.RepoTree, "GET")
	mux.HandleFunc("/:name/blob/:ref/...", d.FileContent, "GET")
	mux.HandleFunc("/:name/log/:ref", d.Log, "GET")
	mux.HandleFunc("/:name/commit/:ref", d.Diff, "GET")
	mux.HandleFunc("/:name/refs", d.Refs, "GET")
	mux.HandleFunc("/:name/...", d.Multiplex, "GET", "POST")

	return mux
}
