package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"

	"git.icyphox.sh/legit/config"
	"git.icyphox.sh/legit/routes"
)

//go:embed templates
var tmplFiles embed.FS

//go:embed static
var staticFiles embed.FS

func main() {
	var cfg string
	flag.StringVar(&cfg, "config", "./config.yaml", "path to config file")
	flag.Parse()

	c, err := config.Read(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := UnveilPaths([]string{
		c.Dirs.Static,
		c.Repo.ScanPath,
		c.Dirs.Templates,
	},
		"r"); err != nil {
		log.Fatalf("unveil: %s", err)
	}

	if c.Dirs.Templates == "" {
		routes.TmplFiles = &tmplFiles
	}
	if c.Dirs.Static == "" {
		routes.StaticFiles = &staticFiles
	}

	mux := routes.Handlers(c)
	addr := fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
	log.Println("starting server on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
