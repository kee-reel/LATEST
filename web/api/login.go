package api

import (
	"fmt"
	"late/models"
	"late/utils"
	"net/http"
)

type APIToken struct {
	*models.Token
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

	ip := getIP(r)
	user, pass_matched, user_exists := c.storage.GetUser(email, pass)
	if !user_exists {
		return nil, EmailUnknown
	}
	if !pass_matched {
		return nil, PasswordWrong
	}
	token := c.storage.GetTokenForConnection(user.Email, ip)
	if token == nil {
		verification_token := c.storage.CreateVerificationToken(email, ip)
		if verification_token == nil {
			return nil, EmailUnknown
		}
		if utils.EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/verify?token=%s", utils.Env("WEB_DOMAIN"), *verification_token)
			msg := fmt.Sprintf(utils.Env("MAIL_VER_MSG"), *ip, verify_link)
			subj := utils.Env("MAIL_VER_SUBJ")
			sendMail(email, &subj, &msg)
			return nil, TokenNotVerified
		} else {
			user_id, is_token_exists := c.storage.VerifyToken(ip, verification_token)
			if !is_token_exists {
				return nil, TokenUnknown
			}
			if user_id == nil {
				return nil, TokenBoundToOtherIP
			}
			token := c.storage.GetTokenForConnection(user, ip)
			if token == nil {
				panic("Can't autoverify token")
			}
		}
	}
	resp := APIToken{token, user}
	return &resp, web_err
}
