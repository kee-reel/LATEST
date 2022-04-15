package api

import (
	"encoding/json"
	"fmt"
	"late/utils"
	"log"
	"net/http"
	"runtime/debug"
)

func TasksFlat(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetTasksFlat, nil)
}

func TasksHierarchy(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetTasksHierarchy, nil)
}

func Solution(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetSolution, PostSolution)
}

func Register(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetRegistration, PostRegistration)
}

func Verify(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetVerify, nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetLogin, nil)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetLogout, nil)
}

func Profile(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetProfile, nil)
}

func Template(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetTemplate, nil)
}

func Languages(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetLanguages, nil)
}

func Restore(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetRestore, PostRestore)
}

func Leaderboard(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetLeaderboard, nil)
}

type WebMethodFunc func(r *http.Request) (interface{}, WebError)

func HandleFunc(w http.ResponseWriter, r *http.Request, get WebMethodFunc, post WebMethodFunc) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	var web_err WebError
	web_err = MethodNotSupported
	var resp interface{}
	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		if get == nil {
			web_err = MethodNotSupported
		} else {
			resp, web_err = get(r)
		}
	case "POST":
		if post == nil {
			web_err = MethodNotSupported
		} else {
			resp, web_err = post(r)
		}
	}

	if resp == nil {
		if web_err == NoError {
			resp = APINoError{}
		} else {
			log.Printf("Failed user request, error code: %d", web_err)
			resp = APIError{web_err}
		}
	}

	switch resp.(type) {
	case *string:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, *resp.(*string))
	default:
		w.Header().Set("Content-Type", "application/json")
		str_json, err := json.Marshal(resp)
		utils.Err(err)
		w.Write(str_json)
	}
}

func RecoverRequest(w http.ResponseWriter) {
	if r := recover(); r != nil {
		debug.PrintStack()
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		log.Printf("[INTERNAL ERROR]: %s", r)
		response := fmt.Sprintf("{\"error\": \"%d\"}", Internal)
		w.Write([]byte(response))
	}
}
