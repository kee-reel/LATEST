package api

import (
	"crypto/tls"
	"late/models"
	"late/security"
	"late/storage"
	"late/utils"
	"net"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

func getUrlParam(r *http.Request, name string) (*string, WebError) {
	params, ok := r.URL.Query()[name]
	var value *string
	if ok && len(params[0]) > 0 {
		value = &params[0]
	}
	return validateParam(&name, value)
}

func getFormParam(r *http.Request, name string) (*string, WebError) {
	value := r.FormValue(name)
	if len(value) == 0 {
		return validateParam(&name, nil)
	}
	return validateParam(&name, &value)
}

func validateParam(name *string, value *string) (*string, WebError) {
	switch *name {
	case "token":
		if value == nil {
			return nil, TokenNotProvided
		}
		if security.IsTokenInvalid(value) {
			return nil, TokenInvalid
		}
	case "email":
		if value == nil {
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
		if !isLanguageSupported(value) {
			return nil, LanguageNotSupported
		}
	case "task_id":
		if value == nil {
			return nil, TaskIdNotProvided
		}
	default:
		panic("Unsupported parameter")
	}
	return value, NoError
}

func getToken(r *http.Request, token_str *string) (*models.Token, WebError) {
	token := storage.GetTokenData(token_str)
	if token == nil {
		return nil, TokenUnknown
	}
	ip := getIP(r)
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
	m.SetHeader("From", utils.Env("MAIL_EMAIL"))
	m.SetHeader("To", *email)
	m.SetHeader("Subject", *subject)

	m.SetBody("text/plain", strings.Replace(*message, "\\n", "\n", -1))
	port, err := strconv.Atoi(utils.Env("MAIL_SERVER_PORT"))
	utils.Err(err)
	d := gomail.NewDialer(utils.Env("MAIL_SERVER"), port, utils.Env("MAIL_EMAIL"), utils.Env("MAIL_PASS"))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err = d.DialAndSend(m)
	utils.Err(err)
}

func genHtmlResp(text []string) string {
	var sb strings.Builder
	sb.WriteString(`<body style="text-align: center; color:#02c2d7; background-color:#1b1b1b; text-shadow: rgba(0, 73, 142, 0.5) 0px 0px 5px;">`)
	for _, s := range text {
		sb.WriteString("<p>")
		sb.WriteString(s)
		sb.WriteString("</p>")
	}
	sb.WriteString(`</body>`)
	return sb.String()
}
