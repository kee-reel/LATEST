package storage

import (
	"database/sql"
	"fmt"
	"late/utils"
	"time"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
)

type Storage struct {
	db               *sql.DB
	kv               redis.Conn
	token_expiration map[TokenType]time.Duration
}

func CreateRedisConn() redis.Conn {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", utils.Env("REDIS_HOST"), utils.Env("REDIS_PORT")))
	utils.Err(err)
	return conn
}

func NewStorage() *Storage {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		utils.Env("DB_HOST"), utils.Env("DB_PORT"), utils.Env("POSTGRES_USER"), utils.Env("POSTGRES_PASSWORD"), utils.Env("POSTGRES_DB")))
	utils.Err(err)

	storage := Storage{
		db,
		CreateRedisConn(),
		makeTokenDurationMap(),
	}
	return &storage
}
