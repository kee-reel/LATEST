package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/gomodule/redigo/redis"
)

func CreateRedisConn() redis.Conn {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", Env("REDIS_HOST"), Env("REDIS_PORT")))
	Err(err)
	return conn
}

func Env(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(errors.New(fmt.Sprintf("Env variable %s not found", key)))
	}
	return val
}

func EnvB(key string) bool {
	return Env(key) == "true"
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Err(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(statement bool) {
	if !statement {
		panic("Assertion failed")
	}
}
