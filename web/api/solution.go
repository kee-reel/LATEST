package api

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"web/models"
	"web/utils"
)

type APISolution struct {
	Text *string `json:"text,omitempty" example:"a = int(input())\nb = int(input())\nprint(a+b)"`
}

// @Tags solution
// @Summary Get last solution of specific task
// @Description Returns last solution text for specified task. If no solution was posted, nothing will be returned.
// @ID get-solution
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   task_id   formData    int  true    "ID of task"
// @Success 200 {object} api.APISolution "Success"
// @Failure 400 {object} api.APIError "Possible error codes: 300, 301, 302, 304"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /solution [get]
func (c *Controller) GetSolution(r *http.Request) (interface{}, WebError) {
	token, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}
	task_id_str, web_err := getUrlParam(r, "task_id")
	if web_err != NoError {
		return nil, web_err
	}
	task_id, err := strconv.Atoi(task_id_str)
	if err != nil {
		return nil, TaskIdInvalid
	}
	resp := APISolution{
		c.storage.GetSolutionText(token.UserId, task_id),
	}
	return resp, NoError
}

var user_tests_re = regexp.MustCompile(`^((-?\d+;)+\n)+$`)

type APITestSuccessResult struct {
	ScoreDiff float32 `json:"score_diff, omitempty" example:"2.5"`
}
type APITestFailResult struct {
	Error     WebError                  `json:"error,omitempty" example:"508"`
	ErrorData *models.SolutionErrorData `json:"error_data,omitempty"`
}

type testResult struct {
	models.TestResult
	Error     WebError `json:"error,omitempty" example:"508"`
	ScoreDiff float32  `json:"score_diff,omitempty" example:"2.5"`
}

// @Tags solution
// @Summary Send solution for specific task
// @Description Receives solution in form of file or plain text.
// @Description Builds solution and then runs. While running it gives various input parameters (through stdin) and expects specific result (from stdout).
// @Description Apart from errors raised due to invalid POST parameters, there are 2 "normal" errors:
// @Description 504 - Solution build error. If this happens, then result will contain: `{"error":508,"error_data":{"msg":"multiline compilation error", "tests_passed":0, "tests_total":15}}`
// @Description 505 - Solution test error. If this happens, then result will contain: `{"error":509,"error_data":{"expected":"expected result", "params":"semicolon separated input parameters", "result":"actual result", "tests_passed":7, "tests_total":15}}`
// @Description 506 - Solution timeout error. If this happens, then result will contain: `{"error":509,"error_data":{"params":"semicolon separated input parameters", "result":"actual result", "time": 0.5, "tests_passed":0, "tests_total":15}}`
// @Description 507 - Solution runtime error. If this happens, then result will contain: `{"error":509,"error_data":{"params":"semicolon separated input parameters", "msg":"actual result", "tests_passed":2, "tests_total":15}}`
// @ID post-solution
// @Produce  json
// @Param   token   query    string  true    "Access token returned by GET /login"
// @Param   lang_id   formData    int  true    "Language id of passing solution returned by GET /languages"
// @Param   task_id   formData    int  true    "ID of task to pass with given solution"
// @Param   source_text   formData    string  false    "Source text of passing solution - must be less than 5000 symbols"
// @Param   source_file   formData    file  false    "File with source text of passing solution"
// @Success 200 {object} api.APITestSuccessResult "Success"
// @Failure 400 {object} api.APITestFailResult "Possible error codes: 300, 301, 302, 304, 4XX, 5XX, 6XX"
// @Failure 500 {object} api.APIInternalError "Server internal bug"
// @Router /solution [post]
func (c *Controller) PostSolution(r *http.Request) (interface{}, WebError) {
	solution, web_err := c.parseSolution(r)
	if web_err != NoError {
		return nil, web_err
	}

	var test_result_raw *models.TestResult
	solution.Id, test_result_raw = c.storage.CreateSolutionAttempt(solution)
	if test_result_raw == nil {
		test_result_raw = c.buildAndTest(solution.Task, solution)
		c.storage.SaveSolutionResult(solution.Id, test_result_raw)
	}
	if test_result_raw.InternalError != nil {
		panic(*test_result_raw.InternalError)
	}

	test_result := testResult{
		TestResult: *test_result_raw,
	}
	if test_result.ErrorData != nil {
		test_result.Error = SolutionTestFail
	}

	percent := float32(1)
	if test_result.ErrorData != nil {
		percent = float32(test_result.ErrorData.TestsPassed) /
			float32(test_result.ErrorData.TestsTotal)
	}
	test_result.ScoreDiff = c.storage.UpdateSolutionAttempt(solution, percent)
	return test_result, test_result.Error
}

func (c *Controller) parseSolution(r *http.Request) (*models.Solution, WebError) {
	err := r.ParseMultipartForm(32 << 20)
	token, web_err := c.getToken(r)
	if web_err != NoError {
		return nil, web_err
	}
	lang_id_str, web_err := getFormParam(r, "lang_id")
	if web_err != NoError {
		return nil, web_err
	}
	lang_id, err := strconv.Atoi(lang_id_str)
	if err != nil {
		return nil, LanguageNotSupported
	}
	if _, ok := c.supported_languages[lang_id]; !ok {
		return nil, LanguageNotSupported
	}
	task_id_str, web_err := getFormParam(r, "task_id")
	if web_err != NoError {
		return nil, web_err
	}
	task_id, err := strconv.Atoi(task_id_str)
	if err != nil {
		return nil, TaskIdInvalid
	}

	task_ids := []int{task_id}
	tasks := c.storage.GetTasks(token, task_ids)
	if len(*tasks) == 0 {
		return nil, TaskNotFound
	}

	var solution models.Solution
	task := (*tasks)[0]

	source_text := r.FormValue("source_text")
	if source_text == "" {
		file, _, err := r.FormFile("source_file")
		if err != nil {
			return nil, SolutionTextNotProvided
		}
		solution_text_bytes, err := ioutil.ReadAll(file)
		utils.Err(err)
		source_text = string(solution_text_bytes)
	}

	if len(source_text) > 10000 {
		return nil, SolutionTextTooLong
	}

	solution.Source = string(source_text)
	solution.Task = &task
	solution.Token = token
	solution.LanguageId = lang_id
	solution.UserId = *c.storage.GetUserIdByEmail(token.Email)

	return &solution, NoError
}

func (c *Controller) buildAndTest(task *models.Task, solution *models.Solution) *models.TestResult {
	complete_solution_source, fixed_tests := c.storage.GetTaskTestData(task.Id)

	runner_data := models.RunnerData{
		Id: solution.Id,
		UserSolution: models.SolutionData{
			Text:      solution.Source,
			Extention: c.supported_languages[solution.LanguageId],
		},
		CompleteSolution: models.SolutionData{
			Text:      complete_solution_source,
			Extention: c.supported_languages[task.LanguageId],
		},
		Tests: models.SolutionTests{},
	}
	if len(fixed_tests) > 0 {
		runner_data.Tests.Fixed = fixed_tests
	}
	rand_tests := generateTests(task)
	if len(rand_tests) > 0 {
		runner_data.Tests.Random = rand_tests
	}

	test_result_raw := c.workers.DoJob(&runner_data)
	if test_result_raw == nil {
		panic(fmt.Sprintf("Timeout while processing solution %d", runner_data.Id))
	}
	return test_result_raw
}

func generateTests(task *models.Task) string {
	result := ""
	if len(task.Input) == 0 {
		return result
	}
	random_tests_count := 10
	test_case_size := 1 // To add '\n' after every test case
	for _, input := range task.Input {
		test_case_size += input.TotalCount
		if input.TotalCount > 1 {
			for _, d := range input.Dimensions {
				if d != 0 {
					test_case_size++
				}
			}
		}
	}
	test_data := make([]string, random_tests_count*test_case_size)
	rand.Seed(time.Now().UnixNano())
	start_index := 0
	for i := 0; i < random_tests_count; i++ {
		for _, input := range task.Input {
			start_index = genTestParam(test_data, input, start_index)
		}
		if i+1 < random_tests_count {
			test_data[start_index] = "\n"
		}
		start_index++
	}
	result = strings.Join(test_data, "")
	return result
}

func genTestParam(test_data []string, param models.TaskParamData, start_index int) int {
	delimiter := ';'
	values_to_generate := 1
	cur_d := 0
	if param.TotalCount > 1 {
		for _, d := range param.Dimensions {
			if d != 0 {
				cur_d = 1 + rand.Intn(d)
				if cur_d == 1 {
					cur_d = 1 + rand.Intn(d)
				}
				test_data[start_index] = fmt.Sprintf("%d%c", cur_d, delimiter)
				start_index++
			}
			values_to_generate *= cur_d
		}
	}
	last_index := start_index + values_to_generate
	switch param.Type {
	case "float", "double":
		min := (*param.FloatRange)[0]
		value_range := math.Abs((*param.FloatRange)[1]-min) + 1
		type_spec := "%f%c"
		for i := start_index; i < last_index; i++ {
			value := min + rand.Float64()*value_range
			test_data[i] = fmt.Sprintf(type_spec, value, delimiter)
		}
	case "int":
		min := (*param.IntRange)[0]
		value_range := utils.Abs((*param.IntRange)[1]-min) + 1
		type_spec := "%d%c"
		for i := start_index; i < last_index; i++ {
			value := min + rand.Intn(value_range)
			test_data[i] = fmt.Sprintf(type_spec, value, delimiter)
		}
	default:
		utils.Err(fmt.Errorf("Unknown parameter type: %s", param.Type))
	}
	return last_index
}
