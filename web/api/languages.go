package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"late/utils"
	"net/http"
	"sort"
)

type APILangsResponse struct {
	Langs *[]string `json:"langs" example:"c,py,pas"'`
}

// @Tags languages
// @Summary Get supported languages
// @Description Returns list of supported languages.
// @ID get-languages
// @Produce  json
// @Success 200 {object} api.APILangsResponse "Success"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /languages [get]
func (c *Controller) GetLanguages(r *http.Request) (interface{}, WebError) {
	resp := APILangsResponse{
		Langs: c.getSupportedLanguages(),
	}
	return resp, NoError
}

func (c *Controller) getSupportedLanguages() *[]string {
	runner_url := fmt.Sprintf("http://%s:%s", utils.Env("RUNNER_HOST"), utils.Env("RUNNER_PORT"))
	response, err := http.Get(runner_url)
	utils.Err(err)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	utils.Err(err)

	var result map[string][]string
	err = json.Unmarshal([]byte(body), &result)
	utils.Err(err)

	langs := result["langs"]
	sort.Strings(langs)
	return &langs
}

func (c *Controller) isLanguageSupported(lang string) bool {
	langs := c.getSupportedLanguages()
	if len(*langs) == 0 {
		return false
	}
	idx := sort.SearchStrings(*langs, lang)
	return idx < len(*langs) && (*langs)[idx] == lang
}
