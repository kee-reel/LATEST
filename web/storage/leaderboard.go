package storage

import (
	"web/models"
	"web/utils"
)

func (s *Storage) GetLeaderboardScore(user_id int) float32 {
	query, err := s.db.Prepare(`SELECT l.score FROM leaderboard as l WHERE l.user_id = $1`)
	utils.Err(err)
	score := float32(0)
	_ = query.QueryRow(user_id).Scan(&score)
	return score
}

func (s *Storage) GetLeaderboard() *models.Leaderboard {
	query, err := s.db.Prepare(`SELECT u.name, l.score FROM leaderboard AS l
		JOIN users AS u ON u.id = l.user_id`)
	utils.Err(err)
	rows, err := query.Query()
	defer rows.Close()
	leaderboard := models.Leaderboard{}
	for rows.Next() {
		var name string
		score := float32(0)
		err := rows.Scan(&name, &score)
		utils.Err(err)
		leaderboard[name] = score
	}
	return &leaderboard
}
