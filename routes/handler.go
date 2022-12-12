package routes

import (
	"net/http"

	"github.com/alexedwards/flow"
	"icyphox.sh/legit/config"
)

func Handlers(c *config.Config) *flow.Mux {
	mux := flow.New()
	d := deps{c}

	mux.NotFound = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		d.Write404(w)
	})

	mux.HandleFunc("/", d.Index, "GET")
	mux.HandleFunc("/:name", d.RepoIndex, "GET")
	mux.HandleFunc("/:name/tree/:ref/...", d.RepoTree, "GET")
	mux.HandleFunc("/:name/blob/:ref/...", d.FileContent, "GET")
	mux.HandleFunc("/:name/log/:ref", d.Log, "GET")
	mux.HandleFunc("/:name/commit/:ref", d.Diff, "GET")
	return mux
}
