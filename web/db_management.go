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
	units := map[int]*Unit{}
	projects := map[int]*Project{}
	for _, task_id := range tasks_id {
		query, err := db.Prepare(`SELECT t.project_id, t.unit_id, t.position, t.extention, t.folder_name,
			t.name, t.description, t.input, t.output
			FROM tasks AS t
			WHERE t.id = $1`)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, err
		}

		var task Task
		var in_params_str []byte
		var project_id int
		var unit_id int
		err = query.QueryRow(task_id).Scan(
			&project_id, &unit_id, &task.Position, &task.Extention, &task.FolderName,
			&task.Name, &task.Desc, &in_params_str, &task.Output)
		if err != nil {
			log.Printf("DB error: %s", err)
			return nil, fmt.Errorf("Task with id %d not found", task_id)
		}

		project, ok := projects[project_id]
		if !ok {
			project, err = GetProject(project_id)
			if err != nil {
				log.Printf("Not found project for task %d", task_id)
				return nil, err
			}
			projects[project_id] = project
		}
		task.Project = project

		unit, ok := units[unit_id]
		if !ok {
			unit, err = GetUnit(unit_id)
			if err != nil {
				log.Printf("Not found unit for task %d", task_id)
				return nil, err
			}
			units[unit_id] = unit
		}
		task.Unit = unit

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
	rows, err := db.Query(`SELECT t.id FROM tasks AS t WHERE t.project_id = $1`, token.Project)
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

func GetProject(project_id int) (*Project, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT s.name, s.folder_name FROM projects AS s WHERE s.id = $1`)
	if err != nil {
		return nil, err
	}

	var project Project
	err = query.QueryRow(project_id).Scan(&project.Name, &project.FolderName)
	if err != nil {
		log.Printf("DB error: %s", err)
		return nil, fmt.Errorf("Project with id %d not found", project_id)
	}
	project.Id = project_id
	return &project, nil
}

func GetUnit(unit_id int) (*Unit, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT w.name, w.next_unit_id, w.folder_name FROM units AS w WHERE w.id = $1`)
	if err != nil {
		return nil, err
	}

	var unit Unit
	err = query.QueryRow(unit_id).Scan(&unit.Name, &unit.NextId, &unit.FolderName)
	if err != nil {
		log.Printf("DB error: %s", err)
		return nil, fmt.Errorf("Unit with id %d not found", unit_id)
	}
	unit.Id = unit_id
	return &unit, nil
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

	query, err := db.Prepare(`SELECT t.id, t.user_id, t.project_id FROM tokens as t WHERE t.token = $1`)
	if err != nil {
		log.Printf("db error on prepare: %s", err)
		return nil, errors.New("Неизвестный токен доступа")
	}
	var token Token
	err = query.QueryRow(token_str).Scan(&token.Id, &token.UserId, &token.Project)
	if err != nil {
		log.Printf("db error on scan: %s", err)
		return nil, errors.New("Неизвестный токен доступа")
	}
	return &token, nil
}
