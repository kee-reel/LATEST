package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var user_tests_re = regexp.MustCompile(`^((-?\d+;)+\n)+$`)

func ParseSolution(r *http.Request) (*Solution, error) {
	err := r.ParseMultipartForm(32 << 20)
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return nil, fmt.Errorf("Token not specified")
	}
	ip := GetIP(r)
	token, err := GetTokenData(params[0], ip)
	if err != nil {
		return nil, err
	}
	task_id_str := r.FormValue("task_id")
	if len(task_id_str) == 0 {
		return nil, fmt.Errorf("Task id is not specified")
	}
	task_id, err := strconv.Atoi(task_id_str)
	if err != nil {
		print(err)
		return nil, fmt.Errorf("Task id must be number")
	}
	task_ids := make([]int, 1)
	task_ids[0] = task_id
	tasks, err := GetTasks(task_ids)
	Err(err)
	if len(*tasks) == 0 {
		return nil, fmt.Errorf("Task not found")
	}

	var solution Solution
	task := (*tasks)[0]

	var solution_text *string
	source_text := r.FormValue("source_text")
	if source_text != "" {
		solution_text = &source_text
	} else {
		file, _, err := r.FormFile("source_file")
		if err != nil {
			return nil, fmt.Errorf("No solution file or text provided")
		}
		raw_data, err := ioutil.ReadAll(file)
		Err(err)
		str_data := string(raw_data)
		solution_text = &str_data
	}

	if len(*solution_text) > 50000 {
		return nil, fmt.Errorf("Solution is too big")
	}

	solution.Source = *solution_text
	solution.TestCases = r.FormValue("test_cases")
	if len(solution.TestCases) > 50000 {
		return nil, fmt.Errorf("Test cases string is too big")
	}
	if len(solution.TestCases) > 0 {
		solution.TestCases = strings.Replace(solution.TestCases, "\r", "", -1)
		matches := user_tests_re.MatchString(solution.TestCases)
		if !matches {
			return nil, fmt.Errorf("Test cases string have incorrect format")
		}
	}

	solution.Task = &task
	solution.Token = token
	solution.IsVerbose = r.FormValue("verbose") == "true"

	return &solution, nil
}

func FillResponse(tasks *[]Task, resp *map[string]interface{}) {
	resp_tasks := map[int]interface{}{}
	resp_units := map[int]interface{}{}
	resp_projects := map[int]interface{}{}
	for _, task := range *tasks {
		_, ok := resp_units[task.Unit.Id]
		if !ok {
			resp_units[task.Unit.Id] = map[string]interface{}{
				"name":        task.Unit.Name,
				"folder_name": task.Unit.FolderName,
			}
		}

		_, ok = resp_projects[task.Project.Id]
		if !ok {
			resp_projects[task.Project.Id] = map[string]interface{}{
				"name":        task.Project.Name,
				"folder_name": task.Project.FolderName,
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
			"number":      task.Position,
			"project_id":  task.Project.Id,
			"unit_id":     task.Unit.Id,
			"name":        task.Name,
			"folder_name": task.FolderName,
			"desc":        task.Desc,
			"language":    task.Extention,
			"input":       task_input,
			"output":      task.Output,
			"is_passed":   task.IsPassed,
		}
	}
	(*resp)["tasks"] = resp_tasks
	(*resp)["units"] = resp_units
	(*resp)["projects"] = resp_projects
}

func FillResponseFolders(tasks *[]Task, resp *map[string]interface{}) {
	for _, task := range *tasks {
		project, ok := (*resp)[task.Project.FolderName].(map[string]interface{})
		if !ok {
			project = map[string]interface{}{
				"id":    task.Project.Id,
				"name":  task.Project.Name,
				"units": map[string]interface{}{},
			}
			(*resp)[task.Project.FolderName] = project
		}

		unit, ok := project["units"].(map[string]interface{})[task.Unit.FolderName].(map[string]interface{})
		if !ok {
			unit = map[string]interface{}{
				"id":    task.Unit.Id,
				"name":  task.Unit.Name,
				"tasks": map[string]interface{}{},
			}
			project["units"].(map[string]interface{})[task.Unit.FolderName] = unit
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
		unit["tasks"].(map[string]interface{})[task.FolderName] = map[string]interface{}{
			"id":        task.Id,
			"number":    task.Position,
			"name":      task.Name,
			"desc":      task.Desc,
			"language":  task.Extention,
			"input":     task_input,
			"output":    task.Output,
			"is_passed": task.IsPassed,
		}
	}
}

func GetSolution(r *http.Request, resp *map[string]interface{}) error {
	query := r.URL.Query()

	params, ok := query["token"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("Token not specified")
	}
	ip := GetIP(r)
	_, err := GetTokenData(params[0], ip)
	if err != nil {
		return err
	}

	task_ids, err := GetTaskIds()
	Err(err)
	if len(*task_ids) == 0 {
		return fmt.Errorf("No tasks were found")
	}

	params, ok = query["folders"]
	is_folder_structure := ok && len(params[0]) >= 1 && params[0] == "true"

	tasks, err := GetTasks(*task_ids)
	Err(err)
	if len(*tasks) == 0 {
		return fmt.Errorf("No tasks were found")
	}

	if is_folder_structure {
		FillResponseFolders(tasks, resp)
	} else {
		FillResponse(tasks, resp)
	}
	return nil
}

func PostSolution(r *http.Request, resp *map[string]interface{}) error {
	solution, err := ParseSolution(r)
	if err != nil {
		return err
	}
	test_result, is_passed := BuildAndTest(solution.Task, solution)
	if is_passed {
		fail_count := GetFailedSolutions(solution)
		(*test_result)["fail_count"] = fail_count
	}
	SaveSolution(solution, is_passed)
	*resp = *test_result
	return nil
}
