package api

import (
	"encoding/json"
	"fmt"
	"late/utils"
	"log"
	"net/http"
	"runtime/debug"
)

func TasksFlatHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetTasksFlat, nil)
}

func TasksHierarchyHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetTasksHierarchy, nil)
}

func SolutionHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, nil, PostSolution)
}

func RegistrationHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetRegistration, PostRegistration)
}

func VerifyHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetVerify, nil)
}

func LoginHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetLogin, nil)
}

func TemplateHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetTemplate, nil)
}

func LanguagesHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetLanguages, nil)
}

func RestoreHandle(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, GetRestore, PostRestore)
}

type WebMethodFunc func(r *http.Request) (interface{}, WebError)

func HandleFunc(w http.ResponseWriter, r *http.Request, get WebMethodFunc, post WebMethodFunc) {
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

	if web_err == NoError {
		if resp == nil {
			resp = APINoError{}
		}
		w.WriteHeader(http.StatusOK)
	} else if resp == nil {
		log.Printf("Failed user request, error code: %d", web_err)
		resp = APIError{web_err}
		w.WriteHeader(http.StatusBadRequest)
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
