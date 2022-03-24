package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const db_name = "tasks.db"
const token_len = 256
const token_chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const token_chars_len = int64(len(token_chars))

func OpenDB() *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Env("DB_HOST"), Env("DB_PORT"), Env("DB_USER"), Env("DB_PASS"), Env("DB_NAME"))
	db, err := sql.Open("postgres", psqlconn)
	Err(err)
	return db
}

func GetTasks(tasks_id []int) *[]Task {
	db := OpenDB()
	defer db.Close()

	tasks := make([]Task, len(tasks_id))
	units := map[int]*Unit{}
	projects := map[int]*Project{}
	for task_index, task_id := range tasks_id {
		query, err := db.Prepare(`SELECT t.project_id, t.unit_id, t.position, t.extention, t.folder_name,
			t.name, t.description, t.input, t.output
			FROM tasks AS t
			WHERE t.id = $1`)
		Err(err)

		var task Task
		var in_params_str []byte
		var project_id int
		var unit_id int
		err = query.QueryRow(task_id).Scan(
			&project_id, &unit_id, &task.Position, &task.Extention, &task.FolderName,
			&task.Name, &task.Desc, &in_params_str, &task.Output)
		ErrMsg(err, "Task not found")

		project, ok := projects[project_id]
		if !ok {
			project, err = GetProject(project_id)
			Err(err)
			projects[project_id] = project
		}
		task.Project = project

		unit, ok := units[unit_id]
		if !ok {
			unit, err = GetUnit(unit_id)
			Err(err)
			units[unit_id] = unit
		}
		task.Unit = unit

		err = json.Unmarshal(in_params_str, &task.Input)
		Err(err)

		query, err = db.Prepare(`SELECT COUNT(*) FROM solutions AS s
			WHERE s.task_id = $1 AND s.is_passed = TRUE`)
		Err(err)

		var passed_count int
		err = query.QueryRow(task_id).Scan(&passed_count)
		Err(err)
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
				Err(fmt.Errorf("Param type \"%s\" mot supported", param_type))
			}
		}

		task.Id = task_id
		tasks[task_index] = task
	}

	return &tasks
}

func GetTaskTestData(task_id int) (*string, *string, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT t.source_code, t.fixed_tests FROM tasks AS t WHERE t.id = $1`)
	Err(err)

	var source_code string
	var fixed_tests string
	err = query.QueryRow(task_id).Scan(&source_code, &fixed_tests)
	Err(err)

	return &source_code, &fixed_tests, nil
}

func GetTaskTemplate(task_id int) (*string, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT t.template_source_code FROM tasks AS t WHERE t.id = $1`)
	Err(err)

	var source_code string
	err = query.QueryRow(task_id).Scan(&source_code)
	if err != nil {
		return nil, fmt.Errorf("Template for this task is not found")
	}

	return &source_code, nil
}

func GetTaskIds(task_str_ids *[]string) (*[]int, error) {
	db := OpenDB()
	defer db.Close()
	var sb strings.Builder
	sb.WriteString("SELECT t.id FROM tasks AS t")
	if len(*task_str_ids) > 0 {
		sb.WriteString(" WHERE ")
		last_i := len(*task_str_ids) - 1
		for i, task_str_id := range *task_str_ids {
			_, err := strconv.Atoi(task_str_id)
			if err != nil {
				return nil, fmt.Errorf("Task id must be a number")
			}
			sb.WriteString("t.id=")
			sb.WriteString(task_str_id)
			if i != last_i {
				sb.WriteString(" OR ")
			}
		}
	}

	rows, err := db.Query(sb.String())
	Err(err)

	defer rows.Close()
	task_ids := []int{}
	for rows.Next() {
		var task_id int
		err := rows.Scan(&task_id)
		Err(err)
		task_ids = append(task_ids, task_id)
	}
	return &task_ids, nil
}

func GetProject(project_id int) (*Project, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT s.name, s.folder_name FROM projects AS s WHERE s.id = $1`)
	Err(err)

	var project Project
	err = query.QueryRow(project_id).Scan(&project.Name, &project.FolderName)
	Err(err)
	project.Id = project_id
	return &project, nil
}

func GetUnit(unit_id int) (*Unit, error) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT u.name, u.project_id, u.folder_name FROM units AS u WHERE u.id = $1`)
	Err(err)

	var unit Unit
	err = query.QueryRow(unit_id).Scan(&unit.Name, &unit.ProjectId, &unit.FolderName)
	Err(err)
	unit.Id = unit_id
	return &unit, nil
}

func SaveSolution(solution *Solution, is_passed bool) error {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`INSERT INTO solutions(token_id, task_id, is_passed) VALUES($1, $2, $3)`)
	Err(err)
	_, err = query.Exec(solution.Token.Id, solution.Task.Id, is_passed)
	Err(err)

	query, err = db.Prepare(`INSERT INTO 
		solutions_sources(token_id, task_id, source_code) VALUES($1, $2, $3)
		ON CONFLICT (token_id, task_id) DO UPDATE SET source_code = $3`)
	Err(err)
	_, err = query.Exec(solution.Token.Id, solution.Task.Id, solution.Source)
	Err(err)
	return nil
}

func GetFailedSolutions(solution *Solution) int {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT COUNT(*) FROM solutions as s
		WHERE s.task_id = $1 AND s.is_passed = FALSE`)
	Err(err)

	var count int
	err = query.QueryRow(solution.Task.Id).Scan(&count)
	Err(err)
	return count
}

func GetTokenForConnection(email string, pass string, ip string) (*Token, error) {
	db := OpenDB()
	defer db.Close()
	var token Token
	token.IP = ip

	query, err := db.Prepare(`SELECT u.id, u.pass FROM users as u WHERE u.email = $1`)
	Err(err)

	var hash string
	err = query.QueryRow(email).Scan(&token.UserId, &hash)
	if err == nil {
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
		if err != nil {
			return nil, fmt.Errorf("Wrong password")
		}
	} else {
		query, err = db.Prepare("INSERT INTO users(email, pass) VALUES($1, $2) RETURNING id")
		Err(err)
		hash_raw, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		Err(err)
		err = query.QueryRow(email, string(hash_raw)).Scan(&token.UserId)
		Err(err)
	}

	query, err = db.Prepare(`SELECT id, token, is_verified FROM tokens WHERE user_id = $1 AND ip = $2`)
	Err(err)

	err = query.QueryRow(token.UserId, token.IP).Scan(&token.Id, &token.Token, &token.IsVerified)
	if err == nil {
		return &token, nil
	}

	// If token not found, then generate new one
	query, err = db.Prepare(`INSERT INTO tokens(token, user_id, ip) VALUES($1, $2, $3) RETURNING id`)
	Err(err)
	token.Token = string(GenerateToken())
	err = query.QueryRow(token.Token, token.UserId, token.IP).Scan(&token.Id)
	Err(err)

	return &token, nil
}

func GetTokenData(token_str string, ip string, verified_only bool) (*Token, error) {
	if len(token_str) == 0 {
		log.Print("No token received")
		return nil, fmt.Errorf("Token not specified")
	}
	if len(token_str) != token_len {
		log.Print("Received malformed token")
		return nil, fmt.Errorf("Unknown token")
	}
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT t.id, t.user_id, t.ip, t.is_verified 
		FROM tokens as t WHERE t.token = $1`)
	if err != nil {
		log.Printf("db error on prepare: %s", err)
		return nil, fmt.Errorf("Unknown token")
	}
	var token Token
	token.Token = token_str
	err = query.QueryRow(token.Token).Scan(&token.Id, &token.UserId, &token.IP, &token.IsVerified)
	if err != nil {
		log.Printf("db error on scan: %s", err)
		return nil, fmt.Errorf("Unknown token")
	}
	if verified_only && !token.IsVerified {
		log.Printf("passed non verified token")
		return nil, fmt.Errorf("Email for this IP is not verified, please open link that was sent to your email")
	}
	if ip != token.IP {
		log.Print("IP for token does not match")
		return nil, fmt.Errorf("Given token is bound to other IP")
	}
	return &token, nil
}

func VerifyToken(token *Token) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`UPDATE tokens SET is_verified = TRUE WHERE id = $1`)
	Err(err)
	_, err = query.Exec(token.Id)
	Err(err)
}
