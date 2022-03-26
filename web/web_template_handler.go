package main

import (
	"net/http"
)

func GetTemplate(r *http.Request, resp *map[string]interface{}) WebError {
	query := r.URL.Query()

	params, ok := query["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	_, web_err := GetTokenData(&params[0], ip, true)
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
