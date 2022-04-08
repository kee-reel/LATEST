package api

import (
	"crypto/tls"
	"late/models"
	"late/utils"
	"net"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
)

func getUrlParam(r *http.Request, name string) (*string, WebError) {
	params, ok := r.URL.Query()[name]
	var value *string
	if ok && len(params[0]) > 0 {
		value = &params[0]
	}
	return validateParam(name, value)
}

func getFormParam(r *http.Request, name string) (*string, WebError) {
	value := r.FormValue(name)
	if len(value) == 0 {
		return validateParam(name, nil)
	}
	return validateParam(name, &value)
}

func validateParam(name *string, value *string) (*string, WebError) {
	switch name {
	case "token":
		if value == nil {
			return nil, TokenNotProvided
		}
		if IsTokenInvalid(token) {
			return nil, TokenInvalid
		}
	case "email":
		if value == 0 {
			return nil, EmailNotProvided
		}
		_, err := mail.ParseAddress(*value)
		if err != nil {
			return nil, EmailInvalid
		}
	case "pass":
		if value == nil {
			return nil, PasswordNotProvided
		}
		if len(*value) < 6 {
			return nil, PasswordInvalid
		}
	case "name":
		if value == nil {
			return nil, NameNotProvided
		}
		if len(*value) > 128 {
			return nil, NameInvalid
		}
	case "lang":
		if value == nil {
			return nil, LanguageNotProvided
		}
		if !isLanguageSupported(&lang) {
			return nil, LanguageNotSupported
		}
	case "task_id":
		if value == nil {
			return nil, TaskIdNotProvided
		}
	default:
		panic("Unsupported parameter")
	}
}

func getToken(r *http.Request, token_str *string) (*models.Token, WebError) {
	token := GetTokenData(token_str)
	if token == nil {
		return nil, TokenUnknown
	}
	ip := GetIP(r)
	if *ip != token.IP {
		return nil, TokenBoundToOtherIP
	}
	return token, NoError
}

func getIP(r *http.Request) *string {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return &ip
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return &ip
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	utils.Err(err)
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return &ip
	}
	panic("Can't resolve client's ip")
}

func sendMail(ip *string, email *string, subject *string, message *string) {
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
