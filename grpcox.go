package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gusaul/grpcox/handler"
)

func main() {
	port := ":6969"
	muxRouter := mux.NewRouter()
	handler.Init(muxRouter)
	srv := &http.Server{
		Handler:      muxRouter,
		Addr:         "0.0.0.0" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Service started on", port)
	log.Fatal(srv.ListenAndServe())
}
