package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"web/limits"
	"web/storage"
	"web/tokens"
	"web/utils"
	"web/workers"
)

type EndpointType int

const (
	Register       EndpointType = 0
	Login                       = 1
	Verify                      = 2
	Restore                     = 3
	Logout                      = 4
	Suspend                     = 5
	TasksFlat                   = 6
	TasksHierarchy              = 7
	Languages                   = 8
	Template                    = 9
	Solution                    = 10
	Leaderboard                 = 11
	Profile                     = 12
	Limits                      = 13
)

type HttpFunc func(w http.ResponseWriter, r *http.Request)

type webMethodFunc func(r *http.Request) (interface{}, WebError)

type endpointData struct {
	path  string
	limit limits.Limit
	get   webMethodFunc
	post  webMethodFunc
}

type Controller struct {
	storage             *storage.Storage
	workers             *workers.Workers
	tokens              *tokens.Tokens
	limits              *limits.Limits
	endpoints_map       map[EndpointType]endpointData
	supported_languages map[int]string
}

func NewController() *Controller {
	s := storage.NewStorage()
	c := Controller{
		s,
		workers.NewWorkers(),
		tokens.NewTokens(s),
		limits.NewLimits(),
		nil,
		s.GetLanguages(),
	}
	if len(c.supported_languages) == 0 {
		panic("No supported languages found")
	}
	c.endpoints_map = makeEndpointDataMap(&c)
	return &c
}

func makeEndpointDataMap(c *Controller) map[EndpointType]endpointData {
	return map[EndpointType]endpointData{
		Register:       {"/register", limits.Limit{0.2, 1}, c.GetRegister, c.PostRegister},
		Login:          {"/login", limits.Limit{0.2, 1}, c.GetLogin, nil},
		Verify:         {"/verify", limits.Limit{0.2, 1}, c.GetVerify, nil},
		Restore:        {"/restore", limits.Limit{0.2, 1}, c.GetRestore, c.PostRestore},
		Logout:         {"/logout", limits.Limit{0.2, 1}, c.GetLogout, nil},
		Suspend:        {"/suspend", limits.Limit{0.2, 1}, c.GetSuspend, c.PostSuspend},
		TasksFlat:      {"/tasks/flat", limits.Limit{0.2, 1}, c.GetTasksFlat, nil},
		TasksHierarchy: {"/tasks/hierarchy", limits.Limit{0.2, 1}, c.GetTasksHierarchy, nil},
		Languages:      {"/languages", limits.Limit{0.2, 1}, c.GetLanguages, nil},
		Template:       {"/template", limits.Limit{1, 2}, c.GetTemplate, nil},
		Solution:       {"/solution", limits.Limit{1, 10}, c.GetSolution, c.PostSolution},
		Leaderboard:    {"/leaderboard", limits.Limit{0.2, 1}, c.GetLeaderboard, nil},
		Profile:        {"/profile", limits.Limit{0.2, 1}, c.GetProfile, nil},
		Limits:         {"/limits", limits.Limit{0.2, 1}, c.GetLimits, nil},
	}
}

func writeError(w http.ResponseWriter, web_err WebError) {
	writeErrorWithData(w, &map[string]interface{}{"error": web_err})
}

func writeErrorWithData(w http.ResponseWriter, data *map[string]interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	str_json, err := json.Marshal(data)
	utils.Err(err)
	w.Write(str_json)
}

func (c *Controller) MakeHandleFuncs() map[string]HttpFunc {
	m := map[string]HttpFunc{}
	for k, v := range c.endpoints_map {
		m[v.path] = c.makeHandleFunc(k, v)
	}
	return m
}

func (c *Controller) makeHandleFunc(e EndpointType, data endpointData) HttpFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var f webMethodFunc
		var is_ip_client_id bool
		switch r.Method {
		case "GET":
			f = data.get
			is_ip_client_id = (e == Login || e == Limits)
		case "POST":
			f = data.post
			is_ip_client_id = (e == Register || e == Restore)
		}
		if f == nil {
			writeError(w, MethodNotSupported)
			return
		}

		var client_id string
		if is_ip_client_id {
			client_id = getIP(r)
		} else {
			var web_err WebError
			client_id, web_err = getUrlParam(r, "token")
			if web_err != NoError {
				writeError(w, web_err)
				return
			}
		}

		need_to_wait := c.limits.HandleCall(int(e), client_id, &data.limit)
		if need_to_wait == 0 {
			c.HandleFunc(w, r, f)
		} else {
			writeErrorWithData(w, &map[string]interface{}{
				"error": LimitsExceeded,
				"wait":  need_to_wait,
			})
		}
	}
}

func (c *Controller) HandleFunc(w http.ResponseWriter, r *http.Request, f webMethodFunc) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	var web_err WebError
	var resp interface{}
	defer RecoverRequest(w)
	resp, web_err = f(r)

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
