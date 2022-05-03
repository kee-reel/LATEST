package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"web/models"
	"web/utils"

	"golang.org/x/crypto/sha3"
)

func (s *Storage) CreateSolutionAttempt(solution *models.Solution) (int64, *models.TestResult) {
	query, err := s.db.Prepare(`SELECT id, response, received_times FROM solutions
		WHERE task_id = $1 AND hash = $2`)
	utils.Err(err)

	var test_result *models.TestResult
	var solution_id int64
	var response *[]byte
	var received_times int
	hash := make([]byte, 64)
	sha3.ShakeSum256(hash, []byte(solution.Source))
	hash_str := fmt.Sprintf("%x", hash)
	err = query.QueryRow(solution.Task.Id, hash_str).Scan(&solution_id, &response, &received_times)
	if err == nil {
		if received_times > s.solution_cache_threshold {
			err = json.Unmarshal(*response, &test_result)
			if err != nil {
				log.Printf("[ERROR] Cant unmarshal cached response: %s\n err: %s", string(*response), err)
				query, err = s.db.Prepare(`UPDATE solutions SET response = NULL, received_times = 0 WHERE id = $1`)
				utils.Err(err)
				_, err = query.Exec(solution_id)
				utils.Err(err)
			}
		} else {
			response = nil
		}
	} else {
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

	return solution_id, test_result
}

func (s *Storage) UpdateSolutionAttempt(sol *models.Solution, completion float32) float32 {
	query, err := s.db.Prepare(`INSERT INTO task_completions(user_id, task_id, completion) VALUES($1, $2, $3) 
		ON CONFLICT (user_id, task_id) DO UPDATE SET completion = GREATEST(task_completions.completion, $3)
		RETURNING (SELECT completion FROM task_completions WHERE user_id = $1 AND task_id = $2) AS old_completion`)
	utils.Err(err)
	old_completion := float32(0)
	err = query.QueryRow(sol.UserId, sol.Task.Id, completion).Scan(&old_completion)

	score_diff := float32(0)
	if old_completion < completion {
		score_diff = float32(sol.Task.Score) * (completion - old_completion)
		query, err = s.db.Prepare(`INSERT INTO
				leaderboard(user_id, score) VALUES($1, $2)
				ON CONFLICT (user_id) DO UPDATE SET score = (leaderboard.score + $2)`)
		utils.Err(err)
		_, err = query.Exec(sol.UserId, score_diff)
		utils.Err(err)
	}

	return score_diff
}

func (s *Storage) SaveSolutionResult(solution_id int64, resp *models.TestResult) {
	// Do not cache solution response if it was internal or timeout error
	if resp.InternalError != nil || (resp.ErrorData != nil && resp.ErrorData.Timeout != nil) {
		return
	}
	query, err := s.db.Prepare(`SELECT response FROM solutions WHERE id = $1`)
	utils.Err(err)

	var response *[]byte
	err = query.QueryRow(solution_id).Scan(&response)
	utils.Err(err)

	data_json, err := json.Marshal(resp)
	utils.Err(err)
	if response == nil || bytes.Equal(*response, data_json) {
		query, err = s.db.Prepare(`UPDATE solutions
			SET received_times = received_times + 1
			WHERE id = $1`)
		utils.Err(err)
		_, err = query.Exec(solution_id)
	} else {
		query, err = s.db.Prepare(`UPDATE solutions
			SET response = $1, received_times = 1
			WHERE id = $2`)
		utils.Err(err)
		_, err = query.Exec(data_json, solution_id)
	}
	utils.Err(err)

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
