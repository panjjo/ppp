package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func restAPI() {
	router := httprouter.New()

	log.Fatal(http.ListenAndServe(config.Sys.ADDR, router))
}
