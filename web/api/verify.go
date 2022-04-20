package api

import (
	"fmt"
	"net/http"
)

// @Tags verify
// @Summary Verifies user connection from new IP
// @Description Usually user makes this request when opening link sent on email.
// @ID get-verify
// @Produce  html
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Success 200 string strgin "Request result described on HTML page"
// @Router /verify [get]
func (c *Controller) GetVerify(r *http.Request) (interface{}, WebError) {
	token, web_err := getUrlParam(r, "token")
	if web_err != NoError {
		return nil, web_err
	}
	ip := getIP(r)
	user_id, is_token_exists := c.storage.VerifyToken(ip, token)
	var resp string
	if !is_token_exists {
		resp = genHtmlResp([]string{
			`Эта ссылка более не действительна.`,
			`Если вы ещё не подтвердили вход с этого IP, то попробуйте вновь войти в свой профиль, чтобы получить новое письмо.`,
		})
	} else if user_id == nil {
		resp = genHtmlResp([]string{
			`Эта ссылка была отправлена для другого IP адреса!`,
			`Если вы хотите подтвердить вход с этого IP, то попробуйте вновь войти в свой профиль, чтобы получить новое письмо.`,
		})
	} else {
		user := c.storage.GetUserById(*user_id)
		resp = genHtmlResp([]string{
			`Теперь вам доступен вход с этого IP адреса!`,
			fmt.Sprintf("%s, теперь вы можете зайти в свой профиль.</p>", user.Name),
		})
	}
	return &resp, NoError
}
