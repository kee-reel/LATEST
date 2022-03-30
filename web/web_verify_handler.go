package main

import (
	"net/http"
)

func GetVerify(r *http.Request, resp *map[string]interface{}) WebError {
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return TokenNotProvided
	}
	ip := GetIP(r)
	return VerifyToken(ip, &params[0])
}
