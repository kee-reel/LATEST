package api

import (
	"net/http"
)

type APILangsResponse struct {
	Langs []string `json:"langs" example:"c,py,pas"'`
}

// @Tags languages
// @Summary Get supported languages
// @Description Returns list of supported languages.
// @ID get-languages
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Success 200 {object} api.APILangsResponse "Success"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /languages [get]
func (c *Controller) GetLanguages(r *http.Request) (interface{}, WebError) {
	_, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}
	j := 0
	langs := make([]string, len(c.supported_languages))
	for k := range c.supported_languages {
		langs[j] = k
		j++
	}
	resp := APILangsResponse{
		Langs: langs,
	}
	return resp, NoError
}
