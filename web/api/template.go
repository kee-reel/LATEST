package api

import (
	"net/http"
	"strconv"
)

type APITemplate struct {
	Template string `json:"template" example:"#include <stdio.h>\nint main() {\n\t\n}"`
}

// @Tags template
// @Summary Returns solution template
// @Description Returns solution template for supported language.
// @Description Also it's possible to provide task_id:
// @Description - if there is no specific template for this task - common one will be returned
// @Description - if there is one - specific template will be returned
// @ID get-template
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   lang   query    string  true    "Language of template"
// @Param   task_id   query    int  false    "ID of task"
// @Success 200 {object} api.APITemplate "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304, 600, 601"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /template [get]
func (c *Controller) GetTemplate(r *http.Request) (interface{}, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	_, web_err = c.getToken(r, token)
	if web_err != NoError {
		return nil, web_err
	}
	lang, web_err := getUrlParam(r, "lang")
	if web_err != NoError {
		return nil, web_err
	}
	var task_id *int
	task_id_str, web_err := getUrlParam(r, "task_id")
	if web_err == NoError {
		task_id_temp, err := strconv.Atoi(*task_id_str)
		if err != nil {
			return nil, TaskIdInvalid
		}
		task_id = &task_id_temp
	}

	resp := APITemplate{
		Template: *c.storage.GetTaskTemplate(lang, task_id),
	}
	return &resp, NoError
}
