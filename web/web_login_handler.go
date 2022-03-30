package main

import (
	"fmt"
	"net/http"
	"net/mail"
)

func GetLogin(r *http.Request, resp *map[string]interface{}) WebError {
	query := r.URL.Query()
	params, ok := query["email"]
	if !ok || len(params[0]) < 1 {
		return EmailNotProvided
	}
	email := params[0]
	_, err := mail.ParseAddress(email)
	if err != nil {
		return EmailInvalid
	}
	params, ok = query["pass"]
	if !ok || len(params[0]) < 1 {
		return PasswordNotProvided
	}
	pass := params[0]
	if len(pass) < 6 {
		return PasswordInvalid
	}

	ip := GetIP(r)
	token, web_err := GetTokenForConnection(&email, &pass, ip)
	if web_err == TokenNotVerified {
		verification_token := CreateVerificationToken(token.UserId, ip)
		if EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/verification?token=%s", Env("WEB_DOMAIN"), *verification_token)
			msg := fmt.Sprintf(Env("MAIL_VER_MSG"), *ip, verify_link)
			subj := Env("MAIL_VER_SUBJ")
			SendMail(ip, &email, &subj, &msg)
		} else {
			web_err = VerifyToken(ip, verification_token)
		}
	} else if web_err == NoError {
		(*resp)["token"] = token.Token
	}
	return web_err
}
