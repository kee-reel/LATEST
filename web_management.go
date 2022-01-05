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

	_ "github.com/mattn/go-sqlite3"
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
	flags, _ := GetQueryParam(r, "gen_doc")
	user_data.GenerateDoc = flags != nil && (*flags)[0] == 1
	if user_data.GenerateDoc {
		user_data.Teacher = r.FormValue("teacher")
		if len(user_data.Teacher) == 0 {
			return nil, nil, errors.New("User teacher is empty")
		}
	}
	tasks_raw := r.FormValue("tasks")
	if len(tasks_raw) == 0 {
		return nil, nil, errors.New("List of tasks is empty")
	}
	tasks := strings.Split(string(tasks_raw), ",")
	solutions := make([]Solution, len(tasks))
	for i, task := range tasks {
		solution := &solutions[i]
		solution.Task, err = strconv.Atoi(task)
		if err != nil {
			return nil, nil, err
		}
		file, _, err := r.FormFile(fmt.Sprintf("source_%s", task))
		if err != nil {
			return nil, nil, err
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, nil, err
		}
		solution.Source = string(data)
		solution.TestCases = r.FormValue(fmt.Sprintf("test_cases_%s", task))
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
	generate_doc := true
	values, _ := GetQueryParam(r, "gen_doc")
	if values != nil {
		log.Printf("Override doc generation: %s", values)
		generate_doc = (*values)[0] == 1
	}
	task_ids, err := GetWorkTasks(token)
	if err != nil {
		return err
	}
	if len(*task_ids) == 0 {
		return errors.New("No tasks were found")
	}
	resp_tasks := map[int]interface{}{}
	pages := map[int]*M{}
	for _, task_id := range *task_ids {
		task, err := GetTask(task_id)
		if err != nil {
			return err
		}
		if generate_doc {
			page_content, err := GenTaskDesc(task, nil, nil)
			if err != nil {
				return err
			}
			pages[task.Number] = page_content
		}
		resp_tasks[task.Number] = map[string]interface{}{
			"task_id": task_id,
			"name":    task.Name,
		}
	}
	if generate_doc {
		pages_content := make([]M, len(pages))
		for num, page := range pages {
			pages_content[num-1] = *page
		}
		suggested_doc_name := fmt.Sprintf("work-desc-%d-%d-%d", token.Subject, token.Work, token.Variant)
		work_desc_filename, err := GenDoc("work-desc", pages_content, nil, &suggested_doc_name)
		if err != nil {
			return err
		}
		(*resp)["link"] = work_desc_filename
	}
	(*resp)["tasks"] = resp_tasks
	return nil
}

func PostSolution(r *http.Request, resp *map[string]interface{}) error {
	solutions, user_data, err := ParseSolution(r)
	if err != nil {
		return err
	}
	pages_content := []M{}
	for _, solution := range *solutions {
		task, err := GetTask(solution.Task)
		if err != nil {
			return err
		}
		test_result, is_user_tests_passed, test_err := BuildAndTest(task, &solution)
		SaveSolution(&solution, is_user_tests_passed, test_err == nil)
		fail_count, err := GetFailedSolutions(&solution)
		if err != nil {
			log.Print(err)
		}
		(*resp)["fail_count"] = fail_count
		if test_err != nil {
			return test_err
		}
		(*resp)["result"] = *test_result
		if user_data.GenerateDoc {
			page_content, err := GenTaskDesc(task, &solution.Source, test_result)
			if err != nil {
				return err
			}
			pages_content = append(pages_content, *page_content)
		}
	}
	if user_data.GenerateDoc {
		gen_result_filename, err := GenDoc("report", pages_content, user_data, nil)
		if err != nil {
			return err
		}
		(*resp)["link"] = gen_result_filename
	}
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
