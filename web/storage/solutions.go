package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"web/models"
	"web/utils"
)

func (s *Storage) SaveSolution(solution *models.Solution) int64 {
	var solution_id int64

	query, err := s.db.Prepare(`SELECT s.id, l.response FROM last_solutions AS l
		WHERE user_id = $1 AND task_id = $2 AND s.hash = $3`)
	utils.Err(err)
	err = query.QueryRow(solution.UserId, solution.Task.Id).Scan(&solution_id)
	utils.Err(err)

	query, err := s.db.Prepare(`INSERT INTO solutions(user_id, task_id, hash) 
		VALUES($1, $2, $3) RETURNING id`)
	utils.Err(err)
	err = query.QueryRow(solution.UserId, solution.Task.Id).Scan(&solution_id)
	utils.Err(err)

	h := sha1.New()
	h.Write([]byte(s))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	query, err = s.db.Prepare(`INSERT INTO 
		last_solutions(solution_id, hash) VALUES($1, $2, $3)
		DO UPDATE SET source_code = $3, response = NULL`)
	utils.Err(err)
	_, err = query.Exec(solution.UserId, solution.Task.Id, solution.Source, sha1_hash)
	utils.Err(err)
	return solution_id
}

func (s *Storage) UpdateSolutionScore(solution *models.Solution, percent float32) float32 {
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
	return score_diff
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
