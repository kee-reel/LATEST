package storage

import (
	"encoding/json"
	"fmt"
	"late/models"
	"late/security"
	"late/utils"
	"log"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

func (s *Storage) GetTokenForConnection(user *models.User, ip *string) *models.Token {
	token := models.Token{
		UserId: user.Id,
		IP:     *ip,
	}

	query, err := s.db.Prepare(`SELECT id, token FROM tokens WHERE user_id = $1 AND ip = $2`)
	utils.Err(err)
	err = query.QueryRow(user.Id, token.IP).Scan(&token.Id, &token.Token)
	if err != nil {
		return nil
	}
	return &token
}

func (s *Storage) GetTokenData(token_str *string) *models.Token {
	query, err := s.db.Prepare(`SELECT t.id, t.user_id, t.ip 
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

func (s *Storage) RemoveToken(token *models.Token) {
	query, err := s.db.Prepare(`DELETE FROM tokens WHERE id = $1`)
	utils.Err(err)
	_, err = query.Exec(token.Id)
	utils.Err(err)
}

type RegistrationData struct {
	Name  string
	Email string
	IP    string
	pass  string
}

func (s *Storage) CreateRegistrationToken(email *string, pass *string, name *string, ip *string) (*string, bool) {
	key := fmt.Sprintf("%s:%s", *email, *ip)
	is_exists, err := redis.Bool(s.kv.Do("EXISTS", key))
	utils.Err(err)
	log.Print(is_exists)

	token := security.GenerateToken()
	hash_raw, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
	utils.Err(err)
	data := RegistrationData{
		*ip,
		*name,
		*email,
		string(hash_raw),
	}
	json_data, err := json.Marshal(data)
	utils.Err(err)
	_, err = s.kv.Do("SET", token, json_data, "EX", s.token_expiration[registerToken])
	utils.Err(err)

	/*
		var user_id int
		err = query.QueryRow(*email).Scan(&user_id)
		if err == nil {
			return nil, false
		}

		var token string
		is_new_token := false
		query, err = s.db.Prepare(`SELECT r.token FROM registration_tokens AS r WHERE r.email = $1 AND r.ip = $2`)
		utils.Err(err)
		err = query.QueryRow(*email, *ip).Scan(&token)
		if err != nil {
			query, err := s.db.Prepare(`INSERT INTO registration_tokens(token, email, ip, pass, name) VALUES($1, $2, $3, $4, $5)`)
			utils.Err(err)
			hash_raw, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
			utils.Err(err)
			token = security.GenerateToken()
			_, err = query.Exec(token, *email, *ip, hash_raw, *name)
			utils.Err(err)
			is_new_token = true
		}
	*/
	return nil, false
}

func (s *Storage) RegisterToken(ip *string, token_str *string) (*models.User, bool) {
	query, err := s.db.Prepare(`SELECT r.email, r.pass, r.name, r.ip FROM registration_tokens as r WHERE r.token = $1`)
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

	query, err = s.db.Prepare(`INSERT INTO users(email, pass, name) VALUES($1, $2, $3) RETURNING id`)
	utils.Err(err)
	err = query.QueryRow(user.Email, pass, user.Name).Scan(&user.Id)
	utils.Err(err)

	query, err = s.db.Prepare(`INSERT INTO tokens(token, user_id, ip) VALUES($1, $2, $3)`)
	utils.Err(err)
	new_token_str := security.GenerateToken()
	_, err = query.Exec(new_token_str, user.Id, *ip)
	utils.Err(err)

	query, err = s.db.Prepare(`DELETE FROM registration_tokens AS r WHERE r.email = $1`)
	utils.Err(err)
	_, err = query.Exec(user.Email)
	utils.Err(err)

	return &user, true
}

func (s *Storage) CreateVerificationToken(email *string, ip *string) *string {
	query, err := s.db.Prepare(`SELECT u.id FROM users as u WHERE u.email = $1`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(*email).Scan(&user_id)
	if err != nil {
		return nil
	}

	query, err = s.db.Prepare(`INSERT INTO 
		verification_tokens(token, ip, user_id) VALUES($1, $2, $3)
		ON CONFLICT (ip, user_id) DO UPDATE SET token = $1`)
	utils.Err(err)
	token := security.GenerateToken()
	_, err = query.Exec(token, *ip, user_id)
	utils.Err(err)
	return &token
}

func (s *Storage) VerifyToken(ip *string, token_str *string) (*int, bool) {
	query, err := s.db.Prepare(`SELECT v.user_id, v.ip FROM verification_tokens as v WHERE v.token = $1`)
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

	query, err = s.db.Prepare(`INSERT INTO tokens(token, user_id, ip) VALUES($1, $2, $3)`)
	utils.Err(err)
	new_token_str := security.GenerateToken()
	_, err = query.Exec(new_token_str, user_id, *ip)
	utils.Err(err)

	query, err = s.db.Prepare(`DELETE FROM verification_tokens AS v WHERE v.user_id = $1 AND v.ip = $2`)
	utils.Err(err)
	_, err = query.Exec(user_id, *ip)
	utils.Err(err)

	return &user_id, true
}

func (s *Storage) CreateResetToken(user_id int, ip *string) *string {
	query, err := s.db.Prepare(`INSERT INTO 
		reset_tokens(token, ip, user_id) VALUES($1, $2, $3)
		ON CONFLICT (ip, user_id) DO UPDATE SET token = $1`)
	utils.Err(err)
	token := security.GenerateToken()
	_, err = query.Exec(token, *ip, user_id)
	utils.Err(err)
	return &token
}

func (s *Storage) ResetToken(ip *string, token_str *string) (*int, bool) {
	query, err := s.db.Prepare(`SELECT v.user_id, v.ip FROM reset_tokens as v WHERE v.token = $1`)
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

	query, err = s.db.Prepare(`DELETE FROM solutions AS s WHERE s.user_id = $1`)
	utils.Err(err)
	_, err = query.Exec(user_id)
	utils.Err(err)

	query, err = s.db.Prepare(`DELETE FROM solutions_sources AS s WHERE s.user_id = $1`)
	utils.Err(err)
	_, err = query.Exec(user_id)
	utils.Err(err)

	query, err = s.db.Prepare(`DELETE FROM leaderboard AS l WHERE l.user_id = $1`)
	utils.Err(err)
	_, err = query.Exec(user_id)
	utils.Err(err)

	return &user_id, true
}

func (s *Storage) CreateRestoreToken(email *string, ip *string, pass *string) (*string, bool) {
	query, err := s.db.Prepare(`SELECT u.id FROM users as u WHERE u.email = $1`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(*email).Scan(&user_id)
	if err != nil {
		return nil, false
	}

	var token string
	is_new_token := false
	query, err = s.db.Prepare(`SELECT r.token FROM restore_tokens AS r WHERE r.user_id = $1 AND r.ip = $2`)
	utils.Err(err)
	err = query.QueryRow(user_id, *ip).Scan(&token)
	if err != nil {
		query, err = s.db.Prepare(`INSERT INTO restore_tokens(token, ip, user_id, pass) VALUES($1, $2, $3, $4)`)
		utils.Err(err)
		hash_raw := security.HashPassword(pass)
		token = security.GenerateToken()
		_, err = query.Exec(token, *ip, user_id, hash_raw)
		utils.Err(err)
		is_new_token = true
	}
	return &token, is_new_token
}

func (s *Storage) RestoreToken(ip *string, token_str *string) (*int, bool) {
	query, err := s.db.Prepare(`SELECT r.user_id, r.ip, r.pass FROM restore_tokens as r WHERE r.token = $1`)
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

	query, err = s.db.Prepare(`UPDATE users SET pass = $1 WHERE id = $2`)
	utils.Err(err)
	_, err = query.Exec(pass, user_id)
	utils.Err(err)

	query, err = s.db.Prepare(`DELETE FROM restore_tokens WHERE user_id = $1`)
	utils.Err(err)
	_, err = query.Exec(user_id)
	utils.Err(err)

	return &user_id, true
}
