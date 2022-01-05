package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const build_script = "./scripts/build_solution.sh"
const test_script = "./scripts/test_solution.sh"
const tasks_path = "./tasks"
const user_solution_src_filename = "solution.c"
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

func GenerateTests(task *Task) string {
	if len(task.Input) == 0 {
		return ""
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
	return strings.Join(test_data, "")
}

func BuildSolution(task *Task, solution *Solution) error {
	solution.Path = fmt.Sprintf("%s/%s", task.Path, user_solution_src_filename)
	err := ioutil.WriteFile(solution.Path, []byte(solution.Source), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/solution-%d-%d-%d-%d-%d.c",
		user_solutions_path, solution.Token.UserId, solution.Token.Subject,
		solution.Token.Work, solution.Token.Variant, task.Number), []byte(solution.Source), 0777)
	if err != nil {
		log.Printf("Can't save solution: %s", err)
	}
	_, err = ExecCmd(fmt.Sprintf("%s %s", build_script, task.Path))
	return err
}

func RunTests(task *Task, user_test_data string, test_data string) (*string, bool, error) {
	is_user_tests_passed := false
	test_data_filename := fmt.Sprintf("./%s/%s", task.Path, user_tests_filename)
	err := ioutil.WriteFile(test_data_filename, []byte(user_test_data), 0777)
	if err != nil {
		return nil, is_user_tests_passed, err
	}
	test_data_filename = fmt.Sprintf("./%s/%s", task.Path, random_tests_filename)
	err = ioutil.WriteFile(test_data_filename, []byte(test_data), 0777)
	if err != nil {
		return nil, is_user_tests_passed, err
	}
	cmd_str := fmt.Sprintf("%s %s", test_script, task.Path)
	out, err := ExecCmd(cmd_str)
	log.Print(out)
	is_user_tests_passed = len(out) != 0 && out[0] == '+'
	if err != nil {
		return nil, is_user_tests_passed, fmt.Errorf("Тест для задачи %d провален.\n%s", task.Number, err.Error())
	}
	read, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", task.Path, test_result_filename))
	if err != nil {
		return nil, is_user_tests_passed, err
	}
	test_result := string(read)
	return &test_result, is_user_tests_passed, nil
}

func BuildAndTest(task *Task, solution *Solution) (*string, bool, error) {
	err := BuildSolution(task, solution)
	if err != nil {
		return nil, false, err
	}
	test_data := GenerateTests(task)
	test_result, is_user_tests_passed, err := RunTests(task, solution.TestCases, test_data)
	if err != nil {
		return nil, is_user_tests_passed, err
	}
	return test_result, is_user_tests_passed, nil
}
