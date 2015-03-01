package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
)

var port = ":8080"
var storage map[string]map[string]string

func generateToken(w http.ResponseWriter, r *http.Request) {
	id := uuid()
	create(id)
	fmt.Fprint(w, id)
}

func collect(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	token := params["token"]

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := string(body)
	if data == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "No value to collect")
		return
	}

	if save(token, data) == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Token not found: %s", token)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func list(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	token := params["token"]

	keys := make([]string, 0, len(storage[token]))
	for k := range storage[token] {
		keys = append(keys, k)
	}

	fmt.Fprint(w, keys)
}

func tokenValues(token string) map[string]string {
	return storage[token]
}

func create(token string) {
	if storage[token] == nil {
		storage[token] = make(map[string]string)
	}
}

func save(token string, data string) string {
	// TODO: better storage
	if storage[token] != nil {
		storage[token][data] = data
		return token
	}
	return ""
}

func initStorage() {
	storage = make(map[string]map[string]string)
}

func uuid() string {
	ut, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(strings.TrimSpace(string(ut)), "-", "", -1)
}

func main() {
	initStorage()
	fmt.Printf("Starting token service on port: %s\n", port)

	router := mux.NewRouter()
	router.HandleFunc("/v1/token", generateToken).Methods("GET")
	router.HandleFunc("/v1/collect/{token:[a-z0-9]+}", collect).Methods("POST")
	router.HandleFunc("/v1/token/{token:[a-z0-9]+}", list).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(port, nil)
}
