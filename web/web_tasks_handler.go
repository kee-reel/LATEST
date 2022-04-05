package main

import (
	"net/http"
	"strings"
)

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

type APIProjectHierarchy struct {
	Id    int    `example:"666"`
	Name  string `example:"Sample project"`
	Units []APIUnitHierarchy
}
type APIUnitHierarchy struct {
	Id         int    `example:"666"`
	FolderName string `json:"folder_name" example:"unit-1"`
	Tasks      []APITaskHierarchy
}
type APIProjectFlat struct {
	Name       string `example:"Sample project"`
	FolderName string `json:"folder_name" example:"sample_project"`
}
type APIUnitFlat struct {
	Name       string `example:"Sample unit"`
	FolderName string `json:"folder_name" example:"unit-1"`
}
type APITaskInput struct {
	Name       string   `example:"task-1"`
	Type       string   `example:"int"`
	Dimensions []int    `example:"5,4"`
	Range      []string `example:"-1000,1000"`
}
type APITaskFlat struct {
	Number     int    `example:"0"`
	UnitId     int    `example:"1"`
	Name       string `example:"Sample task"`
	FolderName string `example:"task-1"`
	Desc       string `example:"Sample description"`
	Input      []APITaskInput
	Output     string `example:"Description of output format"`
	IsPassed   bool   `example:"true"`
}
type APITaskHierarchy struct {
	Id       int    `example:"666"`
	Number   int    `example:"0"`
	UnitId   int    `example:"1"`
	Name     string `example:"Sample task"`
	Desc     string `example:"Sample description"`
	Input    []APITaskInput
	Output   string `example:"Description of output format"`
	IsPassed bool   `example:"true"`
}

type APITasksDataFlat struct {
	Projects map[int]APIProjectFlat
	Units    map[int]APIUnitFlat
	Tasks    map[int]APITaskFlat
}

type APITasksDataHierarchy map[string]APIProjectHierarchy

// @Tags tasks
// @Summary Get tasks data in flat structure
// @Description Returns complete data about existing tasks.
// @Description Result will be formatted in flat structure with integer ID as a key. Example could be found in responses section.
// @ID get-tasks-flat
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   ids   query    string  false    "Comma separated task IDs: \"1,17,104\". If provided - returns data only for specified tasks (including related projects and units). If any of the tasks could not be found - error 402 will be thrown."
// @Success 200 {object} main.APITasksDataFlat "Tasks data. additionalProp here stands for integer IDs"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304, 401, 402"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /tasks/flat [get]
func GetTasksFlat(r *http.Request, resp *map[string]interface{}) WebError {
	query := r.URL.Query()

	params, ok := query["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	token, web_err := GetTokenData(&params[0], ip)
	if web_err != NoError {
		return web_err
	}

	var task_ids *[]int
	var task_str_ids []string
	params, ok = query["ids"]
	if ok && len(params[0]) > 1 {
		task_str_ids = strings.Split(string(params[0]), ",")
	}
	task_ids, web_err = GetTaskIdsById(&task_str_ids)
	if web_err != NoError {
		return web_err
	}

	tasks := GetTasks(token, *task_ids)
	FillResponse(tasks, resp)
	return NoError
}

// @Tags tasks
// @Summary Get tasks data in hierarhical structure
// @Description Returns complete data about existing tasks.
// @Description Result will be formatted in hierarchical structure with folder_name as a key. Example could be found in responses section.
// @ID get-tasks-hierarhy
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   folders   query    string  false    "Comma separated folder names: \"sample-project,unit-1,task-1\". If provided - returns data for specified project/unit/task. Folder names must be specified in strict sequence: project->unit->task."
// @Success 200 {object} main.APITasksDataHierarchy "Tasks data. additionalProp here stands for integer IDs"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304, 8XX"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /tasks/hierarhy [get]
func GetTasksHierarchy(r *http.Request, resp *map[string]interface{}) WebError {
	query := r.URL.Query()

	params, ok := query["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	token, web_err := GetTokenData(&params[0], ip)
	if web_err != NoError {
		return web_err
	}

	var task_ids *[]int
	params, ok = query["folders"]
	var task_folders []string
	if ok && len(params[0]) > 1 {
		task_folders = strings.Split(string(params[0]), ",")
		if len(task_folders) > 3 {
			return TasksFoldersInvalid
		}
	}
	task_ids, web_err = GetTaskIdsByFolder(&task_folders)
	if web_err != NoError {
		return web_err
	}
	tasks := GetTasks(token, *task_ids)
	FillResponseFolders(tasks, resp)
	return NoError
}
