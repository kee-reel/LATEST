package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func GetTemplate(r *http.Request, resp *map[string]interface{}) error {
	query := r.URL.Query()

	params, ok := query["token"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("Token not specified")
	}
	ip := GetIP(r)
	_, err := GetTokenData(params[0], ip, true)
	if err != nil {
		return err
	}

	params, ok = query["task_id"]
	if !ok || len(params[0]) < 1 {
		return fmt.Errorf("task_id is not specified")
	}
	task_id, err := strconv.Atoi(params[0])
	if err != nil {
		return fmt.Errorf("Task id must be a number")
	}

	template, err := GetTaskTemplate(task_id)
	if err != nil {
		return err
	}
	(*resp)["template"] = *template
	return nil
}
