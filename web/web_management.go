package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

var user_tests_re = regexp.MustCompile(`^((-?\d+;)+\n)+$`)

func ParseSolution(r *http.Request) *Solution {
	err := r.ParseMultipartForm(32 << 20)
	Err(err)
	token_str := r.FormValue("token")
	token, err := GetTokenData(token_str)
	Err(err)
	task_id_str := r.FormValue("task_id")
	if len(task_id_str) == 0 {
		panic("Task id is not specified")
	}
	task_id, err := strconv.Atoi(task_id_str)
	Err(err)
	task_ids := make([]int, 1)
	task_ids[0] = task_id
	tasks, err := GetTasks(task_ids, token)
	Err(err)

	var solution Solution
	task := (*tasks)[0]

	var solution_text *string
	source_text := r.FormValue("source_text")
	if source_text != "" {
		solution_text = &source_text
	} else {
		file, _, err := r.FormFile("source_file")
		if err != nil {
			panic("No solution file or text provided")
		}
		raw_data, err := ioutil.ReadAll(file)
		Err(err)
		str_data := string(raw_data)
		solution_text = &str_data
	}

	if len(*solution_text) > 50000 {
		panic("Solution is too big")
	}

	solution.Source = *solution_text
	solution.TestCases = r.FormValue("test_cases")
	if len(solution.TestCases) > 50000 {
		panic("Test cases string is too big")
	}
	if len(solution.TestCases) > 0 {
		solution.TestCases = strings.Replace(solution.TestCases, "\r", "", -1)
		matches := user_tests_re.MatchString(solution.TestCases)
		if !matches {
			panic("Test cases string have incorrect format")
		}
	}

	solution.Task = &task
	solution.Token = token
	solution.IsVerbose = r.FormValue("verbose") == "true"

	return &solution
}

func GetSolution(r *http.Request, resp *map[string]interface{}) {
	var err error
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		panic("Parameter 'token' not specified")
	}
	token, err := GetTokenData(params[0])
	Err(err)
	task_ids, err := GetTasksByToken(token)
	Err(err)
	if len(*task_ids) == 0 {
		panic("No tasks were found")
	}
	tasks, err := GetTasks(*task_ids, token)
	Err(err)

	resp_tasks := map[int]interface{}{}
	resp_units := map[int]interface{}{}
	resp_projects := map[int]interface{}{}
	for _, task := range *tasks {
		_, ok := resp_units[task.Unit.Id]
		if !ok {
			resp_units[task.Unit.Id] = map[string]interface{}{
				"name":    task.Unit.Name,
			}
		}

		_, ok = resp_projects[task.Project.Id]
		if !ok {
			resp_projects[task.Project.Id] = map[string]interface{}{
				"name": task.Project.Name,
			}
		}

		task_input := []map[string]interface{}{}
		for _, input := range task.Input {
			task_input = append(task_input, map[string]interface{}{
				"name":       input.Name,
				"type":       input.Type,
				"dimensions": input.Dimensions,
				"range":      input.Range,
			})
		}
		resp_tasks[task.Id] = map[string]interface{}{
			"number":    task.Position,
			"project":   task.Project.Id,
			"unit":      task.Unit.Id,
			"name":      task.Name,
			"desc":      task.Desc,
			"input":     task_input,
			"output":    task.Output,
			"is_passed": task.IsPassed,
		}
	}
	(*resp)["tasks"] = resp_tasks
	(*resp)["units"] = resp_units
	(*resp)["projects"] = resp_projects
}

func PostSolution(r *http.Request, resp *map[string]interface{}) error {
	solution := ParseSolution(r)
	test_result, test_err := BuildAndTest(solution.Task, solution)
	SaveSolution(solution, test_err == nil)
	if test_result != nil {
		fail_count, err := GetFailedSolutions(solution)
		if err != nil {
			log.Print(err)
		}
		(*test_result)["fail_count"] = fail_count
	}
	if test_err != nil {
		return test_err
	}
	(*resp)["result"] = *test_result
	return nil
}

func ProcessSolution(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	var err error

	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		GetSolution(r, &resp)
	case "POST":
		err = PostSolution(r, &resp)
	default:
		Err(fmt.Errorf("Only GET and POST methods are supported"))
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		resp["error"] = fmt.Sprintf("Error: %s", err.Error())
	}
	jsonResp, err := json.Marshal(resp)
	Err(err)
	log.Printf("[RESP]: %s", jsonResp)
	w.Write(jsonResp)
}

func GetIP(r *http.Request) string {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	Err(err)
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip
	}
	panic("Can't resolve client's ip")
}

func GetLogin(r *http.Request, resp *map[string]interface{}) error {
	query := r.URL.Query()
	params, ok := query["email"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("email is not specified")
	}
	email := params[0]
	params, ok = query["pass"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("pass is not specified")
	}
	pass := params[0]
	if len(pass) < 6 {
		return fmt.Errorf("Password is too weak, please use at least 6 characters")
	}
	ip := GetIP(r)
	token, err := GetTokenForConnection(email, pass, ip)
	if err != nil {
		return err
	}

	(*resp)["token"] = *token
	return nil
}

func ProcessLogin(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	var err error

	defer RecoverRequest(w)
	switch r.Method {
	case "GET":
		err = GetLogin(r, &resp)
	case "POST":
	default:
		err = fmt.Errorf("Only GET and POST methods are supported")
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Failed user request, error: %s", err.Error())
		resp["error"] = fmt.Sprintf("Error: %s", err.Error())
	}
	jsonResp, err := json.Marshal(resp)
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
