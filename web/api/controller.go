package api

import (
	"encoding/json"
	"fmt"
	"web/limits"
	"web/storage"
	"web/tokens"
	"web/utils"
	"web/workers"
	"log"
	"net/http"
	"runtime/debug"
)

type methodType int

const (
	Get  methodType = 0
	Post            = 1
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
)

type Controller struct {
	storage       *storage.Storage
	workers       *workers.Workers
	tokens        *tokens.Tokens
	limits        *limits.Limits
	endpoints_map webMethodFuncMap
}

func NewController() *Controller {
	s := storage.NewStorage()
	c := Controller{
		s,
		workers.NewWorkers(),
		tokens.NewTokens(s),
		limits.NewLimits(
			map[int]limits.Limit{
				Login:          {2, 5},
				Verify:         {2, 5},
				Restore:        {2, 5},
				Logout:         {2, 5},
				Suspend:        {2, 5},
				TasksFlat:      {2, 5},
				TasksHierarchy: {2, 5},
				Languages:      {2, 5},
				Template:       {2, 5},
				Solution:       {2, 5},
				Leaderboard:    {2, 5},
				Profile:        {2, 5},
			},
		),
		webMethodFuncMap{},
	}
	c.endpoints_map = makeHandleFuncMap(&c)
	return &c
}

func makeHandleFuncMap(c *Controller) webMethodFuncMap {
	return webMethodFuncMap{
		Login:          {Get: c.GetLogin},
		Verify:         {Get: c.GetVerify},
		Restore:        {Get: c.GetRestore, Post: c.PostRestore},
		Logout:         {Get: c.GetLogout},
		Suspend:        {Get: c.GetSuspend, Post: c.PostSuspend},
		TasksFlat:      {Get: c.GetTasksFlat},
		TasksHierarchy: {Get: c.GetTasksHierarchy},
		Languages:      {Get: c.GetLanguages},
		Template:       {Get: c.GetTemplate},
		Solution:       {Get: c.GetSolution, Post: c.PostSolution},
		Leaderboard:    {Get: c.GetLeaderboard},
		Profile:        {Get: c.GetProfile},
	}
}

func (c *Controller) MakeHandleFunc(e EndpointType) HttpFunc {
	f_get := c.endpoints_map[e][Get]
	f_post := c.endpoints_map[e][Post]
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		if c.limits.HandleCall(int(e), ip) {
			var f webMethodFunc
			switch r.Method {
			case "GET":
				f = f_get
			case "POST":
				f = f_post
			}
			c.HandleFunc(w, r, f)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			response := fmt.Sprintf("{\"error\": \"%d\"}", LimitsExceeded)
			w.Write([]byte(response))
		}
	}
}

type HttpFunc func(w http.ResponseWriter, r *http.Request)

type webMethodFunc func(r *http.Request) (interface{}, WebError)

type webMethodFuncMap map[EndpointType]map[methodType]webMethodFunc

func (c *Controller) HandleFunc(w http.ResponseWriter, r *http.Request, f webMethodFunc) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	var web_err WebError
	web_err = MethodNotSupported
	var resp interface{}
	defer RecoverRequest(w)
	if f == nil {
		web_err = MethodNotSupported
	} else {
		resp, web_err = f(r)
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
