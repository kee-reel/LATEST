package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

const db_name = "tasks.db"
const db_creation_script = "scripts/create_db.sh"

func InitDB() {
	_, err := ExecCmd(db_creation_script)
	if err != nil {
		panic(err)
	}
}

func OpenDB() *sql.DB {
	db, err := sql.Open("sqlite", db_name)
	if err != nil {
		panic(err)
	}
	return db
}

func GetTasks(tasks_id []int, token *Token) (*[]Task, error) {
	db := OpenDB()
	defer db.Close()

	var tasks []Task
	for _, task_id := range tasks_id {
		query, err := db.Prepare(`SELECT t.subject, t.work, t.variant, t.number, t.extention, t.name, t.desc, t.input, t.output
			FROM task AS t
			WHERE t.id = ?`)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, err
		}

		var task Task
		var in_params_str []byte
		err = query.QueryRow(task_id).Scan(
			&task.Subject, &task.Work, &task.Variant, &task.Number, &task.Extention,
			&task.Name, &task.Desc, &in_params_str, &task.Output)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, fmt.Errorf("Task with id %d not found", task_id)
		}

		err = json.Unmarshal(in_params_str, &task.Input)
		if err != nil {
			log.Printf("JSON error: %s", err)
			return nil, err
		}

		query, err = db.Prepare(`SELECT COUNT(*) FROM solution AS s 
			WHERE s.token_id = ? AND s.task_id = ? AND s.is_passed = 1`)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, err
		}

		var passed_count int
		err = query.QueryRow(token.Id, task_id).Scan(&passed_count)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, fmt.Errorf("Task with id %d not found", task_id)
		}
		task.IsPassed = passed_count > 0

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

		task.Id = task_id
		task.Path = fmt.Sprintf("%s/subject-%d/work-%d/variant-%d/task-%d",
			tasks_path, task.Subject, task.Work, task.Variant, task.Number)
		tasks = append(tasks, task)
	}

	return &tasks, nil
}

func GetTasksByToken(token *Token) (*[]int, error) {
	db := OpenDB()
	defer db.Close()
	rows, err := db.Query(`SELECT t.id FROM task AS t WHERE t.subject = ? AND t.variant = ?`,
		token.Subject, token.Variant)
	if err != nil {
		log.Printf("DB error: %s", err)
		return nil, err
	}

	defer rows.Close()
	tasks := []int{}
	for rows.Next() {
		var task_id int
		err := rows.Scan(&task_id)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, err
		}
		tasks = append(tasks, task_id)
	}
	return &tasks, nil
}

func GetSubject(subject_id int) (*Subject, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT s.name FROM subject AS s WHERE s.id = ?`)
	if err != nil {
		return nil, err
	}

	var subject Subject
	err = query.QueryRow(subject_id).Scan(&subject.Name)
	if err != nil {
		log.Printf("DB error: %s", err)
		return nil, fmt.Errorf("Subject with id %d not found", subject_id)
	}
	subject.Id = subject_id
	return &subject, nil
}

func GetWork(work_id int) (*Work, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT w.name, w.next_work_id FROM work AS w WHERE w.id = ?`)
	if err != nil {
		return nil, err
	}

	var work Work
	err = query.QueryRow(work_id).Scan(&work.Name, &work.NextId)
	if err != nil {
		log.Printf("DB error: %s", err)
		return nil, fmt.Errorf("Work with id %d not found", work_id)
	}
	work.Id = work_id
	return &work, nil
}

func SaveSolution(solution *Solution, is_user_tests_passed bool, is_passed bool) error {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`INSERT INTO solution(token_id, task_id, is_user_tests_passed, is_passed) VALUES(?, ?, ?, ?)`)
	if err != nil {
		log.Print(err)
	}

	_, err = query.Exec(solution.Token.Id, solution.Task.Id, is_user_tests_passed, is_passed)
	if err != nil {
		log.Print(err)
	}
	return nil
}

func GetFailedSolutions(solution *Solution) (int, error) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT COUNT(*) FROM solution as s
		WHERE s.token_id = ? AND s.task_id = ? AND s.is_user_tests_passed = 1 AND s.is_passed = 0`)
	if err != nil {
		return -1, err
	}

	var count int
	err = query.QueryRow(solution.Token.Id, solution.Task.Id).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func GetTokenData(token_str string) (*Token, error) {
	if len(token_str) == 0 {
		log.Print("Received empty token")
		return nil, errors.New("Токен доступа не указан")
	}
	if len(token_str) != 256 {
		log.Print("Received malformed token")
		return nil, errors.New("Неизвестный токен доступа")
	}
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT a.id, a.user_id, a.subject, a.variant FROM access_token as a
		WHERE a.token = ?`)
	if err != nil {
		log.Printf("Got db error: %s", err)
		return nil, errors.New("Неизвестный токен доступа")
	}
	var token Token
	err = query.QueryRow(token_str).Scan(&token.Id, &token.UserId, &token.Subject, &token.Variant)
	if err != nil {
		log.Printf("Got db error: %s", err)
		return nil, errors.New("Неизвестный токен доступа")
	}
	return &token, nil
}
