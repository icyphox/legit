package main

import (
	"flag"
	"log"
	"net/http"

	"icyphox.sh/legit/config"
	"icyphox.sh/legit/routes"
)

func main() {
	var cfg string
	flag.StringVar(&cfg, "config", "./config.yaml", "path to config file")
	flag.Parse()

	c, err := config.Read(cfg)
	if err != nil {
		log.Fatal(err)
	}

	mux := routes.Handlers(c)
	log.Fatal(http.ListenAndServe(":5555", mux))
}
