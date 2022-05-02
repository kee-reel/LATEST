package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"web/models"
	"web/utils"

	"golang.org/x/crypto/sha3"
)

func (s *Storage) SaveSolution(solution *models.Solution) (int64, *models.TestResult) {
	query, err := s.db.Prepare(`SELECT id, response FROM solutions
		WHERE task_id = $1 AND hash = $2`)
	utils.Err(err)
	var solution_id int64
	var response *[]byte
	hash := make([]byte, 64)
	sha3.ShakeSum256(hash, []byte(solution.Source))
	hash_str := fmt.Sprintf("%x", hash)
	log.Print(hash_str)
	err = query.QueryRow(solution.Task.Id, hash_str).Scan(&solution_id, &response)
	if err != nil {
		query, err = s.db.Prepare(`INSERT INTO 
		solutions(task_id, hash, text) VALUES($1, $2, $3)
		ON CONFLICT (task_id, hash) DO NOTHING RETURNING id`)
		utils.Err(err)
		err = query.QueryRow(solution.Task.Id, hash_str, solution.Source).Scan(&solution_id)
		utils.Err(err)
	}

	query, err = s.db.Prepare(`INSERT INTO solution_attempts(user_id, task_id, solution_id) 
		VALUES($1, $2, $3)`)
	utils.Err(err)
	_, err = query.Exec(solution.UserId, solution.Task.Id, solution_id)
	utils.Err(err)

	var test_result *models.TestResult
	if response != nil {
		err = json.Unmarshal(*response, &test_result)
		if err != nil {
			log.Printf("[ERROR] Cant unmarshal cached response: %s\n err: %s", string(*response), err)
			query, err = s.db.Prepare(`UPDATE solution SET response = NULL, received_times = 0 WHERE id = $1`)
			utils.Err(err)
			_, err = query.Exec(solution_id)
			utils.Err(err)
		}
	}

	return solution_id, test_result
}

func (s *Storage) UpdateSolutionScore(solution *models.Solution, response *models.TestResult, percent float32) float32 {
	/*
		query, err := s.db.Prepare(`SELECT MAX(s.completion) FROM solutions AS s
			WHERE s.user_id = $1 AND s.task_id = $2`)
		utils.Err(err)
		best_percent := float32(0)
		err = query.QueryRow(solution.UserId, solution.Task.Id).Scan(&best_percent)

		score_diff := float32(0)
		if best_percent < percent {
			score_diff = float32(solution.Task.Score) * (percent - best_percent)
			query, err = s.db.Prepare(`INSERT INTO
				leaderboard(user_id, project_id, score) VALUES($1, $2, $3)
				ON CONFLICT (user_id, project_id) DO UPDATE SET score = (leaderboard.score + $3)`)
			utils.Err(err)
			_, err = query.Exec(solution.UserId, solution.Task.Project.Id, score_diff)
			utils.Err(err)
		}
	*/

	// Do not cache solution response if it was internal or timeout error
	if response.InternalError == nil && (response.ErrorData == nil || response.ErrorData.Timeout == nil) {
		query, err := s.db.Prepare(`UPDATE solutions
			SET completion = $1, response = $2, received_times = received_times + 1
			WHERE id = $3`)
		utils.Err(err)
		data_json, err := json.Marshal(response)
		utils.Err(err)
		_, err = query.Exec(percent, data_json, solution.Id)
		utils.Err(err)
	}

	return 0
}

func (s *Storage) GetSolutionText(user_id int, task_id int) *string {
	query, err := s.db.Prepare(`SELECT s.source_code FROM last_solutions as s
		WHERE s.user_id = $1 AND s.task_id = $2`)
	utils.Err(err)

	var source_code string
	err = query.QueryRow(user_id, task_id).Scan(&source_code)
	return &source_code
}

func (s *Storage) GetFailedSolutions(solution *models.Solution) int {
	query, err := s.db.Prepare(`SELECT COUNT(*) FROM solutions as s
		WHERE s.task_id = $1 AND s.is_passed = FALSE`)
	utils.Err(err)

	var count int
	err = query.QueryRow(solution.Task.Id).Scan(&count)
	utils.Err(err)
	return count
}
