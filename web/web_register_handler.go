package main

import (
	"fmt"
	"net/http"
	"net/mail"
)

func GetRegistration(r *http.Request, resp *map[string]interface{}) WebError {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	return RegisterToken(ip, &params[0])
}

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
	if len(name) > 50 {
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
