package security

import (
	"crypto/rand"
	"late/utils"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const token_len = 64
const token_chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const token_chars_len = int64(len(token_chars))

func IsTokenInvalid(token string) bool {
	return len(token) != token_len
}

func GenerateToken() string {
	token_raw := make([]byte, token_len)
	for i := 0; i < token_len; i++ {
		char_i, err := rand.Int(rand.Reader, big.NewInt(token_chars_len))
		utils.Err(err)
		token_raw[i] = token_chars[char_i.Int64()]
	}
	return string(token_raw)
}

func HashPassword(pass string) string {
	hash_raw, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	utils.Err(err)
	return string(hash_raw)
}

func CheckPassword(hash string, pass string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}
