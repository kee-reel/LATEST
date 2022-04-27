package api

import (
	"late/tokens"
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
func (c *Controller) GetLogout(r *http.Request) (interface{}, WebError) {
	token, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}
	c.tokens.RemoveToken(tokens.AccessToken, token.Token, token.IP)
	return nil, NoError
}
