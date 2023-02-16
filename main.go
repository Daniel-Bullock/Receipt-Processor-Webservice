package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	http.HandleFunc("/users", getUsers)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	users := []User{
		User{Name: "Alice", Email: "alice@example.com"},
		User{Name: "Bob", Email: "bob@example.com"},
	}
	json.NewEncoder(w).Encode(users)
}
