package api

import (
	"fmt"
	"late/storage"
	"late/utils"
	"net/http"
)

// @Tags restore
// @Summary Confirm user password restore
// @Description Usually user makes this request when opening link sent on email.
// @ID get-restore
// @Produce  json
// @Param   token   query    string  true    "Verification token, sent by POST /verify"
// @Success 200 {object} main.APINoError "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /restore [get]
func GetRestore(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	user_id, is_token_exists := storage.RestoreToken(ip, token_str)
	if !is_token_exists {
		return nil, TokenUnknown
	}
	if user_id == nil {
		return nil, TokenBoundToOtherIP
	}
	return nil, NoError
}

// @Tags restore
// @Summary Restore user password
// @Description On success user will receive confirmation link on specified email.
// @ID post-restore
// @Produce  json
// @Param   email   formData    string  true    "User email"
// @Param   pass   formData    string  true    "New user password"
// @Success 200 {object} main.APINoError "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 100, 101, 102, 200, 201"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /restore [post]
func PostRestore(r *http.Request) (interface{}, WebError) {
	email, web_err := getFormParam(r, "email")
	if web_err != NoError {
		return nil, web_err
	}
	pass, web_err := getFormParam(r, "pass")
	if web_err != NoError {
		return nil, web_err
	}

	ip := getIP(r)
	token := storage.CreateRestoreToken(email, ip, pass)
	if token == nil {
		return nil, EmailUnknown
	}
	if utils.EnvB("MAIL_ENABLED") {
		verify_link := fmt.Sprintf("https://%s/restore?token=%s", utils.Env("WEB_DOMAIN"), *token)
		msg := fmt.Sprintf(utils.Env("MAIL_RESTORE_MSG"), *ip, verify_link)
		subj := utils.Env("MAIL_RESTORE_SUBJ")
		sendMail(ip, email, &subj, &msg)
	} else {
		user_id, is_token_exists := storage.RestoreToken(ip, token)
		if !is_token_exists {
			return nil, TokenUnknown
		}
		if user_id == nil {
			return nil, TokenBoundToOtherIP
		}
	}
	return nil, NoError
}
