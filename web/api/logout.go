package api

import (
	"late/storage"
	"net/http"
)

// @Tags logout
// @Summary Delete existing token on server
// @Description If user wants to logout on current IP - send this request.
// @ID get-logout
// @Produce  json
// @Param   token   query    string  true    "Token, returned by GET /login"
// @Success 200 {object} api.APINoError "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /logout [get]
func GetLogout(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	token, web_err := getToken(r, token_str)
	if web_err != NoError {
		return nil, web_err
	}
	storage.RemoveToken(token)
	return nil, NoError
}
