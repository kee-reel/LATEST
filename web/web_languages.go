package main

import (
	"net/http"
)

func GetLanguages(r *http.Request, resp *map[string]interface{}) WebError {
	langs := GetSupportedLanguages()
	(*resp)["langs"] = *langs
	return NoError
}
