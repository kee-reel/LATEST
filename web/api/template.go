package api

import (
	"late/storage"
	"net/http"
)

type APITemplate struct {
	Template string `json:"template" example:"#include <stdio.h>\nint main() {\n\t\n}"`
}

// @Tags template
// @Summary Returns solution template for specified language
// @Description Template for solution for all supported languages
// @ID get-template
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   lang   formData    string  true    "Language of template"
// @Success 200 {object} api.APITemplate "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304, 600, 601"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /template [get]
func GetTemplate(r *http.Request) (interface{}, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	_, web_err = getToken(r, token)
	if web_err != NoError {
		return nil, web_err
	}
	lang, web_err := getUrlParam(r, "lang")
	if web_err != NoError {
		return nil, web_err
	}

	resp := APITemplate{
		Template: *storage.GetTaskTemplate(lang),
	}
	return &resp, NoError
}
