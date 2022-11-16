package main

import "service/internal/application"

// @title Balance Server API
// @version 1.0
// @description API Server for avito backend internship

// @host

const (
	path = "deployment/service.yml"
)

func main() {
	app := application.Application{}
	app.Build(path)

	app.Run()
}
