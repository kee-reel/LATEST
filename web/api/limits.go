package api

import (
	"net/http"
	"web/limits"
)

// @Tags limits
// @Summary Get service call limits
// @Description Returns map with `rate` and `burst` call limits for each endpoint. Call limits is implemented with "Leaky Bucket" algorithm. `rate` - is how much calls could be made per second. `burst` - how much calls could be made in single burst.
// @ID get-limits
// @Produce  json
// @Success 200 {object} map[string]limits.Limit "Success"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /limits [get]
func (c *Controller) GetLimits(r *http.Request) (interface{}, WebError) {
	resp := map[string]limits.Limit{}
	for _, v := range c.endpoints_map {
		resp[v.path] = v.limit
	}
	return resp, NoError
}
