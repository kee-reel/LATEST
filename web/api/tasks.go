package api

import (
	"late/models"
	"late/storage"
	"net/http"
	"strings"
)

// @Tags tasks
// @Summary Get tasks data in flat structure
// @Description Returns complete data about existing tasks.
// @Description Result will be formatted in flat structure with integer ID as a key. Example could be found in responses section.
// @ID get-tasks-flat
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   ids   query    string  false    "Comma separated task IDs, for example: `ids=1,17,104`. If provided - returns data only for specified tasks (including related projects and units). If any of the tasks could not be found - error 402 will be thrown."
// @Success 200 {object} api.APITasksFlat "Tasks data. additionalProp here stands for integer IDs"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304, 401, 402"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
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

	task_ids_str, ok := r.URL.Query()["ids"]
	var task_ids_str_arr []string
	if ok && len(task_ids_str) > 1 {
		task_ids_str_arr = strings.Split(task_ids_str[0], ",")
	}
	task_ids, is_valid := storage.GetTaskIdsById(&task_ids_str_arr)
	if !is_valid {
		return nil, TaskIdInvalid
	}
	if task_ids == nil {
		return nil, TaskNotFound
	}

	tasks := storage.GetTasks(token, *task_ids)
	resp := MakeFlatResponse(tasks)
	return resp, NoError
}

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
	for i, task := range *tasks {
		_, ok := resp.Units[task.Unit.Id]
		if !ok {
			resp.Units[task.Unit.Id] = task.Unit
		}
		_, ok = resp.Projects[task.Project.Id]
		if !ok {
			resp.Projects[task.Project.Id] = task.Project
		}
		resp.Tasks[task.Id] = &(*tasks)[i]
	}
	return &resp
}

// @Tags tasks
// @Summary Get tasks data in hierarchical structure
// @Description Returns complete data about existing tasks.
// @Description Result will be formatted in hierarchical structure with folder_name as a key. Example could be found in responses section.
// @ID get-tasks-hierarchy
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   folders   query    string  false    "Comma separated folder names, for example: `folders=sample-project,unit-1,task-1`. If provided - returns data for specified project/unit/task. Folder names must be specified in strict sequence: project->unit->task."
// @Success 200 {object} api.APITasksHierarchy "Tasks data. additionalProp here stands for folder_name of project/unit/task"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304, 8XX"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /tasks/hierarchy [get]
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
	task_ids, not_found_index := storage.GetTaskIdsByFolder(&task_folders)
	switch not_found_index {
	case 0:
		return nil, TasksProjectFolderNotFound
	case 1:
		return nil, TasksUnitFolderNotFound
	case 2:
		return nil, TasksTaskFolderNotFound
	}
	if len(*task_ids) == 0 {
		return nil, TaskNotFound
	}
	tasks := storage.GetTasks(token, *task_ids)
	resp := MakeHierarchyResponse(tasks)
	return resp, NoError
}

type APIProjectHierarchy struct {
	models.Project
	Units map[string]APIUnitHierarchy `json:"units"`
}
type APIUnitHierarchy struct {
	models.Unit
	Tasks map[string]*models.Task `json:"tasks"`
}

type APITasksHierarchy map[string]APIProjectHierarchy

func MakeHierarchyResponse(tasks *[]models.Task) interface{} {
	resp := APITasksHierarchy{}
	for i, task := range *tasks {
		project, ok := resp[task.Project.FolderName]
		if !ok {
			project = APIProjectHierarchy{*task.Project, map[string]APIUnitHierarchy{}}
			resp[task.Project.FolderName] = project
		}

		unit, ok := project.Units[task.Unit.FolderName]
		if !ok {
			unit = APIUnitHierarchy{*task.Unit, map[string]*models.Task{}}
			project.Units[task.Unit.FolderName] = unit
		}

		unit.Tasks[task.FolderName] = &(*tasks)[i]
	}
	return resp
}
