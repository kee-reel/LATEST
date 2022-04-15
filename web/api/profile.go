package api

import (
	"fmt"
	"late/storage"
	"net/http"
)

// @Tags profile
// @Summary Get information about user by token
// @Description Returns user's score, email and name by token
// @ID get-profile
// @Produce  json
// @Param   token   query    string  true    "Token, returned by GET /login"
// @Success 200 {object} models.User "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /profile [get]
func GetLeaderboard(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	token, web_err := getToken(r, token_str)
	if web_err != NoError {
		return nil, web_err
	}
	user := storage.GetUserById(token.UserId)
	if user == nil {
		panic(fmt.Sprintf("User %d not found", token.UserId))
	}
	return user, NoError
}
