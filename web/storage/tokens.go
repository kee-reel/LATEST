package storage

import (
	"encoding/json"
	"fmt"
	"late/security"
	"late/utils"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

func (s *Storage) GetToken(token_type TokenType, email string, ip string) *string {
	key := makeKey(token_type, email, ip)
	token, err := redis.String(s.kv.Do("GET", key))
	if err != nil {
		return nil
	}
	return &token
}

func (s *Storage) GetTokenData(token_type TokenType, token string, ip string) (*TokenData, TokenError) {
	key := fmt.Sprintf("%d:%s", token_type, token)
	json_data, err := redis.Bytes(s.kv.Do("GET", key))
	if err != nil {
		return nil, TokenUnknown
	}

	var token_data TokenData
	err = json.Unmarshal(json_data, &token_data)
	utils.Err(err)
	if token_data.IP != ip {
		return nil, WrongIP
	}

	user_id := s.GetUserIdByEmail(token_data.Email)
	if user_id != nil {
		token_data.UserId = *user_id
	}
	return &token_data, NoError
}

func (s *Storage) RemoveToken(token_type TokenType, email string, ip string) bool {
	token := s.GetToken(token_type, email, ip)
	if token == nil {
		return false
	}
	s.kv.Send("MULTI")
	s.kv.Send("DEL", token)
	s.kv.Send("DEL", makeKey(token_type, email, ip))
	_, err := s.kv.Do("EXEC")
	utils.Err(err)
	return true
}

func (s *Storage) CreateToken(token_type TokenType, email string, ip string, args ...string) (*string, TokenError) {
	token := s.GetToken(token_type, email, ip)
	if token != nil {
		return nil, TokenExists
	}

	user_id := s.GetUserIdByEmail(email)
	if token_type == RegisterToken {
		if user_id != nil {
			return nil, EmailTaken
		}
	} else {
		if user_id == nil {
			return nil, EmailUnknown
		}
	}

	var extra_data *map[string]string
	switch token_type {
	case RegisterToken:
		if len(args) != 2 {
			panic("Wrong register args")
		}
		extra_data = &map[string]string{
			"name": args[0],
			"pass": security.HashPassword(args[1]),
		}
	case RestoreToken:
		if len(args) != 1 {
			panic("Wrong restore args")
		}
		extra_data = &map[string]string{
			"pass": security.HashPassword(args[0]),
		}
	}

	token_str := s.addToken(token_type, email, ip, extra_data)
	return &token_str, NoError
}

func (s *Storage) ApplyToken(token_type TokenType, token string, ip string) (*int, TokenError) {
	token_data, token_err := s.GetTokenData(token_type, token, ip)
	if token_err != NoError {
		return nil, token_err
	}

	user_id := s.GetUserIdByEmail(token_data.Email)
	switch token_type {
	case RegisterToken:
		user_id_temp := s.AddUser(token_data.Email, (*token_data.Extra)["pass"], (*token_data.Extra)["name"])
		user_id = &user_id_temp
		_, token_err = s.CreateToken(AccessToken, token_data.Email, token_data.IP)
		if token_err != NoError {
			return nil, token_err
		}
	case VerifyToken:
		_, token_err = s.CreateToken(AccessToken, token_data.Email, token_data.IP)
	case RestoreToken:
		user_id := s.GetUserIdByEmail(token_data.Email)
		s.UpdateUserPassword(*user_id, (*token_data.Extra)["pass"])
	case SuspendToken:
		s.SuspendUser(*user_id)
		keys, err := redis.Strings(s.kv.Do("KEYS", fmt.Sprintf("*:%s:*", token_data.Email)))
		utils.Err(err)
		for _, k := range keys {
			k_arr := strings.Split(k, ":")
			utils.Assert(len(k_arr) == 3)
			t, err := strconv.Atoi(k_arr[0])
			utils.Err(err)
			_ = s.RemoveToken(TokenType(t), k_arr[1], k_arr[2])
		}
	}

	_ = s.RemoveToken(RegisterToken, token_data.Email, token_data.IP)
	return user_id, token_err
}
