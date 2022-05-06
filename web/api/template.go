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
// @Param   lang_id   query    int  true    "Language id of passing solution returned by GET /languages"
// @Param   task_id   query    int  false    "ID of task"
// @Success 200 {object} api.APITemplate "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304, 600, 601"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /template [get]
func (c *Controller) GetTemplate(r *http.Request) (interface{}, WebError) {
	_, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}
	lang_id_str, web_err := getUrlParam(r, "lang_id")
	if web_err != NoError {
		return nil, web_err
	}
	lang_id, err := strconv.Atoi(lang_id_str)
	if err != nil {
		return nil, LanguageNotSupported
	}
	if _, ok := c.supported_languages[lang_id]; !ok {
		return nil, LanguageNotSupported
	}

	var task_id *int
	task_id_str, web_err := getUrlParam(r, "task_id")
	if web_err == NoError {
		task_id_temp, err := strconv.Atoi(task_id_str)
		if err != nil {
			return nil, TaskIdInvalid
		}
		task_id = &task_id_temp
	}

	resp := APITemplate{
		Template: c.storage.GetTaskTemplate(task_id, lang_id),
	}
	return &resp, NoError
}
