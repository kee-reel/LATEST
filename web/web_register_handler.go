package main

import (
	"fmt"
	"net/http"
	"net/mail"
)

// @Tags register
// @Summary Confirm new user registration
// @Description Usually user makes this request when opening link sent on email.
// @ID get-register
// @Produce  json
// @Param   token   query    string  true    "Registration token, sent by POST /register"
// @Success 200 {object} main.APINoError "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /register [get]
func GetRegistration(r *http.Request, resp *map[string]interface{}) WebError {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	if len(params[0]) != token_len {
		return TokenInvalid
	}
	ip := GetIP(r)
	return RegisterToken(ip, &params[0])
}

// @Tags register
// @Summary Register new user
// @Description On success user will receive confirmation link on specified email.
// @ID post-register
// @Produce  json
// @Param   email   formData    string  true    "User email"
// @Param   pass   formData    string  true    "User password"
// @Param   name   formData    string  true    "User name"
// @Success 200 {object} main.APINoError "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 100, 101, 103, 200, 201, 700, 701"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /register [post]
func PostRegistration(r *http.Request, resp *map[string]interface{}) WebError {
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
	name := r.FormValue("name")
	if len(name) == 0 {
		return NameNotProvided
	}
	if len(name) > 128 {
		return NameInvalid
	}

	ip := GetIP(r)
	token, web_err := CreateRegistrationToken(&email, &pass, &name, ip)
	// If user not registered yet
	if web_err == NoError {
		if EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/register?token=%s", Env("WEB_DOMAIN"), *token)
			msg := fmt.Sprintf(Env("MAIL_REG_MSG"), name, *ip, verify_link)
			subj := Env("MAIL_REG_SUBJ")
			SendMail(ip, &email, &subj, &msg)
		} else {
			web_err = RegisterToken(ip, token)
		}
	}
	return web_err
}
