// +build !windows,!plan9

package main

import (
	"log"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/labstack/echo/engine/standard.v3.0.3"
)

func gracefulRun(std *standard.Server) {
	log.Fatal(gracehttp.Serve(std.Server))
}
