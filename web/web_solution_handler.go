package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var user_tests_re = regexp.MustCompile(`^((-?\d+;)+\n)+$`)

func ParseSolution(r *http.Request) (*Solution, WebError) {
	err := r.ParseMultipartForm(32 << 20)
	params, ok := r.URL.Query()["token"]
	if !ok || len(params[0]) < 1 {
		return nil, TokenNotProvided
	}
	ip := GetIP(r)
	token, web_err := GetTokenData(&params[0], ip)
	if web_err != NoError {
		return nil, web_err
	}
	lang := r.FormValue("lang")
	if len(lang) == 0 {
		return nil, LanguageNotProvided
	}

	if !IsLanguageSupported(lang) {
		return nil, LanguageNotSupported
	}

	task_id_str := r.FormValue("task_id")
	if len(task_id_str) == 0 {
		return nil, TaskIdNotProvided
	}
	task_id, err := strconv.Atoi(task_id_str)
	if err != nil {
		return nil, TaskIdInvalid
	}
	task_ids := make([]int, 1)
	task_ids[0] = task_id
	tasks := GetTasks(token, task_ids)
	if len(*tasks) == 0 {
		return nil, TaskNotFound
	}

	var solution Solution
	task := (*tasks)[0]

	var solution_text *string
	source_text := r.FormValue("source_text")
	if source_text != "" {
		solution_text = &source_text
	} else {
		file, _, err := r.FormFile("source_file")
		if err != nil {
			return nil, SolutionTextNotProvided
		}
		raw_data, err := ioutil.ReadAll(file)
		Err(err)
		str_data := string(raw_data)
		solution_text = &str_data
	}

	if len(*solution_text) > 50000 {
		return nil, SolutionTextTooLong
	}

	solution.Source = *solution_text
	solution.TestCases = r.FormValue("test_cases")
	if len(solution.TestCases) > 50000 {
		return nil, SolutionTestsTooLong
	}
	if len(solution.TestCases) > 0 {
		solution.TestCases = strings.Replace(solution.TestCases, "\r", "", -1)
		matches := user_tests_re.MatchString(solution.TestCases)
		if !matches {
			return nil, SolutionTestsInvalid
		}
	}

	solution.Task = &task
	solution.Token = token
	solution.IsVerbose = r.FormValue("verbose") == "true"
	solution.Extention = lang

	return &solution, NoError
}

type APISolutionVerboseResult struct {
	Params string `example:"2;1;7;'"`
	Result string `example:"8"`
}
type APISolutionErrorData struct {
	Stage    string `example:"test"`
	Msg      string `example:"Build fail message"`
	Params   string `example:"2;1;7'"`
	Expected string `example:"8"`
	Result   string `example:"1"`
}

type APISolutionTest struct {
	Error     WebError `example:"508"`
	ErrorData APISolutionErrorData
}

type APISolutionTestSuccess struct {
	Result []APISolutionVerboseResult
}

// @Tags solution
// @Summary Send solution for specific task
// @Description Receives solution in form of file or plain text.
// @Description Builds solution and then runs. While running it gives various input parameters (through stdin) and expects specific result (from stdout).
// @Description Apart from errors raised due to invalid POST parameters, there are 2 "normal" errors:
// @Description 508 - Solution build error. If this happens, then result will contain: `{"error":508,"error_data":{"msg":"multiline compilation error"}}`
// @Description 509 - Solution test error. If this happens, then result will contain: `{"error":509,"error_data":{"expected":"expected result", "params":"semicolon separated input parameters", "result":"actual result"}}`
// @Description If "verbose" flag is "true" then result will contain (if no error occurs): `{"result":[{"params":"semicolon separated input parameters", "result":"actual result"}]}`
// @ID post-solution
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   lang   formData    string  true    "Language of passing solution"
// @Param   task_id   formData    int  true    "ID of task to pass with given solution"
// @Param   source_text   formData    string  false    "Source text of passing solution"
// @Param   source_file   formData    file  false    "File with source text of passing solution"
// @Param   test_cases   formData    string  false    "User test cases for solution"
// @Param   verbose   formData    bool  false    "If specified - when solution is passed, all test results will be returned"
// @Success 200 {object} main.APISolutionTestSuccess "Success"
// @Failure 400 {object} main.APISolutionTest "Possible error codes: 300, 301, 302, 304, 4XX, 5XX, 6XX"
// @Failure 500 {object} main.APIInternalError "Server internal bug"
// @Router /solution [post]
func PostSolution(r *http.Request, resp *map[string]interface{}) WebError {
	solution, err := ParseSolution(r)
	if err != NoError {
		return err
	}
	web_err, err_data, verbose_result := BuildAndTest(solution.Task, solution)
	if web_err != NoError {
		(*resp)["error_data"] = *err_data
		(*resp)["fail_count"] = GetFailedSolutions(solution)
	} else if verbose_result != nil {
		(*resp)["result"] = *verbose_result
	}
	SaveSolution(solution, web_err == NoError)
	return web_err
}
