// main.go
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

func Env(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(errors.New(fmt.Sprintf("Env variable %s not found", key)))
	}
	return val
}

func EnvB(key string) bool {
	return Env(key) == "true"
}

func main() {
	var err error
	entry := Env("WEB_ENTRY")
	http.HandleFunc(entry, ProcessSolution)
	http.HandleFunc(fmt.Sprintf("%slogin", entry), ProcessLogin)
	http.HandleFunc(fmt.Sprintf("%stemplate", entry), ProcessTemplate)
	addr := fmt.Sprintf("%s:%s", Env("WEB_HOST"), Env("WEB_PORT"))
	is_http := EnvB("WEB_HTTP")
	log.Printf("Started listening on %s%s HTTPS(%t)", addr, entry, !is_http)
	if is_http {
		err = http.ListenAndServe(addr, nil)
	} else {
		err = http.ListenAndServeTLS(addr, Env("WEB_CERT_FILE"), Env("WEB_KEY_FILE"), nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
