package main

import (
	"fmt"
	"net/http"
	"net/mail"
)

type APITokenResponse struct {
    Token string `example:"9rzNUDp8bP6VOnGIqOO011f5EB4jk0eN0osZt0KFZHTtWIpiwqzVj2vof5sOq80QIJbne5dHiH5vEUe7uJ42X5X39tHGpt0LTreFOjMkfdn4sB6gzouUHc4tGubhikoKuK05P06W1x0QK0zJzbPaZYG4mfBpfU1u8xbqSPVo8ZI9zumiJUiHC8MbJxMPYsGJjZMChQBtA0NvKuAReS3v1704QBX5zZCAyyNP47VZ51E9MMqVGoZBxFmJ4mCHRBy7"`
}

// @Tags login
// @Summary Get token for registered user
// @Description Returns token that could be used in other requests.
// @ID get-login
// @Produce  json
// @Param   email   query    string  true    "User email address"
// @Param   pass    query    string  true    "User password. Must be at least 6 symbols"
// @Success 200 {object} main.APITokenResponse "Success"
// @Failure 400 {object} main.APIError "Possible error codes: 100, 101, 102, 200, 201, 202, 303"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /login [get]
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
		verification_token, web_err := CreateVerificationToken(&email, ip)
        if web_err != NoError {
            return web_err
        }
		if EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/verification?token=%s", Env("WEB_DOMAIN"), *verification_token)
			msg := fmt.Sprintf(Env("MAIL_VER_MSG"), *ip, verify_link)
			subj := Env("MAIL_VER_SUBJ")
			SendMail(ip, &email, &subj, &msg)
		} else {
			web_err = VerifyToken(ip, verification_token)
		}
	} else if web_err == NoError {
        (*resp)["name"] = GetUserName(token.UserId)
		(*resp)["token"] = token.Token
	}
	return web_err
}
