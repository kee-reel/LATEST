package storage

import (
	"database/sql"
	"fmt"
	"late/utils"
	"time"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

type tokenType int

const (
	registerToken tokenType = 1
	verifyToken             = 2
	accessToken             = 3
	restoreToken            = 4
	deleteToken             = 5
)

type Storage struct {
	db               *sql.DB
	kv               redis.Conn
	token_expiration map[tokenType]time.Duration
}

func NewStorage() *Storage {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		utils.Env("DB_HOST"), utils.Env("DB_PORT"), utils.Env("DB_USER"), utils.Env("DB_PASS"), utils.Env("DB_NAME")))
	utils.Err(err)

	kv, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", utils.Env("REDIS_HOST"), utils.Env("REDIS_PORT")))
	utils.Err(err)
	return &Storage{
		db,
		kv,
		map[tokenType]time.Duration{},
	}
}
