// @title LATEST
// @version 0.5
// @description Main goal of this project is to provide web-service that allows teachers to create programming courses with built-in interactive excercises.
// @description Project page: https://github.com/kee-reel/LATEST

// @contact.name Kee Reel
// @contact.url https://kee-reel.com/about-ru
// @contact.email c4rb0n_unit@protonmail.com
// @host localhost:1473
// @basePath /
package main

import (
	"fmt"
	"log"
	"net/http"
	"web/api"
	"web/utils"
)

type EndpointFunc func(w http.ResponseWriter, r *http.Request)

func main() {
	var err error
	addr := fmt.Sprintf("0.0.0.0:%s", utils.Env("WEB_PORT"))
	c := api.NewController()
	for k, v := range c.MakeHandleFuncs() {
		http.HandleFunc(k, v)
	}
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
