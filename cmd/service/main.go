package main

import (
	"flag"
	"service/internal/application"
)

// @title Balance Server API
// @version 1.0
// @description API Server for avito backend internship

// @host

func main() {
	var confPath string

	flag.StringVar(&confPath, "config", "", "yaml config file")
	flag.Parse()

	app := application.Application{}
	app.Build(confPath)

	app.Run()
}
