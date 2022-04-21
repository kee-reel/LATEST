package api

import (
	"fmt"
	"late/storage"
	"late/utils"
	"net/http"
)

// @Tags reset
// @Summary Resets all user's progress
// @Description Delete all solutions and reset user's score
// @ID get-reset
// @Produce  json
// @Param   token   query    string  true    "Token, returned by POST /reset"
// @Success 200 {object} string "Request result described on HTML page"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /reset [get]
func (c *Controller) GetReset(r *http.Request) (interface{}, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	_, token_err := c.storage.ApplyToken(storage.DeleteToken, token, ip)
	web_err = translateTokenErr(token_err)
	var resp string
	switch web_err {
	case TokenUnknown:
		resp = genHtmlResp([]string{
			`Эта ссылка более не действительна.`,
			`Если вы ещё не сбросили прогресс, то отправьте новый запрос на сброс.`,
		})
	case TokenBoundToOtherIP:
		resp = genHtmlResp([]string{
			`Эта ссылка была отправлена для другого IP адреса!`,
			`Если вы хотите сбросить прогресс с этого IP, то отправьте новый запрос на сброс.`,
		})
	case NoError:
		resp = genHtmlResp([]string{
			"Прогресс успешно сброшен!",
			"Теперь вы можете начать всё с чистого листа.</p>",
		})
	default:
		panic("Not handled error")
	}
	return &resp, web_err
}

// @Tags reset
// @Summary Resets all user's progress
// @Description Sends mail to user's email with confirmation to delete all solutions and reset user's score.
// @ID post-reset
// @Produce  json
// @Param   token   query    string  true    "Token, returned by GET /login"
// @Success 200 {object} api.APINoError "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /reset [post]
func (c *Controller) PostReset(r *http.Request) (interface{}, WebError) {
	token, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}

	delete_token, token_err := c.storage.CreateToken(storage.DeleteToken, token.Email, token.IP)
	web_err = translateTokenErr(token_err)
	if web_err != NoError {
		return nil, web_err
	}

	if utils.EnvB("MAIL_ENABLED") {
		link := fmt.Sprintf("https://%s/reset?token=%s", utils.Env("WEB_DOMAIN"), *delete_token)
		msg := fmt.Sprintf(utils.Env("MAIL_RESET_MSG"), token.IP, link)
		subj := utils.Env("MAIL_RESET_SUBJ")
		sendMail(token.Email, subj, msg)
	}

	_, token_err = c.storage.ApplyToken(storage.DeleteToken, *delete_token, token.IP)
	web_err = translateTokenErr(token_err)
	return nil, web_err
}
