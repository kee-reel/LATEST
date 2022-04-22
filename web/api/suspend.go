package api

import (
	"fmt"
	"late/storage"
	"late/utils"
	"net/http"
)

// @Tags suspend
// @Summary Suspends user's account
// @Description Marks user's account for deletion. In 7 days after suspension accound will be deleted. Deletion clears user's profile, task solutions and score in leaderboard.
// @ID get-suspend
// @Produce  json
// @Param   token   query    string  true    "Token, returned by POST /suspend"
// @Success 200 {object} string "Request result described on HTML page"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /reset [get]
func (c *Controller) GetSuspend(r *http.Request) (interface{}, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	_, token_err := c.storage.ApplyToken(storage.SuspendToken, token, ip)
	web_err = translateTokenErr(token_err)
	var resp string
	switch web_err {
	case TokenUnknown:
		resp = genHtmlResp([]string{
			`Эта ссылка более не действительна.`,
			`Если вы ещё не заморозили аккаунт, то отправьте новый запрос.`,
		})
	case TokenBoundToOtherIP:
		resp = genHtmlResp([]string{
			`Эта ссылка была отправлена для другого IP адреса!`,
			`Если вы хотите заморозить аккаунт с этого IP, то отправьте новый запрос на сброс.`,
		})
	case NoError:
		resp = genHtmlResp([]string{
			"Аккаунт успешно заморожен!",
			"Через семь дней аккаунт будет удалён.",
			"Если вы хотите отменить удаление аккаунта, то сообщите об этом администрации сайта.",
		})
	default:
		panic("Not handled error")
	}
	return &resp, web_err
}

// @Tags suspend
// @Summary Suspends user's account
// @Description Sends email with GET /suspend request that marks user's account for deletion.
// @ID post-suspend
// @Produce  json
// @Param   token   query    string  true    "Token, returned by GET /login"
// @Success 200 {object} api.APINoError "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /reset [post]
func (c *Controller) PostSuspend(r *http.Request) (interface{}, WebError) {
	token, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}

	delete_token, token_err := c.storage.CreateToken(storage.SuspendToken, token.Email, token.IP)
	web_err = translateTokenErr(token_err)
	if web_err != NoError {
		return nil, web_err
	}

	if utils.EnvB("MAIL_ENABLED") {
		link := fmt.Sprintf("https://%s/reset?token=%s", utils.Env("WEB_DOMAIN"), *delete_token)
		msg := fmt.Sprintf(utils.Env("MAIL_RESET_MSG"), token.IP, link)
		subj := utils.Env("MAIL_RESET_SUBJ")
		sendMail(token.Email, subj, msg)
		return nil, NoError
	}

	_, token_err = c.storage.ApplyToken(storage.SuspendToken, *delete_token, token.IP)
	web_err = translateTokenErr(token_err)
	return nil, web_err
}
