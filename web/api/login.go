package api

import (
	"fmt"
	"late/storage"
	"late/utils"
	"net/http"
)

type ResponseToken struct {
	Name  string `example:"User name"`
	Token string `example:"9rzNUDp8bP6VOnGIqOO011f5EB4jk0eN0osZt0KFZHTtWIpiwqzVj2vof5sOq80QIJbne5dHiH5vEUe7uJ42X5X39tHGpt0LTreFOjMkfdn4sB6gzouUHc4tGubhikoKuK05P06W1x0QK0zJzbPaZYG4mfBpfU1u8xbqSPVo8ZI9zumiJUiHC8MbJxMPYsGJjZMChQBtA0NvKuAReS3v1704QBX5zZCAyyNP47VZ51E9MMqVGoZBxFmJ4mCHRBy7"`
}

// @Tags login
// @Summary Get token for registered user
// @Description Returns token that could be used in other requests.
// @Description If user logs in from new IP, verification link will be sent on email.
// @ID get-login
// @Produce  json
// @Param   email   query    string  true    "User email address"
// @Param   pass    query    string  true    "User password. Must be at least 6 symbols"
// @Success 200 {object} main.APIToken "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 100, 101, 102, 200, 201, 202, 303"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /login [get]
func GetLogin(r *http.Request) (interface{}, WebError) {
	email, web_err := getUrlParam(r, "email")
	if web_err != NoError {
		return nil, web_err
	}
	pass, web_err := getUrlParam(r, "email")
	if web_err != NoError {
		return nil, web_err
	}

	ip := getIP(r)
	user, pass_matched := storage.GetUser(email, pass)
	if user == nil {
		return nil, EmailUnknown
	}
	if !pass_matched {
		return nil, PasswordWrong
	}
	token := storage.GetTokenForConnection(user, ip)
	if token == nil {
		verification_token := storage.CreateVerificationToken(email, ip)
		if verification_token == nil {
			return nil, EmailUnknown
		}
		if utils.EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/verify?token=%s", utils.Env("WEB_DOMAIN"), *verification_token)
			msg := fmt.Sprintf(utils.Env("MAIL_VER_MSG"), *ip, verify_link)
			subj := utils.Env("MAIL_VER_SUBJ")
			sendMail(ip, email, &subj, &msg)
		} else {
			user_id, is_token_exists := storage.VerifyToken(ip, verification_token)
			if !is_token_exists {
				return nil, TokenUnknown
			}
			if user_id == nil {
				return nil, TokenBoundToOtherIP
			}
		}
	}
	resp := ResponseToken{
		Token: token.Token,
		Name:  user.Name,
	}
	return &resp, web_err
}
