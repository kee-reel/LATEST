package api

import (
	"encoding/json"
	"fmt"
	"late/storage"
	"late/tokens"
	"late/utils"
	"late/workers"
	"log"
	"net/http"
	"runtime/debug"
)

type Controller struct {
	storage *storage.Storage
	workers *workers.Workers
	tokens  *tokens.Tokens
}

func NewController() *Controller {
	s := storage.NewStorage()
	return &Controller{
		s,
		workers.NewWorkers(),
		tokens.NewTokens(s),
	}
}

func (c *Controller) TasksFlat(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetTasksFlat, nil)
}

func (c *Controller) TasksHierarchy(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetTasksHierarchy, nil)
}

func (c *Controller) Solution(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetSolution, c.PostSolution)
}

func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetRegistration, c.PostRegistration)
}

func (c *Controller) Verify(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetVerify, nil)
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetLogin, nil)
}

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetLogout, nil)
}

func (c *Controller) Profile(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetProfile, nil)
}

func (c *Controller) Template(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetTemplate, nil)
}

func (c *Controller) Languages(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetLanguages, nil)
}

func (c *Controller) Restore(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetRestore, c.PostRestore)
}

func (c *Controller) Leaderboard(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetLeaderboard, nil)
}

func (c *Controller) Suspend(w http.ResponseWriter, r *http.Request) {
	HandleFunc(w, r, c.GetSuspend, c.PostSuspend)
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
