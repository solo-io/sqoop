package main

import (
	"log"
	"github.com/solo-io/qloo/examples/starwars/server"
	"net/http"
	"fmt"
)

func main() {
	port := 1234
	log.Printf("listening on :%v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), server.New()))
}