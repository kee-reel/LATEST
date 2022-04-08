package api

import (
	"late/models"
	"late/storage"
	"net/http"
	"strings"
)

type APITasksFlat struct {
	Projects map[int]*models.Project `json:"projects"`
	Units    map[int]*models.Unit    `json:"units"`
	Tasks    map[int]*models.Task    `json:"tasks"`
}

func MakeFlatResponse(tasks *[]models.Task) interface{} {
	resp := APITasksFlat{
		Projects: map[int]*models.Project{},
		Units:    map[int]*models.Unit{},
		Tasks:    map[int]*models.Task{},
	}
	for _, task := range *tasks {
		_, ok := resp.Units[task.Unit.Id]
		if !ok {
			resp.Units[task.Unit.Id] = task.Unit
		}
		_, ok = resp.Projects[task.Project.Id]
		if !ok {
			resp.Projects[task.Project.Id] = task.Project
		}
		resp.Tasks[task.Id] = &task
	}
	return &resp
}

// @Tags tasks
// @Summary Get tasks data in flat structure
// @Description Returns complete data about existing tasks.
// @Description Result will be formatted in flat structure with integer ID as a key. Example could be found in responses section.
// @ID get-tasks-flat
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   ids   query    string  false    "Comma separated task IDs: \"1,17,104\". If provided - returns data only for specified tasks (including related projects and units). If any of the tasks could not be found - error 402 will be thrown."
// @Success 200 {object} main.APITasksFlat "Tasks data. additionalProp here stands for integer IDs"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304, 401, 402"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /tasks/flat [get]
func GetTasksFlat(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	token, web_err := getToken(r, token_str)
	if web_err != NoError {
		return nil, web_err
	}

	var task_ids *[]int
	var task_str_ids []string
	params, ok := r.URL.Query()["ids"]
	if ok && len(params[0]) > 1 {
		task_str_ids = strings.Split(string(params[0]), ",")
	}
	task_ids, is_valid := storage.GetTaskIdsById(&task_str_ids)
	if is_valid {
		return nil, TaskIdInvalid
	}
	if task_ids == nil {
		return nil, TaskNotFound
	}

	tasks := storage.GetTasks(token, *task_ids)
	resp := MakeFlatResponse(tasks)
	return resp, NoError
}

type APIProjectHierarchy struct {
	models.Project
	Units map[string]APIUnitHierarchy `json:"units"`
}
type APIUnitHierarchy struct {
	models.Unit
	Tasks map[string]*models.Task
}

type APITasksHierarchy map[string]APIProjectHierarchy

func MakeHierarchyResponse(tasks *[]models.Task) interface{} {
	resp := APITasksHierarchy{}
	for _, task := range *tasks {
		project, ok := resp[task.Project.FolderName]
		if !ok {
			project = APIProjectHierarchy{
				*task.Project,
				map[string]APIUnitHierarchy{},
			}
			resp[task.Project.FolderName] = project
		}

		unit, ok := project.Units[task.Unit.FolderName]
		if !ok {
			unit = APIUnitHierarchy{
				*task.Unit,
				map[string]*models.Task{},
			}
			project.Units[task.Unit.FolderName] = unit
		}

		unit.Tasks[task.FolderName] = &task
	}
	return resp
}

// @Tags tasks
// @Summary Get tasks data in hierarhical structure
// @Description Returns complete data about existing tasks.
// @Description Result will be formatted in hierarchical structure with folder_name as a key. Example could be found in responses section.
// @ID get-tasks-hierarhy
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   folders   query    string  false    "Comma separated folder names: \"sample-project,unit-1,task-1\". If provided - returns data for specified project/unit/task. Folder names must be specified in strict sequence: project->unit->task."
// @Success 200 {object} main.APITasksHierarchy "Tasks data. additionalProp here stands for integer IDs"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304, 8XX"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /tasks/hierarhy [get]
func GetTasksHierarchy(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	token, web_err := getToken(r, token_str)
	if web_err != NoError {
		return nil, web_err
	}

	var task_ids *[]int
	params, ok := r.URL.Query()["folders"]
	var task_folders []string
	if ok && len(params[0]) > 1 {
		task_folders = strings.Split(string(params[0]), ",")
		if len(task_folders) > 3 {
			return nil, TasksFoldersInvalid
		}
	}
	task_ids, web_err = storage.GetTaskIdsByFolder(&task_folders)
	if web_err != NoError {
		return nil, web_err
	}
	tasks := storage.GetTasks(token, *task_ids)
	resp := MakeHierarchyResponse(tasks)
	return resp, NoError
}
