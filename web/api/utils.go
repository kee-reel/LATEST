package api

import (
	"crypto/tls"
	"fmt"
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

func getUrlParam(r *http.Request, name string) (string, WebError) {
	params, ok := r.URL.Query()[name]
	var value string
	if ok && len(params[0]) > 0 {
		value = params[0]
	}
	return validateParam(name, value)
}

func getFormParam(r *http.Request, name string) (string, WebError) {
	value := r.FormValue(name)
	return validateParam(name, value)
}

func validateParam(name string, value string) (string, WebError) {
	switch name {
	case "token":
		if len(value) == 0 {
			return "", TokenNotProvided
		}
		if security.IsTokenInvalid(value) {
			return "", TokenInvalid
		}
	case "email":
		if len(value) == 0 {
			return "", EmailNotProvided
		}
		if _, err := mail.ParseAddress(value); err != nil {
			return "", EmailInvalid
		}
	case "pass":
		if len(value) == 0 {
			return "", PasswordNotProvided
		}
		if len(value) < 6 {
			return "", PasswordInvalid
		}
	case "name":
		if len(value) == 0 {
			return "", NameNotProvided
		}
		if len(value) > 128 {
			return "", NameInvalid
		}
	case "lang":
		if len(value) == 0 {
			return "", LanguageNotProvided
		}
	case "task_id":
		if len(value) == 0 {
			return "", TaskIdNotProvided
		}
	default:
		panic("Unsupported parameter")
	}
	return value, NoError
}

func (c *Controller) getToken(r *http.Request) (*models.Token, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}

	ip := getIP(r)
	token_data, token_err := c.storage.GetTokenData(storage.AccessToken, token, ip)
	switch token_err {
	case storage.TokenUnknown:
		return nil, TokenUnknown
	case storage.WrongIP:
		return nil, TokenBoundToOtherIP
	}
	return &models.Token{
		Token:  token,
		Email:  token_data.Email,
		IP:     token_data.IP,
		UserId: token_data.UserId,
	}, NoError
}

func translateTokenErr(token_err storage.TokenError) WebError {
	switch token_err {
	case storage.NoError:
		return NoError
	case storage.TokenUnknown:
		return TokenUnknown
	case storage.TokenExists:
		return TokenExists
	case storage.EmailTaken:
		return EmailTaken
	case storage.EmailUnknown:
		return EmailUnknown
	case storage.WrongIP:
		return TokenBoundToOtherIP
	}
	panic(fmt.Sprintf("Error %s not handled", token_err))
}

func getIP(r *http.Request) string {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	utils.Err(err)
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip
	}
	panic("Can't resolve client's ip")
}

func sendMail(email string, subject string, message string) {
	m := gomail.NewMessage()
	m.SetHeader("From", utils.Env("MAIL_EMAIL"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)

	m.SetBody("text/plain", strings.Replace(message, "\\n", "\n", -1))
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
