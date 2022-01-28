package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func GenerateToken() string {
	token_raw := make([]byte, token_len)
	for i := 0; i < token_len; i++ {
		char_i, err := rand.Int(rand.Reader, big.NewInt(token_chars_len))
		Err(err)
		token_raw[i] = token_chars[char_i.Int64()]
	}
	return string(token_raw)
}

func Err(err error) {
	if err != nil {
		log.Printf("[ERROR] %s", err)
		panic(fmt.Errorf("Internal error"))
	}
}

func ErrMsg(err error, msg string) {
	if err != nil {
		log.Printf("[ERROR] %s", err)
		panic(fmt.Errorf(msg))
	}
}
