package storage

import (
	"late/models"
	"late/security"
	"late/utils"
)

func (s *Storage) AuthenticateUser(email string, pass string) (*models.User, bool) {
	query, err := s.db.Prepare(`SELECT u.id, u.pass FROM users as u WHERE u.email = $1 AND u.is_suspended = FALSE`)
	utils.Err(err)

	var user_id int
	var hash string
	err = query.QueryRow(email).Scan(&user_id, &hash)
	if err != nil {
		return nil, false
	}
	if !security.CheckPassword(hash, pass) {
		return nil, true
	}
	return s.GetUserById(user_id), true
}

func (s *Storage) GetUserById(user_id int) *models.User {
	query, err := s.db.Prepare(`SELECT u.name, u.email, l.score FROM users as u 
		LEFT JOIN leaderboard AS l ON l.user_id = u.id WHERE u.id = $1 AND u.is_suspended = FALSE`)
	utils.Err(err)
	user := models.User{
		Score: 0,
		Id:    user_id,
	}
	var score *float32
	err = query.QueryRow(user_id).Scan(&user.Name, &user.Email, &score)
	if score != nil {
		user.Score = *score
	}
	if err != nil {
		return nil
	}
	user.Id = user_id
	return &user
}

func (s *Storage) GetUserIdByEmail(email string) *int {
	query, err := s.db.Prepare(`SELECT u.id FROM users AS u WHERE u.email = $1 AND u.is_suspended = FALSE`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(email).Scan(&user_id)
	if err != nil {
		return nil
	}
	return &user_id
}

func (s *Storage) AddUser(email string, pass string, name string) int {
	query, err := s.db.Prepare(`INSERT INTO users(email, pass, name) VALUES($1, $2, $3) RETURNING id`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(email, pass, name).Scan(&user_id)
	utils.Err(err)
	return user_id
}

func (s *Storage) SuspendUser(user_id int) {
	query, err := s.db.Prepare(`BEGIN;
	UPDATE users SET is_suspended = TRUE;
	DELETE FROM leaderboard AS l WHERE l.user_id = $1;
	COMMIT;`)
	utils.Err(err)
	_, err = query.Exec(user_id)
	utils.Err(err)
}

func (s *Storage) UpdateUserPassword(user_id int, pass string) {
	query, err := s.db.Prepare(`UPDATE users SET pass = $1 WHERE id = $2`)
	utils.Err(err)
	_, err = query.Exec(pass, user_id)
	utils.Err(err)
}
