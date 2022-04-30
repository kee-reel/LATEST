package api

import (
	"fmt"
	"web/tokens"
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
	user_id, token_err := c.tokens.ApplyToken(tokens.VerifyToken, token, ip)
	web_err = translateTokenErr(token_err)

	var resp string
	switch web_err {
	case TokenUnknown:
		resp = genHtmlResp([]string{
			`Эта ссылка более не действительна.`,
			`Если вы ещё не подтвердили вход с этого IP, то попробуйте вновь войти в свой профиль, чтобы получить новое письмо.`,
		})
	case TokenBoundToOtherIP:
		resp = genHtmlResp([]string{
			`Эта ссылка была отправлена для другого IP адреса!`,
			`Если вы хотите подтвердить вход с этого IP, то попробуйте вновь войти в свой профиль, чтобы получить новое письмо.`,
		})
	case NoError:
		user := c.storage.GetUserById(*user_id)
		resp = genHtmlResp([]string{
			`Теперь вам доступен вход с этого IP адреса!`,
			fmt.Sprintf("%s, теперь вы можете зайти в свой профиль.</p>", user.Name),
		})
	default:
		panic("Not handled error")
	}
	return &resp, NoError
}
