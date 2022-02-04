package main

import (
	"fmt"
	"net/http"
)

func GetLogin(r *http.Request, resp *map[string]interface{}) error {
	query := r.URL.Query()
	params, ok := query["email"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("email is not specified")
	}
	email := params[0]
	params, ok = query["pass"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("pass is not specified")
	}
	pass := params[0]
	if len(pass) < 6 {
		return fmt.Errorf("Password is too weak, please use at least 6 characters")
	}
	ip := GetIP(r)
	token, err := GetTokenForConnection(email, pass, ip)
	if err != nil {
		return err
	}

	(*resp)["token"] = *token
	return nil
}
