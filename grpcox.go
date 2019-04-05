package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gusaul/grpcox/handler"
)

func main() {
	// logging conf
	f, err := os.OpenFile("log/grpcox.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// start app
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
