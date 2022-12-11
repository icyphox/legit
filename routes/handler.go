package routes

import (
	"github.com/alexedwards/flow"
	"icyphox.sh/legit/config"
)

func Handlers(c *config.Config) *flow.Mux {
	mux := flow.New()
	d := deps{c}
	mux.HandleFunc("/:name", d.Repo, "GET")
	mux.HandleFunc("/:name/tree/...", d.Repo, "GET")
	return mux
}
