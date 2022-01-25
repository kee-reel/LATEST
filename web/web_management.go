package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var user_tests_re = regexp.MustCompile(`^((-?\d+;)+\n)+$`)

func ParseSolution(r *http.Request) (*[]Solution, *UserData, error) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, nil, err
	}
	token_str := r.FormValue("token")
	token, err := GetTokenData(token_str)
	if err != nil {
		return nil, nil, err
	}
	var user_data UserData
	tasks_raw := r.FormValue("tasks")
	if len(tasks_raw) == 0 {
		return nil, nil, errors.New("List of tasks is empty")
	}
	var task_ids []int
	for _, task_id_str := range strings.Split(string(tasks_raw), ",") {

		task_id, err := strconv.Atoi(task_id_str)
		if err != nil {
			return nil, nil, err
		}
		task_ids = append(task_ids, task_id)
	}
	tasks, err := GetTasks(task_ids, token)
	if err != nil {
		return nil, nil, err
	}
	is_verbose := r.FormValue("verbose") == "true"
	solutions := make([]Solution, len(*tasks))
	for i, task := range *tasks {
		solution := &solutions[i]
		solution.Task = &task
		file_name := fmt.Sprintf("source_%d", task.Id)
		file, _, err := r.FormFile(file_name)
		if err != nil {
			log.Printf("Can't open file %s", file_name)
			return nil, nil, err
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, nil, err
		}
		solution.Source = string(data)
		solution.TestCases = r.FormValue(fmt.Sprintf("test_cases_%s", task.Id))
		if len(solution.TestCases) > 50000 {
			return nil, nil, errors.New("Test cases string is too big")
		}
		if len(solution.TestCases) > 0 {
			solution.TestCases = strings.Replace(solution.TestCases, "\r", "", -1)
			matches := user_tests_re.MatchString(solution.TestCases)
			if !matches {
				return nil, nil, errors.New("Test cases string have incorrect format")
			}
		}
		solution.Token = token
		solution.IsVerbose = is_verbose
	}
	return &solutions, &user_data, err
}

func GetQueryParam(r *http.Request, key string) (*[]int, error) {
	params, ok := r.URL.Query()[key]
	if !ok || len(params[0]) < 1 {
		return nil, fmt.Errorf("Parameter %s not found in query", key)
	}
	res := make([]int, len(params))
	for i, str_value := range params {
		value, err := strconv.Atoi(str_value)
		if err != nil {
			return nil, err
		}
		res[i] = value
	}
	return &res, nil
}

func GetSolution(r *http.Request, resp *map[string]interface{}) error {
	var err error
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return errors.New("Parameter 'token' not specified")
	}
	token, err := GetTokenData(params[0])
	if err != nil {
		return err
	}
	task_ids, err := GetTasksByToken(token)
	if err != nil {
		return err
	}
	if len(*task_ids) == 0 {
		return errors.New("No tasks were found")
	}
	tasks, err := GetTasks(*task_ids, token)
	if err != nil {
		return err
	}

	resp_tasks := map[int]interface{}{}
	resp_units := map[int]interface{}{}
	resp_projects := map[int]interface{}{}
	for _, task := range *tasks {
		_, ok := resp_units[task.Unit.Id]
		if !ok {
			resp_units[task.Unit.Id] = map[string]interface{}{
				"name":    task.Unit.Name,
				"next_id": task.Unit.NextId,
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
			"project":   task.Project,
			"unit":      task.Unit,
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
	return nil
}

func PostSolution(r *http.Request, resp *map[string]interface{}) error {
	solutions, _, err := ParseSolution(r)
	if err != nil {
		return err
	}
	solution_results := map[int]interface{}{}
	for _, solution := range *solutions {
		test_result, test_err := BuildAndTest(solution.Task, &solution)
		SaveSolution(&solution, test_err == nil)
		if test_result != nil {
			fail_count, err := GetFailedSolutions(&solution)
			if err != nil {
				log.Print(err)
			}
			(*test_result)["fail_count"] = fail_count
		}
		if test_err != nil {
			return test_err
		}
		solution_results[solution.Task.Id] = *test_result
	}
	(*resp)["result"] = solution_results
	return nil
}

func ProcessSolution(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{}
	var err error

	switch r.Method {
	case "GET":
		err = GetSolution(r, &resp)
	case "POST":
		err = PostSolution(r, &resp)
	default:
		err = errors.New("Only GET and POST methods are supported")
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Failed user request, error: %s", err.Error())
		resp["error"] = fmt.Sprintf("Error: %s", err.Error())
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Can't prepare response JSON, error: %s", err.Error())
		jsonResp = []byte(`{"error": "Error happened in response JSON creation"`)
	}
	log.Printf("[RESP]: %s", jsonResp)
	w.Write(jsonResp)
}
