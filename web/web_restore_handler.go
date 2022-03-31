package main

import (
	"fmt"
	"net/http"
	"net/mail"
)

func GetRestore(r *http.Request, resp *map[string]interface{}) WebError {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	return RestoreToken(ip, &params[0])
}

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
