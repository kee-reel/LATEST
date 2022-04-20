package storage

import (
	"encoding/json"
	"fmt"
	"late/security"
	"late/utils"

	"github.com/gomodule/redigo/redis"
)

type tokenData struct {
	IP    string
	Email string
	Extra *string
}

type registrationData struct {
	Name string
	Pass string
}

func (s *Storage) getToken(token_type TokenType, email *string, ip *string) *string {
	key := fmt.Sprintf("%d:%s:%s", token_type, *email, *ip)
	token, err := redis.String(s.kv.Do("GET", key))
	if err != nil {
		return nil
	}
	return &token
}

func makeKey(token_type TokenType, email *string, ip *string) *string {
	return fmt.Sprintf("%d:%s:%s", token_type, *email, *ip)
}

func (s *Storage) addToken(token_type TokenType, email *string, ip *string, data interface{}) *string {
	json_data, err := json.Marshal(data)
	utils.Err(err)
	key := makeKey(token_type, *email, *ip)
	token := security.GenerateToken()
	s.kv.Send("MULTI")
	s.kv.Send("SET", token, json_data, "EX", s.token_expiration[token_type].Seconds())
	s.kv.Send("SET", key, token, "EX", s.token_expiration[token_type].Seconds())
	_, err = s.kv.Do("EXEC")
	utils.Err(err)
	return token
}

func (s *Storage) getTokenData(token *string, ip *string) (*tokenData, bool) {
	json_data, err := redis.Bytes(s.kv.Do("GET", token))
	if err != nil {
		return nil, true
	}

	var token_data tokenData
	err = json.Unmarshal(json_data, &token_data)
	utils.Err(err)
	if token_data.IP != *ip {
		return nil, false
	}

	user_id := s.GetUserIdByEmail(&token_data.Email)
	if user_id == nil {
		utils.Err(fmt.Errorf("Can't find user by email %s", token_data.Email))
	}
	return token_data, true
}

func (s *Storage) RemoveToken(token_type TokenType, email *string, ip *string) bool {
	token := s.getToken(token_type, email, ip)
	if token == nil {
		return false
	}
	s.kv.Send("MULTI")
	s.kv.Send("DEL", token)
	s.kv.Send("DEL", makeKey(token_type, *email, *ip))
	_, err := s.kv.Do("EXEC")
	utils.Err(err)
	return true
}

func (s *Storage) CreateRegistrationToken(email *string, pass *string, name *string, ip *string) (*string, bool) {
	token := s.getToken(RegisterToken, email, ip)
	if token != nil {
		return nil, false
	}

	user_id := s.GetUserIdByEmail(email)
	if user_id == nil {
		return nil, true
	}

	data := registrationData{
		accessData: accessData{
			IP:    *ip,
			Email: *email,
		},
		Name: *name,
		Pass: security.HashPassword(pass),
	}
	return s.addToken(RegisterToken, email, ip, data), false
}

func (s *Storage) ApplyRegisterToken(ip *string, register_token *string) (*int, bool) {
	token_data := s.getTokenData(register_token, ip)
	if token_data == nil {
		return nil, false
	}

	var register_data registrationData
	err = json.Unmarshal(*token_data.Extra, &register_data)
	utils.Err(err)

	query, err := s.db.Prepare(`INSERT INTO users(email, pass, name) VALUES($1, $2, $3) RETURNING id`)
	utils.Err(err)
	var user_id int
	err = query.QueryRow(register_data.Email, register_data.Pass, register_data.Name).Scan(&user_id)
	utils.Err(err)

	s.RemoveToken(RegisterToken, token_data.Email, token_data.IP)

	_ = s.CreateAccessToken(&register_data.Email, ip)
	return &user_id, true
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

func (s *Storage) VerifyToken(ip *string, verify_token *string) (*int, bool) {
	json_data, err := redis.Bytes(s.kv.Do("GET", *verify_token))
	if err != nil {
		return nil, false
	}
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

	_ = s.createAccessToken(&register_data.Email, ip)
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

func (s *Storage) CreateAccessToken(email *string, ip *string) *string {
	access_data := accessData{
		IP:    *ip,
		Email: *email,
	}
	json_data, err = json.Marshal(access_data)
	utils.Err(err)
	access_token := security.GenerateToken()

	s.kv.Send("MULTI")
	s.kv.Send("SET", access_token, json_data, "EX", s.token_expiration[registerToken].Seconds())
	s.kv.Send("SET", *email, access_token, "EX", s.token_expiration[registerToken].Seconds())
	_, err := s.kv.Do("EXEC")
	return access_token
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
