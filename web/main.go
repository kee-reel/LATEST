// @title LATE API
// @version 0.1
// @description Web service that allows to run tests for programms written in almost any language.
// @description Project page: https://github.com/kee-reel/LATE

// @contact.name Kee Reel
// @contact.url https://kee-reel.com/about-ru
// @contact.email c4rb0n_unit@protonmail.com
// @host localhost:12345
// @basePath /
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
	addr := fmt.Sprintf("0.0.0.0:%s", Env("WEB_PORT"))
	http.HandleFunc(fmt.Sprintf("%slogin", entry), LoginHandle)
	http.HandleFunc(fmt.Sprintf("%sverify", entry), VerifyHandle)
	http.HandleFunc(fmt.Sprintf("%stemplate", entry), TemplateHandle)
	http.HandleFunc(fmt.Sprintf("%slanguages", entry), LanguagesHandle)
	http.HandleFunc(fmt.Sprintf("%sregister", entry), RegistrationHandle)
	http.HandleFunc(fmt.Sprintf("%srestore", entry), RestoreHandle)
	http.HandleFunc(fmt.Sprintf("%stasks/flat", entry), TasksFlatHandle)
	http.HandleFunc(fmt.Sprintf("%stasks/hierarchy", entry), TasksHierarchyHandle)
	http.HandleFunc(fmt.Sprintf("%ssolution", entry), SolutionHandle)
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
