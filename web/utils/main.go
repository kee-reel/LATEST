package utils

import (
	"errors"
	"fmt"
	"os"
)

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
