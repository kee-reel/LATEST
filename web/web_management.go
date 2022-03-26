package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func ProcessSolution(w http.ResponseWriter, r *http.Request) {
	var err WebError
	err = MethodNotSupported
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetSolution(r, &resp)
	case "POST":
		err = PostSolution(r, &resp)
	}
	HandleResponse(w, &resp, err)
}

func ProcessVerify(w http.ResponseWriter, r *http.Request) {
	var err WebError
	err = MethodNotSupported
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetVerify(r, &resp)
	}
	HandleResponse(w, &resp, err)
}

func ProcessLogin(w http.ResponseWriter, r *http.Request) {
	var err WebError
	err = MethodNotSupported
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetLogin(r, &resp)
	}
	HandleResponse(w, &resp, err)
}

func ProcessTemplate(w http.ResponseWriter, r *http.Request) {
	var err WebError
	err = MethodNotSupported
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetTemplate(r, &resp)
	}
	HandleResponse(w, &resp, err)
}

func ProcessLanguages(w http.ResponseWriter, r *http.Request) {
	var err WebError
	err = MethodNotSupported
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		langs := GetSupportedLanguages()
		resp["langs"] = *langs
	}
	HandleResponse(w, &resp, err)
}

func HandleResponse(w http.ResponseWriter, resp *map[string]interface{}, web_err WebError) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if web_err != NoError {
		log.Printf("Failed user request, error code: %d", web_err)
		(*resp)["error"] = web_err
	}
	jsonResp, err := json.Marshal(*resp)
	Err(err)
	w.Write(jsonResp)
}

func RecoverRequest(w http.ResponseWriter) {
	r := recover()
	if r != nil {
		debug.PrintStack()
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		log.Printf("[INTERNAL ERROR]: %s", r)
		response := fmt.Sprintf("{\"error\": \"%d\"}", Internal)
		w.Write([]byte(response))
	}
}
