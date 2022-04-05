package main

import (
	"fmt"
	"net/http"
	"net/mail"
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
func GetRestore(r *http.Request, resp *map[string]interface{}) WebError {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	if len(params[0]) != token_len {
		return TokenInvalid
	}
	ip := GetIP(r)
	return RestoreToken(ip, &params[0])
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
func PostRestore(r *http.Request, resp *map[string]interface{}) WebError {
	email := r.FormValue("email")
	if len(email) == 0 {
		return EmailNotProvided
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return EmailInvalid
	}
	pass := r.FormValue("pass")
	if len(pass) == 0 {
		return PasswordNotProvided
	}
	if len(pass) < 6 {
		return PasswordInvalid
	}

	ip := GetIP(r)
	token, web_err := CreateRestoreToken(&email, ip, &pass)
	if web_err == NoError {
		if EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/restore?token=%s", Env("WEB_DOMAIN"), *token)
			msg := fmt.Sprintf(Env("MAIL_RESTORE_MSG"), *ip, verify_link)
			subj := Env("MAIL_RESTORE_SUBJ")
			SendMail(ip, &email, &subj, &msg)
		} else {
			web_err = RestoreToken(ip, token)
		}
	}
	return web_err
}
