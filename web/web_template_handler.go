package main

import (
	"net/http"
)

type APITemplate struct {
	Template string `example:"#include <stdio.h>\nint main() {\n\t\n}"`
}

// @Tags template
// @Summary Returns solution template for specified language
// @Description Template for solution for all supported languages
// @ID get-template
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   lang   formData    string  true    "Language of template"
// @Success 200 {object} main.APITemplate "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304, 600, 601"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /template [get]
func GetTemplate(r *http.Request, resp *map[string]interface{}) WebError {
	query := r.URL.Query()

	params, ok := query["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	_, web_err := GetTokenData(&params[0], ip)
	if web_err != NoError {
		return web_err
	}

	params, ok = query["lang"]
	if !ok || len(params[0]) < 1 {
		return LanguageNotProvided
	}

	lang := params[0]
	if !IsLanguageSupported(lang) {
		return LanguageNotSupported
	}

	template := GetTaskTemplate(lang)
	(*resp)["template"] = *template
	return NoError
}
