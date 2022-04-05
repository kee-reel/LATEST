package main

import (
	"net/http"
)

type APILangsResponse struct {
    Langs []string `example:"c,py,pas"'`
}

// @Tags languages
// @Summary Get supported languages
// @Description Returns list of supported languages.
// @ID get-languages
// @Produce  json
// @Success 200 {object} main.APILangsResponse "Success"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /languages [get]
func GetLanguages(r *http.Request, resp *map[string]interface{}) WebError {
	langs := GetSupportedLanguages()
	(*resp)["langs"] = *langs
	return NoError
}
