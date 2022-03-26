package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var user_tests_re = regexp.MustCompile(`^((-?\d+;)+\n)+$`)

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
			"input":     task_input,
			"output":    task.Output,
			"is_passed": task.IsPassed,
		}
	}
}

func GetSolution(r *http.Request, resp *map[string]interface{}) WebError {
	query := r.URL.Query()

	params, ok := query["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	token, web_err := GetTokenData(&params[0], ip, true)
	if web_err != NoError {
		return web_err
	}

	params, ok = query["folders"]
	is_folder_structure := ok && len(params) >= 1 && params[0] == "true"

	var task_ids *[]int
	if is_folder_structure {
		params, ok = query["task_folders"]
		var task_folders []string
		if ok && len(params[0]) > 1 {
			task_folders = strings.Split(string(params[0]), ",")
			if len(task_folders) > 3 {
				return SolutionTaskFoldersInvalid
			}
		}
		task_ids, web_err = GetTaskIdsByFolder(&task_folders)
	} else {
		var task_str_ids []string
		params, ok = query["task_ids"]
		if ok && len(params[0]) > 1 {
			task_str_ids = strings.Split(string(params[0]), ",")
		}
		task_ids, web_err = GetTaskIdsById(&task_str_ids)
	}
	if web_err != NoError {
		return web_err
	}

	tasks := GetTasks(token, *task_ids)
	if is_folder_structure {
		FillResponseFolders(tasks, resp)
	} else {
		FillResponse(tasks, resp)
	}
	return NoError
}

func ParseSolution(r *http.Request) (*Solution, WebError) {
	err := r.ParseMultipartForm(32 << 20)
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return nil, TokenNotProvided
	}
	ip := GetIP(r)
	token, web_err := GetTokenData(&params[0], ip, true)
	if web_err != NoError {
		return nil, web_err
	}
	lang := r.FormValue("lang")
	if len(lang) == 0 {
		return nil, LanguageNotProvided
	}

	if !IsLanguageSupported(lang) {
		return nil, LanguageNotSupported
	}

	task_id_str := r.FormValue("task_id")
	if len(task_id_str) == 0 {
		return nil, TaskIdNotProvided
	}
	task_id, err := strconv.Atoi(task_id_str)
	if err != nil {
		return nil, TaskIdInvalid
	}
	task_ids := make([]int, 1)
	task_ids[0] = task_id
	tasks := GetTasks(token, task_ids)
	if len(*tasks) == 0 {
		return nil, TaskNotFound
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
			return nil, SolutionTextNotProvided
		}
		raw_data, err := ioutil.ReadAll(file)
		Err(err)
		str_data := string(raw_data)
		solution_text = &str_data
	}

	if len(*solution_text) > 50000 {
		return nil, SolutionTextTooLong
	}

	solution.Source = *solution_text
	solution.TestCases = r.FormValue("test_cases")
	if len(solution.TestCases) > 50000 {
		return nil, SolutionTestsTooLong
	}
	if len(solution.TestCases) > 0 {
		solution.TestCases = strings.Replace(solution.TestCases, "\r", "", -1)
		matches := user_tests_re.MatchString(solution.TestCases)
		if !matches {
			return nil, SolutionTestsInvalid
		}
	}

	solution.Task = &task
	solution.Token = token
	solution.IsVerbose = r.FormValue("verbose") == "true"
	solution.Extention = lang

	return &solution, NoError
}

func PostSolution(r *http.Request, resp *map[string]interface{}) WebError {
	solution, err := ParseSolution(r)
	if err != NoError {
		return err
	}
	test_result, is_passed := BuildAndTest(solution.Task, solution)
	if is_passed {
		fail_count := GetFailedSolutions(solution)
		(*test_result)["fail_count"] = fail_count
	}
	SaveSolution(solution, is_passed)
	*resp = *test_result
	return NoError
}
