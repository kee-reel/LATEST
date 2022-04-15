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
	"fmt"
	"late/api"
	"late/utils"
	"log"
	"net/http"
	"reflect"
	"runtime"
)

type EndpointFunc func(w http.ResponseWriter, r *http.Request)

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func main() {
	var err error
	addr := fmt.Sprintf("0.0.0.0:%s", utils.Env("WEB_PORT"))
	mux := http.NewServeMux()
	mux.HandleFunc("/login", api.Login)
	mux.HandleFunc("/logout", api.Logout)
	mux.HandleFunc("/verify", api.Verify)
	mux.HandleFunc("/restore", api.Restore)
	mux.HandleFunc("/profile", api.Profile)
	mux.HandleFunc("/register", api.Register)
	mux.HandleFunc("/template", api.Template)
	mux.HandleFunc("/solution", api.Solution)
	mux.HandleFunc("/languages", api.Languages)
	mux.HandleFunc("/leaderboard", api.Leaderboard)
	mux.HandleFunc("/tasks/flat", api.TasksFlat)
	mux.HandleFunc("/tasks/hierarchy", api.TasksHierarchy)
	mux.HandleFunc("/reset", api.Reset)
	is_http := utils.EnvB("WEB_HTTP")
	log.Printf("Started listening on %s HTTPS(%t)", addr, !is_http)
	if is_http {
		err = http.ListenAndServe(addr, mux)
	} else {
		err = http.ListenAndServeTLS(addr, utils.Env("WEB_CERT_FILE"), utils.Env("WEB_KEY_FILE"), mux)
	}
	if err != nil {
		log.Fatal(err)
	}
}
