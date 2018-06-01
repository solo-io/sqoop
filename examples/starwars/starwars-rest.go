package main

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/vektah/gqlgen/example/starwars"
)

var baseResolvers = starwars.NewResolver()

// our main function
func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/hero", GetHero).Methods("GET")
	router.HandleFunc("/api/humans/{id}", GetHuman).Methods("GET")
	// needs to be POST because there's a body
	router.HandleFunc("/api/humans/{id}/friends", GetHumanFriends).Methods("POST")
	router.HandleFunc("/api/droids/{id}/friends", GetDroidFriends).Methods("POST")
	log.Printf("listening on :9000")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func GetHero(w http.ResponseWriter, r *http.Request) {
	hero, err := baseResolvers.Query_hero(r.Context(), "")
	if err != nil {
		log.Printf("request failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(hero); err != nil {
		log.Printf("encoding failed: %v", err)
	}
}

func GetHuman(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	friends, err := baseResolvers.Query_human(r.Context(), params["id"])
	if err != nil {
		log.Printf("request failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(friends); err != nil {
		log.Printf("encoding failed: %v", err)
	}
}

func GetHumanFriends(w http.ResponseWriter, r *http.Request) {
	var input starwars.Human
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("decoding json: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params := mux.Vars(r)
	input.ID = params["id"]
	friends, err := baseResolvers.Human_friends(r.Context(), &input)
	if err != nil {
		log.Printf("request failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(friends); err != nil {
		log.Printf("encoding failed: %v", err)
	}
}

func GetDroidFriends(w http.ResponseWriter, r *http.Request) {
	var input starwars.Droid
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("decoding json: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params := mux.Vars(r)
	input.ID = params["id"]
	friends, err := baseResolvers.Droid_friends(r.Context(), &input)
	if err != nil {
		log.Printf("request failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(friends); err != nil {
		log.Printf("encoding failed: %v", err)
	}
}
