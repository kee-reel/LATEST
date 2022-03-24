package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	gomail "gopkg.in/mail.v2"
)

func GetLogin(r *http.Request, resp *map[string]interface{}) error {
	query := r.URL.Query()
	params, ok := query["email"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("email is not specified")
	}
	email := params[0]
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email is malformed")
	}
	params, ok = query["pass"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("pass is not specified")
	}
	pass := params[0]
	if len(pass) < 6 {
		return fmt.Errorf("Password is too weak, please use at least 6 characters")
	}

	ip := GetIP(r)
	token, err := GetTokenForConnection(email, pass, ip)
	if err != nil {
		return err
	}

	if !token.IsVerified {
		if EnvB("MAIL_ENABLE") {
			err := SendMail(email, token)
			if err != nil {
				return err
			}
			return fmt.Errorf("Email for this IP is not verified, please open link that was sent to your email")
		}
		VerifyToken(token)
	}

	(*resp)["token"] = token.Token
	return nil
}

func SendMail(email string, token *Token) error {
	m := gomail.NewMessage()
	m.SetHeader("From", Env("MAIL_EMAIL"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", Env("MAIL_SUBJECT"))
	verify_link := fmt.Sprintf("https://%s%sverify?token=%s", Env("WEB_HOST"), Env("WEB_ENTRY"), token.Token)
	m.SetBody("text/plain",
		fmt.Sprintf(strings.Replace(Env("MAIL_MSG"), "\\n", "\n", -1),
			token.IP, verify_link))
	d := gomail.NewDialer("smtp.yandex.com", 465, Env("MAIL_EMAIL"), Env("MAIL_PASS"))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := d.DialAndSend(m)
	return err
}

func GetVerify(r *http.Request, resp *map[string]interface{}) error {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("Token not specified")
	}
	ip := GetIP(r)
	token, err := GetTokenData(params[0], ip, false)
	if err != nil {
		return err
	}

	if !token.IsVerified {
		VerifyToken(token)
	}
	return nil
}
