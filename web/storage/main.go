package storage

import (
	"database/sql"
	"fmt"
	"late/utils"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func NewStorage() *Storage {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		utils.Env("DB_HOST"), utils.Env("DB_PORT"), utils.Env("POSTGRES_USER"), utils.Env("POSTGRES_PASSWORD"), utils.Env("POSTGRES_DB")))
	utils.Err(err)

	storage := Storage{db}
	return &storage
}
