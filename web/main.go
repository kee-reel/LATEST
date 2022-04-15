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
	entry := utils.Env("WEB_ENTRY")
	addr := fmt.Sprintf("0.0.0.0:%s", utils.Env("WEB_PORT"))
	http.HandleFunc("/login", api.Login)
	http.HandleFunc("/logout", api.Logout)
	http.HandleFunc("/verify", api.Verify)
	http.HandleFunc("/restore", api.Restore)
	http.HandleFunc("/profile", api.Leaderboard)
	http.HandleFunc("/register", api.Register)
	http.HandleFunc("/template", api.Template)
	http.HandleFunc("/solution", api.Solution)
	http.HandleFunc("/languages", api.Languages)
	http.HandleFunc("/leaderboard", api.Profile)
	http.HandleFunc("/tasks/flat", api.TasksFlat)
	http.HandleFunc("/tasks/hierarchy", api.TasksHierarchy)
	is_http := utils.EnvB("WEB_HTTP")
	log.Printf("Started listening on %s%s HTTPS(%t)", addr, entry, !is_http)
	if is_http {
		err = http.ListenAndServe(addr, nil)
	} else {
		err = http.ListenAndServeTLS(addr, utils.Env("WEB_CERT_FILE"), utils.Env("WEB_KEY_FILE"), nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
