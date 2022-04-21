package api

import (
	"fmt"
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
func (c *Controller) GetProfile(r *http.Request) (interface{}, WebError) {
	token, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}
	user := c.storage.GetUserById(token.UserId)
	if user == nil {
		panic(fmt.Sprintf("User %d not found", token.UserId))
	}
	return user, NoError
}
