package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"late/models"
	"late/security"
	"late/utils"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func OpenDB() *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		utils.Env("DB_HOST"), utils.Env("DB_PORT"), utils.Env("DB_USER"), utils.Env("DB_PASS"), utils.Env("DB_NAME"))
	db, err := sql.Open("postgres", psqlconn)
	utils.Err(err)
	return db
}

func GetTasks(token *models.Token, task_ids []int) *[]models.Task {
	db := OpenDB()
	defer db.Close()

	tasks := make([]models.Task, len(task_ids))
	units := map[int]*models.Unit{}
	projects := map[int]*models.Project{}
	for task_index, task_id := range task_ids {
		query, err := db.Prepare(`SELECT t.project_id, t.unit_id, t.position, t.extention, t.folder_name,
			t.name, t.description, t.input, t.output, t.score
			FROM tasks AS t WHERE t.id = $1`)
		utils.Err(err)

		var task models.Task
		var in_params_str []byte
		var project_id int
		var unit_id int
		err = query.QueryRow(task_id).Scan(
			&project_id, &unit_id, &task.Number, &task.Extention, &task.FolderName,
			&task.Name, &task.Desc, &in_params_str, &task.Output, &task.Score)
		utils.Err(err)

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

		query, err = db.Prepare(`SELECT MAX(s.completion) FROM solutions AS s
			WHERE s.user_id = $1 AND s.task_id = $2 GROUP BY s.task_id`)
		utils.Err(err)

		_ = query.QueryRow(token.UserId, task_id).Scan(&task.Completion)
		err = json.Unmarshal(in_params_str, &task.Input)
		utils.Err(err)

		for i := range task.Input {
			input := &task.Input[i]
			param_type := input.Type
			param_range := input.Range
			if len(input.Dimensions) == 0 {
				input.TotalCount = 1
				input.Dimensions = []int{1}
			} else {
				last_d := 0
				input.TotalCount = 1
				for _, d := range input.Dimensions {
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
				utils.Err(fmt.Errorf("Param type \"%s\" not supported", param_type))
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
	utils.Err(err)

	var source_code string
	var fixed_tests string
	err = query.QueryRow(task_id).Scan(&source_code, &fixed_tests)
	utils.Err(err)

	return &source_code, &fixed_tests
}

func GetTaskTemplate(lang *string, task_id *int) *string {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT t.source_code FROM solution_templates AS t WHERE t.extention = $1`)
	utils.Err(err)

	var source_code string
	err = query.QueryRow(*lang).Scan(&source_code)
	utils.Err(err)

	return &source_code
}

func GetTaskIdsByFolder(folder_names *[]string) (*[]int, int) {
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
		utils.Err(err)
		err = query.QueryRow((*folder_names)[0]).Scan(&project_id)
		if err != nil {
			return nil, 0
		}
		if folders_count != 3 {
			sb.WriteString(" WHERE t.project_id = ")
			sb.WriteString(strconv.Itoa(project_id))
		}
	}
	if folders_count >= 2 {
		query, err := db.Prepare("SELECT u.id FROM units AS u WHERE u.project_id = $1 AND u.folder_name = $2")
		utils.Err(err)
		err = query.QueryRow(project_id, (*folder_names)[1]).Scan(&unit_id)
		if err != nil {
			return nil, 1
		}
		if folders_count != 3 {
			sb.WriteString(" AND t.unit_id = ")
			sb.WriteString(strconv.Itoa(unit_id))
		}
	}
	if folders_count == 3 {
		query, err := db.Prepare("SELECT t.id FROM tasks AS t WHERE t.project_id = $1 AND t.unit_id = $2 AND t.folder_name = $3")
		utils.Err(err)
		var task_id int
		err = query.QueryRow(project_id, unit_id, (*folder_names)[2]).Scan(&task_id)
		if err != nil {
			return nil, 2
		}
		task_ids := []int{task_id}
		return &task_ids, -1
	}

	rows, err := db.Query(sb.String())
	utils.Err(err)

	defer rows.Close()
	task_ids := []int{}
	for rows.Next() {
		var task_id int
		err := rows.Scan(&task_id)
		utils.Err(err)
		task_ids = append(task_ids, task_id)
	}
	return &task_ids, -1
}

func GetTaskIdsById(task_str_ids *[]string) (*[]int, bool) {
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
				return nil, false
			}
			sb.WriteString("t.id=")
			sb.WriteString(task_str_id)
			if i != last_i {
				sb.WriteString(" OR ")
			}
		}
	}

	rows, err := db.Query(sb.String())
	utils.Err(err)

	defer rows.Close()
	task_ids := []int{}
	for rows.Next() {
		var task_id int
		err := rows.Scan(&task_id)
		utils.Err(err)
		task_ids = append(task_ids, task_id)
	}
	if len(*task_str_ids) != 0 && len(*task_str_ids) != len(task_ids) {
		log.Print("Requested:", task_str_ids, "Got:", task_ids)
		return nil, true
	}
	return &task_ids, true
}

func GetProject(project_id int) *models.Project {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT s.name, s.folder_name FROM projects AS s WHERE s.id = $1`)
	utils.Err(err)

	var project models.Project
	err = query.QueryRow(project_id).Scan(&project.Name, &project.FolderName)
	utils.Err(err)
	project.Id = project_id
	return &project
}

func GetUnit(unit_id int) *models.Unit {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT u.name, u.project_id, u.folder_name FROM units AS u WHERE u.id = $1`)
	utils.Err(err)

	var unit models.Unit
	err = query.QueryRow(unit_id).Scan(&unit.Name, &unit.ProjectId, &unit.FolderName)
	utils.Err(err)
	unit.Id = unit_id
	return &unit
}

func SaveSolution(solution *models.Solution, percent float32) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT MAX(s.completion) FROM solutions AS s
		WHERE s.user_id = $1 AND s.task_id = $2`)
	utils.Err(err)
	var solution_id int
	var best_percent float32
	err = query.QueryRow(solution.Token.UserId, solution.Task.Id, percent).Scan(&solution_id, &best_percent)

	if best_percent < percent {
		score_diff := float32(solution.Task.Score) * (percent - best_percent)
		query, err = db.Prepare(`INSERT INTO 
			leaderboard(user_id, project_id, score) VALUES($1, $2, $3)
			ON CONFLICT (user_id, project_id) DO UPDATE SET score = (leaderboard.score + $3)`)
		utils.Err(err)
		_, err = query.Exec(solution.Token.UserId, solution.Task.Project.Id, score_diff)
		utils.Err(err)
	}

	query, err = db.Prepare(`INSERT INTO solutions(user_id, task_id, completion) VALUES($1, $2, $3)`)
	utils.Err(err)
	_, err = query.Exec(solution.Token.UserId, solution.Task.Id, percent)
	utils.Err(err)

	query, err = db.Prepare(`INSERT INTO 
		solutions_sources(user_id, task_id, source_code) VALUES($1, $2, $3)
		ON CONFLICT (user_id, task_id) DO UPDATE SET source_code = $3`)
	utils.Err(err)
	_, err = query.Exec(solution.Token.UserId, solution.Task.Id, solution.Source)
	utils.Err(err)
}

func GetFailedSolutions(solution *models.Solution) int {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT COUNT(*) FROM solutions as s
		WHERE s.task_id = $1 AND s.is_passed = FALSE`)
	utils.Err(err)

	var count int
	err = query.QueryRow(solution.Task.Id).Scan(&count)
	utils.Err(err)
	return count
}

func GetTokenForConnection(user *models.User, ip *string) *models.Token {
	db := OpenDB()
	defer db.Close()
	token := models.Token{
		IP: *ip,
	}

	query, err := db.Prepare(`SELECT id, token FROM tokens WHERE user_id = $1 AND ip = $2`)
	utils.Err(err)
	err = query.QueryRow(user.Id, token.IP).Scan(&token.Id, &token.Token)
	if err != nil {
		return nil
	}
	return &token
}

func GetTokenData(token_str *string) *models.Token {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT t.id, t.user_id, t.ip 
		FROM tokens as t WHERE t.token = $1`)
	utils.Err(err)
	var token models.Token
	token.Token = *token_str
	err = query.QueryRow(token.Token).Scan(&token.Id, &token.UserId, &token.IP)
	if err != nil {
		return nil
	}
	return &token
}

func RemoveToken(token *models.Token) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`DELETE FROM tokens WHERE id = $1`)
	utils.Err(err)
	_, err = query.Exec(token.Id)
	utils.Err(err)
}

func CreateRegistrationToken(email *string, pass *string, name *string, ip *string) *string {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT u.id FROM users as u WHERE u.email = $1`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(*email).Scan(&user_id)
	if err == nil {
		return nil
	}

	var token string
	query, err = db.Prepare(`SELECT r.token FROM registration_tokens AS r WHERE r.email = $1 AND r.ip = $2`)
	utils.Err(err)
	err = query.QueryRow(*email, *ip).Scan(&token)
	if err != nil {
		query, err := db.Prepare(`INSERT INTO registration_tokens(token, email, ip, pass, name) VALUES($1, $2, $3, $4, $5)`)
		utils.Err(err)
		hash_raw, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
		utils.Err(err)
		token = security.GenerateToken()
		_, err = query.Exec(token, *email, *ip, hash_raw, *name)
		utils.Err(err)
	}
	return &token
}

func RegisterToken(ip *string, token_str *string) (*models.User, bool) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT r.email, r.pass, r.name, r.ip FROM registration_tokens as r WHERE r.token = $1`)
	utils.Err(err)

	var user models.User
	var pass string
	var ip_from_db string
	err = query.QueryRow(*token_str).Scan(&user.Email, &pass, &user.Name, &ip_from_db)
	if err != nil {
		return nil, false
	}

	if *ip != ip_from_db {
		return nil, true
	}

	query, err = db.Prepare(`INSERT INTO users(email, pass, name) VALUES($1, $2, $3) RETURNING id`)
	utils.Err(err)
	err = query.QueryRow(user.Email, pass, user.Name).Scan(&user.Id)
	utils.Err(err)

	query, err = db.Prepare(`INSERT INTO tokens(token, user_id, ip) VALUES($1, $2, $3)`)
	utils.Err(err)
	new_token_str := security.GenerateToken()
	_, err = query.Exec(new_token_str, user.Id, *ip)
	utils.Err(err)

	query, err = db.Prepare(`DELETE FROM registration_tokens AS r WHERE r.email = $1`)
	utils.Err(err)
	_, err = query.Exec(user.Email)
	utils.Err(err)

	return &user, true
}

func CreateVerificationToken(email *string, ip *string) *string {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT u.id FROM users as u WHERE u.email = $1`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(*email).Scan(&user_id)
	if err != nil {
		return nil
	}

	query, err = db.Prepare(`INSERT INTO 
		verification_tokens(token, ip, user_id) VALUES($1, $2, $3)
		ON CONFLICT (ip, user_id) DO UPDATE SET token = $1`)
	utils.Err(err)
	token := security.GenerateToken()
	_, err = query.Exec(token, *ip, user_id)
	utils.Err(err)
	return &token
}

func VerifyToken(ip *string, token_str *string) (*int, bool) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT v.user_id, v.ip FROM verification_tokens as v WHERE v.token = $1`)
	utils.Err(err)

	var ip_from_db string
	var user_id int
	err = query.QueryRow(*token_str).Scan(&user_id, &ip_from_db)
	if err != nil {
		return nil, false
	}
	if *ip != ip_from_db {
		return nil, true
	}

	query, err = db.Prepare(`INSERT INTO tokens(token, user_id, ip) VALUES($1, $2, $3)`)
	utils.Err(err)
	new_token_str := security.GenerateToken()
	_, err = query.Exec(new_token_str, user_id, *ip)
	utils.Err(err)

	query, err = db.Prepare(`DELETE FROM verification_tokens AS v WHERE v.user_id = $1 AND v.ip = $2`)
	utils.Err(err)
	_, err = query.Exec(user_id, *ip)
	utils.Err(err)

	return &user_id, true
}

func CreateRestoreToken(email *string, ip *string, pass *string) *string {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT u.id FROM users as u WHERE u.email = $1`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(*email).Scan(&user_id)
	if err != nil {
		return nil
	}

	var token string
	query, err = db.Prepare(`SELECT r.token FROM restore_tokens AS r WHERE r.user_id = $1 AND r.ip = $2`)
	utils.Err(err)
	err = query.QueryRow(user_id, *ip).Scan(&token)
	query, err = db.Prepare(`INSERT INTO 
		restore_tokens(token, ip, user_id, pass) VALUES($1, $2, $3, $4)
		ON CONFLICT (ip, user_id) DO UPDATE SET token = $1, pass = $4`)
	utils.Err(err)
	hash_raw := security.HashPassword(pass)
	token = security.GenerateToken()
	_, err = query.Exec(token, *ip, user_id, hash_raw)
	utils.Err(err)
	return &token
}

func RestoreToken(ip *string, token_str *string) (*int, bool) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT r.user_id, r.ip, r.pass FROM restore_tokens as r WHERE r.token = $1`)
	utils.Err(err)

	var user_id int
	var ip_from_db string
	var pass string
	err = query.QueryRow(*token_str).Scan(&user_id, &ip_from_db, &pass)
	if err != nil {
		return nil, false
	}
	if *ip != ip_from_db {
		return nil, true
	}

	query, err = db.Prepare(`UPDATE users SET pass = $1 WHERE id = $2`)
	utils.Err(err)
	_, err = query.Exec(pass, user_id)
	utils.Err(err)

	query, err = db.Prepare(`DELETE FROM restore_tokens WHERE user_id = $1`)
	utils.Err(err)
	_, err = query.Exec(user_id)
	utils.Err(err)

	return &user_id, true
}

func GetUser(email *string, pass *string) (*models.User, bool, bool) {
	db := OpenDB()
	defer db.Close()

	query, err := db.Prepare(`SELECT u.id, u.pass FROM users as u WHERE u.email = $1`)
	utils.Err(err)

	var user_id int
	var hash string
	err = query.QueryRow(*email).Scan(&user_id, &hash)
	if err != nil {
		return nil, false, false
	}
	if !security.CheckPassword(&hash, pass) {
		return nil, false, true
	}
	return GetUserById(user_id), true, true
}

func GetUserById(user_id int) *models.User {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT u.name, u.email FROM users as u WHERE u.id = $1`)
	utils.Err(err)
	var user models.User
	err = query.QueryRow(user_id).Scan(&user.Name, &user.Email)
	if err != nil {
		return nil
	}
	user.Score = GetLeaderboardScore(user_id)
	user.Id = user_id
	return &user
}

func GetLeaderboardScore(user_id int) float32 {
	db := OpenDB()
	defer db.Close()
	query, err := db.Prepare(`SELECT SUM(l.score) FROM leaderboard as l 
		WHERE l.user_id = $1 GROUP BY l.project_id`)
	utils.Err(err)
	var score float32
	_ = query.QueryRow(user_id).Scan(&score)
	return score
}
