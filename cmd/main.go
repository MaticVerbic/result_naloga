package main

import (
	"naloga/config"
	"naloga/server"
	"naloga/services"
)

func main() {
	cfg := config.New()

	f := services.NewFetcher(cfg.URLS, cfg.Log)
	s := server.New(f, cfg.Log)
	s.Serve()
}
