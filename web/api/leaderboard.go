package api

import (
	"net/http"
)

// @Tags leaderboard
// @Summary Get information about users scores
// @Description Returns scores for all users
// @ID get-leaderboard
// @Produce  json
// @Param   token   query    string  true    "Token, returned by GET /login"
// @Success 200 {object} models.Leaderboard "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /leaderboard [get]
func (c *Controller) GetLeaderboard(r *http.Request) (interface{}, WebError) {
	_, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}
	leaderboard := c.storage.GetLeaderboard()
	return leaderboard, NoError
}
