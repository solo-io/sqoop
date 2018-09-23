package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/solo-io/sqoop/examples/starwars/server"
)

func main() {
	port := 1234
	log.Printf("listening on :%v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), server.New()))
}
