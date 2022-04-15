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
func GetReset(r *http.Request) (interface{}, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	user, is_token_exists := storage.ResetToken(ip, token)
	var resp string
	if !is_token_exists {
		resp = genHtmlResp([]string{
			`Эта ссылка более не действительна.`,
			`Если вы ещё не сбросили прогресс, то отправьте новый запрос на сброс.`,
		})
	} else if user == nil {
		resp = genHtmlResp([]string{
			`Эта ссылка была отправлена для другого IP адреса!`,
			`Если вы хотите сбросить прогресс с этого IP, то отправьте новый запрос на сброс.`,
		})
	} else {
		resp = genHtmlResp([]string{
			"Прогресс успешно сброшен!",
			"Теперь вы можете начать всё с чистого листа.</p>",
		})
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
func PostReset(r *http.Request) (interface{}, WebError) {
	token_str, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	token, web_err := getToken(r, token_str)
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	user := storage.GetUserById(token.UserId)
	if user == nil {
		panic("Can't find user with existing token")
	}
	reset_token := storage.CreateResetToken(user.Id, ip)
	if token == nil {
		panic("Can't create reset token")
	}
	if utils.EnvB("MAIL_ENABLED") {
		link := fmt.Sprintf("https://%s/reset?token=%s", utils.Env("WEB_DOMAIN"), *reset_token)
		msg := fmt.Sprintf(utils.Env("MAIL_RESET_MSG"), *ip, link)
		subj := utils.Env("MAIL_RESET_SUBJ")
		sendMail(&user.Email, &subj, &msg)
	} else {
		user, is_token_exists := storage.ResetToken(ip, reset_token)
		if !is_token_exists {
			return nil, TokenUnknown
		}
		if user == nil {
			return nil, TokenBoundToOtherIP
		}
	}
	return nil, NoError
}
