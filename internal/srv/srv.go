package srv

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"ritual/internal/api"
	"ritual/internal/web"
)

var Mux *http.ServeMux

const SocketPath string = "/tmp/ritual.sock"

func MakeMux() {
	Mux = http.NewServeMux()
	web.Register(Mux)
	api.Register(Mux)
}

func WebServe() {
	web.LoadTemplates() // This whole template thing needs to be reevaluated during the web rewrite.
	webSrv := &http.Server{
		Addr:         ":1771",
		Handler:      Mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(webSrv.ListenAndServe())
}

func SocketServe() {
	if err := os.Remove(SocketPath); err != nil {
		log.Fatal(err)
	}

	sockSrv, err := net.Listen("unix", SocketPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(sockSrv, Mux))
}

func NewSocketClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", SocketPath)
			},
		},
	}
}
