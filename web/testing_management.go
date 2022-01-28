package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GenTestParam(test_data []string, param TaskParamData, start_index int) int {
	delimiter := ';'
	values_to_generate := 1
	cur_d := 0
	if param.TotalCount > 1 {
		for _, d := range *param.Dimensions {
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
		value_range := Abs((*param.IntRange)[1]-min) + 1
		type_spec := "%d%c"
		for i := start_index; i < last_index; i++ {
			value := min + rand.Intn(value_range)
			test_data[i] = fmt.Sprintf(type_spec, value, delimiter)
		}
	default:
		panic(fmt.Sprintf("Unknown parameter type: %s", param.Type))
	}
	return last_index
}

func GenerateTests(task *Task) *string {
	result := ""
	if len(task.Input) == 0 {
		return &result
	}
	random_tests_count := 30
	test_case_size := 1 // To add '\n' after every test case
	for _, input := range task.Input {
		test_case_size += input.TotalCount
		if input.TotalCount > 1 {
			for _, d := range *input.Dimensions {
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
			start_index = GenTestParam(test_data, input, start_index)
		}
		test_data[start_index] = "\n"
		start_index++
	}
	result = strings.Join(test_data, "")
	return &result
}

func BuildAndTest(task *Task, solution *Solution) (*map[string]interface{}, error) {
	complete_solution_source, fixed_tests, err := GetTaskTestData(task.Id)
	if err != nil {
		panic(err)
	}
	random_tests := GenerateTests(task)

	runner_url := fmt.Sprintf("http://%s:%s", Env("RUNNER_HOST"), Env("RUNNER_PORT"))
	verbose_text := "false"
	if EnvB("RUNNER_VERBOSE") && solution.IsVerbose {
		verbose_text = "true"
	}
	response, err := http.PostForm(runner_url, url.Values{
		"solution":          {solution.Source},
		"complete_solution": {*complete_solution_source},
		"user_tests":        {solution.TestCases},
		"fixed_tests":       {*fixed_tests},
		"random_tests":      {*random_tests},
		"extention":         {task.Extention},
		"verbose":           {verbose_text},
	})
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var test_result map[string]interface{}
	err = json.Unmarshal([]byte(body), &test_result)
	if err != nil {
		return nil, fmt.Errorf("Internal error while running task %d.\n%s", task.Id, err.Error())
	}

	return &test_result, nil
}
