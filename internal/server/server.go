package server

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"ritual/internal/api"
	"ritual/internal/web"
)

var Mux *http.ServeMux

func MakeMux() {
	Mux = http.NewServeMux()
	web.Register(Mux)
	api.Register(Mux)
}

func WebServe() {
	web.LoadTemplates() // This whole template thing needs to be reevaluated during the web rewrite.
	srv := &http.Server{
		Addr:         ":1771",
		Handler:      Mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func SocketServe() {
	const sockPath = "/tmp/ritual.sock"

	os.Remove(sockPath)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(ln, Mux))
}
