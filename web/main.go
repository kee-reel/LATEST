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

func main() {
	var err error
	entry := utils.Env("WEB_ENTRY")
	addr := fmt.Sprintf("0.0.0.0:%s", utils.Env("WEB_PORT"))
	http.HandleFunc(fmt.Sprintf("%sverify", entry), api.VerifyHandle)
	http.HandleFunc(fmt.Sprintf("%stemplate", entry), api.TemplateHandle)
	http.HandleFunc(fmt.Sprintf("%slanguages", entry), api.LanguagesHandle)
	http.HandleFunc(fmt.Sprintf("%srestore", entry), api.RestoreHandle)
	http.HandleFunc(fmt.Sprintf("%stasks/hierarchy", entry), api.TasksHierarchyHandle)
	http.HandleFunc(fmt.Sprintf("%ssolution", entry), api.SolutionHandle)
	http.HandleFunc(fmt.Sprintf("%slogin", entry), api.LoginHandle)
	http.HandleFunc(fmt.Sprintf("%sregister", entry), api.RegistrationHandle)
	http.HandleFunc(fmt.Sprintf("%stasks/flat", entry), api.TasksFlatHandle)
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
