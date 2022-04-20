package api

import (
	"fmt"
	"late/utils"
	"net/http"
)

// @Tags restore
// @Summary Confirm user password restore
// @Description Usually user makes this request when opening link sent on email.
// @ID get-restore
// @Produce  html
// @Param   token   query    string  true    "Verification token, sent by POST /verify"
// @Success 200 string strgin "Request result described on HTML page"
// @Router /restore [get]
func (c *Controller) GetRestore(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	user_id, is_token_exists := c.storage.RestoreToken(ip, token_str)
	var resp string
	if !is_token_exists {
		resp = genHtmlResp([]string{
			`Эта ссылка более не действительна.`,
			`Если вы ещё не завершили восстановление пароля, то попробуйте вновь отправить запрос на изменение пароля, чтобы получить новое письмо.`,
		})
	} else if user_id == nil {
		resp = genHtmlResp([]string{
			`Эта ссылка была отправлена для другого IP адреса!`,
			`Если вы хотите восстановить пароль с этого IP, то отправьте новый запрос.`,
		})
	} else {
		user := c.storage.GetUserById(*user_id)
		resp = genHtmlResp([]string{
			`Ваш пароль успешно изменён!`,
			fmt.Sprintf("%s, теперь вы можете зайти в свой профиль.</p>", user.Name),
		})
	}
	return &resp, NoError
}

// @Tags restore
// @Summary Restore user password
// @Description On success user will receive confirmation link on specified email.
// @ID post-restore
// @Produce  json
// @Param   email   formData    string  true    "User email"
// @Param   pass   formData    string  true    "New user password"
// @Success 200 {object} api.APINoError "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 100, 101, 102, 200, 201"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /restore [post]
func (c *Controller) PostRestore(r *http.Request) (interface{}, WebError) {
	email, web_err := getFormParam(r, "email")
	if web_err != NoError {
		return nil, web_err
	}
	pass, web_err := getFormParam(r, "pass")
	if web_err != NoError {
		return nil, web_err
	}

	ip := getIP(r)
	token, is_new_token := c.storage.CreateRestoreToken(email, ip, pass)
	if token == nil {
		return nil, EmailUnknown
	}
	if is_new_token {
		if utils.EnvB("MAIL_ENABLED") {
			verify_link := fmt.Sprintf("https://%s/restore?token=%s", utils.Env("WEB_DOMAIN"), *token)
			msg := fmt.Sprintf(utils.Env("MAIL_RESTORE_MSG"), *ip, verify_link)
			subj := utils.Env("MAIL_RESTORE_SUBJ")
			sendMail(email, &subj, &msg)
		} else {
			user_id, is_token_exists := c.storage.RestoreToken(ip, token)
			if !is_token_exists {
				return nil, TokenUnknown
			}
			if user_id == nil {
				return nil, TokenBoundToOtherIP
			}
		}
	}
	return nil, NoError
}
