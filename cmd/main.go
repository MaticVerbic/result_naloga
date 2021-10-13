package main

import (
	"naloga/config"
	"naloga/server"
	"naloga/services"
)

// @title Result test task
// @version 1.0
// @description This is a test task for result.
// @name Result task
// @host localhost:8080
// @BasePath /

// @license.name none

func main() {
	cfg := config.New()

	f := services.NewFetcher(cfg.URLS, cfg.Log)
	s := server.New(f, cfg.Log)
	s.Serve()
}
