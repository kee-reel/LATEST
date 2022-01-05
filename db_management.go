package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const db_name = "tasks.db"
const create_subject_table = `CREATE TABLE IF NOT EXISTS subject (
	id INTEGER PRIMARY KEY,
	name VARCHAR(64))`
const create_work_table = `CREATE TABLE IF NOT EXISTS work (
	id INTEGER,
	subject INTEGER,
	name VARCHAR(64),
	PRIMARY KEY(id, subject))`
const create_variant_table = `CREATE TABLE IF NOT EXISTS variant (
	id INTEGER,
	subject INTEGER,
	work INTEGER,
	name VARCHAR(64),
	PRIMARY KEY(id, subject, work))`
const create_task_table = `CREATE TABLE IF NOT EXISTS task (
	id INTEGER PRIMARY KEY,
	subject INTEGER,
	work INTEGER,
	variant INTEGER,
	number INTEGER,
	name VARCHAR(64),
	desc VARCHAR(1024),
	input VARCHAR(512),
	output VARCHAR(128),
	UNIQUE(subject, work, variant, number))`
const create_solution_table = `CREATE TABLE IF NOT EXISTS solution(
	token VARCHAR(256),
	task_id INTEGER,
	is_user_tests_passed BOOLEAN,
	is_passed BOOLEAN,
	dt TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`
const create_access_token_table = `CREATE TABLE IF NOT EXISTS access_token (
	token VARCHAR(256) PRIMARY KEY,
	user_id INTEGER,
	subject INTEGER,
	work INTEGER,
	variant INTEGER)`
const create_user_table = `CREATE TABLE IF NOT EXISTS user (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	group_name VARCHAR(128),
	number INTEGER,
	name VARCHAR(128),
	last_name VARCHAR(128),
	UNIQUE(group_name, number))`

func InitDB() {
	create_table_queries := []string{
		create_subject_table,
		create_work_table,
		create_variant_table,
		create_task_table,
		create_solution_table,
		create_access_token_table,
		create_user_table,
	}
	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		panic(err)
	}
	tx, err := db.Begin()
	defer tx.Rollback()
	for _, query := range create_table_queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			panic(fmt.Sprintf("Error in query '%s': %s", query, err))
		}
		defer stmt.Close()
		_, err = stmt.Exec()
		if err != nil {
			panic(fmt.Sprintf("Error in query '%s': %s", query, err))
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func GetTask(task_id int) (*Task, error) {
	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query, err := db.Prepare(`SELECT t.subject, t.work, t.variant, t.number, t.name, t.desc, t.input, t.output FROM task AS t WHERE t.id = ?`)
	if err != nil {
		return nil, err
	}

	var task Task
	var in_params_str []byte
	err = query.QueryRow(task_id).Scan(&task.Subject, &task.Work, &task.Variant, &task.Number, &task.Name, &task.Desc, &in_params_str, &task.Output)
	if err != nil {
		return nil, fmt.Errorf("Task with id %d not found", task_id)
	}
	err = json.Unmarshal(in_params_str, &task.Input)
	if err != nil {
		return nil, err
	}

	for i := range task.Input {
		input := &task.Input[i]
		param_type := input.Type
		param_range := input.Range
		if input.Dimensions == nil {
			input.TotalCount = 1
			dimensions := []int{1}
			input.Dimensions = &dimensions
		} else {
			last_d := 0
			input.TotalCount = 1
			for _, d := range *input.Dimensions {
				if d != 0 {
					last_d = d
				}
				input.TotalCount *= last_d
			}
		}
		switch param_type {
		case "float", "double":
			float_range := make([]float64, 2)
			float_range[0], _ = strconv.ParseFloat(param_range[0], 64)
			float_range[1], _ = strconv.ParseFloat(param_range[1], 64)
			input.FloatRange = &float_range
		case "int":
			int_range := make([]int, 2)
			int_range[0], _ = strconv.Atoi(param_range[0])
			int_range[1], _ = strconv.Atoi(param_range[1])
			input.IntRange = &int_range
		default:
			panic(fmt.Sprint("Param type \"%s\" mot supported", param_type))
		}
	}

	task.Path = fmt.Sprintf("%s/subject-%d/work-%d/variant-%d/task-%d", tasks_path, task.Subject, task.Work, task.Variant, task.Number)

	return &task, nil
}

func GetWorkTasks(token *Token) (*[]int, error) {
	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT t.id FROM task AS t 
		WHERE t.subject = ? AND t.work = ? AND t.variant = ?`, token.Subject, token.Work, token.Variant)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	tasks := []int{}
	for rows.Next() {
		var task_id int
		err := rows.Scan(&task_id)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task_id)
	}
	return &tasks, nil
}

func SaveSolution(solution *Solution, is_user_tests_passed bool, is_passed bool) error {
	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		log.Print(err)
	}
	defer db.Close()

	query, err := db.Prepare(`INSERT INTO solution(token, task_id, is_user_tests_passed, is_passed) VALUES(?, ?, ?, ?)`)
	if err != nil {
		log.Print(err)
	}

	_, err = query.Exec(solution.Token.Token, solution.Task, is_user_tests_passed, is_passed)
	if err != nil {
		log.Print(err)
	}
	return nil
}

func GetFailedSolutions(solution *Solution) (int, error) {
	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	query, err := db.Prepare(`SELECT COUNT(*) FROM solution
		WHERE token = ? AND task_id = ? AND is_user_tests_passed = 1 AND is_passed = 0`)
	if err != nil {
		return -1, err
	}

	var count int
	err = query.QueryRow(solution.Token.Token, solution.Task).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func GetTokenData(token_str string) (*Token, error) {
	if len(token_str) == 0 {
		return nil, errors.New("Токен доступа не указан")
	}
	if len(token_str) != 256 {
		return nil, errors.New("Неизвестный токен доступа")
	}
	db, err := sql.Open("sqlite3", db_name)
	if err != nil {
		return nil, errors.New("Неизвестный токен доступа")
	}
	defer db.Close()

	query, err := db.Prepare(`SELECT a.user_id, a.subject, a.work, a.variant FROM access_token as a
		WHERE a.token = ?`)
	if err != nil {
		return nil, errors.New("Неизвестный токен доступа")
	}
	var token Token
	err = query.QueryRow(token_str).Scan(&token.UserId, &token.Subject, &token.Work, &token.Variant)
	if err != nil {
		return nil, errors.New("Неизвестный токен доступа")
	}
	token.Token = token_str
	return &token, nil
}
