package main

import (
	"log"

	"github.com/devlsc/distributed_services_with_go/proglog/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
