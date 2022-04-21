package api

import (
	"fmt"
	"late/storage"
	"late/utils"
	"log"
	"net/http"
)

// @Tags register
// @Summary Confirm new user registration
// @Description Usually user makes this request when opening link sent on email.
// @ID get-register
// @Produce  html
// @Param   token   query    string  true    "Registration token, sent by POST /register"
// @Success 200 string string "Request result described on HTML page"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /register [get]
func (c *Controller) GetRegistration(r *http.Request) (interface{}, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	user_id, token_err := c.storage.ApplyToken(storage.RegisterToken, token, ip)
	web_err = translateTokenErr(token_err)

	var resp string
	switch web_err {
	case TokenUnknown:
		resp = genHtmlResp([]string{
			`Эта ссылка более не действительна.`,
			`Если вы ещё не зарегистрировались, то отправьте новый запрос на регистрацию.`,
		})
	case TokenBoundToOtherIP:
		resp = genHtmlResp([]string{
			`Эта ссылка была отправлена для другого IP адреса!`,
			`Если вы хотите пройти регистрацию с этого IP, то отправьте новый запрос на регистрацию.`,
		})
	case NoError:
		user := c.storage.GetUserById(*user_id)
		resp = genHtmlResp([]string{
			"Регистрация успешно завершена!",
			fmt.Sprintf("%s, теперь вы можете зайти в свой профиль.</p>", user.Name),
		})
	default:
		panic("Not handled error")
	}
	return &resp, web_err
}

// @Tags register
// @Summary Register new user
// @Description On success user will receive confirmation link on specified email.
// @ID post-register
// @Produce  json
// @Param   email   formData    string  true    "User email"
// @Param   pass   formData    string  true    "User password"
// @Param   name   formData    string  true    "User name"
// @Success 200 {object} api.APINoError "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 100, 101, 103, 200, 201, 700, 701"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /register [post]
func (c *Controller) PostRegistration(r *http.Request) (interface{}, WebError) {
	email, web_err := getFormParam(r, "email")
	if web_err != NoError {
		return nil, web_err
	}
	pass, web_err := getFormParam(r, "pass")
	if web_err != NoError {
		return nil, web_err
	}
	name, web_err := getFormParam(r, "name")
	if web_err != NoError {
		return nil, web_err
	}

	ip := getIP(r)
	log.Println(name, pass)
	token, token_err := c.storage.CreateToken(storage.RegisterToken, email, ip, name, pass)
	web_err = translateTokenErr(token_err)
	if web_err != NoError {
		return nil, web_err
	}

	if utils.EnvB("MAIL_ENABLED") {
		verify_link := fmt.Sprintf("https://%s/register?token=%s", utils.Env("WEB_DOMAIN"), *token)
		msg := fmt.Sprintf(utils.Env("MAIL_REG_MSG"), name, ip, verify_link)
		subj := utils.Env("MAIL_REG_SUBJ")
		sendMail(email, subj, msg)
		return nil, NoError
	}

	_, token_err = c.storage.ApplyToken(storage.RegisterToken, *token, ip)
	web_err = translateTokenErr(token_err)
	return nil, web_err
}
