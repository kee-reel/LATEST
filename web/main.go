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
)

type EndpointFunc func(w http.ResponseWriter, r *http.Request)

func main() {
	var err error
	addr := fmt.Sprintf("0.0.0.0:%s", utils.Env("WEB_PORT"))
	c := api.NewController()
	http.HandleFunc("/login", c.Login)
	http.HandleFunc("/logout", c.Logout)
	http.HandleFunc("/verify", c.Verify)
	http.HandleFunc("/restore", c.Restore)
	http.HandleFunc("/profile", c.Profile)
	http.HandleFunc("/register", c.Register)
	http.HandleFunc("/template", c.Template)
	http.HandleFunc("/solution", c.Solution)
	http.HandleFunc("/languages", c.Languages)
	http.HandleFunc("/leaderboard", c.Leaderboard)
	http.HandleFunc("/tasks/flat", c.TasksFlat)
	http.HandleFunc("/tasks/hierarchy", c.TasksHierarchy)
	http.HandleFunc("/reset", c.Reset)

	is_http := utils.EnvB("WEB_HTTP")
	log.Printf("Started listening on %s HTTPS(%t)", addr, !is_http)
	if is_http {
		err = http.ListenAndServe(addr, nil)
	} else {
		err = http.ListenAndServeTLS(addr, utils.Env("WEB_CERT_FILE"), utils.Env("WEB_KEY_FILE"), nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
