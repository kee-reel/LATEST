package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func ProcessSolution(w http.ResponseWriter, r *http.Request) {
	var err error
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetSolution(r, &resp)
	case "POST":
		err = PostSolution(r, &resp)
	default:
		err = fmt.Errorf("Unsupported method")
	}
	HandleResponse(w, &resp, err)
}

func ProcessVerify(w http.ResponseWriter, r *http.Request) {
	var err error
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetVerify(r, &resp)
	default:
		err = fmt.Errorf("Unsupported method")
	}
	HandleResponse(w, &resp, err)
}

func ProcessLogin(w http.ResponseWriter, r *http.Request) {
	var err error
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetLogin(r, &resp)
	default:
		err = fmt.Errorf("Unsupported method")
	}
	HandleResponse(w, &resp, err)
}

func ProcessTemplate(w http.ResponseWriter, r *http.Request) {
	var err error
	resp := map[string]interface{}{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetTemplate(r, &resp)
	default:
		err = fmt.Errorf("Unsupported method")
	}
	HandleResponse(w, &resp, err)
}

func HandleResponse(w http.ResponseWriter, resp *map[string]interface{}, err error) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Failed user request, error: %s", err.Error())
		(*resp)["error"] = fmt.Sprintf("Error: %s", err.Error())
	}
	jsonResp, err := json.Marshal(*resp)
	Err(err)
	log.Printf("[RESP]: %s", jsonResp)
	w.Write(jsonResp)
}

func RecoverRequest(w http.ResponseWriter) {
	if r := recover(); r != nil {
		debug.PrintStack()
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf("{\"error\": \"%s\"}", r)
		log.Printf("[RESP]: %s", response)
		w.Write([]byte(response))
	}
}
