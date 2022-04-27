package api

import (
	"fmt"
	"late/models"
	"late/storage"
	"late/utils"
	"net/http"
)

type APIToken struct {
	Token string `json:"token" example:"9rzNUDp8bP6VOnGIqOO011f5EB4jk0eN0osZt0KFZHTtWIpiwqzVj2vof5sOq80QIJbne5dHiH5vEUe7uJ42X5X39tHGpt0LTreFOjMkfdn4sB6gzouUHc4tGubhikoKuK05P06W1x0QK0zJzbPaZYG4mfBpfU1u8xbqSPVo8ZI9zumiJUiHC8MbJxMPYsGJjZMChQBtA0NvKuAReS3v1704QBX5zZCAyyNP47VZ51E9MMqVGoZBxFmJ4mCHRBy7"`
	*models.User
}

// @Tags login
// @Summary Get token for registered user
// @Description Returns token that could be used in other requests.
// @Description If user logs in from new IP, verification link will be sent on email.
// @ID get-login
// @Produce  json
// @Param   email   query    string  true    "User email address"
// @Param   pass    query    string  true    "User password. Must be at least 6 symbols"
// @Success 200 {object} api.APIToken "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 100, 101, 102, 200, 201, 202, 303"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /login [get]
func (c *Controller) GetLogin(r *http.Request) (interface{}, WebError) {
	email, web_err := getUrlParam(r, "email")
	if web_err != NoError {
		return nil, web_err
	}
	pass, web_err := getUrlParam(r, "pass")
	if web_err != NoError {
		return nil, web_err
	}

	user, user_exists := c.storage.AuthenticateUser(email, pass)
	if !user_exists {
		return nil, EmailUnknown
	}
	if user == nil {
		return nil, PasswordWrong
	}

	ip := getIP(r)
	token := c.storage.GetToken(storage.AccessToken, email, ip)
	if token == nil {
		verification_token, token_err := c.storage.CreateToken(storage.VerifyToken, email, ip)
		web_err = translateTokenErr(token_err)
		if web_err != NoError {
			return nil, web_err
		}

		if utils.EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/verify?token=%s", utils.Env("WEB_DOMAIN"), *verification_token)
			msg := fmt.Sprintf(utils.Env("MAIL_VER_MSG"), ip, verify_link)
			subj := utils.Env("MAIL_VER_SUBJ")
			sendMail(email, subj, msg)
			return nil, TokenNotVerified
		}

		_, token_err = c.storage.ApplyToken(storage.VerifyToken, *verification_token, ip)
		web_err = translateTokenErr(token_err)
		if web_err != NoError {
			return nil, web_err
		}

		token = c.storage.GetToken(storage.AccessToken, email, ip)
		if token == nil {
			panic("Auto verify failed")
		}
	}
	resp := APIToken{
		Token: *token,
		User:  user,
	}
	return &resp, NoError
}
