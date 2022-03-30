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
	http.HandleFunc(entry, SolutionHandle)
	http.HandleFunc(fmt.Sprintf("%slogin", entry), LoginHandle)
	http.HandleFunc(fmt.Sprintf("%sverify", entry), VerifyHandle)
	http.HandleFunc(fmt.Sprintf("%stemplate", entry), TemplateHandle)
	http.HandleFunc(fmt.Sprintf("%slanguages", entry), LanguagesHandle)
	http.HandleFunc(fmt.Sprintf("%sregister", entry), RegistrationHandle)
	http.HandleFunc(fmt.Sprintf("%srestore", entry), RestoreHandle)
	addr := fmt.Sprintf("0.0.0.0:%s", Env("WEB_PORT"))
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
