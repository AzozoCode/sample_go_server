package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	Name string `json:"name"`
}

var cacheUser = make(map[int]User)

var cacheMutex sync.RWMutex

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("POST /user", handleCreateUser)
	mux.HandleFunc("GET /user/{id}", handleGetUser)
	mux.HandleFunc("GET /user", handleGetUsers)
	mux.HandleFunc("DELETE /user/{id}", handleDeleteUser)

	fmt.Println("listening on port:8080")

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "Hello world")
	w.WriteHeader(http.StatusOK)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {

	var user User
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	cacheUser[len(cacheUser)+1] = user
	cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheMutex.RLock()
	user, ok := cacheUser[id]
	cacheMutex.RUnlock()

	if !ok {
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {

	var users = make([]User, 0)

	cacheMutex.RLock()
	for _, value := range cacheUser {
		users = append(users, value)
	}

	cacheMutex.RUnlock()

	//j, err := json.Marshal(users)
	err := json.NewEncoder(w).Encode(users)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//w.Write(j)

}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	delete(cacheUser, id)
	cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}
