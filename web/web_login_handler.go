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
	token, web_err := GetTokenForConnection(email, pass, ip)
	if web_err != NoError {
		return web_err
	}

	if !token.IsVerified {
		if EnvB("MAIL_ENABLE") {
			web_err := SendMail(email, token)
			if web_err != NoError {
				return web_err
			}
			return TokenNotVerified
		}
		VerifyToken(token)
	}

	(*resp)["token"] = token.Token
	return NoError
}

func SendMail(email string, token *Token) WebError {
	m := gomail.NewMessage()
	m.SetHeader("From", Env("MAIL_EMAIL"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", Env("MAIL_SUBJECT"))
	verify_link := fmt.Sprintf("https://%s%sverify?token=%s", Env("WEB_HOST"), Env("WEB_ENTRY"), token.Token)
	m.SetBody("text/plain", fmt.Sprintf(strings.Replace(Env("MAIL_MSG"), "\\n", "\n", -1),
		token.IP, verify_link))
	port, err := strconv.Atoi(Env("MAIL_SERVER_PORT"))
	Err(err)
	d := gomail.NewDialer(Env("MAIL_SERVER"), port, Env("MAIL_EMAIL"), Env("MAIL_PASS"))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err = d.DialAndSend(m)
	Err(err)
	return NoError
}

func GetVerify(r *http.Request, resp *map[string]interface{}) WebError {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	token, err := GetTokenData(params[0], ip, false)
	if err == NoError && !token.IsVerified {
		VerifyToken(token)
	}
	return err
}
