package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/hb-go/conn"
	"github.com/hb-go/conn/benchmark/dashboard"
	"github.com/hb-go/conn/benchmark/handler"
)

func main() {
	go func() {
		http.Handle("/dashboard", http.HandlerFunc(dashboard.Index))
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	srv, _ := conn.NewServer()
	srv.ConnHandler = handler.OnConn
	srv.ReadHandler = handler.Reader
	srv.WriteHandler = handler.Writer

	log.Fatal(srv.ListenAndServe("tcp", ":8080"))

}
