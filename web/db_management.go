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

func GetTasks(token *Token, task_ids []int) *[]Task {
	db := OpenDB()
	defer db.Close()

	tasks := make([]Task, len(task_ids))
	units := map[int]*Unit{}
	projects := map[int]*Project{}
	for task_index, task_id := range task_ids {
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
			project = GetProject(project_id)
			projects[project_id] = project
		}
		task.Project = project

		unit, ok := units[unit_id]
		if !ok {
			unit = GetUnit(unit_id)
			units[unit_id] = unit
		}
		task.Unit = unit

		err = json.Unmarshal(in_params_str, &task.Input)
		Err(err)

		query, err = db.Prepare(`SELECT s.is_passed FROM solutions AS s
			WHERE s.token_id = $1 AND s.task_id = $2 AND s.is_passed = TRUE LIMIT 1`)
		Err(err)

		var passed_count int
		err = query.QueryRow(token.Id, task_id).Scan(&passed_count)
		task.IsPassed = err == nil

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
				Err(fmt.Errorf("Param type \"%s\" not supported", param_type))
			}
		}

		task.Id = task_id
		tasks[task_index] = task
	}

	return &tasks
}

func GetTaskTestData(task_id int) (*string, *string) {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT t.source_code, t.fixed_tests FROM tasks AS t WHERE t.id = $1`)
	Err(err)

	var source_code string
	var fixed_tests string
	err = query.QueryRow(task_id).Scan(&source_code, &fixed_tests)
	Err(err)

	return &source_code, &fixed_tests
}

func GetTaskTemplate(lang string) *string {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT t.source_code FROM solution_templates AS t WHERE t.extention = $1`)
	Err(err)

	var source_code string
	err = query.QueryRow(lang).Scan(&source_code)
	Err(err)

	return &source_code
}

func GetTaskIdsByFolder(folder_names *[]string) (*[]int, WebError) {
	db := OpenDB()
	defer db.Close()
	var sb strings.Builder
	sb.WriteString("SELECT t.id FROM tasks AS t")

	var err error
	var project_id int
	var unit_id int
	folders_count := len(*folder_names)
	if folders_count >= 1 {
		query, err := db.Prepare("SELECT p.id FROM projects AS p WHERE p.folder_name = $1")
		Err(err)
		err = query.QueryRow((*folder_names)[0]).Scan(&project_id)
		if err != nil {
			return nil, SolutionProjectFolderNotFound
		}
		if folders_count != 3 {
			sb.WriteString(" WHERE t.project_id = ")
			sb.WriteString(strconv.Itoa(project_id))
		}
	}
	if folders_count >= 2 {
		query, err := db.Prepare("SELECT u.id FROM units AS u WHERE u.project_id = $1 AND u.folder_name = $2")
		Err(err)
		err = query.QueryRow(project_id, (*folder_names)[1]).Scan(&unit_id)
		if err != nil {
			return nil, SolutionUnitFolderNotFound
		}
		if folders_count != 3 {
			sb.WriteString(" AND t.unit_id = ")
			sb.WriteString(strconv.Itoa(unit_id))
		}
	}
	if folders_count == 3 {
		query, err := db.Prepare("SELECT t.id FROM tasks AS t WHERE t.project_id = $1 AND t.unit_id = $2 AND t.folder_name = $3")
		Err(err)
		var task_id int
		err = query.QueryRow(project_id, unit_id, (*folder_names)[2]).Scan(&task_id)
		if err != nil {
			return nil, SolutionTaskFolderNotFound
		}
		task_ids := []int{task_id}
		return &task_ids, NoError
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
	return &task_ids, NoError
}

func GetTaskIdsById(task_str_ids *[]string) (*[]int, WebError) {
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
				return nil, TaskIdInvalid
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
	if len(*task_str_ids) != 0 && len(*task_str_ids) != len(task_ids) {
		log.Print("Requested:", task_str_ids, "Got:", task_ids)
		return nil, TaskNotFound
	}
	return &task_ids, NoError
}

func GetProject(project_id int) *Project {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT s.name, s.folder_name FROM projects AS s WHERE s.id = $1`)
	Err(err)

	var project Project
	err = query.QueryRow(project_id).Scan(&project.Name, &project.FolderName)
	Err(err)
	project.Id = project_id
	return &project
}

func GetUnit(unit_id int) *Unit {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT u.name, u.project_id, u.folder_name FROM units AS u WHERE u.id = $1`)
	Err(err)

	var unit Unit
	err = query.QueryRow(unit_id).Scan(&unit.Name, &unit.ProjectId, &unit.FolderName)
	Err(err)
	unit.Id = unit_id
	return &unit
}

func SaveSolution(solution *Solution, is_passed bool) {
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

func GetTokenForConnection(email *string, pass *string, ip *string) (*Token, WebError) {
	db := OpenDB()
	defer db.Close()
	var token Token
	token.IP = *ip
	query, err := db.Prepare(`SELECT u.id, u.pass FROM users as u WHERE u.email = $1`)
	Err(err)

	var hash string
	err = query.QueryRow(*email).Scan(&token.UserId, &hash)
	if err != nil {
		return nil, EmailUnknown
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(*pass))
	if err != nil {
		return nil, PasswordWrong
	}

	query, err = db.Prepare(`SELECT id, token FROM tokens WHERE user_id = $1 AND ip = $2`)
	Err(err)
	err = query.QueryRow(token.UserId, token.IP).Scan(&token.Id, &token.Token)
	if err != nil {
		return nil, TokenNotVerified
	}
	return &token, NoError
}

func GetTokenData(token_str *string, ip *string) (*Token, WebError) {
	if len(*token_str) == 0 {
		log.Print("No token received")
		return nil, TokenNotProvided
	}
	if len(*token_str) != token_len {
		log.Print("Received malformed token")
		return nil, TokenInvalid
	}
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT t.id, t.user_id, t.ip 
		FROM tokens as t WHERE t.token = $1`)
	Err(err)
	var token Token
	token.Token = *token_str
	err = query.QueryRow(token.Token).Scan(&token.Id, &token.UserId, &token.IP)
	if err != nil {
		log.Printf("db error on scan: %s", err)
		return nil, TokenUnknown
	}
	if *ip != token.IP {
		log.Print("IP for token does not match")
		return nil, TokenBoundToOtherIP
	}
	return &token, NoError
}

func CreateRegistrationToken(email *string, pass *string, name *string, ip *string) (*string, WebError) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT u.id FROM users as u WHERE u.email = $1`)
	Err(err)
	var user_id int
	err = query.QueryRow(*email).Scan(&user_id)
	if err == nil {
		return nil, EmailTaken
	}

	var token string
	query, err = db.Prepare(`SELECT r.token FROM registration_tokens AS r WHERE r.email = $1 AND r.ip = $2`)
	Err(err)
	err = query.QueryRow(*email, *ip).Scan(&token)
	if err != nil {
		query, err := db.Prepare(`INSERT INTO registration_tokens(token, email, ip, pass, name) VALUES($1, $2, $3, $4, $5)`)
		Err(err)
		hash_raw, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
		Err(err)
		token = string(GenerateToken())
		_, err = query.Exec(token, *email, *ip, hash_raw, *name)
		Err(err)
	}
	return &token, NoError
}

func RegisterToken(ip *string, token_str *string) WebError {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT r.email, r.pass, r.name, r.ip FROM registration_tokens as r WHERE r.token = $1`)
	Err(err)

	var email string
	var pass string
	var name string
	var ip_from_db string
	err = query.QueryRow(*token_str).Scan(&email, &pass, &name, &ip_from_db)
	if err != nil {
		return TokenUnknown
	}

	if *ip != ip_from_db {
		return TokenBoundToOtherIP
	}

	var token Token
	token.IP = *ip
	query, err = db.Prepare(`INSERT INTO users(email, pass, name) VALUES($1, $2, $3) RETURNING id`)
	Err(err)
	err = query.QueryRow(email, pass, name).Scan(&token.UserId)
	Err(err)

	query, err = db.Prepare(`INSERT INTO tokens(token, user_id, ip) VALUES($1, $2, $3) RETURNING id`)
	Err(err)
	token.Token = string(GenerateToken())
	err = query.QueryRow(token.Token, token.UserId, token.IP).Scan(&token.Id)
	Err(err)

	query, err = db.Prepare(`DELETE FROM registration_tokens AS r WHERE r.email = $1`)
	Err(err)
	_, err = query.Exec(email)
	Err(err)

	return NoError
}

func CreateVerificationToken(user_id int, ip *string) *string {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`INSERT INTO 
		verification_tokens(token, ip, user_id) VALUES($1, $2, $3)
		ON CONFLICT (ip, user_id) DO UPDATE SET token = $1`)
	Err(err)
	token := string(GenerateToken())
	_, err = query.Exec(token, *ip, user_id)
	Err(err)
	return &token
}

func VerifyToken(ip *string, token_str *string) WebError {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT v.user_id, v.ip FROM verification_tokens as v WHERE v.token = $1`)
	Err(err)

	var token Token
	err = query.QueryRow(*token_str).Scan(&token.UserId, &token.IP)
	if err != nil {
		return TokenUnknown
	}
	if *ip != token.IP {
		return TokenBoundToOtherIP
	}

	query, err = db.Prepare(`INSERT INTO tokens(token, user_id, ip) VALUES($1, $2, $3) RETURNING id`)
	Err(err)
	token.Token = string(GenerateToken())
	err = query.QueryRow(token.Token, token.UserId, token.IP).Scan(&token.Id)
	Err(err)

	query, err = db.Prepare(`DELETE FROM verification_tokens AS v WHERE r.user_id = $1 AND r.ip = $2`)
	Err(err)
	_, err = query.Exec(token.UserId, token.IP)
	Err(err)

	return NoError
}
