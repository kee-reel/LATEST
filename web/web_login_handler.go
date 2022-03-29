package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	gomail "gopkg.in/mail.v2"
)

func SendMail(ip *string, email *string, subject *string, message *string) {
	m := gomail.NewMessage()
	m.SetHeader("From", Env("MAIL_EMAIL"))
	m.SetHeader("To", *email)
	m.SetHeader("Subject", *subject)

	m.SetBody("text/plain", strings.Replace(*message, "\\n", "\n", -1))
	port, err := strconv.Atoi(Env("MAIL_SERVER_PORT"))
	Err(err)
	d := gomail.NewDialer(Env("MAIL_SERVER"), port, Env("MAIL_EMAIL"), Env("MAIL_PASS"))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err = d.DialAndSend(m)
	Err(err)
}

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

func GetVerify(r *http.Request, resp *map[string]interface{}) WebError {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	return VerifyToken(ip, &params[0])
}

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
