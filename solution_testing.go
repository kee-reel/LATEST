package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

const build_script = "python3 scripts/build_solution.py"
const test_script = "python3 scripts/test_solution.py"
const tasks_path = "./tasks"
const complete_solution_src_filename = "complete_solution"
const user_solution_src_filename = "solution"
const user_solutions_path = "./solutions"
const user_tests_filename = "user_tests.txt"
const random_tests_filename = "random_tests.txt"
const test_result_filename = "test_result.txt"

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func ExecCmd(command_str string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command_str)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out_str := stdout.String()
	err_str := stderr.String()
	if config.Verbose {
		log.Printf("[EXEC]: %s; [OUT]: %s", command_str, out_str)
	} else {
		log.Printf("[EXEC]: %s", command_str)
	}
	if err != nil || err_str != "" {
		log.Printf("[ERR]: %s", err)
		return out_str, errors.New(err_str)
	}
	return out_str, nil
}

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

func BuildSolution(solution *Solution) error {
	solution.Path = fmt.Sprintf("%s/%s.%s", solution.Task.Path, user_solution_src_filename, solution.Task.Extention)
	err := ioutil.WriteFile(solution.Path, []byte(solution.Source), 0777)
	if err != nil {
		return err
	}
	_, err = os.Stat(user_solutions_path)
	if os.IsNotExist(err) {
		os.Mkdir(user_solutions_path, os.FileMode(0777))
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/solution-%d-%d-%d-%d-%d.%s",
		user_solutions_path, solution.Token.UserId, solution.Task.Subject,
		solution.Task.Work, solution.Task.Variant, solution.Task.Number,
		solution.Task.Extention), []byte(solution.Source), 0777)
	if err != nil {
		log.Printf("Can't save solution: %s", err)
	}
	exec_ext, err := ExecCmd(fmt.Sprintf("%s %s %s.%s %s.%s", build_script, solution.Task.Path,
		user_solution_src_filename, solution.Task.Extention,
		complete_solution_src_filename, solution.Task.Extention))
	if err == nil {
		solution.CompleteExecFilename = fmt.Sprintf("%s.%s", complete_solution_src_filename, exec_ext)
		solution.ExecFilename = fmt.Sprintf("%s.%s", user_solution_src_filename, exec_ext)
	}
	return err
}

func RunTests(solution *Solution, test_data *string) (*map[string]interface{}, bool, error) {
	task := solution.Task
	is_user_tests_passed := false
	user_tests_path := fmt.Sprintf("./%s/%s", task.Path, user_tests_filename)
	err := ioutil.WriteFile(user_tests_path, []byte(solution.TestCases), 0777)
	if err != nil {
		return nil, is_user_tests_passed, err
	}
	random_tests_path := fmt.Sprintf("./%s/%s", task.Path, random_tests_filename)
	err = ioutil.WriteFile(random_tests_path, []byte(*test_data), 0777)
	if err != nil {
		_ = os.Remove(user_tests_path)
		return nil, is_user_tests_passed, err
	}
	cmd_str := fmt.Sprintf("%s %s %s %s", test_script, task.Path, solution.CompleteExecFilename, solution.ExecFilename)
	test_result_str, err := ExecCmd(cmd_str)
	// Remove all user files
	_ = os.Remove(solution.Path)
	_ = os.Remove(fmt.Sprintf("%s/%s", task.Path, solution.CompleteExecFilename))
	_ = os.Remove(fmt.Sprintf("%s/%s", task.Path, solution.ExecFilename))
	_ = os.Remove(solution.Path)
	_ = os.Remove(user_tests_path)
	_ = os.Remove(random_tests_path)
	_ = os.Remove(random_tests_path)
	_ = os.Remove(random_tests_path)
	if err != nil {
		return nil, is_user_tests_passed, fmt.Errorf("Internal error while running task %d.\n%s", task.Number, err.Error())
	}
	var test_result map[string]interface{}
	err = json.Unmarshal([]byte(test_result_str), &test_result)
	if err != nil {
		return nil, is_user_tests_passed, fmt.Errorf("Internal error while running task %d.\n%s", task.Number, err.Error())
	}
	return &test_result, is_user_tests_passed, nil
}

func BuildAndTest(task *Task, solution *Solution) (*map[string]interface{}, bool, error) {
	err := BuildSolution(solution)
	if err != nil {
		return nil, false, err
	}
	test_data := GenerateTests(task)
	test_result, is_user_tests_passed, err := RunTests(solution, test_data)
	if err != nil {
		return nil, is_user_tests_passed, err
	}
	return test_result, is_user_tests_passed, nil
}
