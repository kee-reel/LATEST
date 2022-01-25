package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	_ "github.com/lib/pq"
)

const db_name = "tasks.db"

func OpenDB() *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Env("DB_HOST"), Env("DB_PORT"), Env("DB_USER"), Env("DB_PASS"), Env("DB_NAME"))
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}
	return db
}

func GetTasks(tasks_id []int, token *Token) (*[]Task, error) {
	db := OpenDB()
	defer db.Close()

	var tasks []Task
	works := map[int]*Work{}
	subjects := map[int]*Subject{}
	for _, task_id := range tasks_id {
		query, err := db.Prepare(`SELECT t.subject_id, t.work_id, t.position, t.extention, t.folder_name,
			t.name, t.description, t.input, t.output
			FROM tasks AS t
			WHERE t.id = $1`)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, err
		}

		var task Task
		var in_params_str []byte
		var subject_id int
		var work_id int
		err = query.QueryRow(task_id).Scan(
			&subject_id, &work_id, &task.Position, &task.Extention, &task.FolderName,
			&task.Name, &task.Desc, &in_params_str, &task.Output)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, fmt.Errorf("Task with id %d not found", task_id)
		}

		subject, ok := subjects[subject_id]
		if !ok {
			subject, err = GetSubject(subject_id)
			if err != nil {
				log.Printf("Not found subject for task %d", task_id)
				return nil, err
			}
			subjects[subject_id] = subject
		}
		task.Subject = subject

		work, ok := works[work_id]
		if !ok {
			work, err = GetWork(work_id)
			if err != nil {
				log.Printf("Not found work for task %d", task_id)
				return nil, err
			}
			works[work_id] = work
		}
		task.Work = work

		err = json.Unmarshal(in_params_str, &task.Input)
		if err != nil {
			log.Printf("JSON error: %s", err)
			return nil, err
		}

		query, err = db.Prepare(`SELECT COUNT(*) FROM solutions AS s 
			WHERE s.token_id = $1 AND s.task_id = $2 AND s.is_passed = TRUE`)
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
		tasks = append(tasks, task)
	}

	return &tasks, nil
}

func GetTaskTestData(task_id int) (*string, *string, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT t.source_code, t.fixed_tests FROM tasks AS t WHERE t.id = $1`)
	if err != nil {
		return nil, nil, err
	}

	var source_code string
	var fixed_tests string
	err = query.QueryRow(task_id).Scan(&source_code, &fixed_tests)
	if err != nil {
		log.Printf("DB error: %s", err)
		return nil, nil, fmt.Errorf("Task with id %d not found", task_id)
	}
	return &source_code, &fixed_tests, nil
}

func GetTasksByToken(token *Token) (*[]int, error) {
	db := OpenDB()
	defer db.Close()
	rows, err := db.Query(`SELECT t.id FROM tasks AS t WHERE t.subject_id = $1`, token.Subject)
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
	query, err := db.Prepare(`SELECT s.name, s.folder_name FROM subjects AS s WHERE s.id = $1`)
	if err != nil {
		return nil, err
	}

	var subject Subject
	err = query.QueryRow(subject_id).Scan(&subject.Name, &subject.FolderName)
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
	query, err := db.Prepare(`SELECT w.name, w.next_work_id, w.folder_name FROM works AS w WHERE w.id = $1`)
	if err != nil {
		return nil, err
	}

	var work Work
	err = query.QueryRow(work_id).Scan(&work.Name, &work.NextId, &work.FolderName)
	if err != nil {
		log.Printf("DB error: %s", err)
		return nil, fmt.Errorf("Work with id %d not found", work_id)
	}
	work.Id = work_id
	return &work, nil
}

func SaveSolution(solution *Solution, is_passed bool) error {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`INSERT INTO solutions(token_id, task_id, is_passed) VALUES($1, $2, $3)`)
	if err != nil {
		log.Print(err)
	}
	_, err = query.Exec(solution.Token.Id, solution.Task.Id, is_passed)
	if err != nil {
		log.Print(err)
	}

	query, err = db.Prepare(`INSERT INTO 
		solutions_sources(token_id, task_id, source_code) VALUES($1, $2, $3)
		ON CONFLICT (token_id, task_id) DO UPDATE SET source_code = $3`)
	if err != nil {
		log.Print(err)
	}
	_, err = query.Exec(solution.Token.Id, solution.Task.Id, solution.Source)
	if err != nil {
		log.Print(err)
	}
	return nil
}

func GetFailedSolutions(solution *Solution) (int, error) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT COUNT(*) FROM solutions as s
		WHERE s.token_id = $1 AND s.task_id = $2 AND s.is_passed = FALSE`)
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

	query, err := db.Prepare(`SELECT t.id, t.user_id, t.subject_id FROM tokens as t WHERE t.token = $1`)
	if err != nil {
		log.Printf("db error on prepare: %s", err)
		return nil, errors.New("Неизвестный токен доступа")
	}
	var token Token
	err = query.QueryRow(token_str).Scan(&token.Id, &token.UserId, &token.Subject)
	if err != nil {
		log.Printf("db error on scan: %s", err)
		return nil, errors.New("Неизвестный токен доступа")
	}
	return &token, nil
}
