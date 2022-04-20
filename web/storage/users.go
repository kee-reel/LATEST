package storage

import (
	"late/models"
	"late/security"
	"late/utils"
)

func (s *Storage) GetUser(email *string, pass *string) (*models.User, bool, bool) {
	query, err := s.db.Prepare(`SELECT u.id, u.pass FROM users as u WHERE u.email = $1`)
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
	return s.GetUserById(user_id), true, true
}

func (s *Storage) GetUserById(user_id int) *models.User {
	query, err := s.db.Prepare(`SELECT u.name, u.email FROM users as u WHERE u.id = $1`)
	utils.Err(err)
	var user models.User
	err = query.QueryRow(user_id).Scan(&user.Name, &user.Email)
	if err != nil {
		return nil
	}
	user.Score = s.GetLeaderboardScore(user_id)
	user.Id = user_id
	return &user
}
