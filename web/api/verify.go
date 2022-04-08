package api

import (
	"net/http"
)

// @Tags verify
// @Summary Verifies user connection from new IP
// @Description Usually user makes this request when opening link sent on email.
// @ID get-verify
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Success 200 {object} main.APITemplate "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /template [get]
func GetVerify(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam("token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := GetIP(r)
	user_id, is_token_exists := VerifyToken(ip, &params[0])
	if !is_token_exists {
		return nil, TokenUnknown
	}
	if user_id == nil {
		return nil, TokenBoundToOtherIP
	}

	return nil, NoError
}
