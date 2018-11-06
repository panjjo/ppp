package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type restResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	TS   int64       `json:"ts"`
}

func restAPI() {
	router := httprouter.New()

	log.Fatal(http.ListenAndServe(config.Sys.ADDR, router))
}
