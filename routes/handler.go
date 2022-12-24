package routes

import (
	"net/http"

	"git.icyphox.sh/legit/config"
	"git.icyphox.sh/legit/git"
	"github.com/alexedwards/flow"
)

type depsWrapper struct {
	actualDeps deps
	gitsvc     http.Handler
}

// Checks for gitprotocol-http(5) specific smells; if found, passes
// the request on to the git http service, else render the web frontend.
func (dw *depsWrapper) Multiplex(w http.ResponseWriter, r *http.Request) {
	path := flow.Param(r.Context(), "...")

	if r.URL.RawQuery == "service=git-receive-pack" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no pushing allowed!"))
		return
	}

	if path == "info/refs" && r.URL.RawQuery == "service=git-upload-pack" && r.Method == "GET" {
		dw.gitsvc.ServeHTTP(w, r)
	} else if path == "git-upload-pack" && r.Method == "POST" {
		dw.gitsvc.ServeHTTP(w, r)
	} else if r.Method == "GET" {
		dw.actualDeps.RepoIndex(w, r)
	}
}

func Handlers(c *config.Config) *flow.Mux {
	mux := flow.New()
	d := deps{c}

	gitsvc := git.Handler(c.Repo.ScanPath)

	dw := depsWrapper{actualDeps: d, gitsvc: gitsvc}

	mux.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		d.Write404(w)
	})

	mux.HandleFunc("/", d.Index, "GET")
	mux.HandleFunc("/static/:file", d.ServeStatic, "GET")
	mux.HandleFunc("/:name", dw.Multiplex, "GET", "POST")
	mux.HandleFunc("/:name/tree/:ref/...", d.RepoTree, "GET")
	mux.HandleFunc("/:name/blob/:ref/...", d.FileContent, "GET")
	mux.HandleFunc("/:name/log/:ref", d.Log, "GET")
	mux.HandleFunc("/:name/commit/:ref", d.Diff, "GET")
	mux.HandleFunc("/:name/refs", d.Refs, "GET")
	mux.HandleFunc("/:name/...", dw.Multiplex, "GET", "POST")

	return mux
}
