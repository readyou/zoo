package util

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

var Encrypt = &encryptUtil{}

type encryptUtil struct {
}

func (*encryptUtil) EncryptPassword(password string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed)
}

func (*encryptUtil) IsPasswordMatch(hashedPassword, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return false
	}
	return true
}

func (*encryptUtil) RandStr() string {
	byteList := make([]byte, 32)
	_, err := rand.Read(byteList)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(byteList)
}
