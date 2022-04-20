package storage

import (
	"database/sql"
	"fmt"
	"late/utils"
	"time"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

type TokenType int

const (
	RegisterToken TokenType = 1
	VerifyToken             = 2
	AccessToken             = 3
	RestoreToken            = 4
	DeleteToken             = 5
)

type Storage struct {
	db               *sql.DB
	kv               redis.Conn
	token_expiration map[TokenType]time.Duration
}

func NewStorage() *Storage {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		utils.Env("DB_HOST"), utils.Env("DB_PORT"), utils.Env("DB_USER"), utils.Env("DB_PASS"), utils.Env("DB_NAME")))
	utils.Err(err)

	kv, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", utils.Env("REDIS_HOST"), utils.Env("REDIS_PORT")))
	utils.Err(err)
	default_duration, err := time.ParseDuration(utils.Env("WEB_TOKEN_DEFAULT_DURATION"))
	utils.Err(err)
	access_duration, err := time.ParseDuration(utils.Env("WEB_TOKEN_ACCESS_DURATION"))
	utils.Err(err)
	return &Storage{
		db,
		kv,
		map[tokenType]time.Duration{
			registerToken: default_duration,
			verifyToken:   default_duration,
			accessToken:   access_duration,
			restoreToken:  default_duration,
			deleteToken:   default_duration,
		},
	}
}
