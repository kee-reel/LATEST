package tokens

import (
	"encoding/json"
	"fmt"
	"late/security"
	"late/storage"
	"late/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

type TokenData struct {
	IP     string             `json:"ip"`
	Email  string             `json:"email"`
	Extra  *map[string]string `json:"extra"`
	UserId int                `json:"-"`
}

type TokenType int

const (
	RegisterToken TokenType = 1
	VerifyToken             = 2
	AccessToken             = 3
	RestoreToken            = 4
	SuspendToken            = 5
)

type TokenError int

const (
	NoError      TokenError = 0
	TokenUnknown            = 1
	TokenExists             = 2
	EmailTaken              = 3
	EmailUnknown            = 4
	WrongIP                 = 5
)

type Tokens struct {
	s                *storage.Storage
	kv               redis.Conn
	token_expiration map[TokenType]time.Duration
}

func NewTokens(s *storage.Storage) *Tokens {
	t := Tokens{
		s,
		utils.CreateRedisConn(),
		makeTokenDurationMap(),
	}
	return &t
}

func (t *Tokens) GetToken(token_type TokenType, email string, ip string) *string {
	key := makeKey(token_type, email, ip)
	token, err := redis.String(t.kv.Do("GET", key))
	if err != nil {
		return nil
	}
	return &token
}

func (t *Tokens) GetTokenData(token_type TokenType, token string, ip string) (*TokenData, TokenError) {
	key := fmt.Sprintf("%d:%s", token_type, token)
	json_data, err := redis.Bytes(t.kv.Do("GET", key))
	if err != nil {
		return nil, TokenUnknown
	}

	var token_data TokenData
	err = json.Unmarshal(json_data, &token_data)
	utils.Err(err)
	if token_data.IP != ip {
		return nil, WrongIP
	}

	user_id := t.s.GetUserIdByEmail(token_data.Email)
	if user_id != nil {
		token_data.UserId = *user_id
	}
	return &token_data, NoError
}

func (t *Tokens) RemoveToken(token_type TokenType, email string, ip string) bool {
	token := t.GetToken(token_type, email, ip)
	if token == nil {
		return false
	}
	t.kv.Send("MULTI")
	t.kv.Send("DEL", token)
	t.kv.Send("DEL", makeKey(token_type, email, ip))
	_, err := t.kv.Do("EXEC")
	utils.Err(err)
	return true
}

func (t *Tokens) CreateToken(token_type TokenType, email string, ip string, args ...string) (*string, TokenError) {
	token := t.GetToken(token_type, email, ip)
	if token != nil {
		return nil, TokenExists
	}

	user_id := t.s.GetUserIdByEmail(email)
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

	token_str := t.addToken(token_type, email, ip, extra_data)
	return &token_str, NoError
}

func (t *Tokens) ApplyToken(token_type TokenType, token string, ip string) (*int, TokenError) {
	token_data, token_err := t.GetTokenData(token_type, token, ip)
	if token_err != NoError {
		return nil, token_err
	}

	user_id := t.s.GetUserIdByEmail(token_data.Email)
	switch token_type {
	case RegisterToken:
		user_id_temp := t.s.AddUser(token_data.Email, (*token_data.Extra)["pass"], (*token_data.Extra)["name"])
		user_id = &user_id_temp
		_, token_err = t.CreateToken(AccessToken, token_data.Email, token_data.IP)
		if token_err != NoError {
			return nil, token_err
		}
	case VerifyToken:
		_, token_err = t.CreateToken(AccessToken, token_data.Email, token_data.IP)
	case RestoreToken:
		user_id := t.s.GetUserIdByEmail(token_data.Email)
		t.s.UpdateUserPassword(*user_id, (*token_data.Extra)["pass"])
	case SuspendToken:
		t.s.SuspendUser(*user_id)
		keys, err := redis.Strings(t.kv.Do("KEYS", fmt.Sprintf("*:%s:*", token_data.Email)))
		utils.Err(err)
		for _, k := range keys {
			keys_arr := strings.Split(k, ":")
			utils.Assert(len(keys_arr) == 3)
			token_type_key, err := strconv.Atoi(keys_arr[0])
			utils.Err(err)
			_ = t.RemoveToken(TokenType(token_type_key), keys_arr[1], keys_arr[2])
		}
	}

	_ = t.RemoveToken(RegisterToken, token_data.Email, token_data.IP)
	return user_id, token_err
}
